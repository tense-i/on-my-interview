package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"on-my-interview/server/internal/llm"
)

type Config struct {
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

type Extractor struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewExtractor(cfg Config) *Extractor {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = "gpt-4.1-mini"
	}
	return &Extractor{
		baseURL:    baseURL,
		apiKey:     cfg.APIKey,
		model:      model,
		httpClient: httpClient,
	}
}

func (e *Extractor) Name() string {
	return "openai-compatible"
}

func (e *Extractor) Extract(ctx context.Context, post llm.RawPostForLLM) (*llm.StructuredPost, error) {
	requestBody := map[string]any{
		"model": e.model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": buildUserPrompt(post),
			},
		},
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/v1/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(e.apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+e.apiKey)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var completion completionResponse
	if err := json.NewDecoder(resp.Body).Decode(&completion); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("completion returned no choices")
	}

	content := strings.TrimSpace(completion.Choices[0].Message.Content)
	var structured llm.StructuredPost
	if err := json.Unmarshal([]byte(content), &structured); err != nil {
		return nil, fmt.Errorf("decode structured payload: %w", err)
	}
	structured.Usage = completion.Usage
	if err := structured.Validate(); err != nil {
		return nil, fmt.Errorf("validate structured payload: %w", err)
	}
	return &structured, nil
}

func buildUserPrompt(post llm.RawPostForLLM) string {
	return fmt.Sprintf("platform: %s\npost_id: %s\ntitle: %s\nurl: %s\ncontent:\n%s", post.Platform, post.SourcePostID, post.Title, post.PostURL, post.Content)
}

type completionResponse struct {
	Usage   *llm.Usage `json:"usage"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

const systemPrompt = `You extract structured interview experience data and must return valid json only.
Use exactly this schema:
{"schema_version":"v1","company":{"raw_name":"string","normalized_name":"string","confidence":0.0},"post_type":"interview_experience","sentiment":{"label":"positive|neutral|negative|mixed","confidence":0.0,"reason":"string"},"key_events":[{"type":"string","summary":"string","round":"string"}],"questions":[{"order":1,"question":"string","category":"string","tags":["string"],"source_excerpt":"string"}],"overall_tags":["string"]}.
Keep company as an object, sentiment as an object, key_events as objects, and each question item with order/question/category/tags/source_excerpt.
Split each interview question into its own row-sized item.
If some fields are unknown, keep the required object shape and use empty strings, empty arrays, or 0.`
