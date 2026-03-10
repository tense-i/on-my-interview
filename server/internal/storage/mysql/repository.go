package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"on-my-interview/server/internal/llm"
	"on-my-interview/server/internal/storage/repository"
	"on-my-interview/server/internal/usage"
)

type Repository struct {
	db  *sql.DB
	now func() time.Time
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db:  db,
		now: func() time.Time { return time.Now().UTC() },
	}
}

func (r *Repository) CreateJob(ctx context.Context, params repository.CreateJobParams) (repository.CrawlJob, error) {
	now := r.now()
	platformsJSON, err := json.Marshal(params.Platforms)
	if err != nil {
		return repository.CrawlJob{}, err
	}
	keywordsJSON, err := json.Marshal(params.Keywords)
	if err != nil {
		return repository.CrawlJob{}, err
	}

	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO crawl_jobs
			(trigger_type, status, platforms_json, keywords_json, pages, force_reparse, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		params.TriggerType,
		repository.JobStatusPending,
		string(platformsJSON),
		string(keywordsJSON),
		params.Pages,
		params.ForceReparse,
		now,
		now,
	)
	if err != nil {
		return repository.CrawlJob{}, err
	}

	jobID, _ := result.LastInsertId()
	return repository.CrawlJob{
		ID:           jobID,
		TriggerType:  params.TriggerType,
		Status:       repository.JobStatusPending,
		Platforms:    append([]string(nil), params.Platforms...),
		Keywords:     append([]string(nil), params.Keywords...),
		Pages:        params.Pages,
		ForceReparse: params.ForceReparse,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (r *Repository) GetJob(ctx context.Context, jobID int64) (repository.CrawlJob, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, trigger_type, status, platforms_json, keywords_json, pages, force_reparse, stats_json, error_message, started_at, finished_at, created_at, updated_at
		 FROM crawl_jobs WHERE id = ?`,
		jobID,
	)

	job, err := scanJob(row)
	if err != nil {
		return repository.CrawlJob{}, err
	}
	return job, nil
}

func (r *Repository) UpdateJob(ctx context.Context, job repository.CrawlJob) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE crawl_jobs
		    SET status = ?, pages = ?, force_reparse = ?, stats_json = ?, error_message = ?, started_at = ?, finished_at = ?, updated_at = ?
		  WHERE id = ?`,
		job.Status,
		job.Pages,
		job.ForceReparse,
		nullIfEmpty(job.StatsJSON),
		nullIfEmpty(job.ErrorMessage),
		job.StartedAt,
		job.FinishedAt,
		r.now(),
		job.ID,
	)
	return err
}

func (r *Repository) UpsertRawPost(ctx context.Context, input repository.RawPostInput) (repository.UpsertRawPostResult, error) {
	existing, err := r.findRawPostBySource(ctx, input.Platform, input.SourcePostID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return repository.UpsertRawPostResult{}, err
	}

	if errors.Is(err, sql.ErrNoRows) {
		post := repository.RawPost{
			Platform:        input.Platform,
			SourcePostID:    input.SourcePostID,
			Title:           input.Title,
			Content:         input.Content,
			ContentHash:     input.ContentHash,
			AuthorName:      input.AuthorName,
			PostURL:         input.PostURL,
			CompanyNameRaw:  input.CompanyNameRaw,
			CompanyNameNorm: input.CompanyNameNorm,
			SourceCreatedAt: input.SourceCreatedAt,
			SourceEditedAt:  input.SourceEditedAt,
			LastCrawledAt:   r.now(),
			ParseStatus:     repository.ParseStatusPending,
			ParseAttempts:   0,
			RawPayloadJSON:  input.RawPayloadJSON,
		}

		result, insertErr := r.db.ExecContext(
			ctx,
			`INSERT INTO raw_posts
				(platform, source_post_id, title, content, content_hash, author_name, post_url, source_created_at, source_edited_at, last_crawled_at, parse_status, parse_attempts, raw_payload_json)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			post.Platform,
			post.SourcePostID,
			post.Title,
			post.Content,
			post.ContentHash,
			post.AuthorName,
			post.PostURL,
			post.SourceCreatedAt,
			post.SourceEditedAt,
			post.LastCrawledAt,
			post.ParseStatus,
			post.ParseAttempts,
			post.RawPayloadJSON,
		)
		if insertErr != nil {
			return repository.UpsertRawPostResult{}, insertErr
		}

		post.ID, _ = result.LastInsertId()
		return repository.UpsertRawPostResult{
			Disposition: repository.UpsertDispositionInserted,
			Post:        post,
		}, nil
	}

	if existing.SourceEditedAt.Equal(input.SourceEditedAt) && existing.ContentHash == input.ContentHash {
		return repository.UpsertRawPostResult{
			Disposition: repository.UpsertDispositionUnchanged,
			Post:        existing,
		}, nil
	}

	existing.Title = input.Title
	existing.Content = input.Content
	existing.ContentHash = input.ContentHash
	existing.AuthorName = input.AuthorName
	existing.PostURL = input.PostURL
	existing.CompanyNameRaw = input.CompanyNameRaw
	existing.CompanyNameNorm = input.CompanyNameNorm
	existing.SourceCreatedAt = input.SourceCreatedAt
	existing.SourceEditedAt = input.SourceEditedAt
	existing.LastCrawledAt = r.now()
	existing.ParseStatus = repository.ParseStatusPending
	existing.RawPayloadJSON = input.RawPayloadJSON

	_, err = r.db.ExecContext(
		ctx,
		`UPDATE raw_posts SET
			title = ?,
			content = ?,
			content_hash = ?,
			author_name = ?,
			post_url = ?,
			company_name_raw = ?,
			company_name_norm = ?,
			source_created_at = ?,
			source_edited_at = ?,
			last_crawled_at = ?,
			parse_status = ?,
			raw_payload_json = ?
		  WHERE id = ?`,
		existing.Title,
		existing.Content,
		existing.ContentHash,
		existing.AuthorName,
		existing.PostURL,
		existing.CompanyNameRaw,
		existing.CompanyNameNorm,
		existing.SourceCreatedAt,
		existing.SourceEditedAt,
		existing.LastCrawledAt,
		existing.ParseStatus,
		existing.RawPayloadJSON,
		existing.ID,
	)
	if err != nil {
		return repository.UpsertRawPostResult{}, err
	}

	return repository.UpsertRawPostResult{
		Disposition: repository.UpsertDispositionUpdated,
		Post:        existing,
	}, nil
}

func (r *Repository) SaveParseResult(ctx context.Context, input repository.PostParseResultInput) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO post_parse_results
			(raw_post_id, schema_version, llm_provider, llm_model, company_name, sentiment, sentiment_reason, key_events_json, questions_json, tags_json, raw_json, parsed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
			schema_version = VALUES(schema_version),
			llm_provider = VALUES(llm_provider),
			llm_model = VALUES(llm_model),
			company_name = VALUES(company_name),
			sentiment = VALUES(sentiment),
			sentiment_reason = VALUES(sentiment_reason),
			key_events_json = VALUES(key_events_json),
			questions_json = VALUES(questions_json),
			tags_json = VALUES(tags_json),
			raw_json = VALUES(raw_json),
			parsed_at = VALUES(parsed_at)`,
		input.RawPostID,
		input.SchemaVersion,
		input.LLMProvider,
		input.LLMModel,
		nullIfEmpty(input.CompanyName),
		nullIfEmpty(input.Sentiment),
		nullIfEmpty(input.SentimentReason),
		input.KeyEventsJSON,
		input.QuestionsJSON,
		input.TagsJSON,
		input.RawJSON,
		input.ParsedAt,
	)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(
		ctx,
		`UPDATE raw_posts
		    SET parse_status = ?, company_name_norm = ?, parse_attempts = parse_attempts + 1
		  WHERE id = ?`,
		repository.ParseStatusParsed,
		nullIfEmpty(input.CompanyName),
		input.RawPostID,
	)
	return err
}

func (r *Repository) ReplaceQuestions(ctx context.Context, rawPostID int64, questions []repository.InterviewQuestionInput) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `DELETE qt FROM question_tags qt JOIN interview_questions iq ON iq.id = qt.question_id WHERE iq.raw_post_id = ?`, rawPostID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM interview_questions WHERE raw_post_id = ?`, rawPostID); err != nil {
		return err
	}

	for _, question := range questions {
		result, execErr := tx.ExecContext(
			ctx,
			`INSERT INTO interview_questions
				(raw_post_id, platform, company_name, question_text, question_order, category, tags_json, source_excerpt)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			question.RawPostID,
			question.Platform,
			nullIfEmpty(question.CompanyName),
			question.QuestionText,
			question.QuestionOrder,
			nullIfEmpty(question.Category),
			question.TagsJSON,
			question.SourceExcerpt,
		)
		if execErr != nil {
			err = execErr
			return err
		}

		questionID, _ := result.LastInsertId()
		var tags []string
		if unmarshalErr := json.Unmarshal([]byte(question.TagsJSON), &tags); unmarshalErr != nil {
			err = unmarshalErr
			return err
		}
		for _, tag := range tags {
			if _, execErr = tx.ExecContext(ctx, `INSERT INTO question_tags (question_id, tag) VALUES (?, ?)`, questionID, tag); execErr != nil {
				err = execErr
				return err
			}
		}
	}

	err = tx.Commit()
	return err
}

