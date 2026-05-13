package tool

import (
	"strings"
	"testing"
)

func TestBash_ExecutesCommand(t *testing.T) {
	root := t.TempDir()

	result, err := Bash(BashInput{Command: "echo hello"}, root, 30, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
	if strings.TrimSpace(result.Stdout) != "hello" {
		t.Errorf("expected stdout 'hello', got %q", result.Stdout)
	}
}

func TestBash_TimeoutError(t *testing.T) {
	root := t.TempDir()

	_, err := Bash(BashInput{Command: "sleep 10", Timeout: 1}, root, 30, nil)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected timeout error message, got: %v", err)
	}
}

func TestBash_BlockedCommand(t *testing.T) {
	root := t.TempDir()

	_, err := Bash(BashInput{Command: "sudo ls"}, root, 30, nil)
	if err == nil {
		t.Fatal("expected error for blocked command")
	}
	if !strings.Contains(err.Error(), "blocked pattern") {
		t.Errorf("expected blocked pattern error, got: %v", err)
	}
}

func TestBash_WorkspaceDir(t *testing.T) {
	root := t.TempDir()

	result, err := Bash(BashInput{Command: "pwd"}, root, 30, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(result.Stdout) != root {
		t.Errorf("expected pwd to be %q, got %q", root, strings.TrimSpace(result.Stdout))
	}
}

func TestBash_EmptyCommand(t *testing.T) {
	root := t.TempDir()

	_, err := Bash(BashInput{Command: ""}, root, 30, nil)
	if err == nil {
		t.Error("expected error for empty command")
	}
}

func TestBash_ExitCode(t *testing.T) {
	root := t.TempDir()

	result, err := Bash(BashInput{Command: "false"}, root, 30, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExitCode != 1 {
		t.Errorf("expected exit code 1 for 'false', got %d", result.ExitCode)
	}
}

func TestBash_BlockedFromConfig(t *testing.T) {
	root := t.TempDir()

	_, err := Bash(BashInput{Command: "docker ps"}, root, 30, []string{"docker"})
	if err == nil {
		t.Fatal("expected error for blocked command from config")
	}
	if !strings.Contains(err.Error(), "blocked pattern from configuration") {
		t.Errorf("expected config blocked pattern error, got: %v", err)
	}
}

func TestBash_StderrCapture(t *testing.T) {
	root := t.TempDir()

	result, err := Bash(BashInput{Command: "echo error_message >&2"}, root, 30, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(result.Stderr) != "error_message" {
		t.Errorf("expected stderr 'error_message', got %q", result.Stderr)
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
}
