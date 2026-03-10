package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"on-my-interview/server/internal/crawler"
	"on-my-interview/server/internal/llm"
	"on-my-interview/server/internal/storage/repository"
)

type Service struct {
	store     repository.Store
	registry  *crawler.Registry
	extractor llm.Extractor
	llmModel  string
	now       func() time.Time
	queue     chan int64
	startOnce sync.Once
}

func NewService(store repository.Store, registry *crawler.Registry, extractor llm.Extractor, llmModel string) *Service {
	return &Service{
		store:     store,
		registry:  registry,
		extractor: extractor,
		llmModel:  llmModel,
		now:       func() time.Time { return time.Now().UTC() },
		queue:     make(chan int64, 128),
	}
}

func (s *Service) CreateJob(ctx context.Context, params repository.CreateJobParams) (repository.CrawlJob, error) {
	if params.Pages <= 0 {
		params.Pages = 1
	}
	if len(params.Platforms) == 0 {
		return repository.CrawlJob{}, fmt.Errorf("at least one platform is required")
	}
	job, err := s.store.CreateJob(ctx, params)
	if err != nil {
		return repository.CrawlJob{}, err
	}
	s.enqueue(job.ID)
	return job, nil
}

func (s *Service) Start(ctx context.Context) {
	s.startOnce.Do(func() {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case jobID := <-s.queue:
					_ = s.RunJob(ctx, jobID)
				}
			}
		}()
	})
}

func (s *Service) RunJob(ctx context.Context, jobID int64) error {
	job, err := s.store.GetJob(ctx, jobID)
	if err != nil {
		return err
	}

	startedAt := s.now()
	job.Status = repository.JobStatusRunning
	job.StartedAt = &startedAt
	job.ErrorMessage = ""
	if err := s.store.UpdateJob(ctx, job); err != nil {
		return err
	}

	stats := runStats{}
	for _, platform := range job.Platforms {
		platformCrawler, err := s.registry.Get(platform)
		if err != nil {
			return s.failJob(ctx, job, err)
		}
		for _, keyword := range job.Keywords {
			for page := 1; page <= job.Pages; page++ {
				posts, err := platformCrawler.SearchPosts(ctx, crawler.SearchRequest{
					Keyword: keyword,
					Page:    page,
				})
				if err != nil {
					return s.failJob(ctx, job, err)
				}
				for _, crawledPost := range posts {
					if err := s.processPost(ctx, job, crawledPost, &stats); err != nil {
						return s.failJob(ctx, job, err)
					}
				}
			}
		}
	}

	statsJSON, err := json.Marshal(stats)
	if err != nil {
		return s.failJob(ctx, job, err)
	}

	finishedAt := s.now()
	job.Status = repository.JobStatusCompleted
	job.FinishedAt = &finishedAt
	job.StatsJSON = string(statsJSON)
	return s.store.UpdateJob(ctx, job)
}

func (s *Service) processPost(ctx context.Context, job repository.CrawlJob, crawledPost crawler.RawPostInput, stats *runStats) error {
	upsertResult, err := s.store.UpsertRawPost(ctx, repository.RawPostInput{
		Platform:        crawledPost.Platform,
		SourcePostID:    crawledPost.SourcePostID,
		Title:           crawledPost.Title,
		Content:         crawledPost.Content,
		ContentHash:     crawledPost.ContentHash,
		AuthorName:      crawledPost.AuthorName,
		PostURL:         crawledPost.PostURL,
		CompanyNameRaw:  crawledPost.CompanyNameRaw,
		CompanyNameNorm: crawledPost.CompanyNameNorm,
		SourceCreatedAt: crawledPost.SourceCreatedAt,
		SourceEditedAt:  crawledPost.SourceEditedAt,
		RawPayloadJSON:  crawledPost.RawPayloadJSON,
	})
	if err != nil {
		return err
	}

	stats.Total++
	stats.bumpDisposition(upsertResult.Disposition)

	shouldParse := upsertResult.Disposition != repository.UpsertDispositionUnchanged || job.ForceReparse
	if !shouldParse {
		return s.store.RecordJobPost(ctx, repository.JobPostRecord{
			JobID:       job.ID,
			RawPostID:   upsertResult.Post.ID,
			Disposition: string(upsertResult.Disposition),
		})
	}

	structured, err := s.extractor.Extract(ctx, llm.RawPostForLLM{
		Platform:     crawledPost.Platform,
		SourcePostID: crawledPost.SourcePostID,
		Title:        crawledPost.Title,
		Content:      crawledPost.Content,
		PostURL:      crawledPost.PostURL,
	})
	if err != nil {
		stats.ParseFailed++
		if markErr := s.store.MarkParseFailed(ctx, upsertResult.Post.ID, err.Error()); markErr != nil {
			return markErr
		}
		return s.store.RecordJobPost(ctx, repository.JobPostRecord{
			JobID:       job.ID,
			RawPostID:   upsertResult.Post.ID,
			Disposition: "parse_failed",
			Message:     err.Error(),
		})
	}

	parseResult, questions, err := s.buildParseInputs(upsertResult.Post.ID, crawledPost.Platform, structured)
	if err != nil {
		return err
	}

	if err := s.store.SaveParseResult(ctx, parseResult); err != nil {
		return err
	}
	if err := s.store.ReplaceQuestions(ctx, upsertResult.Post.ID, questions); err != nil {
		return err
	}
	if structured.Usage != nil {
		if err := s.store.RecordUsageWindow(ctx, repository.RecordUsageWindowInput{
			Provider:   s.extractor.Name(),
			Model:      s.llmModel,
			RecordedAt: s.now(),
			Usage:      *structured.Usage,
		}); err != nil {
			return err
		}
	}
	stats.Parsed++
	return s.store.RecordJobPost(ctx, repository.JobPostRecord{
		JobID:       job.ID,
		RawPostID:   upsertResult.Post.ID,
		Disposition: string(upsertResult.Disposition),
	})
}

