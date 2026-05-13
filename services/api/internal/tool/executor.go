package tool

import (
	"encoding/json"
	"fmt"
)

// ToolCallRequest represents a tool call request parsed from LLM output.
type ToolCallRequest struct {
	ToolName string          `json:"tool_name"`
	Input    json.RawMessage `json:"input"`
}

// ToolCallResponse represents the result of executing a tool call.
type ToolCallResponse struct {
	ToolName string      `json:"tool_name"`
	Input    json.RawMessage `json:"input"`
	Output   any         `json:"output,omitempty"`
	Error    string      `json:"error,omitempty"`
	Success  bool        `json:"success"`
}

// ToolCallsWrapper is the expected JSON structure from LLM output.
type ToolCallsWrapper struct {
	ToolCalls []ToolCallRequest `json:"tool_calls"`
}

// ParseToolCalls attempts to parse a JSON tool_calls block from text.
// It looks for a JSON object containing a "tool_calls" key.
func ParseToolCalls(text string) ([]ToolCallRequest, error) {
	// Try to find and parse JSON with tool_calls
	var wrapper ToolCallsWrapper
	if err := json.Unmarshal([]byte(text), &wrapper); err == nil && len(wrapper.ToolCalls) > 0 {
		return wrapper.ToolCalls, nil
	}

	// Try to find a JSON block within the text (between { and })
	start := -1
	depth := 0
	for i, ch := range text {
		if ch == '{' {
			if start == -1 {
				start = i
			}
			depth++
		} else if ch == '}' {
			depth--
			if depth == 0 && start != -1 {
				chunk := text[start : i+1]
				var w ToolCallsWrapper
				if err := json.Unmarshal([]byte(chunk), &w); err == nil && len(w.ToolCalls) > 0 {
					return w.ToolCalls, nil
				}
				start = -1
			}
		}
	}

	return nil, nil // no tool calls found
}

// ExecuteToolCall executes a single tool call and returns the response.
func ExecuteToolCall(tc ToolCallRequest, opts ToolOptions) *ToolCallResponse {
	resp := &ToolCallResponse{
		ToolName: tc.ToolName,
		Input:    tc.Input,
	}

	tool, err := GetTool(tc.ToolName)
	if err != nil {
		resp.Error = err.Error()
		resp.Success = false
		return resp
	}

	output, err := tool.Execute(tc.Input, opts.WorkspaceRoot, opts)
	if err != nil {
		resp.Error = err.Error()
		resp.Success = false
		return resp
	}

	resp.Output = output
	resp.Success = true
	return resp
}

// ExecuteToolCalls executes a list of tool calls in order and returns responses.
func ExecuteToolCalls(tcs []ToolCallRequest, opts ToolOptions) []*ToolCallResponse {
	responses := make([]*ToolCallResponse, 0, len(tcs))
	for _, tc := range tcs {
		resp := ExecuteToolCall(tc, opts)
		responses = append(responses, resp)
	}
	return responses
}

// FormatToolOutput serializes the tool output to JSON string.
func FormatToolOutput(resp *ToolCallResponse) string {
	output, err := json.Marshal(resp)
	if err != nil {
		return fmt.Sprintf(`{"tool_name":"%s","error":"failed to serialize output"}`, resp.ToolName)
	}
	return string(output)
}
