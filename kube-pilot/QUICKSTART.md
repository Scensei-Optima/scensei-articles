# Quickstart

This guide spins up a local Kubernetes cluster with [kind](https://kind.sigs.k8s.io), deploys a couple of pods, and uses KubePilot to stream logs from one of them.

## Prerequisites

- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- kubepilot — see [installation](README.md#installation)

## 1. Create a cluster

```bash
kind create cluster --name demo
```

Wait for the control plane to become ready:

```bash
kubectl wait --for=condition=Ready node --all --timeout=60s
```

## 2. Deploy some pods

Start two pods that emit a log line every two seconds:

```bash
kubectl run app-one --image=busybox --restart=Never \
  -- sh -c 'while true; do echo "[app-one] $(date)"; sleep 2; done'

kubectl run app-two --image=busybox --restart=Never \
  -- sh -c 'while true; do echo "[app-two] $(date)"; sleep 2; done'
```

Confirm they are running:

```bash
kubectl get pods
```

```
NAME      READY   STATUS    RESTARTS   AGE
app-one   1/1     Running   0          10s
app-two   1/1     Running   0          8s
```

## 3. Stream logs with KubePilot

```bash
kubepilot -logs
```

KubePilot lists the running pods and lets you pick one with the arrow keys:

```
? Select a Pod to Stream Logs from:
  ▸ app-one
    app-two
```

Select `app-one` and hit Enter. You should see live output:

```
[app-one] Mon Jan  6 12:00:00 UTC 2025
[app-one] Mon Jan  6 12:00:02 UTC 2025
[app-one] Mon Jan  6 12:00:04 UTC 2025
```

Press `Ctrl+C` to stop.

## 4. Clean up

```bash
kind delete cluster --name demo
```
