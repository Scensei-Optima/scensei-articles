package k8s

import (
	"os"
	"testing"
)

func TestCustomKubeconfigPath(t *testing.T) {
	customPath := "/etc/rancher/k3s/config"

	old := os.Getenv(kubeconfigEnv)
	defer os.Setenv(kubeconfigEnv, old)

	err := os.Setenv(kubeconfigEnv, customPath)
	if err != nil {
		t.Fatal(err)
	}

	kubeconfig, err := getKubeConfigPath()
	if err != nil {
		t.Fatal(err)
	}

	if kubeconfig != customPath {
		t.Errorf("wrong retreived kubeconfig path\nexpected: %s\ngot: %s", customPath, kubeconfig)
	}
}
