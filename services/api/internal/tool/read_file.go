package tool

import (
	"fmt"
	"os"
)

const contentTypeBinary = "binary"
const contentTypeText = "text"

// ReadFileInput represents the input for the read_file tool.
type ReadFileInput struct {
	Path string `json:"path"`
}

// ReadFileOutput represents the output from the read_file tool.
type ReadFileOutput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Truncated bool `json:"truncated,omitempty"`
	Size    int    `json:"size"`
	Type    string `json:"type"`
}

// ReadFile reads a file from the workspace, with safety checks and size limits.
func ReadFile(input ReadFileInput, workspaceRoot string, maxBytes int64) (*ReadFileOutput, error) {
	safePath, err := ValidatePath(input.Path, workspaceRoot)
	if err != nil {
		return nil, err
	}

	// Check if file is binary by extension
	if IsBinaryFile(safePath) {
		return nil, fmt.Errorf("cannot read binary file: %s", input.Path)
	}

	// Stat the file
	info, err := os.Stat(safePath)
	if err != nil {
		return nil, fmt.Errorf("cannot access file: %s", input.Path)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file: %s", input.Path)
	}

	// Read the file
	data, err := os.ReadFile(safePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %s", input.Path)
	}

	content := string(data)
	truncated := false

	if int64(len(data)) > maxBytes {
		content = string(data[:maxBytes])
		truncated = true
	}

	// Check for null bytes to detect binary content
	if HasNullBytes([]byte(content), 512) {
		return nil, fmt.Errorf("file appears to be binary: %s", input.Path)
	}

	return &ReadFileOutput{
		Path:      input.Path,
		Content:   content,
		Truncated: truncated,
		Size:      int(info.Size()),
		Type:      contentTypeText,
	}, nil
}
