package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"newsletter-ai-coworker/internal/requests"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
)

type AgentService struct {
	llm            llms.Model
	newsletterSvc  *NewsletterService
	infographicSvc *InfographicService
	sessions       map[string]*requests.SessionData
	mu             sync.RWMutex
}

func NewAgentService(groqAPIKey string, newsletterSvc *NewsletterService, infographicSvc *InfographicService) (*AgentService, error) {
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
		newsletterSvc:  newsletterSvc,
		infographicSvc: infographicSvc,
		sessions:       make(map[string]*requests.SessionData),
	}, nil
}

func (s *AgentService) getOrCreateSession(taskID string) *requests.SessionData {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, exists := s.sessions[taskID]; exists {
		return session
	}

	session := &requests.SessionData{
		ContextID: uuid.New().String(),
		History:   []requests.HistoryMessage{},
		Memory:    memory.NewConversationBuffer(),
	}
	s.sessions[taskID] = session
	return session
}

func (s *AgentService) getSystemPrompt() string {
	return `You are a Newsletter Generation Assistant. Your role is to help users create cool newsletters. with links for references`
}

func (s *AgentService) HandleMessage(ctx context.Context, taskID, userText string) (*requests.TaskResult, error) {
	session := s.getOrCreateSession(taskID)
	fmt.Println(session.History)

	userHistoryMsg := s.createHistoryMessage("user", userText)
	session.History = append(session.History, userHistoryMsg)

	session.Memory.SaveContext(ctx, map[string]any{"input": userText}, map[string]any{})

	messages := s.buildMessages(ctx, session, userText)

	response, err := s.llm.GenerateContent(ctx, messages, llms.WithTemperature(0.7))
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	aiResponse := response.Choices[0].Content

	// fmt.Println("AI Response:", aiResponse)s

	session.Memory.SaveContext(ctx, map[string]any{}, map[string]any{"output": aiResponse})

	result := s.processAIResponse(ctx, session, taskID, aiResponse)

	agentHistoryMsg := s.createHistoryMessage("agent", result.Status.Message.Parts[0].Text)
	agentHistoryMsg.MessageID = result.Status.Message.MessageID
	session.History = append(session.History, agentHistoryMsg)

	fmt.Println("Session History:", result)

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

func (s *AgentService) processAIResponse(ctx context.Context, session *requests.SessionData, taskID, aiResponse string) *requests.TaskResult {
	var artifacts []requests.Artifact
	var finalResponse string
	state := "in_progress"

	// if strings.Contains(aiResponse, "GENERATE_NEWSLETTER|") {
	finalResponse, artifacts = s.handleNewsletterGeneration(ctx, session, aiResponse)
	// finalResponse, artifacts = s.handleInfographicGeneration(session)
	// } else if strings.Contains(aiResponse, "GENERATE_INFOGRAPHIC") {
	// } else {
	// 	finalResponse = aiResponse
	// }

	if s.isConversationComplete(finalResponse) {
		state = "completed"
	}

	return s.buildTaskResult(taskID, session, state, finalResponse, artifacts)
}

func (s *AgentService) handleNewsletterGeneration(ctx context.Context, session *requests.SessionData, aiResponse string) (string, []requests.Artifact) {
	parts := strings.Split(aiResponse, "|")
	if len(parts) < 3 {
		return "I apologize, but I couldn't parse the newsletter details. Let's try again.", nil
	}

	title := strings.TrimSpace(strings.TrimPrefix(parts[1], "Title: "))
	content := strings.TrimSpace(strings.TrimPrefix(parts[2], "Content: "))

	session.Title = title
	session.Content = content

	newsletter, err := s.newsletterSvc.Generate(ctx, title, content)
	if err != nil {
		return fmt.Sprintf("I apologize, but I encountered an error generating the newsletter: %v\n\nWould you like to try again?", err), nil
	}

	session.Newsletter = newsletter

	// res, dataURL := s.handleInfographicGeneration(session)
	// if res != "Success" {
	// 	return res, nil
	// }

	artifact := requests.Artifact{
		ArtifactID: uuid.New().String(),
		Name:       "newsletter",
		Parts: []requests.ResponsePart{
			{
				Kind: "data",
				Data: map[string]interface{}{
					"title":   title,
					"content": newsletter,
				},
			},
		},
	}

	response := fmt.Sprintf("Newsletter generated successfully!\n\n%s\n\n═══════════════════════════════════════\n\nWould you like me to create an infographic for this newsletter?", newsletter)

	return response, []requests.Artifact{artifact}
}

func (s *AgentService) handleInfographicGeneration(session *requests.SessionData) (string, []requests.ResponsePart) {
	if session.Newsletter == "" || session.Title == "" {
		return "I don't have a newsletter to create an infographic for. Would you like to create a new newsletter?", nil
	}

	outputPath := "cache/" + session.ContextID + "_infographic.png"

	err := s.infographicSvc.Generate(session.Title, session.Newsletter, outputPath)
	if err != nil {
		return fmt.Sprintf("I apologize, but I encountered an error creating the infographic: %v\n\nWould you like to create another newsletter?", err), nil
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

func (s *AgentService) buildTaskResult(taskID string, session *requests.SessionData, state, message string, artifacts []requests.Artifact) *requests.TaskResult {
	msgID := uuid.New().String()

	return &requests.TaskResult{
		ID:        taskID,
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
