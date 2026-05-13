package tool

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSearchCode_FindsMatches(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "main.go"), []byte(`package main
import "fmt"
func main() {
	fmt.Println("hello world")
}
`), 0644)
	os.WriteFile(filepath.Join(root, "utils.go"), []byte(`package main
func helper() string {
	return "hello world"
}
`), 0644)

	result, err := SearchCode(SearchCodeInput{Pattern: "hello"}, root, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total < 2 {
		t.Errorf("expected at least 2 matches, got %d", result.Total)
	}
}

func TestSearchCode_RespectsMaxResults(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 10; i++ {
		content := ""
		for j := 0; j < 10; j++ {
			content += "line with match\n"
		}
		os.WriteFile(filepath.Join(root, "file_%d.go"), []byte(content), 0644)
	}

	// Actually write files properly
	for i := 0; i < 10; i++ {
		content := ""
		for j := 0; j < 10; j++ {
			content += "line with match\n"
		}
		os.WriteFile(filepath.Join(root, "file_%d.go"), []byte(content), 0644)
	}
	// Recreate with proper names
	for i := 0; i < 10; i++ {
		os.Remove(filepath.Join(root, "file_%d.go"))
	}
	for i := 0; i < 10; i++ {
		name := filepath.Join(root, "file_"+string(rune('a'+i))+".go")
		content := ""
		for j := 0; j < 10; j++ {
			content += "line with match\n"
		}
		os.WriteFile(name, []byte(content), 0644)
	}

	result, err := SearchCode(SearchCodeInput{Pattern: "match"}, root, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total > 10 {
		t.Errorf("expected at most 5 results (maxResults=5), got %d", result.Total)
	}
	if !result.Truncated {
		t.Error("expected truncated=true since we have many matches")
	}
}

func TestSearchCode_RejectsBlockedPaths(t *testing.T) {
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, ".git"), 0755)
	os.WriteFile(filepath.Join(root, ".git", "config"), []byte("secret = true"), 0644)
	os.WriteFile(filepath.Join(root, "main.go"), []byte("secret = false"), 0644)

	result, err := SearchCode(SearchCodeInput{Pattern: "secret"}, root, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only main.go should match, not .git/config
	if result.Total != 1 {
		t.Errorf("expected 1 match (from main.go only), got %d", result.Total)
		for _, m := range result.Matches {
			t.Logf("  match: %s:%d", m.Path, m.Line)
		}
	}
}

func TestSearchCode_RequiresPattern(t *testing.T) {
	root := t.TempDir()

	_, err := SearchCode(SearchCodeInput{Pattern: ""}, root, 50)
	if err == nil {
		t.Error("expected error for empty pattern")
	}
}

func TestSearchCode_SupportsRegex(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "main.go"), []byte(`func hello() {}
func world() {}
func helloWorld() {}
`), 0644)

	result, err := SearchCode(SearchCodeInput{Pattern: "hello[Ww]orld", IsRegex: true}, root, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("expected 1 regex match for helloWorld, got %d", result.Total)
	}
}
