package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpapi "on-my-interview/server/internal/http"
	"on-my-interview/server/internal/storage/repository"
)

func TestAPI(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	jobService := &fakeJobService{
		createJobResult: repository.CrawlJob{
			ID:          7,
			TriggerType: repository.JobTriggerManual,
			Status:      repository.JobStatusPending,
			Platforms:   []string{"nowcoder"},
			Keywords:    []string{"面经"},
			Pages:       1,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	queryService := &fakeQueryService{
		jobs: []repository.CrawlJob{
			{
				ID:          7,
				TriggerType: repository.JobTriggerManual,
				Status:      repository.JobStatusCompleted,
				Platforms:   []string{"nowcoder"},
				Keywords:    []string{"面经"},
				Pages:       1,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		postList: []repository.PostDetail{
			{
				RawPost: repository.RawPost{
					ID:           42,
					Platform:     "nowcoder",
					SourcePostID: "861032210369871872",
					Title:        "阿里面经",
					Content:      "一面问了 Kafka 为什么能保证高吞吐？",
					PostURL:      "https://www.nowcoder.com/discuss/861032210369871872",
					ParseStatus:  repository.ParseStatusParsed,
				},
				ParseResult: &repository.PostParseResult{
					RawPostID:    42,
					CompanyName:  "阿里巴巴",
					Sentiment:    "positive",
					QuestionsJSON: `[{"question":"Kafka 为什么能保证高吞吐？"}]`,
				},
			},
		},
		postDetail: repository.PostDetail{
			RawPost: repository.RawPost{
				ID:           42,
				Platform:     "nowcoder",
				SourcePostID: "861032210369871872",
				Title:        "阿里面经",
				Content:      "一面问了 Kafka 为什么能保证高吞吐？",
				PostURL:      "https://www.nowcoder.com/discuss/861032210369871872",
				ParseStatus:  repository.ParseStatusParsed,
			},
			ParseResult: &repository.PostParseResult{
				RawPostID:     42,
				CompanyName:   "阿里巴巴",
				Sentiment:     "positive",
				SentimentReason: "整体反馈顺利",
				QuestionsJSON: `[{"question":"Kafka 为什么能保证高吞吐？"}]`,
			},
		},
		questions: []repository.QuestionRecord{
			{
				RawPostID:     42,
				Platform:      "nowcoder",
				CompanyName:   "阿里巴巴",
				QuestionText:  "Kafka 为什么能保证高吞吐？",
				QuestionOrder: 1,
				Category:      "backend",
				Tags:          []string{"kafka", "MQ", "后端", "阿里"},
				SourceExcerpt: "问了 Kafka 为什么吞吐高",
			},
		},
		companies: []repository.CompanySummary{
			{
				CompanyName:   "阿里巴巴",
				PostCount:     1,
				QuestionCount: 1,
			},
		},
		usageWindows: repository.UsageWindowList{
			Items: []repository.UsageWindow{
				{
					WindowStartAt:      time.Date(2026, 3, 10, 16, 0, 0, 0, time.UTC),
					WindowEndAt:        time.Date(2026, 3, 11, 4, 0, 0, 0, time.UTC),
					Timezone:           "Asia/Shanghai",
					Status:             repository.UsageWindowStatusFinalized,
					RequestCount:       2,
					UsageObservedCount: 2,
					PromptTokens:       460000,
					PromptCacheHitTokens: 200000,
					PromptCacheMissTokens: 260000,
					CompletionTokens:   100000,
					TotalTokens:        560000,
					EstimatedCostCNY:   "0.86000000",
				},
			},
			Totals: repository.UsageWindowTotals{
				RequestCount:         2,
				UsageObservedCount:   2,
				PromptTokens:         460000,
				PromptCacheHitTokens: 200000,
				PromptCacheMissTokens: 260000,
				CompletionTokens:     100000,
				TotalTokens:          560000,
				EstimatedCostCNY:     "0.86000000",
			},
		},
	}

	router := httpapi.NewRouter(httpapi.Dependencies{
		JobService:   jobService,
		QueryService: queryService,
	})

	t.Run("create job", func(t *testing.T) {
		body := bytes.NewBufferString(`{"platforms":["nowcoder"],"keywords":["面经"],"pages":1}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/crawl/jobs", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Fatalf("expected 202, got %d", rec.Code)
		}

		var payload map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if int(payload["id"].(float64)) != 7 {
			t.Fatalf("unexpected job response %+v", payload)
		}
	})

	t.Run("list posts", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/posts?platform=nowcoder&company=%E9%98%BF%E9%87%8C%E5%B7%B4%E5%B7%B4&tag=kafka", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if queryService.lastPostFilter.Platform != "nowcoder" || queryService.lastPostFilter.Company != "阿里巴巴" || queryService.lastPostFilter.Tag != "kafka" {
			t.Fatalf("unexpected post filter %+v", queryService.lastPostFilter)
		}
	})

	t.Run("get post detail", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/nowcoder/861032210369871872", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if queryService.lastPostLookupPlatform != "nowcoder" || queryService.lastPostLookupID != "861032210369871872" {
			t.Fatalf("unexpected post lookup %s/%s", queryService.lastPostLookupPlatform, queryService.lastPostLookupID)
		}
	})

	t.Run("list questions", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/questions?company=%E9%98%BF%E9%87%8C%E5%B7%B4%E5%B7%B4&tag=kafka", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if queryService.lastQuestionFilter.Company != "阿里巴巴" || queryService.lastQuestionFilter.Tag != "kafka" {
			t.Fatalf("unexpected question filter %+v", queryService.lastQuestionFilter)
		}
	})

	t.Run("list companies", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/companies", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if !queryService.listCompaniesCalled {
			t.Fatalf("expected list companies to be called")
		}
	})

	t.Run("list usage windows", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/usage/windows?limit=10", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if queryService.lastUsageWindowFilter.Limit != 10 {
			t.Fatalf("unexpected usage filter %+v", queryService.lastUsageWindowFilter)
		}
	})
}

type fakeJobService struct {
	createJobResult repository.CrawlJob
	lastCreateJob   repository.CreateJobParams
	lastReparse     repository.ReparsePostParams
}

func (f *fakeJobService) CreateJob(_ context.Context, params repository.CreateJobParams) (repository.CrawlJob, error) {
	f.lastCreateJob = params
	return f.createJobResult, nil
}

func (f *fakeJobService) ReparsePost(_ context.Context, params repository.ReparsePostParams) error {
	f.lastReparse = params
	return nil
}

type fakeQueryService struct {
	jobs                  []repository.CrawlJob
	postList              []repository.PostDetail
	postDetail            repository.PostDetail
	questions             []repository.QuestionRecord
	companies             []repository.CompanySummary
	usageWindows          repository.UsageWindowList
	lastPostFilter        repository.PostFilter
	lastQuestionFilter    repository.QuestionFilter
	lastUsageWindowFilter repository.UsageWindowFilter
	lastPostLookupPlatform string
	lastPostLookupID      string
	listCompaniesCalled   bool
}

func (f *fakeQueryService) ListJobs(context.Context, repository.ListJobsFilter) ([]repository.CrawlJob, error) {
	return f.jobs, nil
}

func (f *fakeQueryService) GetJob(context.Context, int64) (repository.CrawlJob, error) {
	return f.jobs[0], nil
}

func (f *fakeQueryService) ListPosts(_ context.Context, filter repository.PostFilter) ([]repository.PostDetail, error) {
	f.lastPostFilter = filter
	return f.postList, nil
}

func (f *fakeQueryService) GetPostBySource(_ context.Context, platform, sourcePostID string) (repository.PostDetail, error) {
	f.lastPostLookupPlatform = platform
	f.lastPostLookupID = sourcePostID
	return f.postDetail, nil
}

func (f *fakeQueryService) ListQuestions(_ context.Context, filter repository.QuestionFilter) ([]repository.QuestionRecord, error) {
	f.lastQuestionFilter = filter
	return f.questions, nil
}

func (f *fakeQueryService) ListCompanies(context.Context) ([]repository.CompanySummary, error) {
	f.listCompaniesCalled = true
	return f.companies, nil
}

func (f *fakeQueryService) ListUsageWindows(_ context.Context, filter repository.UsageWindowFilter) (repository.UsageWindowList, error) {
	f.lastUsageWindowFilter = filter
	return f.usageWindows, nil
}
