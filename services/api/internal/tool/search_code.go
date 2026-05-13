package tool

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SearchCodeInput represents the input for the search_code tool.
type SearchCodeInput struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path,omitempty"` // optional: restrict search to a subdirectory
	IsRegex bool   `json:"is_regex,omitempty"`
}

// Match represents a single search match.
type Match struct {
	Path     string `json:"path"`
	Line     int    `json:"line"`
	Snippet  string `json:"snippet"`
}

// SearchCodeOutput represents the output from the search_code tool.
type SearchCodeOutput struct {
	Pattern string  `json:"pattern"`
	Matches []Match `json:"matches"`
	Total   int     `json:"total"`
	Truncated bool  `json:"truncated,omitempty"`
}

// SearchCode searches for a pattern in files within the workspace.
func SearchCode(input SearchCodeInput, workspaceRoot string, maxResults int) (*SearchCodeOutput, error) {
	if input.Pattern == "" {
		return nil, fmt.Errorf("pattern is required")
	}

	searchRoot := workspaceRoot
	if input.Path != "" {
		var err error
		searchRoot, err = ValidatePath(input.Path, workspaceRoot)
		if err != nil {
			return nil, err
		}
		info, err := os.Stat(searchRoot)
		if err != nil {
			return nil, fmt.Errorf("cannot access search path: %s", input.Path)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("search path is not a directory: %s", input.Path)
		}
	}

	var (
		re      *regexp.Regexp
		err     error
	)

	if input.IsRegex {
		re, err = regexp.Compile(input.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %s", err.Error())
		}
	} else {
		// Escape literal string for regex search (case-insensitive)
		pattern := regexp.QuoteMeta(input.Pattern)
		re, err = regexp.Compile("(?i)" + pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid search pattern: %s", err.Error())
		}
	}

	var matches []Match

	err = filepath.WalkDir(searchRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip permission errors
		}

		if d.IsDir() {
			// Skip blocked directories
			if isBlockedDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip blocked files
		if isBlockedFile(d.Name()) {
			return nil
		}

		// Skip binary files by extension
		if IsBinaryFile(path) {
			return nil
		}

		// Check max results reached
		if len(matches) >= maxResults {
			return filepath.SkipAll
		}

		// Read file and search
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable files
		}

		// Skip binary content
		if HasNullBytes(content, 512) {
			return nil
		}

		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if re.MatchString(line) {
				relPath, _ := filepath.Rel(workspaceRoot, path)
				snippet := strings.TrimSpace(line)
				if len(snippet) > 200 {
					snippet = snippet[:200] + "..."
				}
				matches = append(matches, Match{
					Path:    relPath,
					Line:    i + 1,
					Snippet: snippet,
				})
				if len(matches) >= maxResults {
					return filepath.SkipAll
				}
			}
		}

		return nil
	})
	if err != nil {
		// SkipAll error means we hit max results, which is fine
		if err != filepath.SkipAll {
			return nil, fmt.Errorf("error searching code: %s", err.Error())
		}
	}

	truncated := len(matches) >= maxResults

	return &SearchCodeOutput{
		Pattern:   input.Pattern,
		Matches:   matches,
		Total:     len(matches),
		Truncated: truncated,
	}, nil
}
