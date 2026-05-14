package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/config"
)

func TestOpenAIProvider_ChatCompletion_Success(t *testing.T) {
	// Mock server returning a valid OpenAI response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("missing or invalid Authorization header")
		}
		// Return success response
		resp := openAIResponse{
			Choices: []struct {
				Index        int            `json:"index"`
				Message      openAIMessage `json:"message"`
				FinishReason string         `json:"finish_reason"`
			}{
				{Index: 0, Message: openAIMessage{Role: "assistant", Content: "Hello!"}, FinishReason: "stop"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := config.LLMConfig{APIKey: "test-key", Model: "gpt-4", BaseURL: server.URL, Timeout: 30 * time.Second}
	p := NewOpenAIProvider(cfg)

	resp, err := p.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "say hi"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "Hello!" {
		t.Errorf("expected 'Hello!', got '%s'", resp)
	}
}

func TestOpenAIProvider_ChatCompletion_NoAPIKey(t *testing.T) {
	cfg := config.LLMConfig{APIKey: "", Model: "gpt-4", Timeout: 30}
	p := NewOpenAIProvider(cfg)

	_, err := p.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
}

func TestOpenAIProvider_ChatCompletion_DefaultBaseURL(t *testing.T) {
	cfg := config.LLMConfig{APIKey: "test-key", Model: "gpt-4", BaseURL: "", Timeout: 30}
	p := NewOpenAIProvider(cfg)
	if p.baseURL != "https://api.openai.com/v1" {
		t.Errorf("expected default base URL, got %s", p.baseURL)
	}
}

func TestOpenAIProvider_ChatCompletion_CustomBaseURL(t *testing.T) {
	cfg := config.LLMConfig{APIKey: "test-key", Model: "gpt-4", BaseURL: "https://custom.example.com/v1", Timeout: 30}
	p := NewOpenAIProvider(cfg)
	if p.baseURL != "https://custom.example.com/v1" {
		t.Errorf("expected custom base URL, got %s", p.baseURL)
	}
}

func TestOpenAIProvider_ChatCompletion_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(openAIResponse{
			Error: &struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			}{Message: "rate limit exceeded", Type: "rate_limit"},
		})
	}))
	defer server.Close()

	cfg := config.LLMConfig{APIKey: "test-key", Model: "gpt-4", BaseURL: server.URL, Timeout: 30 * time.Second}
	p := NewOpenAIProvider(cfg)

	_, err := p.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected API error")
	}
}

func TestOpenAIProvider_ChatCompletion_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
	}))
	defer server.Close()

	cfg := config.LLMConfig{APIKey: "bad-key", Model: "gpt-4", BaseURL: server.URL, Timeout: 30 * time.Second}
	p := NewOpenAIProvider(cfg)

	_, err := p.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected HTTP error")
	}
}

