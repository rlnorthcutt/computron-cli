package tui

import (
	"fmt"
	"testing"
)

func TestValidShmSize(t *testing.T) {
	valid := []string{"256m", "512M", "1g", "1G", "128k", "128K"}
	invalid := []string{"256", "256mb", "abc", "", "1.5g"}

	for _, s := range valid {
		if !validShmSize(s) {
			t.Errorf("expected %q to be valid", s)
		}
	}
	for _, s := range invalid {
		if validShmSize(s) {
			t.Errorf("expected %q to be invalid", s)
		}
	}
}

func TestNewInstallModelDefaults(t *testing.T) {
	m := NewInstallModel("/tmp/test-config.yaml", nil)
	if m.Phase != PhasePreflight {
		t.Errorf("initial phase should be PhasePreflight, got %d", m.Phase)
	}
	if m.inputs[inputContainerName].Value() != "computron" {
		t.Errorf("default container name should be 'computron'")
	}
	if m.inputs[inputShmSize].Value() != "256m" {
		t.Errorf("default shm size should be '256m'")
	}
}

func TestBuildConfig(t *testing.T) {
	m := NewInstallModel("/tmp/test-config.yaml", nil)
	// Set custom values.
	m.inputs[inputContainerName].SetValue("mycontainer")
	m.inputs[inputSharedDir].SetValue("/data/shared")
	m.inputs[inputStateDir].SetValue("/data/state")
	m.inputs[inputShmSize].SetValue("512m")

	cfg := m.buildConfig()
	if cfg.ContainerName != "mycontainer" {
		t.Errorf("ContainerName: got %q, want 'mycontainer'", cfg.ContainerName)
	}
	if cfg.SharedDir != "/data/shared" {
		t.Errorf("SharedDir: got %q", cfg.SharedDir)
	}
	if cfg.ShmSize != "512m" {
		t.Errorf("ShmSize: got %q", cfg.ShmSize)
	}
}

func TestValidateInputEmpty(t *testing.T) {
	m := NewInstallModel("/tmp/test.yaml", nil)
	m.inputs[inputContainerName].SetValue("")
	if m.validateCurrentInput() {
		t.Error("empty container name should fail validation")
	}
	if m.inputErrs[inputContainerName] == "" {
		t.Error("validation error should be set")
	}
}

func TestValidateInputShmSizeBad(t *testing.T) {
	m := NewInstallModel("/tmp/test.yaml", nil)
	m.inputFocus = inputShmSize
	m.inputs[inputShmSize].SetValue("256mb") // invalid
	if m.validateCurrentInput() {
		t.Error("'256mb' should fail shm size validation")
	}
}

func TestValidateInputShmSizeGood(t *testing.T) {
	m := NewInstallModel("/tmp/test.yaml", nil)
	m.inputFocus = inputShmSize
	m.inputs[inputShmSize].SetValue("256m")
	if !m.validateCurrentInput() {
		t.Error("'256m' should pass validation")
	}
}

func TestPreflightAdvancement(t *testing.T) {
	m := NewInstallModel("/tmp/test.yaml", nil)
	// Simulate all checks arriving.
	m.preflightDone = 3 // one before 4
	cmd := m.maybeAdvancePreflight()
	if cmd != nil {
		t.Error("should not advance until 4 checks complete")
	}
	m.preflightDone = 4
	cmd = m.maybeAdvancePreflight()
	if cmd == nil {
		t.Error("should advance after 4 checks complete")
	}
}

func TestUpdatePreflightEngineOK(t *testing.T) {
	m := NewInstallModel("/tmp/test.yaml", nil)
	// Simulate successful engine check (nil err means permission check fires).
	// We can't test the full async flow, but we can test the state update.
	msg := engineCheckResult{eng: nil, err: nil}
	m.preflightDone = 0
	// eng == nil, so no permission check is fired, preflightDone increments.
	newModel, _ := m.updatePreflight(msg)
	nm := newModel.(InstallModel)
	if nm.preflightDone != 1 {
		t.Errorf("preflightDone should be 1, got %d", nm.preflightDone)
	}
}

func TestUpdatePreflightEngineError(t *testing.T) {
	m := NewInstallModel("/tmp/test.yaml", nil)
	m.preflightDone = 4 // all other checks done
	msg := engineCheckResult{eng: nil, err: fmt.Errorf("no engine")}
	newModel, _ := m.updatePreflight(msg)
	nm := newModel.(InstallModel)
	// allPreflightDone will be fired next tick, but we test direct path:
	// Set preflightDone high and fire allPreflightDone directly.
	nm.engErr = msg.err
	nm2, _ := nm.updatePreflight(allPreflightDone{})
	final := nm2.(InstallModel)
	if final.Phase != PhaseError {
		t.Errorf("should enter error phase when engine missing, got %d", final.Phase)
	}
}