func (r *Repository) MarkParseFailed(ctx context.Context, rawPostID int64, _ string) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE raw_posts
		    SET parse_status = ?, parse_attempts = parse_attempts + 1
		  WHERE id = ?`,
		repository.ParseStatusFailed,
		rawPostID,
	)
	return err
}

func (r *Repository) RecordJobPost(ctx context.Context, record repository.JobPostRecord) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO crawl_job_posts (job_id, raw_post_id, disposition, message) VALUES (?, ?, ?, ?)`,
		record.JobID,
		record.RawPostID,
		record.Disposition,
		nullIfEmpty(record.Message),
	)
	return err
}

func (r *Repository) RecordUsageWindow(ctx context.Context, input repository.RecordUsageWindowInput) (err error) {
	window, err := usage.WindowForTime(input.RecordedAt)
	if err != nil {
		return err
	}

	cost, err := calculateEstimatedCostCNY(input.Model, input.Usage)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = finalizeExpiredUsageWindows(ctx, tx, input.RecordedAt, r.now()); err != nil {
		return err
	}

	now := r.now()
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO llm_usage_windows
			(window_start_at, window_end_at, timezone, status, request_count, usage_observed_count, prompt_tokens, prompt_cache_hit_tokens, prompt_cache_miss_tokens, completion_tokens, total_tokens, estimated_cost_cny, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
			status = VALUES(status),
			request_count = request_count + VALUES(request_count),
			usage_observed_count = usage_observed_count + VALUES(usage_observed_count),
			prompt_tokens = prompt_tokens + VALUES(prompt_tokens),
			prompt_cache_hit_tokens = prompt_cache_hit_tokens + VALUES(prompt_cache_hit_tokens),
			prompt_cache_miss_tokens = prompt_cache_miss_tokens + VALUES(prompt_cache_miss_tokens),
			completion_tokens = completion_tokens + VALUES(completion_tokens),
			total_tokens = total_tokens + VALUES(total_tokens),
			estimated_cost_cny = CAST(estimated_cost_cny + VALUES(estimated_cost_cny) AS DECIMAL(20,8)),
			finalized_at = NULL,
			updated_at = VALUES(updated_at)`,
		window.StartAt,
		window.EndAt,
		window.Timezone,
		repository.UsageWindowStatusAggregating,
		int64(1),
		int64(1),
		input.Usage.PromptTokens,
		input.Usage.PromptCacheHitTokens,
		input.Usage.PromptCacheMissTokens,
		input.Usage.CompletionTokens,
		input.Usage.TotalTokens,
		cost,
		now,
		now,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) ListJobs(ctx context.Context, filter repository.ListJobsFilter) ([]repository.CrawlJob, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, trigger_type, status, platforms_json, keywords_json, pages, force_reparse, stats_json, error_message, started_at, finished_at, created_at, updated_at
		 FROM crawl_jobs
		 ORDER BY id DESC
		 LIMIT ? OFFSET ?`,
		normalizeLimit(filter.Limit),
		normalizeOffset(filter.Offset),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]repository.CrawlJob, 0)
	for rows.Next() {
		job, scanErr := scanJob(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

func (r *Repository) ListPosts(ctx context.Context, filter repository.PostFilter) ([]repository.PostDetail, error) {
	query := strings.Builder{}
	query.WriteString(`SELECT rp.id, rp.platform, rp.source_post_id, rp.title, rp.content, rp.content_hash, rp.author_name, rp.post_url, rp.company_name_raw, rp.company_name_norm, rp.source_created_at, rp.source_edited_at, rp.last_crawled_at, rp.parse_status, rp.parse_attempts, rp.raw_payload_json,
		ppr.raw_post_id, ppr.schema_version, ppr.llm_provider, ppr.llm_model, ppr.company_name, ppr.sentiment, ppr.sentiment_reason, ppr.key_events_json, ppr.questions_json, ppr.tags_json, ppr.raw_json, ppr.parsed_at
		FROM raw_posts rp
		LEFT JOIN post_parse_results ppr ON ppr.raw_post_id = rp.id
		WHERE 1=1`)

	args := make([]any, 0, 6)
	if filter.Platform != "" {
		query.WriteString(` AND rp.platform = ?`)
		args = append(args, filter.Platform)
	}
	if filter.Company != "" {
		query.WriteString(` AND COALESCE(NULLIF(ppr.company_name, ''), NULLIF(rp.company_name_norm, ''), NULLIF(rp.company_name_raw, '')) = ?`)
		args = append(args, filter.Company)
	}
	if filter.Tag != "" {
		query.WriteString(` AND EXISTS (SELECT 1 FROM interview_questions iq WHERE iq.raw_post_id = rp.id AND JSON_SEARCH(iq.tags_json, 'one', ?) IS NOT NULL)`)
		args = append(args, filter.Tag)
	}
	query.WriteString(` ORDER BY rp.source_edited_at DESC LIMIT ? OFFSET ?`)
	args = append(args, normalizeLimit(filter.Limit), normalizeOffset(filter.Offset))

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]repository.PostDetail, 0)
	for rows.Next() {
		post, scanErr := scanPostDetail(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		posts = append(posts, post)
	}
	return posts, rows.Err()
}

func (r *Repository) GetPostBySource(ctx context.Context, platform, sourcePostID string) (repository.PostDetail, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT rp.id, rp.platform, rp.source_post_id, rp.title, rp.content, rp.content_hash, rp.author_name, rp.post_url, rp.company_name_raw, rp.company_name_norm, rp.source_created_at, rp.source_edited_at, rp.last_crawled_at, rp.parse_status, rp.parse_attempts, rp.raw_payload_json,
			ppr.raw_post_id, ppr.schema_version, ppr.llm_provider, ppr.llm_model, ppr.company_name, ppr.sentiment, ppr.sentiment_reason, ppr.key_events_json, ppr.questions_json, ppr.tags_json, ppr.raw_json, ppr.parsed_at
		   FROM raw_posts rp
		   LEFT JOIN post_parse_results ppr ON ppr.raw_post_id = rp.id
		  WHERE rp.platform = ? AND rp.source_post_id = ?`,
		platform,
		sourcePostID,
	)
	return scanPostDetail(row)
}

func (r *Repository) ListQuestions(ctx context.Context, filter repository.QuestionFilter) ([]repository.QuestionRecord, error) {
	query := strings.Builder{}
	query.WriteString(`SELECT raw_post_id, platform, company_name, question_text, question_order, category, tags_json, source_excerpt
		FROM interview_questions
		WHERE 1=1`)

	args := make([]any, 0, 5)
	if filter.Platform != "" {
		query.WriteString(` AND platform = ?`)
		args = append(args, filter.Platform)
	}
	if filter.Company != "" {
		query.WriteString(` AND company_name = ?`)
		args = append(args, filter.Company)
	}
	if filter.Tag != "" {
		query.WriteString(` AND JSON_SEARCH(tags_json, 'one', ?) IS NOT NULL`)
		args = append(args, filter.Tag)
	}
	if filter.Query != "" {
		query.WriteString(` AND question_text LIKE ?`)
		args = append(args, "%"+filter.Query+"%")
	}
	query.WriteString(` ORDER BY raw_post_id DESC, question_order ASC LIMIT ? OFFSET ?`)
	args = append(args, normalizeLimit(filter.Limit), normalizeOffset(filter.Offset))

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	questions := make([]repository.QuestionRecord, 0)
	for rows.Next() {
		var (
			record  repository.QuestionRecord
			tagJSON string
		)
		if err := rows.Scan(
			&record.RawPostID,
			&record.Platform,
			&record.CompanyName,
			&record.QuestionText,
			&record.QuestionOrder,
			&record.Category,
			&tagJSON,
			&record.SourceExcerpt,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(tagJSON), &record.Tags); err != nil {
			return nil, err
		}
		questions = append(questions, record)
	}
	return questions, rows.Err()
}

func (r *Repository) ListCompanies(ctx context.Context) ([]repository.CompanySummary, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT company_name, COUNT(DISTINCT raw_post_id) AS post_count, COUNT(*) AS question_count
		   FROM interview_questions
		  WHERE company_name IS NOT NULL AND company_name <> ''
		  GROUP BY company_name
		  ORDER BY question_count DESC, post_count DESC, company_name ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	companies := make([]repository.CompanySummary, 0)
	for rows.Next() {
		var company repository.CompanySummary
		if err := rows.Scan(&company.CompanyName, &company.PostCount, &company.QuestionCount); err != nil {
			return nil, err
		}
		companies = append(companies, company)
	}
	return companies, rows.Err()
}

func (r *Repository) ListUsageWindows(ctx context.Context, filter repository.UsageWindowFilter) (repository.UsageWindowList, error) {
	if err := finalizeExpiredUsageWindows(ctx, r.db, r.now(), r.now()); err != nil {
		return repository.UsageWindowList{}, err
	}

	query := strings.Builder{}
	query.WriteString(`SELECT window_start_at, window_end_at, timezone, status, request_count, usage_observed_count, prompt_tokens, prompt_cache_hit_tokens, prompt_cache_miss_tokens, completion_tokens, total_tokens, estimated_cost_cny, finalized_at
		FROM llm_usage_windows
		WHERE status = 'finalized'`)
	args := make([]any, 0, 4)
	if filter.From != nil {
		query.WriteString(` AND window_start_at >= ?`)
		args = append(args, *filter.From)
	}
	if filter.To != nil {
		query.WriteString(` AND window_end_at <= ?`)
		args = append(args, *filter.To)
	}
	query.WriteString(` ORDER BY window_start_at DESC LIMIT ? OFFSET ?`)
	args = append(args, normalizeLimit(filter.Limit), normalizeOffset(filter.Offset))

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return repository.UsageWindowList{}, err
	}
	defer rows.Close()

	items := make([]repository.UsageWindow, 0)
	for rows.Next() {
		item, scanErr := scanUsageWindow(rows)
		if scanErr != nil {
			return repository.UsageWindowList{}, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return repository.UsageWindowList{}, err
	}

	totalQuery := strings.Builder{}
	totalQuery.WriteString(`SELECT COALESCE(SUM(request_count), 0), COALESCE(SUM(usage_observed_count), 0), COALESCE(SUM(prompt_tokens), 0), COALESCE(SUM(prompt_cache_hit_tokens), 0), COALESCE(SUM(prompt_cache_miss_tokens), 0), COALESCE(SUM(completion_tokens), 0), COALESCE(SUM(total_tokens), 0), COALESCE(CAST(SUM(estimated_cost_cny) AS CHAR), '0.00000000')
		FROM llm_usage_windows
		WHERE status = 'finalized'`)
	totalArgs := make([]any, 0, 2)
	if filter.From != nil {
		totalQuery.WriteString(` AND window_start_at >= ?`)
		totalArgs = append(totalArgs, *filter.From)
	}
	if filter.To != nil {
		totalQuery.WriteString(` AND window_end_at <= ?`)
		totalArgs = append(totalArgs, *filter.To)
	}

	var totals repository.UsageWindowTotals
	if err := r.db.QueryRowContext(ctx, totalQuery.String(), totalArgs...).Scan(
		&totals.RequestCount,
		&totals.UsageObservedCount,
		&totals.PromptTokens,
		&totals.PromptCacheHitTokens,
		&totals.PromptCacheMissTokens,
		&totals.CompletionTokens,
		&totals.TotalTokens,
		&totals.EstimatedCostCNY,
	); err != nil {
		return repository.UsageWindowList{}, err
	}

	return repository.UsageWindowList{
		Items:  items,
		Totals: totals,
	}, nil
}

func (r *Repository) findRawPostBySource(ctx context.Context, platform, sourcePostID string) (repository.RawPost, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, platform, source_post_id, title, content, content_hash, author_name, post_url, company_name_raw, company_name_norm, source_created_at, source_edited_at, last_crawled_at, parse_status, parse_attempts, raw_payload_json
		 FROM raw_posts WHERE platform = ? AND source_post_id = ?`,
		platform,
		sourcePostID,
	)
	return scanRawPost(row)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanRawPost(row scanner) (repository.RawPost, error) {
	var (
		post            repository.RawPost
		companyNameRaw  sql.NullString
		companyNameNorm sql.NullString
	)
	err := row.Scan(
		&post.ID,
		&post.Platform,
		&post.SourcePostID,
		&post.Title,
		&post.Content,
		&post.ContentHash,
		&post.AuthorName,
		&post.PostURL,
		&companyNameRaw,
		&companyNameNorm,
		&post.SourceCreatedAt,
		&post.SourceEditedAt,
		&post.LastCrawledAt,
		&post.ParseStatus,
		&post.ParseAttempts,
		&post.RawPayloadJSON,
	)
	if err != nil {
		return repository.RawPost{}, err
	}
	post.CompanyNameRaw = companyNameRaw.String
	post.CompanyNameNorm = companyNameNorm.String
	return post, nil
}

func scanJob(row scanner) (repository.CrawlJob, error) {
	var (
		job           repository.CrawlJob
		platformsJSON string
		keywordsJSON  string
		statsJSON     sql.NullString
		errorMessage  sql.NullString
		startedAt     sql.NullTime
		finishedAt    sql.NullTime
	)
	if err := row.Scan(
		&job.ID,
		&job.TriggerType,
		&job.Status,
		&platformsJSON,
		&keywordsJSON,
		&job.Pages,
		&job.ForceReparse,
		&statsJSON,
		&errorMessage,
		&startedAt,
		&finishedAt,
		&job.CreatedAt,
		&job.UpdatedAt,
	); err != nil {
		return repository.CrawlJob{}, err
	}
	if err := json.Unmarshal([]byte(platformsJSON), &job.Platforms); err != nil {
		return repository.CrawlJob{}, err
	}
	if err := json.Unmarshal([]byte(keywordsJSON), &job.Keywords); err != nil {
		return repository.CrawlJob{}, err
	}
	job.StatsJSON = statsJSON.String
	job.ErrorMessage = errorMessage.String
	if startedAt.Valid {
		value := startedAt.Time
		job.StartedAt = &value
	}
	if finishedAt.Valid {
		value := finishedAt.Time
		job.FinishedAt = &value
	}
	return job, nil
}

func scanPostDetail(row scanner) (repository.PostDetail, error) {
	post, err := scanRawPostWithParse(row)
	if err != nil {
		return repository.PostDetail{}, err
	}
	return post, nil
}

func scanRawPostWithParse(row scanner) (repository.PostDetail, error) {
	var (
		detail          repository.PostDetail
		companyNameRaw  sql.NullString
		companyNameNorm sql.NullString
		parseRawPostID  sql.NullInt64
		schemaVersion   sql.NullString
		llmProvider     sql.NullString
		llmModel        sql.NullString
		companyName     sql.NullString
		sentiment       sql.NullString
		sentimentReason sql.NullString
		keyEventsJSON   sql.NullString
		questionsJSON   sql.NullString
		tagsJSON        sql.NullString
		rawJSON         sql.NullString
		parsedAt        sql.NullTime
	)

	if err := row.Scan(
		&detail.RawPost.ID,
		&detail.RawPost.Platform,
		&detail.RawPost.SourcePostID,
		&detail.RawPost.Title,
		&detail.RawPost.Content,
		&detail.RawPost.ContentHash,
		&detail.RawPost.AuthorName,
		&detail.RawPost.PostURL,
		&companyNameRaw,
		&companyNameNorm,
		&detail.RawPost.SourceCreatedAt,
		&detail.RawPost.SourceEditedAt,
		&detail.RawPost.LastCrawledAt,
		&detail.RawPost.ParseStatus,
		&detail.RawPost.ParseAttempts,
		&detail.RawPost.RawPayloadJSON,
		&parseRawPostID,
		&schemaVersion,
		&llmProvider,
		&llmModel,
		&companyName,
		&sentiment,
		&sentimentReason,
		&keyEventsJSON,
		&questionsJSON,
		&tagsJSON,
		&rawJSON,
		&parsedAt,
	); err != nil {
		return repository.PostDetail{}, err
	}
	detail.RawPost.CompanyNameRaw = companyNameRaw.String
	detail.RawPost.CompanyNameNorm = companyNameNorm.String

	if parseRawPostID.Valid {
		detail.ParseResult = &repository.PostParseResult{
			RawPostID:       parseRawPostID.Int64,
			SchemaVersion:   schemaVersion.String,
			LLMProvider:     llmProvider.String,
			LLMModel:        llmModel.String,
			CompanyName:     companyName.String,
			Sentiment:       sentiment.String,
			SentimentReason: sentimentReason.String,
			KeyEventsJSON:   keyEventsJSON.String,
			QuestionsJSON:   questionsJSON.String,
			TagsJSON:        tagsJSON.String,
			RawJSON:         rawJSON.String,
		}
		if parsedAt.Valid {
			detail.ParseResult.ParsedAt = parsedAt.Time
		}
	}

	return detail, nil
}

func scanUsageWindow(row scanner) (repository.UsageWindow, error) {
	var (
		item        repository.UsageWindow
		cost        sql.NullString
		finalizedAt sql.NullTime
	)
	if err := row.Scan(
		&item.WindowStartAt,
		&item.WindowEndAt,
		&item.Timezone,
		&item.Status,
		&item.RequestCount,
		&item.UsageObservedCount,
		&item.PromptTokens,
		&item.PromptCacheHitTokens,
		&item.PromptCacheMissTokens,
		&item.CompletionTokens,
		&item.TotalTokens,
		&cost,
		&finalizedAt,
	); err != nil {
		return repository.UsageWindow{}, err
	}
	item.EstimatedCostCNY = cost.String
	if finalizedAt.Valid {
		value := finalizedAt.Time
		item.FinalizedAt = &value
	}
	return item, nil
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	return limit
}

func normalizeOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}

func nullIfEmpty(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

type execContext interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func finalizeExpiredUsageWindows(ctx context.Context, runner execContext, cutoff, finalizedAt time.Time) error {
	_, err := runner.ExecContext(
		ctx,
		`UPDATE llm_usage_windows SET status = ?, finalized_at = ?, updated_at = ? WHERE status = ? AND window_end_at <= ?`,
		repository.UsageWindowStatusFinalized,
		finalizedAt,
		finalizedAt,
		repository.UsageWindowStatusAggregating,
		cutoff,
	)
	return err
}

func calculateEstimatedCostCNY(model string, observed llm.Usage) (string, error) {
	switch strings.TrimSpace(model) {
	case "deepseek-chat":
		return usage.CalculateDeepSeekChatCostCNY(observed)
	default:
		return "", errors.New("unsupported llm model for usage cost aggregation")
	}
}

var _ repository.Store = (*Repository)(nil)
