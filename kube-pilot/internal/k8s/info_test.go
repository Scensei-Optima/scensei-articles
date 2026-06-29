package k8s_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/scensei-articles/kubepilot/internal/k8s"
)

func makeDiagPod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web-abc",
			Namespace: "scensei",
		},
		Spec: corev1.PodSpec{
			NodeName: "node-1",
			Containers: []corev1.Container{
				{
					Name:  "app",
					Image: "myrepo/myapp:1.2.3",
					Env: []corev1.EnvVar{
						{Name: "PORT", Value: "8080"},
						{Name: "ENV", Value: "production"},
					},
				},
			},
			Volumes: []corev1.Volume{
				{Name: "config", VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{}}},
				{Name: "data", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{}}},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			Conditions: []corev1.PodCondition{
				{Type: corev1.PodReady, Status: corev1.ConditionTrue},
			},
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "app", Ready: true},
			},
		},
	}
}

func TestGetPodDiagnostics_BasicFields(t *testing.T) {
	c := newTestClient(t, makeDiagPod())

	diag, err := c.GetPodDiagnostics("scensei", "web-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if diag.Node != "node-1" {
		t.Errorf("node: got %q, want %q", diag.Node, "node-1")
	}
	if diag.Phase != "Running" {
		t.Errorf("phase: got %q, want %q", diag.Phase, "Running")
	}
	if !diag.Ready {
		t.Error("expected Ready=true")
	}
}

func TestGetPodDiagnostics_ContainerImageAndEnv(t *testing.T) {
	c := newTestClient(t, makeDiagPod())

	diag, err := c.GetPodDiagnostics("scensei", "web-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(diag.Containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(diag.Containers))
	}
	ct := diag.Containers[0]
	if ct.Image != "myrepo/myapp:1.2.3" {
		t.Errorf("image: got %q, want %q", ct.Image, "myrepo/myapp:1.2.3")
	}
	if len(ct.Env) != 2 {
		t.Fatalf("expected 2 env vars, got %d", len(ct.Env))
	}
	if ct.Env[0].Name != "PORT" || ct.Env[0].Value != "8080" {
		t.Errorf("unexpected first env var: %+v", ct.Env[0])
	}
}

func TestGetPodDiagnostics_Volumes(t *testing.T) {
	c := newTestClient(t, makeDiagPod())

	diag, err := c.GetPodDiagnostics("scensei", "web-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(diag.Volumes) != 2 {
		t.Fatalf("expected 2 volumes, got %d", len(diag.Volumes))
	}
	if diag.Volumes[0].Type != "configMap" {
		t.Errorf("volume[0] type: got %q, want %q", diag.Volumes[0].Type, "configMap")
	}
	if diag.Volumes[1].Type != "persistentVolumeClaim" {
		t.Errorf("volume[1] type: got %q, want %q", diag.Volumes[1].Type, "persistentVolumeClaim")
	}
}

func TestGetPodDiagnostics_EnvValueFrom(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "ref-pod", Namespace: "scensei"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:  "app",
				Image: "myrepo/myapp:1.0.0",
				Env: []corev1.EnvVar{
					{
						Name: "DB_PASSWORD",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: "db-secret"},
								Key:                  "password",
							},
						},
					},
					{
						Name: "APP_CONFIG",
						ValueFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: "app-config"},
								Key:                  "config.yaml",
							},
						},
					},
					{Name: "PLAIN", Value: "hello"},
				},
			}},
		},
	}
	c := newTestClient(t, pod)

	diag, err := c.GetPodDiagnostics("scensei", "ref-pod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	env := diag.Containers[0].Env
	if env[0].Value != "<secret:db-secret/password>" {
		t.Errorf("secret ref: got %q", env[0].Value)
	}
	if env[1].Value != "<configMap:app-config/config.yaml>" {
		t.Errorf("configmap ref: got %q", env[1].Value)
	}
	if env[2].Value != "hello" {
		t.Errorf("plain value: got %q", env[2].Value)
	}
}

func TestGetPodDiagnostics_NotFound(t *testing.T) {
	c := newTestClient(t)

	_, err := c.GetPodDiagnostics("scensei", "missing-pod")
	if err == nil {
		t.Error("expected error for missing pod, got nil")
	}
}

func TestPodDiagnostics_PrintText(t *testing.T) {
	c := newTestClient(t, makeDiagPod())

	diag, err := c.GetPodDiagnostics("scensei", "web-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	diag.PrintText(&buf)
	out := buf.String()

	for _, want := range []string{"web-abc", "node-1", "Running", "myrepo/myapp:1.2.3", "PORT=8080", "configMap", "persistentVolumeClaim"} {
		if !strings.Contains(out, want) {
			t.Errorf("PrintText output missing %q\nGot:\n%s", want, out)
		}
	}
}

func TestPodDiagnostics_JSONRoundtrip(t *testing.T) {
	c := newTestClient(t, makeDiagPod())

	diag, err := c.GetPodDiagnostics("scensei", "web-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := json.Marshal(diag)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var got k8s.PodDiagnostics
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if got.Node != diag.Node {
		t.Errorf("node after roundtrip: got %q, want %q", got.Node, diag.Node)
	}
	if len(got.Containers) != len(diag.Containers) {
		t.Errorf("containers after roundtrip: got %d, want %d", len(got.Containers), len(diag.Containers))
	}
}
