package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"on-my-interview/server/internal/storage/repository"
)

type JobService interface {
	CreateJob(ctx context.Context, params repository.CreateJobParams) (repository.CrawlJob, error)
	ReparsePost(ctx context.Context, params repository.ReparsePostParams) error
}

type QueryService interface {
	ListJobs(ctx context.Context, filter repository.ListJobsFilter) ([]repository.CrawlJob, error)
	GetJob(ctx context.Context, jobID int64) (repository.CrawlJob, error)
	ListPosts(ctx context.Context, filter repository.PostFilter) ([]repository.PostDetail, error)
	GetPostBySource(ctx context.Context, platform, sourcePostID string) (repository.PostDetail, error)
	ListQuestions(ctx context.Context, filter repository.QuestionFilter) ([]repository.QuestionRecord, error)
	ListCompanies(ctx context.Context) ([]repository.CompanySummary, error)
	ListUsageWindows(ctx context.Context, filter repository.UsageWindowFilter) (repository.UsageWindowList, error)
}

type Dependencies struct {
	JobService   JobService
	QueryService QueryService
}

type rawPostResponse struct {
	ID              int64                  `json:"id"`
	Platform        string                 `json:"platform"`
	SourcePostID    string                 `json:"source_post_id"`
	Title           string                 `json:"title"`
	Content         string                 `json:"content"`
	ContentHash     string                 `json:"content_hash"`
	AuthorName      string                 `json:"author_name"`
	PostURL         string                 `json:"post_url"`
	CompanyNameRaw  string                 `json:"company_name_raw,omitempty"`
	CompanyNameNorm string                 `json:"company_name_norm,omitempty"`
	SourceCreatedAt time.Time              `json:"source_created_at"`
	SourceEditedAt  time.Time              `json:"source_edited_at"`
	LastCrawledAt   time.Time              `json:"last_crawled_at"`
	ParseStatus     repository.ParseStatus `json:"parse_status"`
	ParseAttempts   int                    `json:"parse_attempts"`
	RawPayloadJSON  string                 `json:"raw_payload_json,omitempty"`
}

type parseResultResponse struct {
	RawPostID       int64     `json:"raw_post_id"`
	SchemaVersion   string    `json:"schema_version"`
	LLMProvider     string    `json:"llm_provider"`
	LLMModel        string    `json:"llm_model"`
	CompanyName     string    `json:"company_name,omitempty"`
	Sentiment       string    `json:"sentiment,omitempty"`
	SentimentReason string    `json:"sentiment_reason,omitempty"`
	KeyEventsJSON   string    `json:"key_events_json,omitempty"`
	QuestionsJSON   string    `json:"questions_json,omitempty"`
	TagsJSON        string    `json:"tags_json,omitempty"`
	RawJSON         string    `json:"raw_json,omitempty"`
	ParsedAt        time.Time `json:"parsed_at"`
}

type postDetailResponse struct {
	RawPost     rawPostResponse      `json:"raw_post"`
	ParseResult *parseResultResponse `json:"parse_result,omitempty"`
}

type questionResponse struct {
	RawPostID     int64    `json:"raw_post_id"`
	Platform      string   `json:"platform"`
	CompanyName   string   `json:"company_name,omitempty"`
	QuestionText  string   `json:"question_text"`
	QuestionOrder int      `json:"question_order"`
	Category      string   `json:"category,omitempty"`
	Tags          []string `json:"tags"`
	SourceExcerpt string   `json:"source_excerpt"`
}

type companyResponse struct {
	CompanyName   string `json:"company_name"`
	PostCount     int    `json:"post_count"`
	QuestionCount int    `json:"question_count"`
}

type usageWindowResponse struct {
	WindowStartAt      time.Time `json:"window_start_at"`
	WindowEndAt        time.Time `json:"window_end_at"`
	Timezone           string    `json:"timezone"`
	RequestCount       int64     `json:"request_count"`
	UsageObservedCount int64     `json:"usage_observed_count"`
	PromptTokens       int64     `json:"prompt_tokens"`
	PromptCacheHitTokens int64   `json:"prompt_cache_hit_tokens"`
	PromptCacheMissTokens int64  `json:"prompt_cache_miss_tokens"`
	CompletionTokens   int64     `json:"completion_tokens"`
	TotalTokens        int64     `json:"total_tokens"`
	EstimatedCostCNY   string    `json:"estimated_cost_cny"`
	FinalizedAt        *time.Time `json:"finalized_at,omitempty"`
}

type usageWindowTotalsResponse struct {
	RequestCount         int64  `json:"request_count"`
	UsageObservedCount   int64  `json:"usage_observed_count"`
	PromptTokens         int64  `json:"prompt_tokens"`
	PromptCacheHitTokens int64  `json:"prompt_cache_hit_tokens"`
	PromptCacheMissTokens int64 `json:"prompt_cache_miss_tokens"`
	CompletionTokens     int64  `json:"completion_tokens"`
	TotalTokens          int64  `json:"total_tokens"`
	EstimatedCostCNY     string `json:"estimated_cost_cny"`
}

