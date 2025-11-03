package services

import (
	"blogpost-ai-coworker/internal/requests"
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type BlogPostService struct {
	llm llms.Model
}

func NewBlogPostService(groqAPIKey string) *BlogPostService {
	llm, err := openai.New(
		openai.WithToken(groqAPIKey),
		openai.WithBaseURL("https://api.groq.com/openai/v1"),
		openai.WithModel("groq/compound"),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create blogpost service: %v", err))
	}

	return &BlogPostService{llm: llm}
}

func (s *BlogPostService) getSystemPrompt() string {
	return `GENERATE_NEWSLETTER|Title: `
}

func (s *BlogPostService) Generate(ctx context.Context, session *requests.SessionData, title string) (string, error) {
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, s.getSystemPrompt()),
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a blog writer. Create engaging, well-structured blog posts."),
		llms.TextParts(llms.ChatMessageTypeHuman, fmt.Sprintf(`Create a cool blog post based on the following:

Title and Content: %s

Generate a well-structured blogpost with:
- A catchy headline
- An engaging introduction
-if you notice any code snippets, include it in the blog explanation
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
