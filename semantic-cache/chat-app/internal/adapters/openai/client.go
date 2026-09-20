package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL, apiKey, model string
	http                   *http.Client
}

func NewEmbedder(baseURL, apiKey, model string) *Client {
	return newClient(baseURL, apiKey, model, 30*time.Second)
}

func NewGenerator(baseURL, apiKey, model string) *Client {
	return newClient(baseURL, apiKey, model, 120*time.Second)
}

func newClient(baseURL, apiKey, model string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) Embed(ctx context.Context, text string) ([]float32, error) {
	var out struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := c.post(ctx, "/embeddings", map[string]any{"model": c.model, "input": text}, &out); err != nil {
		return nil, err
	}
	if len(out.Data) != 1 || len(out.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("OpenAI returned no embedding")
	}
	return out.Data[0].Embedding, nil
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	input := map[string]any{
		"model":    c.model,
		"messages": []map[string]string{{"role": "user", "content": prompt}},
		"stream":   false,
	}
	if err := c.post(ctx, "/chat/completions", input, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 || strings.TrimSpace(out.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("OpenAI returned an empty response")
	}
	return out.Choices[0].Message.Content, nil
}

func (c *Client) post(ctx context.Context, path string, input, output any) error {
	body, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("marshal OpenAI request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create OpenAI request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("call OpenAI endpoint %s: %w", path, err)
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		responseBody, _ := io.ReadAll(io.LimitReader(res.Body, 4<<10))
		if detail := strings.TrimSpace(string(responseBody)); detail != "" {
			return fmt.Errorf("OpenAI endpoint %s returned %s: %s", path, res.Status, detail)
		}
		return fmt.Errorf("OpenAI endpoint %s returned %s", path, res.Status)
	}
	if err := json.NewDecoder(res.Body).Decode(output); err != nil {
		return fmt.Errorf("decode OpenAI endpoint %s response: %w", path, err)
	}
	return nil
}
