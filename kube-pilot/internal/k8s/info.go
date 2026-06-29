package k8s

import (
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ContainerDiag struct {
	Name  string   `json:"name"`
	Image string   `json:"image"`
	Ready bool     `json:"ready"`
	Env   []EnvVar `json:"env"`
}

type VolumeDiag struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type PodDiagnostics struct {
	Pod        string          `json:"pod"`
	Namespace  string          `json:"namespace"`
	Node       string          `json:"node"`
	Phase      string          `json:"phase"`
	Ready      bool            `json:"ready"`
	Containers []ContainerDiag `json:"containers"`
	Volumes    []VolumeDiag    `json:"volumes"`
}

func (c *Client) GetPodDiagnostics(ns, podName string) (*PodDiagnostics, error) {
	pod, err := c.Clientset.CoreV1().Pods(ns).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get pod %s: %w", podName, err)
	}

	readiness := map[string]bool{}
	for _, cs := range pod.Status.ContainerStatuses {
		readiness[cs.Name] = cs.Ready
	}

	diag := &PodDiagnostics{
		Pod:       pod.Name,
		Namespace: pod.Namespace,
		Node:      pod.Spec.NodeName,
		Phase:     string(pod.Status.Phase),
		Ready:     isPodReady(pod),
	}

	for _, ct := range pod.Spec.Containers {
		cd := ContainerDiag{
			Name:  ct.Name,
			Image: ct.Image,
			Ready: readiness[ct.Name],
		}
		for _, e := range ct.Env {
			cd.Env = append(cd.Env, EnvVar{Name: e.Name, Value: resolveEnvValue(e)})
		}
		diag.Containers = append(diag.Containers, cd)
	}

	for _, v := range pod.Spec.Volumes {
		diag.Volumes = append(diag.Volumes, VolumeDiag{
			Name: v.Name,
			Type: volumeType(v),
		})
	}

	return diag, nil
}

func isPodReady(pod *corev1.Pod) bool {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == corev1.PodReady {
			return cond.Status == corev1.ConditionTrue
		}
	}
	return false
}

func resolveEnvValue(e corev1.EnvVar) string {
	if e.ValueFrom == nil {
		return e.Value
	}
	switch {
	case e.ValueFrom.SecretKeyRef != nil:
		return fmt.Sprintf("<secret:%s/%s>", e.ValueFrom.SecretKeyRef.Name, e.ValueFrom.SecretKeyRef.Key)
	case e.ValueFrom.ConfigMapKeyRef != nil:
		return fmt.Sprintf("<configMap:%s/%s>", e.ValueFrom.ConfigMapKeyRef.Name, e.ValueFrom.ConfigMapKeyRef.Key)
	case e.ValueFrom.FieldRef != nil:
		return fmt.Sprintf("<fieldRef:%s>", e.ValueFrom.FieldRef.FieldPath)
	case e.ValueFrom.ResourceFieldRef != nil:
		return fmt.Sprintf("<resourceField:%s>", e.ValueFrom.ResourceFieldRef.Resource)
	default:
		return "<unknown source>"
	}
}

func volumeType(v corev1.Volume) string {
	switch {
	case v.ConfigMap != nil:
		return "configMap"
	case v.Secret != nil:
		return "secret"
	case v.PersistentVolumeClaim != nil:
		return "persistentVolumeClaim"
	case v.EmptyDir != nil:
		return "emptyDir"
	case v.HostPath != nil:
		return "hostPath"
	case v.Projected != nil:
		return "projected"
	default:
		return "unknown"
	}
}

func (d *PodDiagnostics) PrintText(w io.Writer) {
	fmt.Fprintf(w, "Pod:       %s\n", d.Pod)
	fmt.Fprintf(w, "Namespace: %s\n", d.Namespace)
	fmt.Fprintf(w, "Node:      %s\n", d.Node)
	fmt.Fprintf(w, "Phase:     %s\n", d.Phase)
	fmt.Fprintf(w, "Ready:     %v\n", d.Ready)

	for _, ct := range d.Containers {
		fmt.Fprintf(w, "\n--- Container: %s ---\n", ct.Name)
		fmt.Fprintf(w, "Image: %s\n", ct.Image)
		fmt.Fprintf(w, "Ready: %v\n", ct.Ready)
		if len(ct.Env) > 0 {
			fmt.Fprintln(w, "Env:")
			for _, e := range ct.Env {
				fmt.Fprintf(w, "  %s=%s\n", e.Name, e.Value)
			}
		}
	}

	if len(d.Volumes) > 0 {
		fmt.Fprintln(w, "\n--- Volumes ---")
		for _, v := range d.Volumes {
			fmt.Fprintf(w, "  %s (%s)\n", v.Name, v.Type)
		}
	}
}
