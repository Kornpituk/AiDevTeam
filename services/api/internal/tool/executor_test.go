package tool

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestExecuteToolCalls_Batch(t *testing.T) {
	root := t.TempDir()
	// Create a test file to read
	if err := os.WriteFile(filepath.Join(root, "test.txt"), []byte("hello world"), 0644); err != nil {
		t.Fatalf("cannot create test file: %v", err)
	}

	tcs := []ToolCallRequest{
		{ToolName: "list_files", Input: json.RawMessage(`{"path":"."}`)},
		{ToolName: "read_file", Input: json.RawMessage(`{"path":"test.txt"}`)},
	}

	opts := ToolOptions{WorkspaceRoot: root}
	responses := ExecuteToolCalls(tcs, opts)

	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(responses))
	}

	// First response: list_files should succeed
	if !responses[0].Success {
		t.Errorf("list_files should succeed, got error: %s", responses[0].Error)
	}
	if responses[0].ToolName != "list_files" {
		t.Errorf("expected tool_name 'list_files', got %q", responses[0].ToolName)
	}

	// Second response: read_file should succeed
	if !responses[1].Success {
		t.Errorf("read_file should succeed, got error: %s", responses[1].Error)
	}
	if responses[1].ToolName != "read_file" {
		t.Errorf("expected tool_name 'read_file', got %q", responses[1].ToolName)
	}
}

func TestFormatToolOutput_MarshalError(t *testing.T) {
	// Create a ToolCallResponse with an Output field that cannot be marshalled to JSON
	// (a channel is not JSON-serializable, causing json.Marshal to fail)
	ch := make(chan int)
	resp := &ToolCallResponse{
		ToolName: "test_tool",
		Input:    json.RawMessage(`{}`),
		Output:   ch, // channels cannot be marshalled to JSON
		Success:  true,
	}

	output := FormatToolOutput(resp)

	// Should contain the fallback error message
	expected := `{"tool_name":"test_tool","error":"failed to serialize output"}`
	if output != expected {
		t.Errorf("expected fallback output %q, got %q", expected, output)
	}
}
