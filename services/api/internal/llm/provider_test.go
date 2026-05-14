package llm

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/config"
)

func TestFakeProvider_ChatCompletion_ReturnsResponse(t *testing.T) {
	p := NewFakeProvider()
	resp, err := p.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hello"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == "" {
		t.Fatal("expected non-empty response")
	}
}

func TestFakeProvider_ChatCompletion_ReturnsError(t *testing.T) {
	p := NewFakeProvider()
	p.ReturnError = fmt.Errorf("simulated error")
	_, err := p.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hello"}})
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}

func TestFakeProvider_ChatCompletionWithTools_ReturnsResponse(t *testing.T) {
	p := NewFakeProvider()
	resp, err := p.ChatCompletionWithTools(context.Background(), []Message{{Role: "user", Content: "hello"}}, nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content == "" {
		t.Fatal("expected non-empty content")
	}
}

func TestFakeProvider_ChatCompletion_ProcessesMessages(t *testing.T) {
	p := NewFakeProvider()
	resp, err := p.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hello world"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp, "11") { // len("hello world") = 11, so "[Context processed: 11 chars]"
		t.Errorf("expected response to include context length info, got: %s", resp)
	}
}

func TestNewProviderFromConfig_OpenAI(t *testing.T) {
	cfg := config.LLMConfig{Provider: "openai", APIKey: "test-key", Model: "gpt-4", Timeout: 30}
	p, err := NewProviderFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := p.(*OpenAIProvider); !ok {
		t.Errorf("expected OpenAIProvider, got %T", p)
	}
}

func TestNewProviderFromConfig_Fake(t *testing.T) {
	cfg := config.LLMConfig{Provider: "fake"}
	p, err := NewProviderFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := p.(*FakeProvider); !ok {
		t.Errorf("expected FakeProvider, got %T", p)
	}
}

func TestNewProviderFromConfig_InvalidDefaultsToOpenAI(t *testing.T) {
	cfg := config.LLMConfig{Provider: "invalid", APIKey: "test-key"}
	p, err := NewProviderFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := p.(*OpenAIProvider); !ok {
		t.Errorf("expected OpenAIProvider (fallback), got %T", p)
	}
}

func TestFakeProvider_ChatCompletionWithTools_ReturnsError(t *testing.T) {
	p := NewFakeProvider()
	p.ReturnError = fmt.Errorf("simulated tool error")
	_, err := p.ChatCompletionWithTools(context.Background(), []Message{{Role: "user", Content: "hello"}}, nil, "")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}

func TestFakeProvider_EmptyResponses(t *testing.T) {
	p := NewFakeProvider()
	// Set empty responses list
	p.Responses = []string{}

	// Should not panic and should return default response
	resp, err := p.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hello"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == "" {
		t.Fatal("expected non-empty response even with empty responses list")
	}

	// ChatCompletionWithTools should also work
	resp2, err := p.ChatCompletionWithTools(context.Background(), []Message{{Role: "user", Content: "hello"}}, nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp2.Content == "" {
		t.Fatal("expected non-empty content even with empty responses list")
	}
}


