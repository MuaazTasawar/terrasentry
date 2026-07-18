package controller

import (
	appsv1 "k8s.io/api/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// SetupWithManager registers the DriftReconciler with the controller-runtime
// manager so it starts receiving Deployment events (create/update/delete).
func (r *DriftReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Named("drift-controller").
		Complete(r)
}
