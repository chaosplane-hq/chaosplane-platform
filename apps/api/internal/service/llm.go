package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
)

type LLMClient struct {
	cfg    *config.Config
	client *http.Client
}

func NewLLMClient(cfg *config.Config) *LLMClient {
	return &LLMClient{cfg: cfg, client: &http.Client{Timeout: 60 * time.Second}}
}

func (c *LLMClient) IsConfigured() bool {
	return c.cfg.LLMAPIKey != ""
}

type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLMRequest struct {
	Model    string       `json:"model"`
	Messages []LLMMessage `json:"messages"`
}

type LLMResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *LLMClient) Complete(ctx context.Context, systemPrompt string, messages []LLMMessage) (string, error) {
	if !c.IsConfigured() {
		return "AI response requires LLM_API_KEY configuration.", nil
	}

	allMessages := []LLMMessage{{Role: "system", Content: systemPrompt}}
	allMessages = append(allMessages, messages...)

	reqBody := LLMRequest{
		Model:    c.cfg.LLMModel,
		Messages: allMessages,
	}

	endpoint := c.resolveEndpoint()
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal llm request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create llm request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.LLMAPIKey)

	if c.cfg.LLMProvider == "anthropic" {
		req.Header.Set("x-api-key", c.cfg.LLMAPIKey)
		req.Header.Set("anthropic-version", "2023-06-01")
		req.Header.Del("Authorization")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm api request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read llm response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("llm api error %d: %s", resp.StatusCode, string(respBody))
	}

	if c.cfg.LLMProvider == "anthropic" {
		return c.parseAnthropicResponse(respBody)
	}
	return c.parseOpenAIResponse(respBody)
}

func (c *LLMClient) parseOpenAIResponse(body []byte) (string, error) {
	var resp LLMResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("parse openai response: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in openai response")
	}
	return resp.Choices[0].Message.Content, nil
}

func (c *LLMClient) parseAnthropicResponse(body []byte) (string, error) {
	var resp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("parse anthropic response: %w", err)
	}
	if len(resp.Content) == 0 {
		return "", fmt.Errorf("no content in anthropic response")
	}
	return resp.Content[0].Text, nil
}

func (c *LLMClient) resolveEndpoint() string {
	switch c.cfg.LLMProvider {
	case "anthropic":
		return "https://api.anthropic.com/v1/messages"
	default:
		return "https://api.openai.com/v1/chat/completions"
	}
}
