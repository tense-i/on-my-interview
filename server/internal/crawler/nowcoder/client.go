package nowcoder

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"on-my-interview/server/internal/crawler"
)

const defaultBaseURL = "https://gw-c.nowcoder.com"

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultBaseURL
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

func (c *Client) Platform() string {
	return "nowcoder"
}

func (c *Client) SearchPosts(ctx context.Context, req crawler.SearchRequest) ([]crawler.RawPostInput, error) {
	body, err := json.Marshal(map[string]any{
		"type":  "all",
		"query": req.Keyword,
		"page":  req.Page,
		"tag":   []string{},
		"order": "create",
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/sparta/pc/search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}

	var payload searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if !payload.Success {
		return nil, fmt.Errorf("search response unsuccessful")
	}

	posts := make([]crawler.RawPostInput, 0, len(payload.Data.Records))
	for _, record := range payload.Data.Records {
		content := record.Data.ContentData
		if strings.TrimSpace(content.ID) == "" || strings.TrimSpace(content.Title) == "" || strings.TrimSpace(content.Content) == "" {
			continue
		}
		rawJSON, err := json.Marshal(record.Data)
		if err != nil {
			return nil, fmt.Errorf("marshal raw payload: %w", err)
		}
		posts = append(posts, crawler.RawPostInput{
			Platform:        c.Platform(),
			SourcePostID:    content.ID,
			Title:           content.Title,
			Content:         content.Content,
			ContentHash:     contentHash(content.Title, content.Content),
			AuthorName:      record.Data.UserBrief.Nickname,
			PostURL:         "https://www.nowcoder.com/discuss/" + content.ID,
			SourceCreatedAt: time.UnixMilli(content.CreateTime).UTC(),
			SourceEditedAt:  time.UnixMilli(content.EditTime).UTC(),
			RawPayloadJSON:  string(rawJSON),
		})
	}

	return posts, nil
}

func contentHash(title, content string) string {
	sum := sha256.Sum256([]byte(title + "\n" + content))
	return hex.EncodeToString(sum[:])
}

type searchResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Records []searchRecord `json:"records"`
	} `json:"data"`
}

type searchRecord struct {
	Data struct {
		UserBrief struct {
			Nickname string `json:"nickname"`
		} `json:"userBrief"`
		ContentData struct {
			ID         string `json:"id"`
			Title      string `json:"title"`
			Content    string `json:"content"`
			CreateTime int64  `json:"createTime"`
			EditTime   int64  `json:"editTime"`
		} `json:"contentData"`
	} `json:"data"`
}
