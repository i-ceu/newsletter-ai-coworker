package services

import (
	"context"
	"fmt"
	"newsletter-ai-coworker/internal/requests"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type NewsletterService struct {
	llm llms.Model
}

func NewNewsletterService(groqAPIKey string) *NewsletterService {
	llm, err := openai.New(
		openai.WithToken(groqAPIKey),
		openai.WithBaseURL("https://api.groq.com/openai/v1"),
		openai.WithModel("groq/compound"),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create newsletter service: %v", err))
	}

	return &NewsletterService{llm: llm}
}

func (s *NewsletterService) getSystemPrompt() string {
	return `GENERATE_NEWSLETTER|Title: `
}

func (s *NewsletterService) Generate(ctx context.Context, session *requests.SessionData, title string) (string, error) {
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, s.getSystemPrompt()),
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a newsletter writer. Create engaging, well-structured newsletters."),
		llms.TextParts(llms.ChatMessageTypeHuman, fmt.Sprintf(`Create a cool newsletter based on the following:

Title and Content: %s

Generate a well-structured newsletter with:
- A catchy headline
- An engaging introduction
- 3-4 key points or sections with clear headings (use format "## Heading")
- A conclusion or call-to-action

Keep it concise and professional.`, title)),
	}

	memoryVars, err := session.Memory.LoadMemoryVariables(ctx, map[string]any{})
	if err == nil {
		if history, ok := memoryVars["history"].(string); ok && history != "" {
			messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, history))
		}
	}

	messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, title))

	response, err := s.llm.GenerateContent(ctx, messages, llms.WithTemperature(0.7))
	if err != nil {
		return "", err
	}

	return response.Choices[0].Content, nil
}
