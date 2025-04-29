package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// OllamaClient is a modular, robust client for the Ollama API
// All configuration is via environment variables (OLLAMA_BASE_URL)

type OllamaClient interface {
	Generate(ctx context.Context, req OllamaGenerateRequest) (OllamaGenerateResponse, error)
	// Add more methods for chat, embeddings, etc.
}

type ollamaClientImpl struct {
	baseURL string
}

func NewOllamaClient() OllamaClient {
	base := os.Getenv("OLLAMA_BASE_URL")
	if base == "" {
		base = "http://localhost:11434"
	}
	return &ollamaClientImpl{baseURL: base}
}

type OllamaGenerateRequest struct {
	Model   string   `json:"model"`
	Prompt  string   `json:"prompt"`
	Stream  bool     `json:"stream,omitempty"`
	Images  []string `json:"images,omitempty"`
	Format  string   `json:"format,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
	System  string   `json:"system,omitempty"`
	Template string  `json:"template,omitempty"`
	Raw     bool     `json:"raw,omitempty"`
	KeepAlive string `json:"keep_alive,omitempty"`
}

type OllamaGenerateResponse struct {
	Model         string `json:"model"`
	Response      string `json:"response"`
	EvalDuration  int64  `json:"eval_duration"`
	TotalDuration int64  `json:"total_duration"`
	Done          bool   `json:"done"`
	// Add more fields as needed
}

func (c *ollamaClientImpl) Generate(ctx context.Context, req OllamaGenerateRequest) (OllamaGenerateResponse, error) {
	var resp OllamaGenerateResponse
	url := fmt.Sprintf("%s/api/generate", c.baseURL)
	body, err := json.Marshal(req)
	if err != nil {
		return resp, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return resp, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(httpReq)
	if err != nil {
		return resp, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return resp, errors.New("ollama: non-200 response")
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return resp, err
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}
