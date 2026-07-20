package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/MuaazTasawar/terrasentry/api/internal/db"
)

// statefulSetSnapshot captures the fields we care about for drift detection
// on StatefulSets. Volume claim templates are included because, unlike
// Deployments, changing storage size/class on a StatefulSet is a common
// source of unauthorized manual edits and can't be rolled back cheaply.
type statefulSetSnapshot struct {
	Replicas         int32             `json:"replicas"`
	Image            string            `json:"image"`
	CPULimit         string            `json:"cpu_limit"`
	MemLimit         string            `json:"mem_limit"`
	Labels           map[string]string `json:"labels"`
	VolumeClaimSizes map[string]string `json:"volume_claim_sizes"`
}

// StatefulSetDriftReconciler watches StatefulSets and records DriftEvents
// when live state diverges from the last snapshot we took. Mirrors
// DriftReconciler's snapshot/diff/annotation approach for Deployments.
type StatefulSetDriftReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	OnDriftFound func(ctx context.Context, event db.DriftEvent) error
}

func (r *StatefulSetDriftReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var sts appsv1.StatefulSet
	if err := r.Get(ctx, req.NamespacedName, &sts); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to fetch statefulset: %w", err)
	}

	current := snapshotFromStatefulSet(&sts)

	previousRaw, hasSnapshot := sts.Annotations[LastAppliedAnnotation]
	if !hasSnapshot {
		return ctrl.Result{}, r.saveSnapshot(ctx, &sts, current)
	}

	var previous statefulSetSnapshot
	if err := json.Unmarshal([]byte(previousRaw), &previous); err != nil {
		log.Printf("warning: could not parse previous snapshot for %s/%s, resetting baseline: %v",
			sts.Namespace, sts.Name, err)
		return ctrl.Result{}, r.saveSnapshot(ctx, &sts, current)
	}

	diff := diffStatefulSetSnapshots(previous, current)
	if diff == "" {
		return ctrl.Result{}, nil
	}

	log.Printf("drift detected on statefulset %s/%s: %s", sts.Namespace, sts.Name, diff)

	if r.OnDriftFound != nil {
		event := db.DriftEvent{
			ResourceKind: "StatefulSet",
			ResourceName: sts.Name,
			Namespace:    sts.Namespace,
			Diff:         diff,
		}
		if err := r.OnDriftFound(ctx, event); err != nil {
			log.Printf("failed to persist drift event: %v", err)
		}
	}

	return ctrl.Result{}, r.saveSnapshot(ctx, &sts, current)
}

func (r *StatefulSetDriftReconciler) saveSnapshot(ctx context.Context, sts *appsv1.StatefulSet, snap statefulSetSnapshot) error {
	raw, err := json.Marshal(snap)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest := &appsv1.StatefulSet{}
		if err := r.Get(ctx, client.ObjectKeyFromObject(sts), latest); err != nil {
			return err
		}

		if latest.Annotations == nil {
			latest.Annotations = map[string]string{}
		}
		latest.Annotations[LastAppliedAnnotation] = string(raw)

		return r.Update(ctx, latest)
	})
}

func snapshotFromStatefulSet(sts *appsv1.StatefulSet) statefulSetSnapshot {
	snap := statefulSetSnapshot{
		Labels:           sts.Spec.Template.Labels,
		VolumeClaimSizes: map[string]string{},
	}

	if sts.Spec.Replicas != nil {
		snap.Replicas = *sts.Spec.Replicas
	}

	if len(sts.Spec.Template.Spec.Containers) > 0 {
		c := sts.Spec.Template.Spec.Containers[0]
		snap.Image = c.Image

		if cpu := c.Resources.Limits.Cpu(); cpu != nil {
			snap.CPULimit = cpu.String()
		}
		if mem := c.Resources.Limits.Memory(); mem != nil {
			snap.MemLimit = mem.String()
		}
	}

	for _, vct := range sts.Spec.VolumeClaimTemplates {
		if storage, ok := vct.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
			snap.VolumeClaimSizes[vct.Name] = storage.String()
		}
	}

	return snap
}

func diffStatefulSetSnapshots(prev, curr statefulSetSnapshot) string {
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

	claimNames := map[string]bool{}
	for name := range prev.VolumeClaimSizes {
		claimNames[name] = true
	}
	for name := range curr.VolumeClaimSizes {
		claimNames[name] = true
	}
	for _, name := range sortedKeys(claimNames) {
		prevSize := prev.VolumeClaimSizes[name]
		currSize := curr.VolumeClaimSizes[name]
		if prevSize != currSize {
			diffs = append(diffs, fmt.Sprintf("volume_claim[%s]: %s -> %s", name, prevSize, currSize))
		}
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

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j-1] > keys[j]; j-- {
			keys[j-1], keys[j] = keys[j], keys[j-1]
		}
	}
	return keys
}
