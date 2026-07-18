package controller

import (
	appsv1 "k8s.io/api/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// SetupWithManager registers the DriftReconciler with the controller-runtime
// manager so it starts receiving Deployment events. We filter to only
// generation-changing events (i.e. real spec changes) — status-only updates
// (replica rollout progress, observedGeneration, etc.) don't bump generation
// and would otherwise trigger redundant reconciles for the same drift.
func (r *DriftReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Named("drift-controller").
		Complete(r)
}
