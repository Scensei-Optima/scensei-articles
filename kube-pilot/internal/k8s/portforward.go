package k8s

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

func (c *Client) PortForward(ns, podName, ports string) error {
	stopCh := make(chan struct{}, 1)
	readyCh := make(chan struct{})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigs)
	go func() {
		<-sigs
		close(stopCh)
	}()

	roundTripper, upgrader, err := spdy.RoundTripperFor(c.Config)
	if err != nil {
		return fmt.Errorf("failed to create SPDY transport: %w", err)
	}

	serverURL, err := url.Parse(c.Config.Host)
	if err != nil {
		return fmt.Errorf("failed to parse API server URL: %w", err)
	}
	serverURL.Path = fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", ns, podName)

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, serverURL)

	fw, err := portforward.New(dialer, []string{ports}, stopCh, readyCh, os.Stdout, os.Stderr)
	if err != nil {
		return fmt.Errorf("failed to initialize port-forwarder: %w", err)
	}

	if err := fw.ForwardPorts(); err != nil {
		select {
		case <-stopCh:
			return nil
		default:
			return fmt.Errorf("port-forward failed: %w", err)
		}
	}
	return nil
}
