package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"on-my-interview/server/internal/crawler"
	"on-my-interview/server/internal/llm"
	"on-my-interview/server/internal/storage/repository"
)

func TestWorkerRunJob(t *testing.T) {
	t.Parallel()

	store := newFakeStore()
	store.createJobResult = repository.CrawlJob{
		ID:           1,
		TriggerType:  repository.JobTriggerManual,
		Status:       repository.JobStatusPending,
		Platforms:    []string{"nowcoder"},
		Keywords:     []string{"面经"},
		Pages:        1,
		ForceReparse: false,
	}
	store.upsertResults["p1"] = repository.UpsertRawPostResult{
		Disposition: repository.UpsertDispositionInserted,
		Post: repository.RawPost{
			ID:           101,
			Platform:     "nowcoder",
			SourcePostID: "p1",
		},
	}
	store.upsertResults["p2"] = repository.UpsertRawPostResult{
		Disposition: repository.UpsertDispositionUnchanged,
		Post: repository.RawPost{
			ID:           102,
			Platform:     "nowcoder",
			SourcePostID: "p2",
		},
	}
	store.upsertResults["p3"] = repository.UpsertRawPostResult{
		Disposition: repository.UpsertDispositionUpdated,
		Post: repository.RawPost{
			ID:           103,
			Platform:     "nowcoder",
			SourcePostID: "p3",
		},
	}
	store.upsertResults["p4"] = repository.UpsertRawPostResult{
		Disposition: repository.UpsertDispositionInserted,
		Post: repository.RawPost{
			ID:           104,
			Platform:     "nowcoder",
			SourcePostID: "p4",
		},
	}

	registry := crawler.NewRegistry(fakeCrawler{
		platform: "nowcoder",
		posts: []crawler.RawPostInput{
			newCrawlerPost("p1"),
			newCrawlerPost("p2"),
			newCrawlerPost("p3"),
			newCrawlerPost("p4"),
		},
	})

	extractor := fakeExtractor{
		results: map[string]*llm.StructuredPost{
			"p1": structuredPost("阿里巴巴"),
			"p3": structuredPost("阿里巴巴"),
		},
		errors: map[string]error{
			"p4": errors.New("provider timeout"),
		},
	}

	service := NewService(store, registry, extractor, "gpt-test")

	job, err := service.CreateJob(context.Background(), repository.CreateJobParams{
		TriggerType:  repository.JobTriggerManual,
		Platforms:    []string{"nowcoder"},
		Keywords:     []string{"面经"},
		Pages:        1,
		ForceReparse: false,
	})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	if err := service.RunJob(context.Background(), job.ID); err != nil {
		t.Fatalf("RunJob: %v", err)
	}

	if len(store.savedParseResults) != 2 {
		t.Fatalf("expected 2 parse results, got %d", len(store.savedParseResults))
	}
	if len(store.recordedUsageInputs) != 2 {
		t.Fatalf("expected 2 recorded usage inputs, got %d", len(store.recordedUsageInputs))
	}
	if len(store.savedQuestions) != 2 {
		t.Fatalf("expected 2 saved question groups, got %d", len(store.savedQuestions))
	}
	if len(store.failedPosts) != 1 || store.failedPosts[0] != 104 {
		t.Fatalf("expected post 104 to be marked failed, got %+v", store.failedPosts)
	}
	if len(store.recordedJobPosts) != 4 {
		t.Fatalf("expected 4 job post records, got %d", len(store.recordedJobPosts))
	}

	finalJob := store.jobs[job.ID]
	if finalJob.Status != repository.JobStatusCompleted {
		t.Fatalf("expected completed job status, got %s", finalJob.Status)
	}
	if finalJob.StartedAt == nil || finalJob.FinishedAt == nil {
		t.Fatalf("expected started and finished timestamps to be set")
	}
}

func newCrawlerPost(sourceID string) crawler.RawPostInput {
	return crawler.RawPostInput{
		Platform:        "nowcoder",
		SourcePostID:    sourceID,
		Title:           "阿里面经",
		Content:         "问了 Kafka 为什么能保证高吞吐？",
		ContentHash:     sourceID + "-hash",
		AuthorName:      "alice",
		PostURL:         "https://www.nowcoder.com/discuss/" + sourceID,
		SourceCreatedAt: time.Unix(0, 0).UTC(),
		SourceEditedAt:  time.Unix(60, 0).UTC(),
		RawPayloadJSON:  `{"id":"` + sourceID + `"}`,
	}
}

func structuredPost(company string) *llm.StructuredPost {
	return &llm.StructuredPost{
		SchemaVersion: "v1",
		Company: llm.Company{
			NormalizedName: company,
			Confidence:     0.9,
		},
		Sentiment: llm.Sentiment{
			Label: "positive",
		},
		Questions: []llm.StructuredQuestion{
			{
				Order:         1,
				Question:      "Kafka 为什么能保证高吞吐？",
				Category:      "backend",
				Tags:          []string{"kafka", "MQ", "后端", "阿里"},
				SourceExcerpt: "问了 Kafka 为什么吞吐高",
			},
		},
		Usage: &llm.Usage{
			PromptTokens:          120,
			PromptCacheHitTokens:  20,
			PromptCacheMissTokens: 100,
			CompletionTokens:      30,
			TotalTokens:           150,
		},
	}
}

