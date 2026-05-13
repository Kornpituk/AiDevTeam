package llm

import (
	"context"
	"fmt"
)

type FakeProvider struct {
	Responses     []string
	ResponseIndex int
	ReturnError   error
}

func NewFakeProvider() *FakeProvider {
	return &FakeProvider{
		Responses: []string{
			"Fake LLM response: Task analyzed successfully.",
		},
	}
}

func (f *FakeProvider) ChatCompletion(ctx context.Context, messages []Message) (string, error) {
	if f.ReturnError != nil {
		return "", f.ReturnError
	}

	if len(f.Responses) == 0 {
		return "Default fake response: Step executed successfully.", nil
	}

	response := f.Responses[f.ResponseIndex%len(f.Responses)]
	f.ResponseIndex++

	if len(messages) > 0 {
		lastMessage := messages[len(messages)-1]
		return fmt.Sprintf("%s\n\n[Context processed: %d chars]", response, len(lastMessage.Content)), nil
	}

	return response, nil
}

func (f *FakeProvider) ChatCompletionWithTools(ctx context.Context, messages []Message, tools []ToolDefinition, toolChoice string) (*ChatCompletionResponse, error) {
	content, err := f.ChatCompletion(ctx, messages)
	if err != nil {
		return nil, err
	}
	return &ChatCompletionResponse{Content: content}, nil
}
