package tool

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFileInput represents the input for the write_file tool.
type WriteFileInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// WriteFileOutput represents the output from the write_file tool.
type WriteFileOutput struct {
	Path string `json:"path"`
	Size int    `json:"size"`
	Type string `json:"type"` // always "text"
}

// WriteFile creates or overwrites a file within the workspace.
func WriteFile(input WriteFileInput, workspaceRoot string, maxBytes int64) (*WriteFileOutput, error) {
	if input.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	safePath, err := ValidateWritePath(input.Path, workspaceRoot)
	if err != nil {
		return nil, err
	}

	// Reject binary files by extension
	if IsBinaryFile(input.Path) {
		return nil, fmt.Errorf("cannot write binary file: %s", input.Path)
	}

	content := input.Content

	// Reject content with null bytes
	if HasNullBytes([]byte(content), len([]byte(content))) {
		return nil, fmt.Errorf("content contains null bytes (binary content): %s", input.Path)
	}

	// Check content length against maxBytes
	contentSize := int64(len(content))
	if maxBytes > 0 && contentSize > maxBytes {
		return nil, fmt.Errorf("content exceeds maximum size of %d bytes: %d", maxBytes, contentSize)
	}

	// Create parent directories if they don't exist
	parentDir := filepath.Dir(safePath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create parent directories: %s", err.Error())
	}

	// Write the file
	if err := os.WriteFile(safePath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("cannot write file: %s", err.Error())
	}

	return &WriteFileOutput{
		Path: input.Path,
		Size: len(content),
		Type: contentTypeText,
	}, nil
}
