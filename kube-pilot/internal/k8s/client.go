package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	Clientset kubernetes.Interface
	Config    *rest.Config
}

const kubeconfigEnv = "KUBECONFIG"

func getDefaultKubeconfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}

	return filepath.Join(home, ".kube", "config"), nil
}

func getKubeConfigPath() (string, error) {
	if kubeconfig, ok := os.LookupEnv(kubeconfigEnv); ok {
		return kubeconfig, nil
	}
	return getDefaultKubeconfigPath()
}

func NewClient() (*Client, error) {
	kubeconfig, err := getKubeConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to obtain kubeconfig file path: %w", err)
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig from %s: %w", kubeconfig, err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return &Client{Clientset: clientset, Config: config}, nil
}

func (c *Client) ListPods(namespace, labelSelector string, remotePortOverride int) (map[string]int32, error) {
	lOpts := metav1.ListOptions{}
	if labelSelector != "" {
		lOpts.LabelSelector = labelSelector
	}

	pods, err := c.Clientset.CoreV1().Pods(namespace).List(context.TODO(), lOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	result := make(map[string]int32, len(pods.Items))
	for _, p := range pods.Items {
		var port int32
		if remotePortOverride > 0 {
			port = int32(remotePortOverride)
		} else if len(p.Spec.Containers) > 0 && len(p.Spec.Containers[0].Ports) > 0 {
			port = p.Spec.Containers[0].Ports[0].ContainerPort
		} else {
			port = 80
		}
		result[p.Name] = port
	}

	return result, nil
}
