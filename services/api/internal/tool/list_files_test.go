package tool

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListFiles_ListsAllowedDirectory(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(root, "README.md"), []byte("# Project"), 0644)
	os.Mkdir(filepath.Join(root, "src"), 0755)
	os.WriteFile(filepath.Join(root, "src", "app.go"), []byte("package app"), 0644)

	result, err := ListFiles(ListFilesInput{Path: ".", Depth: 1}, root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("expected 3 entries (main.go, README.md, src), got %d: %+v", result.Total, result.Entries)
	}
}

func TestListFiles_DoesNotIncludeBlockedPaths(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "main.go"), []byte("package main"), 0644)
	os.Mkdir(filepath.Join(root, ".git"), 0755)
	os.WriteFile(filepath.Join(root, ".git", "config"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(root, ".env"), []byte("SECRET=key"), 0644)
	os.Mkdir(filepath.Join(root, "node_modules"), 0755)
	os.WriteFile(filepath.Join(root, "node_modules", "pkg.js"), []byte("test"), 0644)

	result, err := ListFiles(ListFilesInput{Path: ".", Depth: 1}, root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only main.go should be visible
	if result.Total != 1 {
		t.Errorf("expected 1 entry (main.go only), got %d", result.Total)
		for _, e := range result.Entries {
			t.Logf("  entry: %s (dir=%v)", e.Name, e.IsDir)
		}
	}
}

func TestListFiles_HandlesMissingPath(t *testing.T) {
	root := t.TempDir()

	_, err := ListFiles(ListFilesInput{Path: "nonexistent"}, root)
	if err == nil {
		t.Error("expected error for missing path")
	}
}

func TestListFiles_RejectsFilePath(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "file.txt"), []byte("test"), 0644)

	_, err := ListFiles(ListFilesInput{Path: "file.txt"}, root)
	if err == nil {
		t.Error("expected error when path is a file, not directory")
	}
}

func TestListFiles_DepthRecursion(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "a", "b", "c"), 0755)
	os.WriteFile(filepath.Join(root, "a", "b", "deep.txt"), []byte("deep"), 0644)

	// Depth 1: only immediate children
	result, err := ListFiles(ListFilesInput{Path: ".", Depth: 1}, root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("depth 1: expected 1 entry (a), got %d", result.Total)
	}

	// Depth 3: should see a, a/b, a/b/c, a/b/deep.txt
	result, err = ListFiles(ListFilesInput{Path: ".", Depth: 3}, root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 4 {
		t.Errorf("depth 3: expected 4 entries, got %d: %+v", result.Total, result.Entries)
	}
}
