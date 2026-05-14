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
