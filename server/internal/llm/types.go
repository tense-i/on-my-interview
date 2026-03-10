package llm

import "context"

type Extractor interface {
	Name() string
	Extract(ctx context.Context, post RawPostForLLM) (*StructuredPost, error)
}

type RawPostForLLM struct {
	Platform     string
	SourcePostID string
	Title        string
	Content      string
	PostURL      string
}

type Company struct {
	RawName        string  `json:"raw_name"`
	NormalizedName string  `json:"normalized_name"`
	Confidence     float64 `json:"confidence"`
}

type Sentiment struct {
	Label      string  `json:"label"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

type KeyEvent struct {
	Type    string `json:"type"`
	Summary string `json:"summary"`
	Round   string `json:"round"`
}

type StructuredQuestion struct {
	Order         int      `json:"order"`
	Question      string   `json:"question"`
	Category      string   `json:"category"`
	Tags          []string `json:"tags"`
	SourceExcerpt string   `json:"source_excerpt"`
}

type Usage struct {
	PromptTokens          int64 `json:"prompt_tokens"`
	PromptCacheHitTokens  int64 `json:"prompt_cache_hit_tokens"`
	PromptCacheMissTokens int64 `json:"prompt_cache_miss_tokens"`
	CompletionTokens      int64 `json:"completion_tokens"`
	TotalTokens           int64 `json:"total_tokens"`
}

type StructuredPost struct {
	SchemaVersion string               `json:"schema_version"`
	Company       Company              `json:"company"`
	PostType      string               `json:"post_type"`
	Sentiment     Sentiment            `json:"sentiment"`
	KeyEvents     []KeyEvent           `json:"key_events"`
	Questions     []StructuredQuestion `json:"questions"`
	OverallTags   []string             `json:"overall_tags"`
	Usage         *Usage               `json:"-"`
}

type QuestionRow struct {
	RawPostID    int64
	Platform     string
	CompanyName  string
	QuestionText string
	QuestionOrder int
	Category     string
	Tags         []string
	SourceExcerpt string
}
