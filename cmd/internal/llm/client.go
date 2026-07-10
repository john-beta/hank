package llm

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// OpenAIClient implements Client using the OpenAI Responses API.
type OpenAIClient struct {
	client *openai.Client
}

// compile-time check that OpenAIClient satisfies the Client interface.
var _ Client = (*OpenAIClient)(nil)

// NewOpenAIClient builds a client authenticated with the given API key.
func NewOpenAIClient(apiKey string) *OpenAIClient {
	c := openai.NewClient(option.WithAPIKey(apiKey))
	return &OpenAIClient{client: &c}
}
