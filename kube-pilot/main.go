package main

import (
	"context"
	"flag" // Added for CLI arguments
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"syscall"

	"github.com/manifoldco/promptui"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/util/term"
)

func main() {
	// 1. Define CLI Flags
	localPort := flag.Int("local", 8080, "The local port to bind to")
	namespace := flag.String("n", "default", "The Kubernetes namespace")
	label := flag.String("l", "", "The label selector for pods")
	remotePortFlag := flag.Int("remote", 0, "The remote port (overrides pod spec)")
	attachMode := flag.Bool("attach", false, "Execute an interactive shell in the pod instead of port-forwarding")
	logMode := flag.Bool("logs", false, "Stream logs from the pod")

	flag.Parse()

	// Mutual Exclusivity Check
	if *attachMode && *logMode {
		fmt.Println("❌ Error: You cannot use -attach and -logs at the same time.")
		os.Exit(1)
	}

	printLogo()

	// 2. Setup Kubernetes Client (Standard boilerplate)
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, _ := clientcmd.BuildConfigFromFlags("", kubeconfig)
	clientset, _ := kubernetes.NewForConfig(config)

	// 3. Fetch Pods using the dereferenced pointers
	fmt.Printf("Searching in namespace: %s\n", *namespace)

	var lOpts metav1.ListOptions

	if *label != "" {
		fmt.Printf("> Using label filter: \"%s\"\n", *label)
		lOpts = metav1.ListOptions{LabelSelector: *label}
	}

	pods, err := clientset.CoreV1().Pods(*namespace).List(context.TODO(), lOpts)
	if err != nil {
		panic(err)
	}

	fetchedPodData := make(map[string]int32)
	for _, p := range pods.Items {
		var port int32

		// 1. Check if user provided a manual override flag
		if *remotePortFlag > 0 {
			port = int32(*remotePortFlag)
		} else if len(p.Spec.Containers) > 0 && len(p.Spec.Containers[0].Ports) > 0 {
			// 2. Try to grab it from the spec
			port = p.Spec.Containers[0].Ports[0].ContainerPort
		} else {
			// 3. Fallback to a common default (like 80) or skip
			port = 80
		}

		fetchedPodData[p.Name] = port
	}

	if len(fetchedPodData) == 0 {
		fmt.Printf("No pods found in namespace %s\n", *namespace)
		return
	}

	var displayLabel string

	if *attachMode {
		displayLabel = "Select a Pod to Attach to"
	} else if *logMode {
		displayLabel = "Select a Pod to Stream Logs from"
	} else {
		displayLabel = "Select a Pod to Port-Forward to"
	}

	// 4. Interactive Menu
	prompt := promptui.Select{
		Label: displayLabel,
		Items: slices.Collect(maps.Keys(fetchedPodData)),
	}

	_, selectedPod, _ := prompt.Run()
	remotePort := fetchedPodData[selectedPod]

	if *attachMode {
		fmt.Printf("💻 Attaching to %s...\n", selectedPod)
		executeRemoteShell(config, clientset, *namespace, selectedPod)
		return
	}

	if *logMode {
		fmt.Printf("📋 Streaming logs from %s...\n", selectedPod)
		streamLogs(clientset, *namespace, selectedPod)
		return
	}

	// 5. Establish Tunnel
	// Note: We use the *localPort pointer from the flags
	ports := fmt.Sprintf("%d:%d", *localPort, remotePort)
	fmt.Printf("🚀 Tunneling: http://localhost:%d -> %s:%d (Ctrl+C to stop)\n", *localPort, selectedPod, remotePort)

	establishPortForward(config, *namespace, selectedPod, ports)
}

func establishPortForward(config *rest.Config, ns, podName, ports string) {
	stopCh := make(chan struct{}, 1)
	readyCh := make(chan struct{})

	// Handle Ctrl+C
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		close(stopCh)
	}()

	// Build the Request URL
	roundTripper, upgrader, _ := spdy.RoundTripperFor(config)
	serverURL, _ := url.Parse(config.Host)
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", ns, podName)
	serverURL.Path = path

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, serverURL)

	fw, _ := portforward.New(dialer, []string{ports}, stopCh, readyCh, os.Stdout, os.Stderr)

	if err := fw.ForwardPorts(); err != nil {
		panic(err)
	}
}

