package tool

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFile_ReadsAllowedFile(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "test.txt"), []byte("hello world"), 0644)

	result, err := ReadFile(ReadFileInput{Path: "test.txt"}, root, 1048576)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Content != "hello world" {
		t.Errorf("expected content 'hello world', got %q", result.Content)
	}
	if result.Size != 11 {
		t.Errorf("expected size 11, got %d", result.Size)
	}
	if result.Truncated {
		t.Error("expected not truncated")
	}
}

func TestReadFile_TruncatesLargeOutput(t *testing.T) {
	root := t.TempDir()
	content := strings.Repeat("A", 100)
	os.WriteFile(filepath.Join(root, "large.txt"), []byte(content), 0644)

	result, err := ReadFile(ReadFileInput{Path: "large.txt"}, root, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Content) != 50 {
		t.Errorf("expected 50 chars, got %d", len(result.Content))
	}
	if !result.Truncated {
		t.Error("expected truncated=true")
	}
}

func TestReadFile_RejectsBinaryFile(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "image.png"), []byte("fake png content"), 0644)

	_, err := ReadFile(ReadFileInput{Path: "image.png"}, root, 1048576)
	if err == nil {
		t.Error("expected error for binary file by extension")
	}
}

func TestReadFile_RejectsEnvFile(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, ".env"), []byte("SECRET=key"), 0644)

	_, err := ReadFile(ReadFileInput{Path: ".env"}, root, 1048576)
	if err == nil {
		t.Error("expected error for .env file")
	}
}

func TestReadFile_RejectsDirectory(t *testing.T) {
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, "mydir"), 0755)

	_, err := ReadFile(ReadFileInput{Path: "mydir"}, root, 1048576)
	if err == nil {
		t.Error("expected error for directory")
	}
}

func TestReadFile_RejectsMissingFile(t *testing.T) {
	root := t.TempDir()

	_, err := ReadFile(ReadFileInput{Path: "nonexistent.txt"}, root, 1048576)
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestReadFile_DetectsBinaryContent(t *testing.T) {
	root := t.TempDir()
	// Create a file with null bytes but a .txt extension
	binaryContent := []byte("normal\ncontent\x00with nulls\nmore text")
	os.WriteFile(filepath.Join(root, "binary.txt"), binaryContent, 0644)

	_, err := ReadFile(ReadFileInput{Path: "binary.txt"}, root, 1048576)
	if err == nil {
		t.Error("expected error for binary content with null bytes")
	}
}

func TestReadFile_RejectsDotGitFile(t *testing.T) {
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, ".git"), 0755)
	os.WriteFile(filepath.Join(root, ".git", "HEAD"), []byte("ref: main"), 0644)

	_, err := ReadFile(ReadFileInput{Path: ".git/HEAD"}, root, 1048576)
	if err == nil {
		t.Error("expected error for .git file")
	}
}
