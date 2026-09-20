package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEmbedUsesOpenAIEmbeddingsContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/embeddings" {
			t.Fatalf("request = %s %s, want POST /v1/embeddings", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q, want Bearer test-key", got)
		}
		var input struct {
			Model string `json:"model"`
			Input string `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		if input.Model != "text-embedding-3-small" || input.Input != "hello" {
			t.Fatalf("input = %+v", input)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"embedding":[0.25,0.5]}]}`))
	}))
	defer server.Close()

	got, err := NewEmbedder(server.URL+"/v1", "test-key", "text-embedding-3-small").Embed(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != 0.25 || got[1] != 0.5 {
		t.Fatalf("embedding = %v", got)
	}
}

func TestGenerateUsesOpenAIChatCompletionsContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("request = %s %s, want POST /v1/chat/completions", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q, want Bearer test-key", got)
		}
		var input struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
			Stream bool `json:"stream"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		if input.Model != "gpt-4o-mini" || len(input.Messages) != 1 || input.Messages[0].Role != "user" || input.Messages[0].Content != "hello" || input.Stream {
			t.Fatalf("input = %+v", input)
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"answer"}}]}`))
	}))
	defer server.Close()

	got, err := NewGenerator(server.URL+"/v1", "test-key", "gpt-4o-mini").Generate(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if got != "answer" {
		t.Fatalf("answer = %q, want answer", got)
	}
}

func TestGenerateIncludesOpenAIStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"message":"rate limited"}}`))
	}))
	defer server.Close()

	_, err := NewGenerator(server.URL, "test-key", "model").Generate(context.Background(), "hello")
	if err == nil || !strings.Contains(err.Error(), "429 Too Many Requests") || !strings.Contains(err.Error(), "rate limited") {
		t.Fatalf("error = %v, want OpenAI status and detail", err)
	}
}
