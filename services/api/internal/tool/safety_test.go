package tool

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePath_RejectsAbsolutePaths(t *testing.T) {
	root := "/tmp/test-workspace"
	_, err := ValidatePath("/etc/passwd", root)
	if err == nil {
		t.Error("expected error for absolute path")
	}
	if err != nil && err.Error() != "absolute paths are not allowed: /etc/passwd" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidatePath_RejectsPathTraversal(t *testing.T) {
	root := "/tmp/test-workspace"
	_, err := ValidatePath("../../../etc/passwd", root)
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestValidatePath_RejectsEnvFile(t *testing.T) {
	// Create temp workspace
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, ".env"), []byte("SECRET=key"), 0644)

	_, err := ValidatePath(".env", root)
	if err == nil {
		t.Error("expected error for .env file")
	}
}

func TestValidatePath_RejectsEnvLocalFile(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, ".env.local"), []byte("SECRET=key"), 0644)

	_, err := ValidatePath(".env.local", root)
	if err == nil {
		t.Error("expected error for .env.local file")
	}
}

func TestValidatePath_RejectsEnvPrefixFile(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, ".env.production"), []byte("SECRET=key"), 0644)

	_, err := ValidatePath(".env.production", root)
	if err == nil {
		t.Error("expected error for .env.* file")
	}
}

func TestValidatePath_RejectsDotGitDirectory(t *testing.T) {
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, ".git"), 0755)
	os.WriteFile(filepath.Join(root, ".git", "config"), []byte("test"), 0644)

	_, err := ValidatePath(".git/config", root)
	if err == nil {
		t.Error("expected error for .git directory")
	}
}

func TestValidatePath_RejectsNodeModules(t *testing.T) {
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, "node_modules"), 0755)
	os.WriteFile(filepath.Join(root, "node_modules", "lodash.js"), []byte("test"), 0644)

	_, err := ValidatePath("node_modules/lodash.js", root)
	if err == nil {
		t.Error("expected error for node_modules")
	}
}

func TestValidatePath_RejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outsideDir := t.TempDir()
	os.WriteFile(filepath.Join(outsideDir, "secret.txt"), []byte("secret"), 0644)

	// Create a symlink inside root that points outside
	os.Symlink(outsideDir, filepath.Join(root, "escape"))

	_, err := ValidatePath("escape/secret.txt", root)
	if err == nil {
		t.Error("expected error for symlink escaping root")
	}
}

