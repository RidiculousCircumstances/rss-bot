package summary

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"log"
	"rss-bot/internal/config"
	"strings"
	"sync"
)

type OpenAISummarizer struct {
	client  *openai.Client
	prompt  string
	enabled bool
	mu      sync.Mutex
}

func New(apiKey string, prompt string) *OpenAISummarizer {
	summarizer := &OpenAISummarizer{
		client: openai.NewClient(apiKey),
		prompt: prompt,
	}

	log.Printf("openai summarizer enabled: %v", apiKey != "")

	if apiKey != "" {
		summarizer.enabled = true
	}

	return summarizer
}

func (s *OpenAISummarizer) Summarize(ctx context.Context, text string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.enabled {
		return "", nil
	}

	request := openai.ChatCompletionRequest{
		Model: config.Get().OpenAIModel,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf("%s%s", text, s.prompt),
			},
		},
		MaxTokens:   256,
		Temperature: 0.7,
		TopP:        1,
	}

	resp, err := s.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return "", err
	}

	stringSummary := strings.TrimSpace(resp.Choices[0].Message.Content)

	if strings.HasSuffix(stringSummary, ".") {
		return stringSummary, nil
	}

	sentences := strings.Split(stringSummary, ".")
	sentences = sentences[:len(sentences)-1]

	return strings.Join(sentences, ".") + ".", nil
}
