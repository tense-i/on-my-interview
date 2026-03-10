package openai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"on-my-interview/server/internal/llm"
)

func TestExtractorExtract(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected authorization header %q", got)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		var request struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
			ResponseFormat struct {
				Type string `json:"type"`
			} `json:"response_format"`
		}
		if err := json.Unmarshal(body, &request); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if request.ResponseFormat.Type != "json_object" {
			t.Fatalf("unexpected response format %+v", request.ResponseFormat)
		}
		if len(request.Messages) != 2 {
			t.Fatalf("expected 2 messages, got %d", len(request.Messages))
		}
		system := request.Messages[0].Content
		for _, snippet := range []string{
			"valid json only",
			`"company":{"raw_name":"string","normalized_name":"string","confidence":0.0}`,
			`"sentiment":{"label":"positive|neutral|negative|mixed","confidence":0.0,"reason":"string"}`,
			`"questions":[{"order":1,"question":"string","category":"string","tags":["string"],"source_excerpt":"string"}]`,
		} {
			if !strings.Contains(system, snippet) {
				t.Fatalf("system prompt missing schema snippet %q in %q", snippet, system)
			}
		}

		response := map[string]any{
			"usage": map[string]any{
				"prompt_tokens":            120,
				"prompt_cache_hit_tokens":  40,
				"prompt_cache_miss_tokens": 80,
				"completion_tokens":        36,
				"total_tokens":             156,
			},
			"choices": []any{
				map[string]any{
					"message": map[string]any{
						"content": `{"schema_version":"v1","company":{"raw_name":"阿里淘天","normalized_name":"阿里巴巴","confidence":0.92},"post_type":"interview_experience","sentiment":{"label":"positive","confidence":0.81,"reason":"整体反馈顺利"},"key_events":[{"type":"interview_round","summary":"一面偏后端","round":"first"}],"questions":[{"order":1,"question":"Kafka 为什么能保证高吞吐？","category":"backend","tags":["kafka","MQ","后端","阿里"],"source_excerpt":"问了 Kafka 为什么吞吐高"}],"overall_tags":["阿里","后端"]}`,
					},
				},
			},
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	extractor := NewExtractor(Config{
		BaseURL:    server.URL,
		APIKey:     "test-key",
		Model:      "gpt-test",
		HTTPClient: server.Client(),
	})

	post, err := extractor.Extract(context.Background(), llm.RawPostForLLM{
		Platform:     "nowcoder",
		SourcePostID: "861032210369871872",
		Title:        "阿里面经",
		Content:      "一面问了 Kafka 为什么能保证高吞吐？",
		PostURL:      "https://www.nowcoder.com/discuss/861032210369871872",
	})
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}

	if post.Company.NormalizedName != "阿里巴巴" {
		t.Fatalf("unexpected company %+v", post.Company)
	}
	if post.Sentiment.Label != "positive" {
		t.Fatalf("unexpected sentiment %+v", post.Sentiment)
	}
	if post.Usage == nil {
		t.Fatalf("expected usage to be decoded")
	}
	if post.Usage.PromptTokens != 120 || post.Usage.PromptCacheHitTokens != 40 || post.Usage.PromptCacheMissTokens != 80 || post.Usage.CompletionTokens != 36 || post.Usage.TotalTokens != 156 {
		t.Fatalf("unexpected usage %+v", post.Usage)
	}
	if len(post.Questions) != 1 {
		t.Fatalf("expected 1 question, got %d", len(post.Questions))
	}

	rows := llm.FlattenQuestions(42, "nowcoder", post.Company.NormalizedName, post.Questions)
	if len(rows) != 1 {
		t.Fatalf("expected 1 flattened question row, got %d", len(rows))
	}
	if rows[0].QuestionText != "Kafka 为什么能保证高吞吐？" {
		t.Fatalf("unexpected flattened question %+v", rows[0])
	}
	if len(rows[0].Tags) != 4 {
		t.Fatalf("expected tags to be preserved, got %+v", rows[0].Tags)
	}
}