func TestValidatePath_AllowsValidPath(t *testing.T) {
	root := t.TempDir()
	resolvedRoot, _ := filepath.EvalSymlinks(root)
	os.MkdirAll(filepath.Join(root, "src", "main"), 0755)
	os.WriteFile(filepath.Join(root, "src", "main", "app.go"), []byte("package main"), 0644)

	path, err := ValidatePath("src/main/app.go", root)
	if err != nil {
		t.Errorf("expected no error for valid path, got: %v", err)
	}
	expected := filepath.Join(resolvedRoot, "src", "main", "app.go")
	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestIsBinaryFile_DetectsByExtension(t *testing.T) {
	binaryFiles := []string{"image.png", "photo.jpg", "icon.svg", "font.woff2", "archive.zip", "binary.exe", "library.so", "secret.db", "data.sqlite"}
	textFiles := []string{"main.go", "README.md", "test.txt", "index.html", "style.css", "script.js", "config.yaml", "Dockerfile"}

	for _, f := range binaryFiles {
		if !IsBinaryFile(f) {
			t.Errorf("expected %q to be detected as binary", f)
		}
	}
	for _, f := range textFiles {
		if IsBinaryFile(f) {
			t.Errorf("expected %q to be detected as text", f)
		}
	}
}

func TestHasNullBytes_DetectsBinary(t *testing.T) {
	textContent := []byte("Hello, this is normal text content.\nWith multiple lines.\x09")
	binaryContent := []byte("Some text\x00with null bytes")

	if HasNullBytes(textContent, 512) {
		t.Error("expected no null bytes in text content")
	}
	if !HasNullBytes(binaryContent, 512) {
		t.Error("expected null bytes detected in binary content")
	}
}

func TestHasNullBytes_OnlyChecksSpecifiedBytes(t *testing.T) {
	content := []byte("AAAA\x00BBBB")
	if !HasNullBytes(content, 5) {
		t.Error("expected null byte detected within first 5 bytes")
	}
	if HasNullBytes(content, 4) {
		t.Error("expected no null byte within first 4 bytes")
	}
}

// --- ValidateWritePath tests ---

func TestValidateWritePath_RejectsAbsolutePaths(t *testing.T) {
	root := "/tmp/test-workspace"
	_, err := ValidateWritePath("/etc/passwd", root)
	if err == nil {
		t.Error("expected error for absolute path")
	}
	if err != nil && err.Error() != "absolute paths are not allowed: /etc/passwd" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidateWritePath_RejectsPathTraversal(t *testing.T) {
	root := "/tmp/test-workspace"
	_, err := ValidateWritePath("../../../etc/passwd", root)
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestValidateWritePath_RejectsDotEnvFile(t *testing.T) {
	root := t.TempDir()

	_, err := ValidateWritePath(".env", root)
	if err == nil {
		t.Error("expected error for .env file")
	}
}

func TestValidateWritePath_RejectsDotGitDirectory(t *testing.T) {
	root := t.TempDir()

	_, err := ValidateWritePath(".git/HEAD", root)
	if err == nil {
		t.Error("expected error for .git directory")
	}
}

func TestValidateWritePath_RejectsNodeModules(t *testing.T) {
	root := t.TempDir()

	_, err := ValidateWritePath("node_modules/foo.js", root)
	if err == nil {
		t.Error("expected error for node_modules")
	}
}

func TestValidateWritePath_AllowsValidNewPath(t *testing.T) {
	root := t.TempDir()

	path, err := ValidateWritePath("src/main/app.go", root)
	if err != nil {
		t.Errorf("expected no error for valid path, got: %v", err)
	}
	expected := filepath.Join(root, "src/main/app.go")
	// Resolve symlinks
	resolvedRoot, _ := filepath.EvalSymlinks(root)
	expected = filepath.Join(resolvedRoot, "src", "main", "app.go")
	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestValidateWritePath_RejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outsideDir := t.TempDir()

	// Create a symlink inside root that points outside
	os.Symlink(outsideDir, filepath.Join(root, "escape"))

	_, err := ValidateWritePath("escape/secret.txt", root)
	if err == nil {
		t.Error("expected error for symlink escaping root")
	}
}

// --- IsBashCommandAllowed tests ---

func TestIsBashCommandAllowed_AllowsSimpleCommand(t *testing.T) {
	err := IsBashCommandAllowed("echo hello")
	if err != nil {
		t.Errorf("expected no error for simple command, got: %v", err)
	}
}

func TestIsBashCommandAllowed_BlocksSudo(t *testing.T) {
	err := IsBashCommandAllowed("sudo rm -rf /")
	if err == nil {
		t.Error("expected error for sudo command")
	}
}

func TestIsBashCommandAllowed_BlocksShutdown(t *testing.T) {
	err := IsBashCommandAllowed("shutdown -h now")
	if err == nil {
		t.Error("expected error for shutdown")
	}
}

func TestIsBashCommandAllowed_BlocksDd(t *testing.T) {
	err := IsBashCommandAllowed("dd if=/dev/zero of=/dev/sda")
	if err == nil {
		t.Error("expected error for dd command")
	}
}

func TestIsBashCommandAllowed_BlocksEmptyCommand(t *testing.T) {
	err := IsBashCommandAllowed("")
	if err == nil {
		t.Error("expected error for empty command")
	}
}

// --- IsGitCommandAllowed tests ---

func TestIsGitCommandAllowed_AllowsStatus(t *testing.T) {
	err := IsGitCommandAllowed([]string{"status"})
	if err != nil {
		t.Errorf("expected no error for git status, got: %v", err)
	}
}

func TestIsGitCommandAllowed_BlocksPush(t *testing.T) {
	err := IsGitCommandAllowed([]string{"push"})
	if err == nil {
		t.Error("expected error for git push")
	}
}

func TestIsGitCommandAllowed_AllowsLog(t *testing.T) {
	err := IsGitCommandAllowed([]string{"log", "--oneline", "-5"})
	if err != nil {
		t.Errorf("expected no error for git log, got: %v", err)
	}
}

func TestIsGitCommandAllowed_BlocksReset(t *testing.T) {
	err := IsGitCommandAllowed([]string{"reset", "--hard"})
	if err == nil {
		t.Error("expected error for git reset")
	}
}

func TestIsGitCommandAllowed_BlocksRebase(t *testing.T) {
	err := IsGitCommandAllowed([]string{"rebase", "main"})
	if err == nil {
		t.Error("expected error for git rebase")
	}
}

func TestIsGitCommandAllowed_BlocksClean(t *testing.T) {
	err := IsGitCommandAllowed([]string{"clean", "-fd"})
	if err == nil {
		t.Error("expected error for git clean")
	}
}

func TestIsGitCommandAllowed_BlocksConfigWrite(t *testing.T) {
	err := IsGitCommandAllowed([]string{"config", "--add", "user.name", "test"})
	if err == nil {
		t.Error("expected error for git config write operation")
	}
}

func TestIsGitCommandAllowed_BlocksConfigSensitiveKey(t *testing.T) {
	err := IsGitCommandAllowed([]string{"config", "credential.helper", "store"})
	if err == nil {
		t.Error("expected error for sensitive git config key")
	}
}

func TestIsGitCommandAllowed_BlocksTagDelete(t *testing.T) {
	err := IsGitCommandAllowed([]string{"tag", "-d", "v1.0"})
	if err == nil {
		t.Error("expected error for git tag -d")
	}
}

func TestIsGitCommandAllowed_BlocksSubmoduleUpdate(t *testing.T) {
	err := IsGitCommandAllowed([]string{"submodule", "update", "--init"})
	if err == nil {
		t.Error("expected error for git submodule update")
	}
}

func TestIsGitCommandAllowed_NoArgs(t *testing.T) {
	err := IsGitCommandAllowed([]string{})
	if err == nil {
		t.Error("expected error for empty args")
	}
}

func TestIsGitCommandAllowed_AllowsAdd(t *testing.T) {
	err := IsGitCommandAllowed([]string{"add", "file.txt"})
	if err != nil {
		t.Errorf("expected no error for git add, got: %v", err)
	}
}

func TestIsGitCommandAllowed_AllowsBranch(t *testing.T) {
	err := IsGitCommandAllowed([]string{"branch", "feature"})
	if err != nil {
		t.Errorf("expected no error for git branch, got: %v", err)
	}
}

func TestIsGitCommandAllowed_BlocksFetch(t *testing.T) {
	err := IsGitCommandAllowed([]string{"fetch", "origin"})
	if err == nil {
		t.Error("expected error for git fetch")
	}
}

func TestIsGitCommandAllowed_BlocksGc(t *testing.T) {
	err := IsGitCommandAllowed([]string{"gc"})
	if err == nil {
		t.Error("expected error for git gc")
	}
}
