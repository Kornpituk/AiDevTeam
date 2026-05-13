package tool

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGit_InitRepo(t *testing.T) {
	root := t.TempDir()

	result, err := Git(GitInput{Args: []string{"init"}}, root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", result.ExitCode, result.Stderr)
	}

	// Verify .git directory was created
	if _, err := os.Stat(filepath.Join(root, ".git")); os.IsNotExist(err) {
		t.Error(".git directory was not created")
	}
}

func TestGit_Status(t *testing.T) {
	root := t.TempDir()
	initGitRepo(t, root)

	result, err := Git(GitInput{Args: []string{"status"}}, root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", result.ExitCode, result.Stderr)
	}
}

func TestGit_PushBlocked(t *testing.T) {
	root := t.TempDir()

	_, err := Git(GitInput{Args: []string{"push"}}, root)
	if err == nil {
		t.Fatal("expected error for git push")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Errorf("expected 'not allowed' error, got: %v", err)
	}
}

func TestGit_ResetBlocked(t *testing.T) {
	root := t.TempDir()

	_, err := Git(GitInput{Args: []string{"reset", "--hard"}}, root)
	if err == nil {
		t.Fatal("expected error for git reset")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Errorf("expected 'not allowed' error, got: %v", err)
	}
}

func TestGit_CleanBlocked(t *testing.T) {
	root := t.TempDir()

	_, err := Git(GitInput{Args: []string{"clean", "-fd"}}, root)
	if err == nil {
		t.Fatal("expected error for git clean")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Errorf("expected 'not allowed' error, got: %v", err)
	}
}

func TestGit_NoArgs(t *testing.T) {
	root := t.TempDir()

	_, err := Git(GitInput{Args: []string{}}, root)
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestGit_RebaseBlocked(t *testing.T) {
	root := t.TempDir()

	_, err := Git(GitInput{Args: []string{"rebase", "main"}}, root)
	if err == nil {
		t.Fatal("expected error for git rebase")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Errorf("expected 'not allowed' error, got: %v", err)
	}
}

func TestGit_InitAndAdd(t *testing.T) {
	root := t.TempDir()

	// Initialize repo
	result, err := Git(GitInput{Args: []string{"init"}}, root)
	if err != nil {
		t.Fatalf("git init failed: %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("git init exit code %d, stderr: %s", result.ExitCode, result.Stderr)
	}

	// Create a file
	if err := os.WriteFile(filepath.Join(root, "test.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("cannot create test file: %v", err)
	}

	// Git add
	result, err = Git(GitInput{Args: []string{"add", "test.txt"}}, root)
	if err != nil {
		t.Fatalf("git add failed: %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("git add exit code %d, stderr: %s", result.ExitCode, result.Stderr)
	}

	// Git status should show staged file
	result, err = Git(GitInput{Args: []string{"status", "--short"}}, root)
	if err != nil {
		t.Fatalf("git status failed: %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("git status exit code %d, stderr: %s", result.ExitCode, result.Stderr)
	}
	if !strings.Contains(result.Stdout, "test.txt") {
		t.Errorf("expected test.txt in status output, got: %s", result.Stdout)
	}
}

func TestGit_PullBlocked(t *testing.T) {
	root := t.TempDir()

	_, err := Git(GitInput{Args: []string{"pull"}}, root)
	if err == nil {
		t.Fatal("expected error for git pull")
	}
}

func TestGit_MergeBlocked(t *testing.T) {
	root := t.TempDir()

	_, err := Git(GitInput{Args: []string{"merge", "main"}}, root)
	if err == nil {
		t.Fatal("expected error for git merge")
	}
}

func TestGit_AddAndStatus(t *testing.T) {
	root := t.TempDir()
	initGitRepo(t, root)

	// Create a file and add it
	os.WriteFile(filepath.Join(root, "new.txt"), []byte("content"), 0644)

	result, err := Git(GitInput{Args: []string{"add", "new.txt"}}, root)
	if err != nil {
		t.Fatalf("git add failed: %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("git add exit code %d, stderr: %s", result.ExitCode, result.Stderr)
	}

	// Check staged status
	result, err = Git(GitInput{Args: []string{"status", "--short"}}, root)
	if err != nil {
		t.Fatalf("git status failed: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", result.ExitCode, result.Stderr)
	}
	if !strings.Contains(result.Stdout, "A") || !strings.Contains(result.Stdout, "new.txt") {
		t.Errorf("expected staged new.txt in status, got: %s", result.Stdout)
	}
}

// initGitRepo initializes a git repository in the given directory.
func initGitRepo(t *testing.T, root string) {
	t.Helper()
	_, err := Git(GitInput{Args: []string{"init"}}, root)
	if err != nil {
		t.Fatalf("cannot init git repo: %v", err)
	}
}
