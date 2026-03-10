package repository

import (
	"time"

	"on-my-interview/server/internal/llm"
)

type ParseStatus string

const (
	ParseStatusPending ParseStatus = "pending"
	ParseStatusParsed  ParseStatus = "parsed"
	ParseStatusFailed  ParseStatus = "failed"
)

type UpsertDisposition string

const (
	UpsertDispositionInserted  UpsertDisposition = "inserted"
	UpsertDispositionUpdated   UpsertDisposition = "updated"
	UpsertDispositionUnchanged UpsertDisposition = "unchanged"
)

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

type RawPost struct {
	ID              int64
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
	LastCrawledAt   time.Time
	ParseStatus     ParseStatus
	ParseAttempts   int
	RawPayloadJSON  string
}

type UpsertRawPostResult struct {
	Disposition UpsertDisposition
	Post        RawPost
}

type JobTrigger string

const (
	JobTriggerManual    JobTrigger = "manual"
	JobTriggerScheduler JobTrigger = "scheduler"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

type CreateJobParams struct {
	TriggerType  JobTrigger
	Platforms    []string
	Keywords     []string
	Pages        int
	ForceReparse bool
}

type CrawlJob struct {
	ID           int64
	TriggerType  JobTrigger
	Status       JobStatus
	Platforms    []string
	Keywords     []string
	Pages        int
	ForceReparse bool
	StatsJSON    string
	ErrorMessage string
	StartedAt    *time.Time
	FinishedAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PostParseResultInput struct {
	RawPostID        int64
	SchemaVersion    string
	LLMProvider      string
	LLMModel         string
	CompanyName      string
	Sentiment        string
	SentimentReason  string
	KeyEventsJSON    string
	QuestionsJSON    string
	TagsJSON         string
	RawJSON          string
	ParsedAt         time.Time
}

type InterviewQuestionInput struct {
	RawPostID     int64
	Platform      string
	CompanyName   string
	QuestionText  string
	QuestionOrder int
	Category      string
	TagsJSON      string
	SourceExcerpt string
}

type JobPostRecord struct {
	JobID       int64
	RawPostID   int64
	Disposition string
	Message     string
}

type ReparsePostParams struct {
	Platform     string
	SourcePostID string
}

type ListJobsFilter struct {
	Limit  int
	Offset int
}

type PostFilter struct {
	Platform string
	Company  string
	Tag      string
	Limit    int
	Offset   int
}

type QuestionFilter struct {
	Platform string
	Company  string
	Tag      string
	Query    string
	Limit    int
	Offset   int
}

type PostParseResult struct {
	RawPostID        int64
	SchemaVersion    string
	LLMProvider      string
	LLMModel         string
	CompanyName      string
	Sentiment        string
	SentimentReason  string
	KeyEventsJSON    string
	QuestionsJSON    string
	TagsJSON         string
	RawJSON          string
	ParsedAt         time.Time
}

type PostDetail struct {
	RawPost     RawPost
	ParseResult *PostParseResult
}

type QuestionRecord struct {
	RawPostID     int64
	Platform      string
	CompanyName   string
	QuestionText  string
	QuestionOrder int
	Category      string
	Tags          []string
	SourceExcerpt string
}

type CompanySummary struct {
	CompanyName   string
	PostCount     int
	QuestionCount int
}

type UsageWindowStatus string

const (
	UsageWindowStatusAggregating UsageWindowStatus = "aggregating"
	UsageWindowStatusFinalized   UsageWindowStatus = "finalized"
)

type RecordUsageWindowInput struct {
	Provider   string
	Model      string
	RecordedAt time.Time
	Usage      llm.Usage
}

type UsageWindowFilter struct {
	Limit  int
	Offset int
	From   *time.Time
	To     *time.Time
}

type UsageWindow struct {
	WindowStartAt      time.Time
	WindowEndAt        time.Time
	Timezone           string
	Status             UsageWindowStatus
	RequestCount       int64
	UsageObservedCount int64
	PromptTokens       int64
	PromptCacheHitTokens int64
	PromptCacheMissTokens int64
	CompletionTokens   int64
	TotalTokens        int64
	EstimatedCostCNY   string
	FinalizedAt        *time.Time
}

type UsageWindowTotals struct {
	RequestCount         int64
	UsageObservedCount   int64
	PromptTokens         int64
	PromptCacheHitTokens int64
	PromptCacheMissTokens int64
	CompletionTokens     int64
	TotalTokens          int64
	EstimatedCostCNY     string
}

type UsageWindowList struct {
	Items  []UsageWindow
	Totals UsageWindowTotals
}
