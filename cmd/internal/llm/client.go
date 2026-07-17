package llm

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// OpenAIClient implements Client using the OpenAI Responses API (not Chat
// Completions).
type OpenAIClient struct {
	client *openai.Client
}

var _ Client = (*OpenAIClient)(nil)

func NewOpenAIClient(apiKey string) *OpenAIClient {
	c := openai.NewClient(option.WithAPIKey(apiKey))
	return &OpenAIClient{client: &c}
}
