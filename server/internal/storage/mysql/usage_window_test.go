package mysql

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"on-my-interview/server/internal/llm"
	"on-my-interview/server/internal/storage/repository"
)

func TestRecordUsageWindow(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 3, 11, 3, 30, 0, 0, time.UTC)
	windowStart := time.Date(2026, 3, 10, 16, 0, 0, 0, time.UTC)
	windowEnd := time.Date(2026, 3, 11, 4, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE llm_usage_windows SET status = \\?, finalized_at = \\?, updated_at = \\? WHERE status = \\? AND window_end_at <= \\?").
		WithArgs(repository.UsageWindowStatusFinalized, sqlmock.AnyArg(), sqlmock.AnyArg(), repository.UsageWindowStatusAggregating, now).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO llm_usage_windows").
		WithArgs(
			windowStart,
			windowEnd,
			"Asia/Shanghai",
			repository.UsageWindowStatusAggregating,
			int64(1),
			int64(1),
			int64(230000),
			int64(100000),
			int64(130000),
			int64(50000),
			int64(280000),
			"0.43000000",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := NewRepository(db)
	repo.now = func() time.Time { return now }

	err = repo.RecordUsageWindow(context.Background(), repository.RecordUsageWindowInput{
		Provider:   "openai-compatible",
		Model:      "deepseek-chat",
		RecordedAt: now,
		Usage: llm.Usage{
			PromptTokens:          230000,
			PromptCacheHitTokens:  100000,
			PromptCacheMissTokens: 130000,
			CompletionTokens:      50000,
			TotalTokens:           280000,
		},
	})
	if err != nil {
		t.Fatalf("RecordUsageWindow: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestListUsageWindows(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 3, 11, 12, 1, 0, 0, time.UTC)
	windowStart := time.Date(2026, 3, 10, 16, 0, 0, 0, time.UTC)
	windowEnd := time.Date(2026, 3, 11, 4, 0, 0, 0, time.UTC)
	finalizedAt := time.Date(2026, 3, 11, 4, 0, 1, 0, time.UTC)

	mock.ExpectExec("UPDATE llm_usage_windows SET status = \\?, finalized_at = \\?, updated_at = \\? WHERE status = \\? AND window_end_at <= \\?").
		WithArgs(repository.UsageWindowStatusFinalized, sqlmock.AnyArg(), sqlmock.AnyArg(), repository.UsageWindowStatusAggregating, now).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rows := sqlmock.NewRows([]string{
		"window_start_at",
		"window_end_at",
		"timezone",
		"status",
		"request_count",
		"usage_observed_count",
		"prompt_tokens",
		"prompt_cache_hit_tokens",
		"prompt_cache_miss_tokens",
		"completion_tokens",
		"total_tokens",
		"estimated_cost_cny",
		"finalized_at",
	}).AddRow(
		windowStart,
		windowEnd,
		"Asia/Shanghai",
		repository.UsageWindowStatusFinalized,
		2,
		2,
		460000,
		200000,
		260000,
		100000,
		560000,
		"0.86000000",
		finalizedAt,
	)

	mock.ExpectQuery("SELECT window_start_at, window_end_at, timezone, status, request_count, usage_observed_count, prompt_tokens, prompt_cache_hit_tokens, prompt_cache_miss_tokens, completion_tokens, total_tokens, estimated_cost_cny, finalized_at FROM llm_usage_windows").
		WithArgs(normalizeLimit(10), normalizeOffset(0)).
		WillReturnRows(rows)

	totalRows := sqlmock.NewRows([]string{
		"request_count",
		"usage_observed_count",
		"prompt_tokens",
		"prompt_cache_hit_tokens",
		"prompt_cache_miss_tokens",
		"completion_tokens",
		"total_tokens",
		"estimated_cost_cny",
	}).AddRow(
		2,
		2,
		460000,
		200000,
		260000,
		100000,
		560000,
		"0.86000000",
	)

	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(request_count\\), 0\\), COALESCE\\(SUM\\(usage_observed_count\\), 0\\), COALESCE\\(SUM\\(prompt_tokens\\), 0\\), COALESCE\\(SUM\\(prompt_cache_hit_tokens\\), 0\\), COALESCE\\(SUM\\(prompt_cache_miss_tokens\\), 0\\), COALESCE\\(SUM\\(completion_tokens\\), 0\\), COALESCE\\(SUM\\(total_tokens\\), 0\\), COALESCE\\(CAST\\(SUM\\(estimated_cost_cny\\) AS CHAR\\), '0.00000000'\\) FROM llm_usage_windows").
		WillReturnRows(totalRows)

	repo := NewRepository(db)
	repo.now = func() time.Time { return now }

	result, err := repo.ListUsageWindows(context.Background(), repository.UsageWindowFilter{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListUsageWindows: %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 usage window, got %d", len(result.Items))
	}
	if result.Items[0].EstimatedCostCNY != "0.86000000" {
		t.Fatalf("unexpected item %+v", result.Items[0])
	}
	if result.Totals.EstimatedCostCNY != "0.86000000" {
		t.Fatalf("unexpected totals %+v", result.Totals)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
