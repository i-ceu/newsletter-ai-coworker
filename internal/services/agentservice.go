package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"blogpost-ai-coworker/internal/requests"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
)

type AgentService struct {
	llm            llms.Model
	blogpostSvc    *BlogPostService
	infographicSvc *InfographicService
	sessions       map[string]*requests.SessionData
	mu             sync.RWMutex
}

func NewAgentService(groqAPIKey string, blogpostSvc *BlogPostService, infographicSvc *InfographicService) (*AgentService, error) {
	llm, err := openai.New(
		openai.WithToken(groqAPIKey),
		openai.WithBaseURL("https://api.groq.com/openai/v1"),
		openai.WithModel("groq/compound"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM: %w", err)
	}

	return &AgentService{
		llm:            llm,
		blogpostSvc:    blogpostSvc,
		infographicSvc: infographicSvc,
		sessions:       make(map[string]*requests.SessionData),
	}, nil
}

func (s *AgentService) getOrCreateSession(MessageID string) *requests.SessionData {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, exists := s.sessions[MessageID]; exists {
		return session
	}

	session := &requests.SessionData{
		ContextID: uuid.New().String(),
		History:   []requests.HistoryMessage{},
		Memory:    memory.NewConversationBuffer(),
	}
	s.sessions[MessageID] = session
	return session
}

func (s *AgentService) getSystemPrompt() string {
	return `GENERATE_NEWSLETTER|Title: `
}

func (s *AgentService) HandleMessage(ctx context.Context, MessageID, userText string) (*requests.TaskResult, error) {
	session := s.getOrCreateSession(MessageID)

	userHistoryMsg := s.createHistoryMessage("user", userText)
	session.History = append(session.History, userHistoryMsg)

	session.Memory.SaveContext(ctx, map[string]any{"input": userText}, map[string]any{})

	// aiResponse := response.Choices[0].Content

	session.Memory.SaveContext(ctx, map[string]any{}, map[string]any{"output": userText})

	result := s.processAIResponse(ctx, session, MessageID, userText)

	agentHistoryMsg := s.createHistoryMessage("agent", result.Artifacts[0].Parts[0].Text)
	agentHistoryMsg.MessageID = result.Artifacts[0].ArtifactID
	session.History = append(session.History, agentHistoryMsg)

	return result, nil
}

func (s *AgentService) buildMessages(ctx context.Context, session *requests.SessionData, userText string) []llms.MessageContent {
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, s.getSystemPrompt()),
	}

	memoryVars, err := session.Memory.LoadMemoryVariables(ctx, map[string]any{})
	if err == nil {
		if history, ok := memoryVars["history"].(string); ok && history != "" {
			messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, history))
		}
	}

	messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, userText))

	return messages
}

func (s *AgentService) processAIResponse(ctx context.Context, session *requests.SessionData, MessageID, userText string) *requests.TaskResult {
	var artifacts []requests.Artifact
	var finalResponse string
	state := "completed"

	// if strings.Contains(aiResponse, "GENERATE_NEWSLETTER|") {
	finalResponse, artifacts = s.handleBlogPostGeneration(ctx, session, userText)
	// finalResponse, artifacts = s.handleInfographicGeneration(session)
	// } else if strings.Contains(aiResponse, "GENERATE_INFOGRAPHIC") {
	// } else {
	// 	finalResponse = aiResponse
	// }

	if s.isConversationComplete(finalResponse) {
		state = "completed"
	}

	return s.buildTaskResult(session, state, finalResponse, artifacts)
}

func (s *AgentService) handleBlogPostGeneration(ctx context.Context, session *requests.SessionData, userText string) (string, []requests.Artifact) {
	// parts := strings.Split(aiResponse, "|")
	// if len(parts) < 3 {
	// 	return "I apologize, but I couldn't parse the blogpost details. Let's try again.", nil
	// }

	title := userText
	// content := strings.TrimSpace(strings.TrimPrefix(parts[2], "Content: "))

	session.Title = title
	// session.Content = content

	blogpost, err := s.blogpostSvc.Generate(ctx, session, title)
	if err != nil {
		return fmt.Sprintf("I apologize, but I encountered an error generating the blogpost: %v\n\nWould you like to try again?", err), nil
	}

	session.BlogPost = blogpost

	// res, dataURL := s.handleInfographicGeneration(session)
	// if res != "Success" {
	// 	return res, nil
	// }

	artifact := requests.Artifact{
		ArtifactID: uuid.New().String(),
		Name:       "blogpost",
		Parts: []requests.ResponsePart{
			{
				Kind: "data",
				Data: map[string]interface{}{
					"title":   title,
					"content": blogpost,
				},
			},
		},
	}

	response := "BlogPost generated successfully!"

	return response, []requests.Artifact{artifact}
}

func (s *AgentService) handleInfographicGeneration(session *requests.SessionData) (string, []requests.ResponsePart) {
	if session.BlogPost == "" || session.Title == "" {
		return "I don't have a blogpost to create an infographic for. Would you like to create a new blogpost?", nil
	}

	outputPath := "cache/" + session.ContextID + "_infographic.png"

	err := s.infographicSvc.Generate(session.Title, session.BlogPost, outputPath)
	if err != nil {
		return fmt.Sprintf("I apologize, but I encountered an error creating the infographic: %v\n\nWould you like to create another blogpost?", err), nil
	}

	artifactPart := []requests.ResponsePart{
		{
			Kind:    "file",
			FileURL: outputPath,
		},
	}

	response := "Success"

	return response, artifactPart
}

func (s *AgentService) isConversationComplete(response string) bool {
	lower := strings.ToLower(response)
	return strings.Contains(lower, "goodbye") || strings.Contains(lower, "have a great day")
}

func (s *AgentService) createHistoryMessage(role, text string) requests.HistoryMessage {
	return requests.HistoryMessage{
		MessageID: uuid.New().String(),
		Role:      role,
		Parts: []requests.ResponsePart{
			{Kind: "text", Text: text},
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

func (s *AgentService) buildTaskResult(session *requests.SessionData, state, message string, artifacts []requests.Artifact) *requests.TaskResult {
	ID := uuid.New().String()
	msgID := uuid.New().String()

	return &requests.TaskResult{
		ID:        ID,
		ContextID: session.ContextID,
		Status: &requests.TaskStatus{
			State:     state,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Message: &requests.ResponseMessage{
				MessageID: msgID,
				Role:      "agent",
				Parts: []requests.ResponsePart{
					{Kind: "text", Text: message},
				},
				Kind: "message",
			},
		},

		Artifacts: artifacts,
		History:   session.History,
		Kind:      "task",
	}
}
