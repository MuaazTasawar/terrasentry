package controller

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDiffSnapshots_NoChange(t *testing.T) {
	snap := deploymentSnapshot{
		Replicas: 3,
		Image:    "nginx:1.25",
		CPULimit: "250m",
		MemLimit: "128Mi",
	}

	got := diffSnapshots(snap, snap)
	if got != "" {
		t.Errorf("expected empty diff for identical snapshots, got %q", got)
	}
}

func TestDiffSnapshots_ReplicasChanged(t *testing.T) {
	prev := deploymentSnapshot{Replicas: 2, Image: "nginx:1.25"}
	curr := deploymentSnapshot{Replicas: 5, Image: "nginx:1.25"}

	got := diffSnapshots(prev, curr)
	want := "replicas: 2 -> 5"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDiffSnapshots_ImageChanged(t *testing.T) {
	prev := deploymentSnapshot{Replicas: 2, Image: "nginx:1.25"}
	curr := deploymentSnapshot{Replicas: 2, Image: "nginx:1.27"}

	got := diffSnapshots(prev, curr)
	want := "image: nginx:1.25 -> nginx:1.27"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDiffSnapshots_MultipleFieldsChanged(t *testing.T) {
	prev := deploymentSnapshot{
		Replicas: 2, Image: "nginx:1.25", CPULimit: "100m", MemLimit: "64Mi",
	}
	curr := deploymentSnapshot{
		Replicas: 4, Image: "nginx:1.27", CPULimit: "100m", MemLimit: "64Mi",
	}

	got := diffSnapshots(prev, curr)
	want := "replicas: 2 -> 4; image: nginx:1.25 -> nginx:1.27"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDiffSnapshots_CPUAndMemChanged(t *testing.T) {
	prev := deploymentSnapshot{Replicas: 2, CPULimit: "100m", MemLimit: "64Mi"}
	curr := deploymentSnapshot{Replicas: 2, CPULimit: "500m", MemLimit: "256Mi"}

	got := diffSnapshots(prev, curr)
	want := "cpu_limit: 100m -> 500m; mem_limit: 64Mi -> 256Mi"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSnapshotFromDeployment_ExtractsCoreFields(t *testing.T) {
	replicas := int32(3)
	deploy := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "demo"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "nginx:1.25",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("250m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
							},
						},
					},
				},
			},
		},
	}

	snap := snapshotFromDeployment(deploy)

	if snap.Replicas != 3 {
		t.Errorf("expected replicas 3, got %d", snap.Replicas)
	}
	if snap.Image != "nginx:1.25" {
		t.Errorf("expected image nginx:1.25, got %s", snap.Image)
	}
	if snap.CPULimit != "250m" {
		t.Errorf("expected cpu limit 250m, got %s", snap.CPULimit)
	}
	if snap.MemLimit != "128Mi" {
		t.Errorf("expected mem limit 128Mi, got %s", snap.MemLimit)
	}
	if snap.Labels["app"] != "demo" {
		t.Errorf("expected label app=demo, got %v", snap.Labels)
	}
}

func TestSnapshotFromDeployment_NilReplicas(t *testing.T) {
	// A Deployment with no explicit replicas set (nil pointer) shouldn't panic
	// and should default to zero rather than dereferencing a nil pointer.
	deploy := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: nil,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Image: "nginx:1.25"}},
				},
			},
		},
	}

	snap := snapshotFromDeployment(deploy)
	if snap.Replicas != 0 {
		t.Errorf("expected replicas 0 for nil pointer, got %d", snap.Replicas)
	}
}

func TestSnapshotFromDeployment_NoContainers(t *testing.T) {
	// Defensive case: a pod spec with zero containers shouldn't panic when
	// indexing into Containers[0].
	replicas := int32(1)
	deploy := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{Containers: []corev1.Container{}},
			},
		},
	}

	snap := snapshotFromDeployment(deploy)
	if snap.Image != "" {
		t.Errorf("expected empty image for no containers, got %s", snap.Image)
	}
}
