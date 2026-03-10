package crawler

import (
	"context"
	"time"
)

type SearchRequest struct {
	Keyword string
	Page    int
}

type RawPostInput struct {
	Platform        string
	SourcePostID    string
	Title           string
	Content         string
	ContentHash     string
	AuthorName      string
	PostURL         string
	CompanyNameRaw  string
	CompanyNameNorm string
	SourceCreatedAt time.Time
	SourceEditedAt  time.Time
	RawPayloadJSON  string
}

type PlatformCrawler interface {
	Platform() string
	SearchPosts(ctx context.Context, req SearchRequest) ([]RawPostInput, error)
}
