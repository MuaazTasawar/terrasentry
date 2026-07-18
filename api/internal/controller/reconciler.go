package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/MuaazTasawar/terrasentry/api/internal/db"
)

// LastAppliedAnnotation stores the config snapshot we compare against on
// each reconcile to detect drift. In a real GitOps setup this would come
// from the last applied Terraform/kubectl state; here we snapshot it the
// first time we see a Deployment and compare on every subsequent event.
const LastAppliedAnnotation = "terrasentry.io/last-applied-spec"

// deploymentSnapshot captures the fields we care about for drift detection.
// Kept intentionally small (replicas, image, resources) for v1 — this is
// the set of fields most likely to drift from an unauthorized manual change.
type deploymentSnapshot struct {
	Replicas int32             `json:"replicas"`
	Image    string            `json:"image"`
	CPULimit string            `json:"cpu_limit"`
	MemLimit string            `json:"mem_limit"`
	Labels   map[string]string `json:"labels"`
}

// DriftReconciler watches Deployments and records DriftEvents when live
// state diverges from the last snapshot we took.
type DriftReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	OnDriftFound func(ctx context.Context, event db.DriftEvent) error
}

func (r *DriftReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var deploy appsv1.Deployment
	if err := r.Get(ctx, req.NamespacedName, &deploy); err != nil {
		if errors.IsNotFound(err) {
			// Deployment was deleted — nothing to reconcile.
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to fetch deployment: %w", err)
	}

	current := snapshotFromDeployment(&deploy)

	previousRaw, hasSnapshot := deploy.Annotations[LastAppliedAnnotation]
	if !hasSnapshot {
		// First time seeing this Deployment: record the baseline, no drift yet.
		return ctrl.Result{}, r.saveSnapshot(ctx, &deploy, current)
	}

	var previous deploymentSnapshot
	if err := json.Unmarshal([]byte(previousRaw), &previous); err != nil {
		log.Printf("warning: could not parse previous snapshot for %s/%s, resetting baseline: %v",
			deploy.Namespace, deploy.Name, err)
		return ctrl.Result{}, r.saveSnapshot(ctx, &deploy, current)
	}

	diff := diffSnapshots(previous, current)
	if diff == "" {
		// No drift detected — nothing to do.
		return ctrl.Result{}, nil
	}

	log.Printf("drift detected on %s/%s: %s", deploy.Namespace, deploy.Name, diff)

	if r.OnDriftFound != nil {
		event := db.DriftEvent{
			ResourceKind: "Deployment",
			ResourceName: deploy.Name,
			Namespace:    deploy.Namespace,
			Diff:         diff,
		}
		if err := r.OnDriftFound(ctx, event); err != nil {
			log.Printf("failed to persist drift event: %v", err)
		}
	}

	// Update the baseline to the current state so we don't re-report the
	// same drift on every future reconcile loop.
	return ctrl.Result{}, r.saveSnapshot(ctx, &deploy, current)
}

func (r *DriftReconciler) saveSnapshot(ctx context.Context, deploy *appsv1.Deployment, snap deploymentSnapshot) error {
	raw, err := json.Marshal(snap)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	if deploy.Annotations == nil {
		deploy.Annotations = map[string]string{}
	}
	deploy.Annotations[LastAppliedAnnotation] = string(raw)

	if err := r.Update(ctx, deploy); err != nil {
		return fmt.Errorf("failed to save snapshot annotation: %w", err)
	}
	return nil
}

func snapshotFromDeployment(deploy *appsv1.Deployment) deploymentSnapshot {
	snap := deploymentSnapshot{
		Labels: deploy.Spec.Template.Labels,
	}

	if deploy.Spec.Replicas != nil {
		snap.Replicas = *deploy.Spec.Replicas
	}

	if len(deploy.Spec.Template.Spec.Containers) > 0 {
		c := deploy.Spec.Template.Spec.Containers[0]
		snap.Image = c.Image

		if cpu := c.Resources.Limits.Cpu(); cpu != nil {
			snap.CPULimit = cpu.String()
		}
		if mem := c.Resources.Limits.Memory(); mem != nil {
			snap.MemLimit = mem.String()
		}
	}

	return snap
}

func diffSnapshots(prev, curr deploymentSnapshot) string {
	var diffs []string

	if prev.Replicas != curr.Replicas {
		diffs = append(diffs, fmt.Sprintf("replicas: %d -> %d", prev.Replicas, curr.Replicas))
	}
	if prev.Image != curr.Image {
		diffs = append(diffs, fmt.Sprintf("image: %s -> %s", prev.Image, curr.Image))
	}
	if prev.CPULimit != curr.CPULimit {
		diffs = append(diffs, fmt.Sprintf("cpu_limit: %s -> %s", prev.CPULimit, curr.CPULimit))
	}
	if prev.MemLimit != curr.MemLimit {
		diffs = append(diffs, fmt.Sprintf("mem_limit: %s -> %s", prev.MemLimit, curr.MemLimit))
	}

	if len(diffs) == 0 {
		return ""
	}

	result := diffs[0]
	for _, d := range diffs[1:] {
		result += "; " + d
	}
	return result
}
