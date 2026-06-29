# Scensei Articles

Code and materials accompanying the [Scensei engineering blog](https://www.scensei.com/blog).

## KubePilot

> A lightweight Kubernetes CLI for on-prem teams. Select a pod, port-forward it, attach a shell, or tail its logs — no dashboard required.

KubePilot is a standalone tool built and maintained by [Scensei](https://www.scensei.com). It ships as a single binary and is designed for fast incident response in environments where your workloads run in the `scensei` namespace.

**Quick install (Linux)**
```bash
VERSION=0.5.0
wget -qO- https://github.com/Scensei-Optima/scensei-articles/releases/download/${VERSION}/kubepilot-linux-${VERSION}.tar.gz | tar -xz
sudo mv kubepilot /usr/local/bin/
```

→ [Full documentation and installation for all platforms](kube-pilot/README.md)  
→ [Quickstart: spin up a local cluster and try it in 5 minutes](kube-pilot/QUICKSTART.md)

---

## Articles

| Article | Code |
|---|---|
| [Reference architecture for automated EKS deployment](https://www.scensei.com/post/reference-architecture-for-automated-eks-deployment) | [eks/](eks/) |
| KubePilot — a minimal Kubernetes CLI for on-prem teams *(coming soon)* | [kube-pilot/](kube-pilot/) |

---