type fakeCrawler struct {
	platform string
	posts    []crawler.RawPostInput
}

func (f fakeCrawler) Platform() string { return f.platform }

func (f fakeCrawler) SearchPosts(context.Context, crawler.SearchRequest) ([]crawler.RawPostInput, error) {
	return append([]crawler.RawPostInput(nil), f.posts...), nil
}

type fakeExtractor struct {
	results map[string]*llm.StructuredPost
	errors  map[string]error
}

func (f fakeExtractor) Name() string { return "fake-extractor" }

func (f fakeExtractor) Extract(_ context.Context, post llm.RawPostForLLM) (*llm.StructuredPost, error) {
	if err := f.errors[post.SourcePostID]; err != nil {
		return nil, err
	}
	result := f.results[post.SourcePostID]
	if result == nil {
		return nil, errors.New("missing mock extractor result")
	}
	return result, nil
}

type fakeStore struct {
	createJobResult   repository.CrawlJob
	jobs              map[int64]repository.CrawlJob
	upsertResults     map[string]repository.UpsertRawPostResult
	savedParseResults []repository.PostParseResultInput
	savedQuestions    [][]repository.InterviewQuestionInput
	failedPosts       []int64
	recordedJobPosts  []repository.JobPostRecord
	recordedUsageInputs []repository.RecordUsageWindowInput
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		jobs:          map[int64]repository.CrawlJob{},
		upsertResults: map[string]repository.UpsertRawPostResult{},
	}
}

func (f *fakeStore) CreateJob(_ context.Context, params repository.CreateJobParams) (repository.CrawlJob, error) {
	job := f.createJobResult
	job.Platforms = append([]string(nil), params.Platforms...)
	job.Keywords = append([]string(nil), params.Keywords...)
	job.Pages = params.Pages
	job.ForceReparse = params.ForceReparse
	f.jobs[job.ID] = job
	return job, nil
}

func (f *fakeStore) GetJob(_ context.Context, jobID int64) (repository.CrawlJob, error) {
	job, ok := f.jobs[jobID]
	if !ok {
		return repository.CrawlJob{}, errors.New("job not found")
	}
	return job, nil
}

func (f *fakeStore) ListJobs(context.Context, repository.ListJobsFilter) ([]repository.CrawlJob, error) {
	result := make([]repository.CrawlJob, 0, len(f.jobs))
	for _, job := range f.jobs {
		result = append(result, job)
	}
	return result, nil
}

func (f *fakeStore) UpdateJob(_ context.Context, job repository.CrawlJob) error {
	f.jobs[job.ID] = job
	return nil
}

func (f *fakeStore) UpsertRawPost(_ context.Context, input repository.RawPostInput) (repository.UpsertRawPostResult, error) {
	result, ok := f.upsertResults[input.SourcePostID]
	if !ok {
		return repository.UpsertRawPostResult{}, errors.New("unexpected upsert input")
	}
	return result, nil
}

func (f *fakeStore) SaveParseResult(_ context.Context, input repository.PostParseResultInput) error {
	f.savedParseResults = append(f.savedParseResults, input)
	return nil
}

func (f *fakeStore) ReplaceQuestions(_ context.Context, rawPostID int64, questions []repository.InterviewQuestionInput) error {
	f.savedQuestions = append(f.savedQuestions, append([]repository.InterviewQuestionInput(nil), questions...))
	return nil
}

func (f *fakeStore) MarkParseFailed(_ context.Context, rawPostID int64, _ string) error {
	f.failedPosts = append(f.failedPosts, rawPostID)
	return nil
}

func (f *fakeStore) RecordJobPost(_ context.Context, record repository.JobPostRecord) error {
	f.recordedJobPosts = append(f.recordedJobPosts, record)
	return nil
}

func (f *fakeStore) ListPosts(context.Context, repository.PostFilter) ([]repository.PostDetail, error) {
	return nil, nil
}

func (f *fakeStore) GetPostBySource(_ context.Context, platform, sourcePostID string) (repository.PostDetail, error) {
	for _, result := range f.upsertResults {
		if result.Post.Platform == platform && result.Post.SourcePostID == sourcePostID {
			return repository.PostDetail{RawPost: result.Post}, nil
		}
	}
	return repository.PostDetail{}, errors.New("post not found")
}

func (f *fakeStore) ListQuestions(context.Context, repository.QuestionFilter) ([]repository.QuestionRecord, error) {
	return nil, nil
}

func (f *fakeStore) ListCompanies(context.Context) ([]repository.CompanySummary, error) {
	return nil, nil
}

func (f *fakeStore) RecordUsageWindow(_ context.Context, input repository.RecordUsageWindowInput) error {
	f.recordedUsageInputs = append(f.recordedUsageInputs, input)
	return nil
}

func (f *fakeStore) ListUsageWindows(context.Context, repository.UsageWindowFilter) (repository.UsageWindowList, error) {
	return repository.UsageWindowList{}, nil
}
