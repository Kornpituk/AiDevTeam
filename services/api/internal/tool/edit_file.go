package tool

import (
	"fmt"
	"os"
	"strings"
)

// EditFileInput represents the input for the edit_file tool.
type EditFileInput struct {
	Path      string `json:"path"`
	OldString string `json:"old_string"`
	NewString string `json:"new_string"`
}

// EditFileOutput represents the output from the edit_file tool.
type EditFileOutput struct {
	Path          string `json:"path"`
	ReplacedCount int    `json:"replaced_count"`
	Type          string `json:"type"` // "text"
}

// EditFile performs a find-and-replace operation on a file within the workspace.
func EditFile(input EditFileInput, workspaceRoot string, maxBytes int64) (*EditFileOutput, error) {
	if input.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	safePath, err := ValidateWritePath(input.Path, workspaceRoot)
	if err != nil {
		return nil, err
	}

	// Check file exists and is not a directory
	info, err := os.Stat(safePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file does not exist: %s", input.Path)
		}
		return nil, fmt.Errorf("cannot access file: %s", input.Path)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file: %s", input.Path)
	}

	// Reject binary files by extension
	if IsBinaryFile(input.Path) {
		return nil, fmt.Errorf("cannot edit binary file: %s", input.Path)
	}

	// Read the file content
	data, err := os.ReadFile(safePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %s", input.Path)
	}

	// Check for null bytes
	if HasNullBytes(data, len(data)) {
		return nil, fmt.Errorf("file appears to be binary: %s", input.Path)
	}

	// Check size limits
	if maxBytes > 0 && int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("file exceeds maximum size of %d bytes: %d", maxBytes, len(data))
	}

	// Validate that OldString is not empty
	if input.OldString == "" {
		return nil, fmt.Errorf("old_string is required")
	}

	// Perform the replacement
	oldContent := string(data)
	if !strings.Contains(oldContent, input.OldString) {
		return nil, fmt.Errorf("no occurrences of old_string found in file: %s", input.Path)
	}

	count := strings.Count(oldContent, input.OldString)
	newContent := strings.ReplaceAll(oldContent, input.OldString, input.NewString)

	// Write back the modified content
	if err := os.WriteFile(safePath, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("cannot write file: %s", err.Error())
	}

	return &EditFileOutput{
		Path:          input.Path,
		ReplacedCount: count,
		Type:          contentTypeText,
	}, nil
}