func TestOpenAIProvider_ChatCompletionWithTools_Success(t *testing.T) {
	// Mock server returning tool call
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIResponse{
			Choices: []struct {
				Index        int            `json:"index"`
				Message      openAIMessage `json:"message"`
				FinishReason string         `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: openAIMessage{
						Role: "assistant",
						Content: "",
						ToolCalls: []openAIToolCall{
							{ID: "call_1", Type: "function", Function: openAIFuncCall{Name: "test_tool", Arguments: `{"key":"value"}`}},
						},
					},
					FinishReason: "tool_calls",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := config.LLMConfig{APIKey: "test-key", Model: "gpt-4", BaseURL: server.URL, Timeout: 30 * time.Second}
	p := NewOpenAIProvider(cfg)

	resp, err := p.ChatCompletionWithTools(context.Background(), []Message{{Role: "user", Content: "use tool"}}, []ToolDefinition{{Name: "test_tool"}}, "auto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].ToolName != "test_tool" {
		t.Errorf("expected tool name 'test_tool', got '%s'", resp.ToolCalls[0].ToolName)
	}
}

func TestOpenAIProvider_ChatCompletionWithTools_ToolChoiceNone(t *testing.T) {
	// When toolChoice="none", the tools field should NOT be included in the request,
	// and the provider should not attempt to parse tool calls from the response.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Decode request to verify tools field is not present
		var req openAIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Tools != nil {
			t.Error("tools should be nil when toolChoice is 'none'")
		}
		if req.ToolChoice != "none" {
			t.Errorf("expected tool_choice 'none', got %v", req.ToolChoice)
		}

		// Return a text response (no tool calls)
		resp := openAIResponse{
			Choices: []struct {
				Index        int            `json:"index"`
				Message      openAIMessage `json:"message"`
				FinishReason string         `json:"finish_reason"`
			}{
				{Index: 0, Message: openAIMessage{Role: "assistant", Content: "Text response without tools."}, FinishReason: "stop"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := config.LLMConfig{APIKey: "test-key", Model: "gpt-4", BaseURL: server.URL, Timeout: 30 * time.Second}
	p := NewOpenAIProvider(cfg)

	resp, err := p.ChatCompletionWithTools(context.Background(), []Message{{Role: "user", Content: "hello"}}, []ToolDefinition{{Name: "some_tool"}}, "none")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "Text response without tools." {
		t.Errorf("expected text response, got %q", resp.Content)
	}
	if len(resp.ToolCalls) != 0 {
		t.Errorf("expected 0 tool calls with toolChoice='none', got %d", len(resp.ToolCalls))
	}
}

func TestOpenAIProvider_ChatCompletionWithTools_ToolChoiceRequired(t *testing.T) {
	// When toolChoice="required", the tool_choice field should be set to "required" in the request body.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req openAIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.ToolChoice != "required" {
			t.Errorf("expected tool_choice 'required', got %v", req.ToolChoice)
		}
		if req.Tools == nil || len(req.Tools) == 0 {
			t.Error("tools should be present when toolChoice is 'required'")
		}

		// Return a tool call response
		resp := openAIResponse{
			Choices: []struct {
				Index        int            `json:"index"`
				Message      openAIMessage `json:"message"`
				FinishReason string         `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: openAIMessage{
						Role:    "assistant",
						Content: "",
						ToolCalls: []openAIToolCall{
							{ID: "call_1", Type: "function", Function: openAIFuncCall{Name: "test_tool", Arguments: `{"key":"value"}`}},
						},
					},
					FinishReason: "tool_calls",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := config.LLMConfig{APIKey: "test-key", Model: "gpt-4", BaseURL: server.URL, Timeout: 30 * time.Second}
	p := NewOpenAIProvider(cfg)

	resp, err := p.ChatCompletionWithTools(context.Background(), []Message{{Role: "user", Content: "use tool"}}, []ToolDefinition{{Name: "test_tool"}}, "required")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].ToolName != "test_tool" {
		t.Errorf("expected tool name 'test_tool', got %q", resp.ToolCalls[0].ToolName)
	}
}

func TestOpenAIProvider_ChatCompletionWithTools_MultipleToolCalls(t *testing.T) {
	// LLM returns 2 tool calls in one response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIResponse{
			Choices: []struct {
				Index        int            `json:"index"`
				Message      openAIMessage `json:"message"`
				FinishReason string         `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: openAIMessage{
						Role:    "assistant",
						Content: "",
						ToolCalls: []openAIToolCall{
							{ID: "call_1", Type: "function", Function: openAIFuncCall{Name: "read_file", Arguments: `{"path":"a.txt"}`}},
							{ID: "call_2", Type: "function", Function: openAIFuncCall{Name: "list_files", Arguments: `{"path":"."}`}},
						},
					},
					FinishReason: "tool_calls",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := config.LLMConfig{APIKey: "test-key", Model: "gpt-4", BaseURL: server.URL, Timeout: 30 * time.Second}
	p := NewOpenAIProvider(cfg)

	resp, err := p.ChatCompletionWithTools(context.Background(), []Message{{Role: "user", Content: "use tools"}}, []ToolDefinition{{Name: "read_file"}, {Name: "list_files"}}, "auto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.ToolCalls) != 2 {
		t.Fatalf("expected 2 tool calls, got %d", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].ToolName != "read_file" {
		t.Errorf("first tool should be 'read_file', got %q", resp.ToolCalls[0].ToolName)
	}
	if resp.ToolCalls[1].ToolName != "list_files" {
		t.Errorf("second tool should be 'list_files', got %q", resp.ToolCalls[1].ToolName)
	}
}

func TestOpenAIProvider_ContextCancellation(t *testing.T) {
	// Context is cancelled during the request -> should return error ASAP
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wait indefinitely (server will be closed by defer)
		<-r.Context().Done()
	}))
	defer server.Close()

	cfg := config.LLMConfig{APIKey: "test-key", Model: "gpt-4", BaseURL: server.URL, Timeout: 30 * time.Second}
	p := NewOpenAIProvider(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := p.ChatCompletionWithTools(ctx, []Message{{Role: "user", Content: "hi"}}, nil, "")
	if err == nil {
		t.Fatal("expected error due to cancelled context")
	}
}

func TestOpenAIProvider_ZeroChoices(t *testing.T) {
	// API returns "choices": [] -> should handle gracefully with an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIResponse{
			Choices: []struct {
				Index        int            `json:"index"`
				Message      openAIMessage `json:"message"`
				FinishReason string         `json:"finish_reason"`
			}{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := config.LLMConfig{APIKey: "test-key", Model: "gpt-4", BaseURL: server.URL, Timeout: 30 * time.Second}
	p := NewOpenAIProvider(cfg)

	_, err := p.ChatCompletionWithTools(context.Background(), []Message{{Role: "user", Content: "hi"}}, nil, "")
	if err == nil {
		t.Fatal("expected error for zero choices")
	}
	if err.Error() != "no choices in response" {
		t.Errorf("expected 'no choices in response', got: %v", err)
	}
}

func TestNewProviderWithTimeout_OverridesTimeout(t *testing.T) {
	cfg := config.LLMConfig{Provider: "openai", APIKey: "test-key", Timeout: 30}
	p, err := NewProviderWithTimeout(cfg, 120*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	op, ok := p.(*OpenAIProvider)
	if !ok {
		t.Fatalf("expected OpenAIProvider, got %T", p)
	}
	if op.client.Timeout != 120*time.Second {
		t.Errorf("expected timeout 120s, got %v", op.client.Timeout)
	}
}