func executeRemoteShell(config *rest.Config, clientset *kubernetes.Clientset, ns, podName string) {
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(ns).
		SubResource("exec")

	// Specify we want an interactive TTY
	option := &corev1.PodExecOptions{
		Command: []string{"/bin/sh"},
		Stdin:   true,
		Stdout:  true,
		Stderr:  true,
		TTY:     true,
	}
	req.VersionedParams(option, scheme.ParameterCodec)

	exec, _ := remotecommand.NewSPDYExecutor(config, "POST", req.URL())

	// Put the local terminal into Raw Mode to handle interactive input correctly
	t := term.TTY{In: os.Stdin, Out: os.Stdout, Raw: true}

	// This starts the session and blocks until the user exits
	_ = t.Safe(func() error {
		return exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
			Stdin:  os.Stdin,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Tty:    true,
		})
	})
}

func streamLogs(clientset *kubernetes.Clientset, ns, podName string) {
	podLogOpts := &corev1.PodLogOptions{
		Follow:    true,
		TailLines: int64Ptr(100), // Show last 100 lines before tailing
	}

	req := clientset.CoreV1().Pods(ns).GetLogs(podName, podLogOpts)
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		fmt.Printf("Error opening log stream: %v\n", err)
		return
	}
	defer podLogs.Close()

	// Stream logs directly to terminal output
	_, err = io.Copy(os.Stdout, podLogs)
	if err != nil {
		fmt.Printf("Error during log streaming: %v\n", err)
	}
}

func int64Ptr(i int64) *int64 { return &i }

func printLogo() {
	logo := `
..-=++******+==-:..                     .:-.                                                                                                         
=################**+-:.             .:+*###+:.                                                                                                      
+############%%%%%%%%#*+-:.       .-*####%%###+.             .:-----:..                                                                         
:###########*==+++**##%%%#*=:. .=*######*=-:-#*:           :+########*:                                                                        
.=##########+.    ...:-+*#%%#*=*#####*-:.    .*%*:        .*%#*:...:-:.    .:-====-:.     .-====-:.     :--:.:-===-:.    .:-====--:.    .:-===-:.    .---:            
 .+##########-           .:-+*#%####=:.      :*##*:        .=#%#+-:.      :+###***##-.   :+###+++##*-.  -###*#***###+.  :*##*++***-. .=###+++###=.   -##%+            
  .=#########+.             .=######=:        -####+.        .-+*#%##*=:  .*%#*:......  .*%#*-:::-#%#-  -###*:...=###-  -#%#+-:...   .+###-:::-*%%=. :###=            
   .-*########=            .=###+--+##+-.    :*#####:           .:-+*#%#- :###+         -####********-  -###=    -###-  .-+*####*=: .####********=.  :###=            
     .+########-          -###=.   :-*#*-..+######-       .::..    -##%+ .*%#*:.   ...  :*%#*:........  -###=    -###- ... ..:=##%*..+%#*-........   :###=            
       :+######*:        .*%*-        .:=*+*#######:      :###****###+:  :*#%#*+**#-    :+###*+++**=.   -#%#=    -#%%- .=**++=+###=. .+###*++++*+.   -#%%+            
         :+#####*-        -#*:            .:=*#####=.      .:-=+++==-:.    .:-=+++=-.    .:-=++++=-:.   :===:    :===:  :==++++=-:      .-==+++==:.  :===-            
           .-*##%#=.    .+#:                .-+*+-.                                                                                                         
             .:=*#%*-.  .*=                    .                                   +---------------------------------------------------------------+            
                .:-+*+-:.+:                                                        |                 Welcome to Kubepilot by Scensei               |            
                    .:--::.                                                        +---------------------------------------------------------------+            
`
	fmt.Println(logo)
}
