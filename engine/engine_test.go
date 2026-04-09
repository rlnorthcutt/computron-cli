package engine

import (
	"os/exec"
	"strings"
	"testing"
)

// TestDetectNone verifies that Detect returns an error when neither docker nor
// podman is on PATH. We override lookPath to simulate a missing binary.
func TestDetectNone(t *testing.T) {
	orig := lookPath
	lookPath = func(file string) (string, error) {
		return "", exec.ErrNotFound
	}
	defer func() { lookPath = orig }()

	_, err := Detect()
	if err == nil {
		t.Fatal("expected error when no engine available")
	}
}

// TestDetectPodmanFirst verifies that Podman is preferred when both are available.
func TestDetectPodmanFirst(t *testing.T) {
	orig := lookPath
	lookPath = func(file string) (string, error) {
		// Both available.
		return "/usr/bin/" + file, nil
	}
	defer func() { lookPath = orig }()

	eng, err := Detect()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eng.Name() != "podman" {
		t.Errorf("expected podman, got %q", eng.Name())
	}
}

// TestDetectDockerFallback verifies Docker is returned when Podman is absent.
func TestDetectDockerFallback(t *testing.T) {
	orig := lookPath
	lookPath = func(file string) (string, error) {
		if file == "podman" {
			return "", exec.ErrNotFound
		}
		return "/usr/bin/docker", nil
	}
	defer func() { lookPath = orig }()

	eng, err := Detect()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eng.Name() != "docker" {
		t.Errorf("expected docker, got %q", eng.Name())
	}
}

// TestBuildRunArgsLinux verifies correct docker run args on Linux (with :Z mounts).
func TestBuildRunArgsLinux(t *testing.T) {
	opts := RunOptions{
		Name:      "computron",
		Image:     "ghcr.io/example/img:latest",
		ShmSize:   "256m",
		SharedDir: "/home/user/Computron",
		StateDir:  "/home/user/Computron/state",
		Network:   "host",
		Restart:   "always",
		Platform:  "linux",
	}

	args := buildRunArgs("docker", opts)

	assertContains(t, args, "--shm-size=256m")
	assertContains(t, args, "--network")
	assertContains(t, args, "host")
	assertContains(t, args, "/home/user/Computron:/home/computron:Z")
	assertContains(t, args, "/home/user/Computron/state:/var/lib/computron:Z")
	assertNotContains(t, args, "OLLAMA_HOST")
}

// TestBuildRunArgsMacOS verifies macOS-specific behaviour (no :Z, OLLAMA_HOST env).
func TestBuildRunArgsMacOS(t *testing.T) {
	opts := RunOptions{
		Name:       "computron",
		Image:      "ghcr.io/example/img:latest",
		ShmSize:    "256m",
		SharedDir:  "/Users/user/Computron",
		StateDir:   "/Users/user/Computron/state",
		OllamaHost: "host.docker.internal:11434",
		Platform:   "darwin",
	}

	args := buildRunArgs("docker", opts)

	assertNotContains(t, args, "/Users/user/Computron:/home/computron:Z")
	assertContains(t, args, "/Users/user/Computron:/home/computron")
	assertContainsPrefix(t, args, "OLLAMA_HOST=http://host.docker.internal:11434")
}

// TestBuildPodmanRunArgsHasReplace verifies --replace is present for Podman.
func TestBuildPodmanRunArgsHasReplace(t *testing.T) {
	opts := RunOptions{
		Name:      "computron",
		Image:     "ghcr.io/example/img:latest",
		ShmSize:   "256m",
		SharedDir: "/home/user/Computron",
		StateDir:  "/home/user/Computron/state",
		Platform:  "linux",
	}

	args := buildPodmanRunArgs(opts)
	assertContains(t, args, "--replace")
}

// TestOllamaHostLinux checks the Linux Ollama host value.
func TestOllamaHostLinux(t *testing.T) {
	// We can only reliably test the current platform, but we verify the function
	// returns a non-empty host:port string.
	h := OllamaHost()
	if h == "" {
		t.Fatal("OllamaHost should not be empty")
	}
	// Must contain a port.
	if !strings.Contains(h, ":") {
		t.Errorf("OllamaHost should be host:port, got %q", h)
	}
}

// --- helpers ---

func assertContains(t *testing.T, slice []string, want string) {
	t.Helper()
	for _, s := range slice {
		if s == want {
			return
		}
	}
	t.Errorf("args %v: expected to contain %q", slice, want)
}

func assertNotContains(t *testing.T, slice []string, unwanted string) {
	t.Helper()
	for _, s := range slice {
		if s == unwanted {
			t.Errorf("args %v: should not contain %q", slice, unwanted)
			return
		}
	}
}

func assertContainsPrefix(t *testing.T, slice []string, prefix string) {
	t.Helper()
	for _, s := range slice {
		if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
			return
		}
	}
	t.Errorf("args %v: expected an element with prefix %q", slice, prefix)
}

