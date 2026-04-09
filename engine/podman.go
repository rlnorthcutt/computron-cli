package engine

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// PodmanEngine implements Engine using the podman CLI.
type PodmanEngine struct{}

func (e *PodmanEngine) Name() string { return "podman" }

func (e *PodmanEngine) IsAvailable() bool {
	_, err := lookPath("podman")
	return err == nil
}

// HasPermission always returns true for rootless Podman.
func (e *PodmanEngine) HasPermission() bool { return true }

func (e *PodmanEngine) ContainerExists(name string) (bool, error) {
	args := []string{"ps", "-a", "--filter", "name=^" + name + "$", "--format", "{{.Names}}"}
	cmd := buildCmd("podman", args...)
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("podman ps: %w", err)
	}
	return strings.TrimSpace(string(out)) == name, nil
}

func (e *PodmanEngine) PullImage(image string, msgs chan<- string) error {
	cmd := buildCmd("podman", "pull", image)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("podman pull: %w", err)
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if msgs != nil {
			msgs <- scanner.Text()
		}
	}
	return cmd.Wait()
}

func (e *PodmanEngine) RunContainer(opts RunOptions) error {
	// Podman uses --replace instead of pre-removing the container.
	args := buildPodmanRunArgs(opts)
	cmd := buildCmd("podman", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("podman run: %w", err)
	}
	return nil
}

func (e *PodmanEngine) StopContainer(name string) error {
	cmd := buildCmd("podman", "stop", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("podman stop: %w", err)
	}
	return nil
}

func (e *PodmanEngine) StartContainer(name string) error {
	cmd := buildCmd("podman", "start", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("podman start: %w", err)
	}
	return nil
}

func (e *PodmanEngine) RemoveContainer(name string) error {
	cmd := buildCmd("podman", "rm", "-f", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("podman rm: %w", err)
	}
	return nil
}

func (e *PodmanEngine) ContainerStatus(name string) (string, error) {
	cmd := buildCmd("podman", "inspect", "--format", "{{.State.Status}}", name)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("podman inspect: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (e *PodmanEngine) TailLogs(name string, follow bool, tail int) error {
	args := []string{"logs"}
	if follow {
		args = append(args, "--follow")
	}
	args = append(args, "--tail", fmt.Sprintf("%d", tail), name)
	cmd := buildCmd("podman", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Version returns the Podman version string.
func (e *PodmanEngine) Version() string {
	cmd := exec.Command("podman", "version", "--format", "{{.Version}}")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

// ImageDigest returns the repo digest of the given image.
func (e *PodmanEngine) ImageDigest(image string) string {
	cmd := exec.Command("podman", "inspect", "--format", "{{index .RepoDigests 0}}", image)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// buildPodmanRunArgs builds args for `podman run` — uses --replace instead of pre-remove.
func buildPodmanRunArgs(opts RunOptions) []string {
	network := opts.Network
	if network == "" {
		network = "host"
	}
	restart := opts.Restart
	if restart == "" {
		restart = "always"
	}

	args := []string{
		"run", "-d",
		"--replace",
		"--name", opts.Name,
		"--restart", restart,
		"--shm-size=" + opts.ShmSize,
		"--network", network,
	}

	sharedMount := opts.SharedDir + ":/home/computron"
	stateMount := opts.StateDir + ":/var/lib/computron"
	if opts.Platform == "linux" {
		sharedMount += ":Z"
		stateMount += ":Z"
	}
	args = append(args, "-v", sharedMount, "-v", stateMount)

	if opts.Platform == "darwin" {
		ollamaHost := opts.OllamaHost
		if ollamaHost == "" {
			ollamaHost = "http://host.docker.internal:11434"
		} else if !strings.HasPrefix(ollamaHost, "http") {
			ollamaHost = "http://" + ollamaHost
		}
		args = append(args, "-e", "OLLAMA_HOST="+ollamaHost)
	}

	// Always set PORT so the container binds to the configured web UI port.
	port := opts.WebUIPort
	if port == "" {
		port = "8080"
	}
	args = append(args, "-e", "PORT="+port)

	args = append(args, opts.Image)
	return args
}

// Ensure PodmanEngine implements Engine at compile time.
var _ Engine = (*PodmanEngine)(nil)
