package k8s_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"

	"github.com/scensei-articles/kubepilot/internal/k8s"
)

func newTestClient(t *testing.T, objects ...runtime.Object) *k8s.Client {
	t.Helper()
	return &k8s.Client{
		Clientset: fake.NewSimpleClientset(objects...),
		Config:    &rest.Config{Host: "http://fake"},
	}
}

func makePod(name, namespace string, port int32, labels map[string]string) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
	if port > 0 {
		pod.Spec.Containers = []corev1.Container{{
			Ports: []corev1.ContainerPort{{ContainerPort: port}},
		}}
	}
	return pod
}

func TestListPods_PortFromSpec(t *testing.T) {
	c := newTestClient(t, makePod("web", "default", 3000, nil))

	pods, err := c.ListPods("default", "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := pods["web"], int32(3000); got != want {
		t.Errorf("port: got %d, want %d", got, want)
	}
}

func TestListPods_FallbackTo80WhenNoPortsDefined(t *testing.T) {
	c := newTestClient(t, makePod("web", "default", 0, nil))

	pods, err := c.ListPods("default", "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := pods["web"], int32(80); got != want {
		t.Errorf("port: got %d, want %d", got, want)
	}
}

func TestListPods_RemotePortOverrideWinsOverSpec(t *testing.T) {
	c := newTestClient(t, makePod("web", "default", 3000, nil))

	pods, err := c.ListPods("default", "", 9090)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := pods["web"], int32(9090); got != want {
		t.Errorf("port: got %d, want %d", got, want)
	}
}

func TestListPods_LabelSelectorFilters(t *testing.T) {
	c := newTestClient(t,
		makePod("web", "default", 3000, map[string]string{"app": "web"}),
		makePod("db", "default", 5432, map[string]string{"app": "db"}),
	)

	pods, err := c.ListPods("default", "app=web", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pods) != 1 {
		t.Fatalf("expected 1 pod, got %d", len(pods))
	}
	if _, ok := pods["web"]; !ok {
		t.Error("expected pod 'web' in result")
	}
}

func TestListPods_OnlyReturnsPodsFromRequestedNamespace(t *testing.T) {
	c := newTestClient(t,
		makePod("web", "default", 3000, nil),
		makePod("api", "staging", 8080, nil),
	)

	pods, err := c.ListPods("default", "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pods) != 1 {
		t.Errorf("expected 1 pod, got %d: %v", len(pods), pods)
	}
}

func TestListPods_ReturnsEmptyMapWhenNoneExist(t *testing.T) {
	c := newTestClient(t)

	pods, err := c.ListPods("default", "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pods) != 0 {
		t.Errorf("expected empty map, got %d entries", len(pods))
	}
}

func TestListPods_MultiplePodsAllReturned(t *testing.T) {
	c := newTestClient(t,
		makePod("web", "default", 3000, nil),
		makePod("api", "default", 8080, nil),
		makePod("worker", "default", 0, nil),
	)

	pods, err := c.ListPods("default", "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pods) != 3 {
		t.Errorf("expected 3 pods, got %d", len(pods))
	}
	if pods["web"] != 3000 || pods["api"] != 8080 || pods["worker"] != 80 {
		t.Errorf("unexpected ports: %v", pods)
	}
}
