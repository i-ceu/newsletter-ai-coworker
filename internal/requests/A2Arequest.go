package requests

import (
	"github.com/tmc/langchaingo/memory"
)

// A2A request structs
type A2ARequest struct {
	JsonRPC string  `validate:"required"`
	ID      string  `validate:"required"`
	Method  string  `validate:"required"`
	Params  *Params `validate:"required"`
}

type Params struct {
	Message       *ReqMessage    `json:"message"`
	Configuration *Configuration `json:"configuration"`
}

type ReqMessage struct {
	Kind      string        `json:"kind"`
	Role      string        `json:"role"`
	Parts     []MessagePart `json:"parts"`
	MessageID string        `json:"messageId"`
}
type MessagePart struct {
	Kind string `json:"kind"`
	Text string `json:"text"`
}
type Configuration struct {
	Blocking bool `json:"blocking"`
}

// response
type A2AResponse struct {
	JsonRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Result  *TaskResult `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type TaskResult struct {
	ID        string           `json:"id"`
	ContextID string           `json:"contextId"`
	Status    *TaskStatus      `json:"status"`
	Artifacts []Artifact       `json:"artifacts,omitempty"`
	History   []HistoryMessage `json:"history"`
	Kind      string           `json:"kind"`
}

type TaskStatus struct {
	State     string           `json:"state"`
	Timestamp string           `json:"timestamp"`
	Message   *ResponseMessage `json:"message"`
}

type ResponseMessage struct {
	MessageID string         `json:"messageId"`
	Role      string         `json:"role"`
	Parts     []ResponsePart `json:"parts"`
	Kind      string         `json:"kind"`
}

type ResponsePart struct {
	Kind    string                 `json:"kind"`
	Text    string                 `json:"text,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
	FileURL string                 `json:"file_url,omitempty"`
}

type Artifact struct {
	ArtifactID string         `json:"artifactId"`
	Name       string         `json:"name"`
	Parts      []ResponsePart `json:"parts"`
}

type HistoryMessage struct {
	MessageID string         `json:"messageId"`
	Role      string         `json:"role"`
	Parts     []ResponsePart `json:"parts"`
	Timestamp string         `json:"timestamp"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Session management
type SessionData struct {
	ContextID string
	History   []HistoryMessage
	Memory    *memory.ConversationBuffer
	Title     string
	Content   string
	BlogPost  string
}