func (s *Service) ReparsePost(ctx context.Context, params repository.ReparsePostParams) error {
	postDetail, err := s.store.GetPostBySource(ctx, params.Platform, params.SourcePostID)
	if err != nil {
		return err
	}
	if s.extractor == nil {
		return fmt.Errorf("extractor is not configured")
	}

	structured, err := s.extractor.Extract(ctx, llm.RawPostForLLM{
		Platform:     postDetail.RawPost.Platform,
		SourcePostID: postDetail.RawPost.SourcePostID,
		Title:        postDetail.RawPost.Title,
		Content:      postDetail.RawPost.Content,
		PostURL:      postDetail.RawPost.PostURL,
	})
	if err != nil {
		return s.store.MarkParseFailed(ctx, postDetail.RawPost.ID, err.Error())
	}

	parseResult, questions, err := s.buildParseInputs(postDetail.RawPost.ID, postDetail.RawPost.Platform, structured)
	if err != nil {
		return err
	}
	if err := s.store.SaveParseResult(ctx, parseResult); err != nil {
		return err
	}
	if structured.Usage != nil {
		if err := s.store.RecordUsageWindow(ctx, repository.RecordUsageWindowInput{
			Provider:   s.extractor.Name(),
			Model:      s.llmModel,
			RecordedAt: s.now(),
			Usage:      *structured.Usage,
		}); err != nil {
			return err
		}
	}
	return s.store.ReplaceQuestions(ctx, postDetail.RawPost.ID, questions)
}

func (s *Service) buildParseInputs(rawPostID int64, platform string, structured *llm.StructuredPost) (repository.PostParseResultInput, []repository.InterviewQuestionInput, error) {
	keyEventsJSON, err := json.Marshal(structured.KeyEvents)
	if err != nil {
		return repository.PostParseResultInput{}, nil, err
	}
	questionsJSON, err := json.Marshal(structured.Questions)
	if err != nil {
		return repository.PostParseResultInput{}, nil, err
	}
	tagsJSON, err := json.Marshal(structured.OverallTags)
	if err != nil {
		return repository.PostParseResultInput{}, nil, err
	}
	rawJSON, err := json.Marshal(structured)
	if err != nil {
		return repository.PostParseResultInput{}, nil, err
	}

	rows := llm.FlattenQuestions(rawPostID, platform, structured.Company.NormalizedName, structured.Questions)
	questionInputs := make([]repository.InterviewQuestionInput, 0, len(rows))
	for _, row := range rows {
		rowTagsJSON, err := json.Marshal(row.Tags)
		if err != nil {
			return repository.PostParseResultInput{}, nil, err
		}
		questionInputs = append(questionInputs, repository.InterviewQuestionInput{
			RawPostID:     row.RawPostID,
			Platform:      row.Platform,
			CompanyName:   row.CompanyName,
			QuestionText:  row.QuestionText,
			QuestionOrder: row.QuestionOrder,
			Category:      row.Category,
			TagsJSON:      string(rowTagsJSON),
			SourceExcerpt: row.SourceExcerpt,
		})
	}

	return repository.PostParseResultInput{
		RawPostID:       rawPostID,
		SchemaVersion:   structured.SchemaVersion,
		LLMProvider:     s.extractor.Name(),
		LLMModel:        s.llmModel,
		CompanyName:     structured.Company.NormalizedName,
		Sentiment:       structured.Sentiment.Label,
		SentimentReason: structured.Sentiment.Reason,
		KeyEventsJSON:   string(keyEventsJSON),
		QuestionsJSON:   string(questionsJSON),
		TagsJSON:        string(tagsJSON),
		RawJSON:         string(rawJSON),
		ParsedAt:        s.now(),
	}, questionInputs, nil
}

func (s *Service) failJob(ctx context.Context, job repository.CrawlJob, runErr error) error {
	finishedAt := s.now()
	job.Status = repository.JobStatusFailed
	job.FinishedAt = &finishedAt
	job.ErrorMessage = runErr.Error()
	if err := s.store.UpdateJob(ctx, job); err != nil {
		return err
	}
	return runErr
}

type runStats struct {
	Total       int `json:"total"`
	Inserted    int `json:"inserted"`
	Updated     int `json:"updated"`
	Unchanged   int `json:"unchanged"`
	Parsed      int `json:"parsed"`
	ParseFailed int `json:"parse_failed"`
}

func (s *runStats) bumpDisposition(disposition repository.UpsertDisposition) {
	switch disposition {
	case repository.UpsertDispositionInserted:
		s.Inserted++
	case repository.UpsertDispositionUpdated:
		s.Updated++
	case repository.UpsertDispositionUnchanged:
		s.Unchanged++
	}
}

func (s *Service) enqueue(jobID int64) {
	select {
	case s.queue <- jobID:
	default:
		go func() {
			s.queue <- jobID
		}()
	}
}
