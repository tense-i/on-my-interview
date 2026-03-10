package repository

import "context"

type Store interface {
	CreateJob(ctx context.Context, params CreateJobParams) (CrawlJob, error)
	GetJob(ctx context.Context, jobID int64) (CrawlJob, error)
	UpdateJob(ctx context.Context, job CrawlJob) error
	UpsertRawPost(ctx context.Context, input RawPostInput) (UpsertRawPostResult, error)
	SaveParseResult(ctx context.Context, input PostParseResultInput) error
	ReplaceQuestions(ctx context.Context, rawPostID int64, questions []InterviewQuestionInput) error
	MarkParseFailed(ctx context.Context, rawPostID int64, message string) error
	RecordJobPost(ctx context.Context, record JobPostRecord) error
	RecordUsageWindow(ctx context.Context, input RecordUsageWindowInput) error
	ListJobs(ctx context.Context, filter ListJobsFilter) ([]CrawlJob, error)
	ListPosts(ctx context.Context, filter PostFilter) ([]PostDetail, error)
	GetPostBySource(ctx context.Context, platform, sourcePostID string) (PostDetail, error)
	ListQuestions(ctx context.Context, filter QuestionFilter) ([]QuestionRecord, error)
	ListCompanies(ctx context.Context) ([]CompanySummary, error)
	ListUsageWindows(ctx context.Context, filter UsageWindowFilter) (UsageWindowList, error)
}
