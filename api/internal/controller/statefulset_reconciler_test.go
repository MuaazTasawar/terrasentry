package controller

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDiffStatefulSetSnapshots_NoChange(t *testing.T) {
	snap := statefulSetSnapshot{
		Replicas: 3, Image: "postgres:16", CPULimit: "500m", MemLimit: "512Mi",
		VolumeClaimSizes: map[string]string{"data": "10Gi"},
	}
	got := diffStatefulSetSnapshots(snap, snap)
	if got != "" {
		t.Errorf("expected empty diff for identical snapshots, got %q", got)
	}
}

func TestDiffStatefulSetSnapshots_ReplicasChanged(t *testing.T) {
	prev := statefulSetSnapshot{Replicas: 3, Image: "postgres:16"}
	curr := statefulSetSnapshot{Replicas: 5, Image: "postgres:16"}
	got := diffStatefulSetSnapshots(prev, curr)
	want := "replicas: 3 -> 5"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDiffStatefulSetSnapshots_VolumeClaimSizeChanged(t *testing.T) {
	prev := statefulSetSnapshot{Replicas: 3, VolumeClaimSizes: map[string]string{"data": "10Gi"}}
	curr := statefulSetSnapshot{Replicas: 3, VolumeClaimSizes: map[string]string{"data": "20Gi"}}
	got := diffStatefulSetSnapshots(prev, curr)
	want := "volume_claim[data]: 10Gi -> 20Gi"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDiffStatefulSetSnapshots_VolumeClaimAddedAndRemoved(t *testing.T) {
	prev := statefulSetSnapshot{VolumeClaimSizes: map[string]string{"data": "10Gi"}}
	curr := statefulSetSnapshot{VolumeClaimSizes: map[string]string{"data": "10Gi", "logs": "5Gi"}}
	got := diffStatefulSetSnapshots(prev, curr)
	want := "volume_claim[logs]:  -> 5Gi"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDiffStatefulSetSnapshots_MultipleFieldsChanged(t *testing.T) {
	prev := statefulSetSnapshot{Replicas: 3, Image: "postgres:15", VolumeClaimSizes: map[string]string{"data": "10Gi"}}
	curr := statefulSetSnapshot{Replicas: 3, Image: "postgres:16", VolumeClaimSizes: map[string]string{"data": "20Gi"}}
	got := diffStatefulSetSnapshots(prev, curr)
	want := "image: postgres:15 -> postgres:16; volume_claim[data]: 10Gi -> 20Gi"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSnapshotFromStatefulSet_ExtractsCoreFields(t *testing.T) {
	replicas := int32(3)
	sts := &appsv1.StatefulSet{
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "postgres"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "postgres:16",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "data"},
					Spec: corev1.PersistentVolumeClaimSpec{
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("10Gi")},
						},
					},
				},
			},
		},
	}

	snap := snapshotFromStatefulSet(sts)
	if snap.Replicas != 3 {
		t.Errorf("expected replicas 3, got %d", snap.Replicas)
	}
	if snap.Image != "postgres:16" {
		t.Errorf("expected image postgres:16, got %s", snap.Image)
	}
	if snap.VolumeClaimSizes["data"] != "10Gi" {
		t.Errorf("expected volume claim 'data' size 10Gi, got %s", snap.VolumeClaimSizes["data"])
	}
	if snap.Labels["app"] != "postgres" {
		t.Errorf("expected label app=postgres, got %v", snap.Labels)
	}
}

func TestSnapshotFromStatefulSet_NilReplicas(t *testing.T) {
	sts := &appsv1.StatefulSet{
		Spec: appsv1.StatefulSetSpec{
			Replicas: nil,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{Containers: []corev1.Container{{Image: "postgres:16"}}},
			},
		},
	}
	snap := snapshotFromStatefulSet(sts)
	if snap.Replicas != 0 {
		t.Errorf("expected replicas 0 for nil pointer, got %d", snap.Replicas)
	}
}

func TestSnapshotFromStatefulSet_NoVolumeClaimTemplates(t *testing.T) {
	replicas := int32(1)
	sts := &appsv1.StatefulSet{
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{Containers: []corev1.Container{{Image: "redis:7"}}},
			},
			VolumeClaimTemplates: nil,
		},
	}
	snap := snapshotFromStatefulSet(sts)
	if snap.VolumeClaimSizes == nil {
		t.Fatal("expected non-nil (empty) VolumeClaimSizes map, got nil")
	}
	if len(snap.VolumeClaimSizes) != 0 {
		t.Errorf("expected no volume claims, got %v", snap.VolumeClaimSizes)
	}
}

func TestSnapshotFromStatefulSet_NoContainers(t *testing.T) {
	replicas := int32(1)
	sts := &appsv1.StatefulSet{
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{}}},
		},
	}
	snap := snapshotFromStatefulSet(sts)
	if snap.Image != "" {
		t.Errorf("expected empty image for no containers, got %s", snap.Image)
	}
}
