# computron

[![Go](https://img.shields.io/badge/go-1.25-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-7C3AED)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macos-6B7280)](#install)

CLI for installing, managing, and monitoring [Computron](https://github.com/lefoulkrod/computron_9000) — a containerized AI assistant. Wraps Docker or Podman with a Bubble Tea TUI and guided install wizard.

---

![screenshot](computron-cli.png)
<!-- Replace with an actual recording once available. Suggested tool: vhs (https://github.com/charmbracelet/vhs) -->

---

## Features

- **Guided install wizard** — preflight checks, config form, confirm panel, spinner-driven install steps
- **Update in place** — pulls the latest image and recreates the container, preserving config
- **Multi-instance** — run more than one Computron container on different ports from a single binary
- **Health check** — `doctor` runs 10 parallel checks and prints a full pass/warn/fail report
- **Docker + Podman** — auto-detected; SELinux and host-networking differences handled per OS
- **Backup on uninstall** — optionally archives data directories to a `.tar.gz` before removal
- **`--no-color`** — safe to pipe; exits non-zero on actual failures, not on warnings

---

## Install

### Homebrew (macOS / Linux)

```sh
# tap not yet published — build from source in the meantime
```

### Build from source

Requires Go 1.21+ and Docker or Podman.

```sh
git clone https://github.com/rlnorthcutt/computron-cli
cd computron-cli
go build -ldflags="-s -w" -o computron .
sudo mv computron /usr/local/bin/
```

### Verify

```sh
computron --version
```

---

## Quickstart

```sh
# 1. Install Computron (interactive wizard)
computron install

# 2. Check everything is healthy
computron doctor

# 3. Open the web UI
#    http://localhost:8080
```

That's it. The wizard detects your container engine, checks Ollama reachability, walks you through the config, and starts the container.

---

## Usage

```
computron <command> [flags]
```

### Commands

| Command | Description |
|---|---|
| `install` | Interactive TUI wizard — install or update an existing instance |
| `update` | Pull the latest image and recreate the container |
| `start` | Start a stopped container |
| `stop` | Gracefully stop the running container |
| `restart` | Stop then start |
| `status` | Print a status table (container, dirs, Ollama, web UI port) |
| `logs` | Tail container logs |
| `doctor` | Run 10 parallel health checks and print a report |
| `config show` | Pretty-print the saved config |
| `config set <key> <value>` | Update a single config key |
| `config path` | Print the config file path |
| `uninstall` | Stop, remove container, optionally delete data dirs |

### Global flags

```
--config string   Config file path (default: ~/.config/computron-cli/config.yaml)
--no-color        Disable color output
--debug           Print raw engine commands and stderr
--version         Print version and exit
```

### Examples

```sh
# Tail logs and follow
computron logs --follow --tail 100

# Run on a non-default port (set during install or via config)
computron config set web_ui_port 8081

# Check status and exit 1 if not running (useful in scripts)
computron status

# Uninstall without interactive prompts
computron uninstall
```

### Multiple instances

If Computron is already installed, `computron install` asks whether to update the existing instance or install a new one. New instances get a unique container name (`computron2`, `computron3`, …), separate data directories, and an incremented port.

Commands that need an instance (e.g. `start`, `status`) show a picker when more than one instance exists.

### Config file

```
~/.config/computron-cli/instances/<container-name>.yaml
```

```yaml
container_name: computron
shared_dir: /home/user/Computron
state_dir: /home/user/Computron/state
shm_size: 256m
web_ui_port: "8080"
engine: docker
image: ghcr.io/lefoulkrod/computron_9000:container-distro-latest
installed_at: 2025-01-15T10:30:00Z
```

---

## Contributing

1. Fork and clone the repo
2. `go test ./...` — all tests must pass
3. `go vet ./...` — no vet errors
4. Open a PR against `main`

No Docker SDK — engine calls shell out intentionally to keep the binary dependency-free. Keep that constraint when adding engine features.

Platform rule: never add a Linux-only flag without a `runtime.GOOS` guard. The `:Z` SELinux volume flag and the Ollama host differ between Linux and macOS — see `CLAUDE.md` for the full platform table.

---

## License

MIT © [rlnorthcutt](https://github.com/rlnorthcutt)
