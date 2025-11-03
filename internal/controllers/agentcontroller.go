package controllers

import (
	"blogpost-ai-coworker/internal/requests"
	"blogpost-ai-coworker/internal/services"
	"encoding/json"
	"fmt"
	"net/http"
)

type AgentController struct {
	agentService *services.AgentService
}

func NewAgentController(agentService *services.AgentService) *AgentController {
	return &AgentController{
		agentService: agentService,
	}
}

func (c *AgentController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req requests.A2ARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.sendError(w, "", -32700, "Parse error")
		return
	}

	if req.JsonRPC != "2.0" {
		c.sendError(w, req.ID, -32600, "Invalid Request")
		return
	}

	if req.Method != "message/send" {
		c.sendError(w, req.ID, -32601, "Method not found")
		return
	}

	if req.Params == nil || req.Params.Message == nil {
		c.sendError(w, req.ID, -32602, "Invalid params")
		return
	}

	userText := c.extractUserText(req.Params.Message.Parts)

	fmt.Println(req.Params.Message.MessageID)
	result, err := c.agentService.HandleMessage(r.Context(), req.Params.Message.MessageID, userText)
	if err != nil {
		c.sendError(w, req.ID, -32000, err.Error())
		return
	}

	c.sendSuccess(w, req.ID, result)
}

func (c *AgentController) extractUserText(parts []requests.MessagePart) string {
	return parts[0].Text
}

func (c *AgentController) sendSuccess(w http.ResponseWriter, id string, result *requests.TaskResult) {
	resp := requests.A2AResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (c *AgentController) sendError(w http.ResponseWriter, id string, code int, message string) {
	fmt.Println("sending error:", message)
	resp := requests.A2AResponse{
		JsonRPC: "2.0",
		ID:      id,
		Error: &requests.RPCError{
			Code:    code,
			Message: message,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
