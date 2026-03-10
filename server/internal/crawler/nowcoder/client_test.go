package nowcoder

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"on-my-interview/server/internal/crawler"
)

func TestSearchPosts(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/sparta/pc/search" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		response := map[string]any{
			"success": true,
			"data": map[string]any{
				"records": []any{
					map[string]any{
						"data": map[string]any{
							"userBrief": map[string]any{
								"nickname": "ghost",
							},
							"contentData": map[string]any{
								"id":         "",
								"title":      "",
								"content":    "",
								"createTime": float64(0),
								"editTime":   float64(0),
							},
						},
					},
					map[string]any{
						"data": map[string]any{
							"userBrief": map[string]any{
								"nickname": "alice",
							},
							"contentData": map[string]any{
								"id":         "861032210369871872",
								"title":      "阿里面经",
								"content":    "一面问了 Kafka 和 MQ",
								"createTime": float64(1773158076000),
								"editTime":   float64(1773159076000),
							},
						},
					},
				},
			},
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	posts, err := client.SearchPosts(context.Background(), crawler.SearchRequest{
		Keyword: "面经",
		Page:    1,
	})
	if err != nil {
		t.Fatalf("SearchPosts: %v", err)
	}

	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}

	post := posts[0]
	if post.Platform != "nowcoder" {
		t.Fatalf("expected platform nowcoder, got %q", post.Platform)
	}
	if post.SourcePostID != "861032210369871872" {
		t.Fatalf("unexpected source post id %q", post.SourcePostID)
	}
	if post.AuthorName != "alice" {
		t.Fatalf("unexpected author %q", post.AuthorName)
	}
	if post.PostURL != "https://www.nowcoder.com/discuss/861032210369871872" {
		t.Fatalf("unexpected post url %q", post.PostURL)
	}
	if post.Title != "阿里面经" || post.Content != "一面问了 Kafka 和 MQ" {
		t.Fatalf("unexpected post content: %+v", post)
	}
	if post.SourceCreatedAt.IsZero() || post.SourceEditedAt.IsZero() {
		t.Fatalf("expected source timestamps to be set")
	}
	if got := post.SourceCreatedAt.UTC(); !got.Equal(time.UnixMilli(1773158076000).UTC()) {
		t.Fatalf("unexpected created time %s", got)
	}
	if post.RawPayloadJSON == "" {
		t.Fatalf("expected raw payload json")
	}
}
