package controller

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// SetupWithManager registers the DriftReconciler with the controller-runtime
// manager so it starts receiving Deployment events. We filter to only
// generation-changing events (i.e. real spec changes) — status-only updates
// don't bump generation and would otherwise trigger redundant reconciles.
func (r *DriftReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Named("drift-controller").
		Complete(r)
}

// SetupWithManager registers the StatefulSetDriftReconciler, following the
// same generation-changed filtering as the Deployment reconciler above.
func (r *StatefulSetDriftReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.StatefulSet{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Named("statefulset-drift-controller").
		Complete(r)
}

// SetupWithManager registers the ConfigMapDriftReconciler. ConfigMaps have
// no generation field, so GenerationChangedPredicate doesn't apply here —
// every update to a ConfigMap's data is a real, reconcile-worthy change.
func (r *ConfigMapDriftReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		Named("configmap-drift-controller").
		Complete(r)
}
