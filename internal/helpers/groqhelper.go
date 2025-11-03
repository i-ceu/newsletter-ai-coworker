package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"blogpost-ai-coworker/internal/requests"
)

func CallGroqAPI(prompt string) (string, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	url := "https://api.groq.com/openai/v1/chat/completions"

	request := requests.GroqRequest{
		Model: "groq/compound",
		Messages: []requests.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Response_Format: requests.ResponseFormat{Type: "json_object"},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response requests.GroqResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}

func ParseGroqResponse(jsonString string) ([]requests.GroqTasksResponse, error) {
	cleanJSON := removeJSONComments(jsonString)

	var tasks []requests.GroqTasksResponse
	err := json.Unmarshal([]byte(cleanJSON), &tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return tasks, nil
}

func removeJSONComments(jsonStr string) string {
	lines := strings.Split(jsonStr, "\n")
	var cleanLines []string

	for _, line := range lines {
		// Remove lines that start with // (comments)
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "//") {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, " ")
}