type usageWindowListResponse struct {
	Items  []usageWindowResponse      `json:"items"`
	Totals usageWindowTotalsResponse  `json:"totals"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func parseInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func parseRFC3339(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func toPostDetailResponse(detail repository.PostDetail) postDetailResponse {
	response := postDetailResponse{
		RawPost: rawPostResponse{
			ID:              detail.RawPost.ID,
			Platform:        detail.RawPost.Platform,
			SourcePostID:    detail.RawPost.SourcePostID,
			Title:           detail.RawPost.Title,
			Content:         detail.RawPost.Content,
			ContentHash:     detail.RawPost.ContentHash,
			AuthorName:      detail.RawPost.AuthorName,
			PostURL:         detail.RawPost.PostURL,
			CompanyNameRaw:  detail.RawPost.CompanyNameRaw,
			CompanyNameNorm: detail.RawPost.CompanyNameNorm,
			SourceCreatedAt: detail.RawPost.SourceCreatedAt,
			SourceEditedAt:  detail.RawPost.SourceEditedAt,
			LastCrawledAt:   detail.RawPost.LastCrawledAt,
			ParseStatus:     detail.RawPost.ParseStatus,
			ParseAttempts:   detail.RawPost.ParseAttempts,
			RawPayloadJSON:  detail.RawPost.RawPayloadJSON,
		},
	}
	if detail.ParseResult != nil {
		response.ParseResult = &parseResultResponse{
			RawPostID:       detail.ParseResult.RawPostID,
			SchemaVersion:   detail.ParseResult.SchemaVersion,
			LLMProvider:     detail.ParseResult.LLMProvider,
			LLMModel:        detail.ParseResult.LLMModel,
			CompanyName:     detail.ParseResult.CompanyName,
			Sentiment:       detail.ParseResult.Sentiment,
			SentimentReason: detail.ParseResult.SentimentReason,
			KeyEventsJSON:   detail.ParseResult.KeyEventsJSON,
			QuestionsJSON:   detail.ParseResult.QuestionsJSON,
			TagsJSON:        detail.ParseResult.TagsJSON,
			RawJSON:         detail.ParseResult.RawJSON,
			ParsedAt:        detail.ParseResult.ParsedAt,
		}
	}
	return response
}

func toQuestionResponse(question repository.QuestionRecord) questionResponse {
	return questionResponse{
		RawPostID:     question.RawPostID,
		Platform:      question.Platform,
		CompanyName:   question.CompanyName,
		QuestionText:  question.QuestionText,
		QuestionOrder: question.QuestionOrder,
		Category:      question.Category,
		Tags:          append([]string(nil), question.Tags...),
		SourceExcerpt: question.SourceExcerpt,
	}
}

func toCompanyResponse(company repository.CompanySummary) companyResponse {
	return companyResponse{
		CompanyName:   company.CompanyName,
		PostCount:     company.PostCount,
		QuestionCount: company.QuestionCount,
	}
}

func toUsageWindowListResponse(result repository.UsageWindowList) usageWindowListResponse {
	items := make([]usageWindowResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toUsageWindowResponse(item))
	}
	return usageWindowListResponse{
		Items: items,
		Totals: usageWindowTotalsResponse{
			RequestCount:         result.Totals.RequestCount,
			UsageObservedCount:   result.Totals.UsageObservedCount,
			PromptTokens:         result.Totals.PromptTokens,
			PromptCacheHitTokens: result.Totals.PromptCacheHitTokens,
			PromptCacheMissTokens: result.Totals.PromptCacheMissTokens,
			CompletionTokens:     result.Totals.CompletionTokens,
			TotalTokens:          result.Totals.TotalTokens,
			EstimatedCostCNY:     result.Totals.EstimatedCostCNY,
		},
	}
}

func toUsageWindowResponse(item repository.UsageWindow) usageWindowResponse {
	location, err := time.LoadLocation(item.Timezone)
	if err != nil {
		location = time.UTC
	}

	response := usageWindowResponse{
		WindowStartAt:      item.WindowStartAt.In(location),
		WindowEndAt:        item.WindowEndAt.In(location),
		Timezone:           item.Timezone,
		RequestCount:       item.RequestCount,
		UsageObservedCount: item.UsageObservedCount,
		PromptTokens:       item.PromptTokens,
		PromptCacheHitTokens: item.PromptCacheHitTokens,
		PromptCacheMissTokens: item.PromptCacheMissTokens,
		CompletionTokens:   item.CompletionTokens,
		TotalTokens:        item.TotalTokens,
		EstimatedCostCNY:   item.EstimatedCostCNY,
	}
	if item.FinalizedAt != nil {
		value := item.FinalizedAt.In(location)
		response.FinalizedAt = &value
	}
	return response
}
