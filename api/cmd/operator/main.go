package main

import (
	"context"
	"log"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/MuaazTasawar/terrasentry/api/internal/config"
	"github.com/MuaazTasawar/terrasentry/api/internal/controller"
	"github.com/MuaazTasawar/terrasentry/api/internal/db"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(appsv1.AddToScheme(scheme))
}

func main() {
	cfg := config.Load()
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("operator: database connection failed: %v", err)
	}
	defer pool.Close()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		log.Fatalf("operator: unable to start manager: %v", err)
	}

	// Shared by all three reconcilers below — same insert, different
	// ResourceKind depending on which watcher found the drift.
	persistDrift := func(ctx context.Context, event db.DriftEvent) error {
		_, err := pool.Exec(ctx,
			`INSERT INTO drift_events (resource_kind, resource_name, namespace, diff)
			 VALUES ($1, $2, $3, $4)`,
			event.ResourceKind, event.ResourceName, event.Namespace, event.Diff,
		)
		return err
	}

	deploymentReconciler := &controller.DriftReconciler{
		Client: mgr.GetClient(), Scheme: mgr.GetScheme(), OnDriftFound: persistDrift,
	}
	if err := deploymentReconciler.SetupWithManager(mgr); err != nil {
		log.Fatalf("operator: unable to set up deployment drift controller: %v", err)
	}

	statefulSetReconciler := &controller.StatefulSetDriftReconciler{
		Client: mgr.GetClient(), Scheme: mgr.GetScheme(), OnDriftFound: persistDrift,
	}
	if err := statefulSetReconciler.SetupWithManager(mgr); err != nil {
		log.Fatalf("operator: unable to set up statefulset drift controller: %v", err)
	}

	configMapReconciler := &controller.ConfigMapDriftReconciler{
		Client: mgr.GetClient(), Scheme: mgr.GetScheme(), OnDriftFound: persistDrift,
	}
	if err := configMapReconciler.SetupWithManager(mgr); err != nil {
		log.Fatalf("operator: unable to set up configmap drift controller: %v", err)
	}

	log.Println("terrasentry operator started — watching Deployments, StatefulSets, and ConfigMaps for drift")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Fatalf("operator: manager exited with error: %v", err)
	}
}
