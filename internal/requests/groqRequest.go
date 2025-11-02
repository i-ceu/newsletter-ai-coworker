package requests

type GroqRequest struct {
	Messages        []Message      `json:"messages"`
	Model           string         `json:"model"`
	Response_Format ResponseFormat `json:"response_format"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ResponseFormat struct {
	Type string `json:"type"`
}

type GroqResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type GroqTasksResponse struct {
	TaskTitle       string `json:"task_title"`
	TaskDescription string `json:"task_description"`
}
