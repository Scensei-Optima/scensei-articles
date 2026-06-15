package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/scensei-articles/kubepilot/internal/k8s"
	"github.com/scensei-articles/kubepilot/internal/ui"
)

func main() {
	localPort := flag.Int("local", 8080, "The local port to bind to")
	namespace := flag.String("n", "default", "The Kubernetes namespace")
	label := flag.String("l", "", "The label selector for pods")
	remotePortFlag := flag.Int("remote", 0, "The remote port (overrides pod spec)")
	attachMode := flag.Bool("attach", false, "Execute an interactive shell in the pod instead of port-forwarding")
	logMode := flag.Bool("logs", false, "Stream logs from the pod")

	flag.Parse()

	if *attachMode && *logMode {
		fmt.Fprintln(os.Stderr, "Error: cannot use -attach and -logs at the same time.")
		os.Exit(1)
	}

	ui.PrintLogo()

	client, err := k8s.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Searching in namespace: %s\n", *namespace)
	if *label != "" {
		fmt.Printf("> Using label filter: \"%s\"\n", *label)
	}

	pods, err := client.ListPods(*namespace, *label, *remotePortFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(pods) == 0 {
		fmt.Printf("No pods found in namespace %s\n", *namespace)
		return
	}

	var selectorLabel string
	switch {
	case *attachMode:
		selectorLabel = "Select a Pod to Attach to"
	case *logMode:
		selectorLabel = "Select a Pod to Stream Logs from"
	default:
		selectorLabel = "Select a Pod to Port-Forward to"
	}

	selectedPod, remotePort, err := ui.SelectPod(pods, selectorLabel)
	if err != nil {
		fmt.Println("\nCancelled.")
		return
	}

	switch {
	case *attachMode:
		fmt.Printf("💻 Attaching to %s...\n", selectedPod)
		if err := client.ExecShell(*namespace, selectedPod); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case *logMode:
		fmt.Printf("📋 Streaming logs from %s...\n", selectedPod)
		if err := client.StreamLogs(*namespace, selectedPod); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	default:
		ports := fmt.Sprintf("%d:%d", *localPort, remotePort)
		fmt.Printf("🚀 Tunneling: http://localhost:%d -> %s:%d (Ctrl+C to stop)\n", *localPort, selectedPod, remotePort)
		if err := client.PortForward(*namespace, selectedPod, ports); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}
