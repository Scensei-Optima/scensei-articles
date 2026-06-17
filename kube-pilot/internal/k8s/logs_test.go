package k8s_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/scensei-articles/kubepilot/internal/k8s"
)

func newHTTPTestClient(t *testing.T, handler http.HandlerFunc) (*k8s.Client, func()) {
	t.Helper()
	ts := httptest.NewServer(handler)
	cs, err := kubernetes.NewForConfig(&rest.Config{Host: ts.URL})
	if err != nil {
		ts.Close()
		t.Fatalf("failed to create clientset: %v", err)
	}
	c := &k8s.Client{
		Clientset: cs,
		Config:    &rest.Config{Host: ts.URL},
	}
	return c, ts.Close
}

func TestStreamLogs_ReturnsNilOnCleanEOF(t *testing.T) {
	c, teardown := newHTTPTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "2024-01-01 app started")
		fmt.Fprintln(w, "2024-01-01 listening on :8080")
	})
	defer teardown()

	if err := c.StreamLogs("default", "test-pod", 100); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStreamLogs_ReturnsErrorOnServerFailure(t *testing.T) {
	c, teardown := newHTTPTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	})
	defer teardown()

	if err := c.StreamLogs("default", "test-pod", 100); err == nil {
		t.Error("expected error from 500 response, got nil")
	}
}

func TestStreamLogs_ReturnsErrorOnUnauthorized(t *testing.T) {
	c, teardown := newHTTPTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})
	defer teardown()

	if err := c.StreamLogs("default", "test-pod", 100); err == nil {
		t.Error("expected error from 401 response, got nil")
	}
}
