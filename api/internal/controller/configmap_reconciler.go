package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/MuaazTasawar/terrasentry/api/internal/db"
)

// configMapSnapshot captures the fields we care about for drift detection
// on ConfigMaps. Unlike Deployments/StatefulSets there's no replicas/image
// to track — the entire surface area that matters is the data keys/values
// themselves, since that's what workloads actually read.
type configMapSnapshot struct {
	Data   map[string]string `json:"data"`
	Labels map[string]string `json:"labels"`
}

// ConfigMapDriftReconciler watches ConfigMaps and records DriftEvents when
// live data diverges from the last snapshot we took. Mirrors
// DriftReconciler's snapshot/diff/annotation approach for Deployments.
type ConfigMapDriftReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	OnDriftFound func(ctx context.Context, event db.DriftEvent) error
}

func (r *ConfigMapDriftReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var cm corev1.ConfigMap
	if err := r.Get(ctx, req.NamespacedName, &cm); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to fetch configmap: %w", err)
	}

	current := snapshotFromConfigMap(&cm)

	previousRaw, hasSnapshot := cm.Annotations[LastAppliedAnnotation]
	if !hasSnapshot {
		return ctrl.Result{}, r.saveSnapshot(ctx, &cm, current)
	}

	var previous configMapSnapshot
	if err := json.Unmarshal([]byte(previousRaw), &previous); err != nil {
		log.Printf("warning: could not parse previous snapshot for %s/%s, resetting baseline: %v",
			cm.Namespace, cm.Name, err)
		return ctrl.Result{}, r.saveSnapshot(ctx, &cm, current)
	}

	diff := diffConfigMapSnapshots(previous, current)
	if diff == "" {
		return ctrl.Result{}, nil
	}

	log.Printf("drift detected on configmap %s/%s: %s", cm.Namespace, cm.Name, diff)

	if r.OnDriftFound != nil {
		event := db.DriftEvent{
			ResourceKind: "ConfigMap",
			ResourceName: cm.Name,
			Namespace:    cm.Namespace,
			Diff:         diff,
		}
		if err := r.OnDriftFound(ctx, event); err != nil {
			log.Printf("failed to persist drift event: %v", err)
		}
	}

	return ctrl.Result{}, r.saveSnapshot(ctx, &cm, current)
}

func (r *ConfigMapDriftReconciler) saveSnapshot(ctx context.Context, cm *corev1.ConfigMap, snap configMapSnapshot) error {
	raw, err := json.Marshal(snap)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest := &corev1.ConfigMap{}
		if err := r.Get(ctx, client.ObjectKeyFromObject(cm), latest); err != nil {
			return err
		}

		if latest.Annotations == nil {
			latest.Annotations = map[string]string{}
		}
		latest.Annotations[LastAppliedAnnotation] = string(raw)

		return r.Update(ctx, latest)
	})
}

func snapshotFromConfigMap(cm *corev1.ConfigMap) configMapSnapshot {
	data := map[string]string{}
	for k, v := range cm.Data {
		data[k] = v
	}
	return configMapSnapshot{Data: data, Labels: cm.Labels}
}

func diffConfigMapSnapshots(prev, curr configMapSnapshot) string {
	var diffs []string

	keySet := map[string]bool{}
	for k := range prev.Data {
		keySet[k] = true
	}
	for k := range curr.Data {
		keySet[k] = true
	}

	for _, key := range sortedKeys(keySet) {
		prevVal, hadKey := prev.Data[key]
		currVal, hasKey := curr.Data[key]

		switch {
		case hadKey && !hasKey:
			diffs = append(diffs, fmt.Sprintf("data[%s]: removed", key))
		case !hadKey && hasKey:
			diffs = append(diffs, fmt.Sprintf("data[%s]: added", key))
		case hadKey && hasKey && prevVal != currVal:
			diffs = append(diffs, fmt.Sprintf("data[%s]: changed", key))
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
