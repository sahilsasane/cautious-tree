package llm

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
)

var (
	ErrNoResponseFromGemini = errors.New("no response from Gemini")
)

type ChatSession struct {
	Messages []Data
}

type Data struct {
	Role  string              `json:"role"`
	Parts []map[string]string `json:"parts"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func NewChatSession() *ChatSession {
	return &ChatSession{
		Messages: []Data{},
	}
}

func (c *ChatSession) AddUserMessage(text string) {
	c.Messages = append(c.Messages, Data{
		Role:  "user",
		Parts: []map[string]string{{"text": text}},
	})
}

func (c *ChatSession) AddModelMessage(text string) {
	c.Messages = append(c.Messages, Data{
		Role:  "model",
		Parts: []map[string]string{{"text": text}},
	})
}

func GetGeminiResponse(messages []Data) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=" + apiKey

	// Construct request payload
	payload := map[string]interface{}{
		"contents": messages,
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", nil
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var geminiRes GeminiResponse
	if err := json.Unmarshal(body, &geminiRes); err != nil {
		return "", err
	}

	if len(geminiRes.Candidates) > 0 && len(geminiRes.Candidates[0].Content.Parts) > 0 {
		return geminiRes.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", ErrNoResponseFromGemini
}
