package llm

import (
	"context"
	"encoding/json"
)

// Message represents a chat message in LLM conversations.
type Message struct {
	Role    string
	Content string
}

// ToolCall represents a native tool call from the LLM.
type ToolCall struct {
	ID       string
	ToolName string
	Input    json.RawMessage
}

// ToolDefinition describes a tool to the LLM for function calling.
type ToolDefinition struct {
	Name        string
	Description string
	InputSchema any
}

// ChatCompletionResponse wraps both text and tool calls from the LLM.
type ChatCompletionResponse struct {
	Content   string
	ToolCalls []ToolCall
}

type LLMProvider interface {
	ChatCompletion(ctx context.Context, messages []Message) (string, error)
	// ChatCompletionWithTools sends messages + tool definitions to the LLM.
	// The LLM may return text and/or native tool calls.
	// toolChoice controls whether the model should use tools: "" or "auto" (default),
	// "none" (no tools), or "required" (force at least one tool call).
	ChatCompletionWithTools(ctx context.Context, messages []Message, tools []ToolDefinition, toolChoice string) (*ChatCompletionResponse, error)
}

type ProviderType string

const (
	ProviderTypeOpenAI ProviderType = "openai"
	ProviderTypeFake   ProviderType = "fake"
)
