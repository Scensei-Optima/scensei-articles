# KubePilot

An interactive CLI for Kubernetes. Select a pod from a live list and port-forward it, open a shell, or tail its logs — all without writing a single `kubectl` command.

## How does it compare to k9s?

[k9s](https://k9scli.io) is excellent — if you work with Kubernetes daily it is hard to beat. KubePilot is not trying to replace it. They solve different problems.

| | k9s | KubePilot |
|---|---|---|
| **Scope** | Full cluster management dashboard | Port-forward, shell, and logs only |
| **Workflow** | Keep the terminal window open, navigate the UI | Launch, do one thing, exit |
| **Learning curve** | Keyboard shortcuts, views, filters, plugins | Arrow keys to pick a pod, one flag per action |
| **Best for** | Deep cluster inspection and day-to-day ops | Quick access from a script, CI, or a fresh machine |

If you already have k9s open, use it. KubePilot is for the moments when you just need to forward a port or tail some logs without context-switching into a full dashboard — or when you are on a machine where installing k9s feels like overkill.

## Installation

### Linux

```bash
VERSION=v1.0.0
wget -qO- https://github.com/scensei-articles/kubepilot/releases/download/${VERSION}/kubepilot-linux-${VERSION}.tar.gz \
  | tar -xz
sudo mv kubepilot /usr/local/bin/
```

### macOS

```bash
VERSION=v1.0.0
wget -qO- https://github.com/scensei-articles/kubepilot/releases/download/${VERSION}/kubepilot-macos-${VERSION}.tar.gz \
  | tar -xz
sudo mv kubepilot /usr/local/bin/
```

### Windows

```powershell
$VERSION = "v1.0.0"
Invoke-WebRequest `
  -Uri "https://github.com/scensei-articles/kubepilot/releases/download/$VERSION/kubepilot-windows-$VERSION.zip" `
  -OutFile kubepilot.zip
Expand-Archive kubepilot.zip -DestinationPath .
```

Move `kubepilot.exe` to a directory on your `PATH`.

### Build from source

Requires Go 1.21+.

```bash
git clone https://github.com/scensei-articles/kubepilot
cd kubepilot/kube-pilot
go build -o kubepilot .
```

## Requirements

- A valid kubeconfig at `~/.kube/config`
- Network access to the Kubernetes API server

## Usage

```
kubepilot [flags]
```

Run without flags to list pods in the `default` namespace and start port-forwarding the selected one.

### Flags

| Flag | Default | Description |
|---|---|---|
| `-n <namespace>` | `default` | Namespace to search |
| `-l <selector>` | | Label selector, e.g. `app=web` |
| `-local <port>` | `8080` | Local port to bind |
| `-remote <port>` | | Override the remote port (skips auto-detection from pod spec) |
| `-attach` | | Open an interactive shell in the selected pod |
| `-logs` | | Stream logs from the selected pod |

`-attach` and `-logs` are mutually exclusive.

### Port forwarding

```bash
# Default namespace, port 8080
kubepilot

# Specific namespace and local port
kubepilot -n production -local 3000

# Filter by label
kubepilot -n production -l app=api
```

### Interactive shell

```bash
kubepilot -n staging -attach
```

Drops into `/bin/sh` inside the selected pod. Press `Ctrl+C` or type `exit` to detach.

### Log streaming

```bash
kubepilot -n production -logs
kubepilot -n production -l app=worker -logs
```

Shows the last 100 lines then tails live. Press `Ctrl+C` to stop.

## Development

```bash
git clone https://github.com/scensei-articles/kubepilot
cd kubepilot/kube-pilot
go test ./...
```

### Pre-commit hooks

Install [pre-commit](https://pre-commit.com), then from the repo root:

```bash
pre-commit install
```

Runs `go fmt` and `go vet` on every commit.
