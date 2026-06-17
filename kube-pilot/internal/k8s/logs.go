package k8s

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	corev1 "k8s.io/api/core/v1"
)

func (c *Client) StreamLogs(ns, podName string, tail int64) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	podLogs, err := c.Clientset.CoreV1().Pods(ns).GetLogs(podName, &corev1.PodLogOptions{
		Follow:    true,
		TailLines: &tail,
	}).Stream(ctx)
	if err != nil {
		return fmt.Errorf("failed to open log stream: %w", err)
	}
	defer podLogs.Close()

	if _, err = io.Copy(os.Stdout, podLogs); err != nil && ctx.Err() == nil {
		return fmt.Errorf("log stream interrupted: %w", err)
	}
	return nil
}
