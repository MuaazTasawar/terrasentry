package controller

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDiffConfigMapSnapshots_NoChange(t *testing.T) {
	snap := configMapSnapshot{Data: map[string]string{"LOG_LEVEL": "info", "TIMEOUT": "30s"}}
	got := diffConfigMapSnapshots(snap, snap)
	if got != "" {
		t.Errorf("expected empty diff for identical snapshots, got %q", got)
	}
}

func TestDiffConfigMapSnapshots_ValueChanged(t *testing.T) {
	prev := configMapSnapshot{Data: map[string]string{"LOG_LEVEL": "info"}}
	curr := configMapSnapshot{Data: map[string]string{"LOG_LEVEL": "debug"}}
	got := diffConfigMapSnapshots(prev, curr)
	want := "data[LOG_LEVEL]: changed"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDiffConfigMapSnapshots_KeyAdded(t *testing.T) {
	prev := configMapSnapshot{Data: map[string]string{"LOG_LEVEL": "info"}}
	curr := configMapSnapshot{Data: map[string]string{"LOG_LEVEL": "info", "FEATURE_FLAG_X": "true"}}
	got := diffConfigMapSnapshots(prev, curr)
	want := "data[FEATURE_FLAG_X]: added"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDiffConfigMapSnapshots_KeyRemoved(t *testing.T) {
	prev := configMapSnapshot{Data: map[string]string{"LOG_LEVEL": "info", "LEGACY_FLAG": "true"}}
	curr := configMapSnapshot{Data: map[string]string{"LOG_LEVEL": "info"}}
	got := diffConfigMapSnapshots(prev, curr)
	want := "data[LEGACY_FLAG]: removed"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDiffConfigMapSnapshots_MultipleKeysChanged(t *testing.T) {
	prev := configMapSnapshot{Data: map[string]string{"A_KEY": "1", "B_KEY": "2", "C_KEY": "3"}}
	curr := configMapSnapshot{Data: map[string]string{"A_KEY": "9", "B_KEY": "2", "C_KEY": "9"}}
	got := diffConfigMapSnapshots(prev, curr)
	want := "data[A_KEY]: changed; data[C_KEY]: changed"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDiffConfigMapSnapshots_EmptyToEmpty(t *testing.T) {
	prev := configMapSnapshot{Data: map[string]string{}}
	curr := configMapSnapshot{Data: map[string]string{}}
	got := diffConfigMapSnapshots(prev, curr)
	if got != "" {
		t.Errorf("expected empty diff for two empty data maps, got %q", got)
	}
}

func TestSnapshotFromConfigMap_ExtractsDataAndLabels(t *testing.T) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "api"}},
		Data:       map[string]string{"LOG_LEVEL": "info"},
	}
	snap := snapshotFromConfigMap(cm)
	if snap.Data["LOG_LEVEL"] != "info" {
		t.Errorf("expected data LOG_LEVEL=info, got %v", snap.Data)
	}
	if snap.Labels["app"] != "api" {
		t.Errorf("expected label app=api, got %v", snap.Labels)
	}
}

func TestSnapshotFromConfigMap_NilData(t *testing.T) {
	cm := &corev1.ConfigMap{Data: nil}
	snap := snapshotFromConfigMap(cm)
	if snap.Data == nil {
		t.Fatal("expected non-nil (empty) Data map, got nil")
	}
	if len(snap.Data) != 0 {
		t.Errorf("expected empty data map, got %v", snap.Data)
	}
}
