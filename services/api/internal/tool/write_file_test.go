package tool

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFile_CreatesNewFile(t *testing.T) {
	root := t.TempDir()

	result, err := WriteFile(WriteFileInput{Path: "newfile.txt", Content: "hello world"}, root, 1048576)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Path != "newfile.txt" {
		t.Errorf("expected path 'newfile.txt', got %q", result.Path)
	}
	if result.Size != 11 {
		t.Errorf("expected size 11, got %d", result.Size)
	}
	if result.Type != "text" {
		t.Errorf("expected type 'text', got %q", result.Type)
	}

	// Verify file was actually written
	data, err := os.ReadFile(filepath.Join(root, "newfile.txt"))
	if err != nil {
		t.Fatalf("cannot read written file: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("expected content 'hello world', got %q", string(data))
	}
}

func TestWriteFile_OverwritesExistingFile(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "existing.txt"), []byte("original content"), 0644)

	result, err := WriteFile(WriteFileInput{Path: "existing.txt", Content: "overwritten content"}, root, 1048576)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Size != 19 {
		t.Errorf("expected size 19, got %d", result.Size)
	}

	// Verify content was overwritten
	data, err := os.ReadFile(filepath.Join(root, "existing.txt"))
	if err != nil {
		t.Fatalf("cannot read written file: %v", err)
	}
	if string(data) != "overwritten content" {
		t.Errorf("expected 'overwritten content', got %q", string(data))
	}
}

func TestWriteFile_CreatesParentDirectories(t *testing.T) {
	root := t.TempDir()

	result, err := WriteFile(WriteFileInput{Path: "a/b/c/nested.txt", Content: "nested file"}, root, 1048576)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Path != "a/b/c/nested.txt" {
		t.Errorf("expected path 'a/b/c/nested.txt', got %q", result.Path)
	}

	// Verify file was actually written
	data, err := os.ReadFile(filepath.Join(root, "a", "b", "c", "nested.txt"))
	if err != nil {
		t.Fatalf("cannot read written file: %v", err)
	}
	if string(data) != "nested file" {
		t.Errorf("expected 'nested file', got %q", string(data))
	}
}

func TestWriteFile_RejectsBinaryExtension(t *testing.T) {
	root := t.TempDir()

	_, err := WriteFile(WriteFileInput{Path: "image.png", Content: "fake png"}, root, 1048576)
	if err == nil {
		t.Error("expected error for binary file extension")
	}
}

func TestWriteFile_RejectsDotEnvFile(t *testing.T) {
	root := t.TempDir()

	_, err := WriteFile(WriteFileInput{Path: ".env", Content: "SECRET=key"}, root, 1048576)
	if err == nil {
		t.Error("expected error for .env file")
	}
}

func TestWriteFile_RejectsPathTraversal(t *testing.T) {
	root := t.TempDir()

	_, err := WriteFile(WriteFileInput{Path: "../../outside.txt", Content: "outside"}, root, 1048576)
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestWriteFile_RejectsNullBytes(t *testing.T) {
	root := t.TempDir()

	content := "normal\x00with null"
	_, err := WriteFile(WriteFileInput{Path: "file.txt", Content: content}, root, 1048576)
	if err == nil {
		t.Error("expected error for null bytes in content")
	}
}

func TestWriteFile_RejectsContentTooLarge(t *testing.T) {
	root := t.TempDir()

	content := strings.Repeat("A", 100)
	_, err := WriteFile(WriteFileInput{Path: "large.txt", Content: content}, root, 50)
	if err == nil {
		t.Error("expected error for content exceeding maxBytes")
	}
}

func TestWriteFile_RejectsEmptyPath(t *testing.T) {
	root := t.TempDir()

	_, err := WriteFile(WriteFileInput{Path: "", Content: "test"}, root, 1048576)
	if err == nil {
		t.Error("expected error for empty path")
	}
}

func TestWriteFile_BlockedDirDist(t *testing.T) {
	root := t.TempDir()

	_, err := WriteFile(WriteFileInput{Path: "dist/output.txt", Content: "test"}, root, 1048576)
	if err == nil {
		t.Error("expected error for dist directory")
	}
}

func TestWriteFile_BlockedDirBuild(t *testing.T) {
	root := t.TempDir()

	_, err := WriteFile(WriteFileInput{Path: "build/output.txt", Content: "test"}, root, 1048576)
	if err == nil {
		t.Error("expected error for build directory")
	}
}

func TestWriteFile_BlockedDirVendor(t *testing.T) {
	root := t.TempDir()

	_, err := WriteFile(WriteFileInput{Path: "vendor/lib/file.txt", Content: "test"}, root, 1048576)
	if err == nil {
		t.Error("expected error for vendor directory")
	}
}
