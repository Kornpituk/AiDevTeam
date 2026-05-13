package tool

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEditFile_ReplacesContent(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "test.txt"), []byte("hello world"), 0644)

	result, err := EditFile(EditFileInput{
		Path:      "test.txt",
		OldString: "world",
		NewString: "there",
	}, root, 1048576)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ReplacedCount != 1 {
		t.Errorf("expected 1 replacement, got %d", result.ReplacedCount)
	}
	if result.Path != "test.txt" {
		t.Errorf("expected path 'test.txt', got %q", result.Path)
	}

	// Verify content
	data, err := os.ReadFile(filepath.Join(root, "test.txt"))
	if err != nil {
		t.Fatalf("cannot read file: %v", err)
	}
	if string(data) != "hello there" {
		t.Errorf("expected 'hello there', got %q", string(data))
	}
}

func TestEditFile_ReplacesMultipleOccurrences(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "test.txt"), []byte("foo foo foo bar"), 0644)

	result, err := EditFile(EditFileInput{
		Path:      "test.txt",
		OldString: "foo",
		NewString: "baz",
	}, root, 1048576)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ReplacedCount != 3 {
		t.Errorf("expected 3 replacements, got %d", result.ReplacedCount)
	}

	data, err := os.ReadFile(filepath.Join(root, "test.txt"))
	if err != nil {
		t.Fatalf("cannot read file: %v", err)
	}
	if string(data) != "baz baz baz bar" {
		t.Errorf("expected 'baz baz baz bar', got %q", string(data))
	}
}

func TestEditFile_NoMatchReturnsError(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "test.txt"), []byte("hello world"), 0644)

	_, err := EditFile(EditFileInput{
		Path:      "test.txt",
		OldString: "nonexistent",
		NewString: "replacement",
	}, root, 1048576)
	if err == nil {
		t.Error("expected error when old_string not found")
	}
}

func TestEditFile_RejectsBinaryExtension(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "image.png"), []byte("fake png"), 0644)

	_, err := EditFile(EditFileInput{
		Path:      "image.png",
		OldString: "fake",
		NewString: "real",
	}, root, 1048576)
	if err == nil {
		t.Error("expected error for binary file extension")
	}
}

func TestEditFile_RejectsPathTraversal(t *testing.T) {
	root := t.TempDir()

	_, err := EditFile(EditFileInput{
		Path:      "../../outside.txt",
		OldString: "test",
		NewString: "replaced",
	}, root, 1048576)
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestEditFile_RejectsNonExistentFile(t *testing.T) {
	root := t.TempDir()

	_, err := EditFile(EditFileInput{
		Path:      "nonexistent.txt",
		OldString: "test",
		NewString: "replaced",
	}, root, 1048576)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestEditFile_RejectsEmptyPath(t *testing.T) {
	root := t.TempDir()

	_, err := EditFile(EditFileInput{
		Path:      "",
		OldString: "test",
		NewString: "replaced",
	}, root, 1048576)
	if err == nil {
		t.Error("expected error for empty path")
	}
}

func TestEditFile_RejectsEmptyOldString(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "test.txt"), []byte("hello world"), 0644)

	_, err := EditFile(EditFileInput{
		Path:      "test.txt",
		OldString: "",
		NewString: "replacement",
	}, root, 1048576)
	if err == nil {
		t.Error("expected error for empty old_string")
	}
}

func TestEditFile_RejectsDotEnvFile(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, ".env"), []byte("SECRET=key"), 0644)

	_, err := EditFile(EditFileInput{
		Path:      ".env",
		OldString: "SECRET",
		NewString: "NOT_SECRET",
	}, root, 1048576)
	if err == nil {
		t.Error("expected error for .env file")
	}
}
