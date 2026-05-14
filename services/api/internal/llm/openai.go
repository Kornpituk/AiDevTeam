package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Kornpituk/AiDevTeam/services/api/internal/config"
)

type OpenAIProvider struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

type openAIMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []openAIToolCall `json:"tool_calls,omitempty"`
}

type openAITool struct {
	Type     string     `json:"type"`
	Function openAIFunc `json:"function"`
}

type openAIFunc struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

type openAIToolCall struct {
	ID       string        `json:"id"`
	Type     string        `json:"type"`
	Function openAIFuncCall `json:"function"`
}

type openAIFuncCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
	Tools       []openAITool    `json:"tools,omitempty"`
	ToolChoice  any             `json:"tool_choice,omitempty"`
}

type openAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int            `json:"index"`
		Message      openAIMessage `json:"message"`
		FinishReason string         `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func NewOpenAIProvider(cfg config.LLMConfig) *OpenAIProvider {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &OpenAIProvider{
		apiKey:  cfg.APIKey,
		model:   cfg.Model,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (p *OpenAIProvider) ChatCompletion(ctx context.Context, messages []Message) (string, error) {
	resp, err := p.ChatCompletionWithTools(ctx, messages, nil, "")
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

func (p *OpenAIProvider) ChatCompletionWithTools(ctx context.Context, messages []Message, tools []ToolDefinition, toolChoice string) (*ChatCompletionResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("LLM_API_KEY not configured")
	}

	openAIMsgs := make([]openAIMessage, len(messages))
	for i, m := range messages {
		openAIMsgs[i] = openAIMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	openAITools := make([]openAITool, len(tools))
	for i, t := range tools {
		openAITools[i] = openAITool{
			Type: "function",
			Function: openAIFunc{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.InputSchema,
			},
		}
	}

	reqBody := openAIRequest{
		Model:       p.model,
		Messages:    openAIMsgs,
		Temperature: 0.7,
	}

	if toolChoice != "" {
		if toolChoice == "none" || toolChoice == "auto" || toolChoice == "required" {
			reqBody.ToolChoice = toolChoice
		}
	}

	// Only include tools in the request if toolChoice is not "none"
	if toolChoice != "none" {
		reqBody.Tools = openAITools
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var result openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("API error: %s", result.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := result.Choices[0]
	chatResp := &ChatCompletionResponse{
		Content: choice.Message.Content,
	}

	if len(choice.Message.ToolCalls) > 0 {
		toolCalls := make([]ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			toolCalls[i] = ToolCall{
				ID:       tc.ID,
				ToolName: tc.Function.Name,
				Input:    json.RawMessage(tc.Function.Arguments),
			}
		}
		chatResp.ToolCalls = toolCalls
	}

	return chatResp, nil
}

func NewProviderFromConfig(cfg config.LLMConfig) (LLMProvider, error) {
	switch ProviderType(cfg.Provider) {
	case ProviderTypeOpenAI:
		return NewOpenAIProvider(cfg), nil
	case ProviderTypeFake:
		return NewFakeProvider(), nil
	default:
		return NewOpenAIProvider(cfg), nil
	}
}

func NewProviderWithTimeout(cfg config.LLMConfig, timeout time.Duration) (LLMProvider, error) {
	cfg.Timeout = timeout
	return NewProviderFromConfig(cfg)
}
