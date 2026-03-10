package mysql

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"on-my-interview/server/internal/storage/repository"
)

func TestUpsertRawPost(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	editedAt := createdAt.Add(10 * time.Minute)

	tests := []struct {
		name        string
		existing    *repository.RawPost
		want        repository.UpsertDisposition
		updatedHash string
	}{
		{
			name:     "insert new post",
			existing: nil,
			want:     repository.UpsertDispositionInserted,
		},
		{
			name: "keep unchanged post",
			existing: &repository.RawPost{
				ID:              11,
				Platform:        "nowcoder",
				SourcePostID:    "861032210369871872",
				Title:           "阿里面经",
				Content:         "问了 Kafka 和 MQ",
				ContentHash:     "hash-1",
				AuthorName:      "alice",
				PostURL:         "https://www.nowcoder.com/discuss/861032210369871872",
				SourceCreatedAt: createdAt,
				SourceEditedAt:  editedAt,
			},
			want: repository.UpsertDispositionUnchanged,
		},
		{
			name: "update changed post",
			existing: &repository.RawPost{
				ID:              12,
				Platform:        "nowcoder",
				SourcePostID:    "861032210369871872",
				Title:           "阿里面经",
				Content:         "旧内容",
				ContentHash:     "hash-old",
				AuthorName:      "alice",
				PostURL:         "https://www.nowcoder.com/discuss/861032210369871872",
				SourceCreatedAt: createdAt,
				SourceEditedAt:  editedAt,
			},
			want:        repository.UpsertDispositionUpdated,
			updatedHash: "hash-new",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			defer db.Close()

			repo := NewRepository(db)
			input := repository.RawPostInput{
				Platform:        "nowcoder",
				SourcePostID:    "861032210369871872",
				Title:           "阿里面经",
				Content:         "问了 Kafka 和 MQ",
				ContentHash:     "hash-1",
				AuthorName:      "alice",
				PostURL:         "https://www.nowcoder.com/discuss/861032210369871872",
				SourceCreatedAt: createdAt,
				SourceEditedAt:  editedAt,
				RawPayloadJSON:  `{"id":"861032210369871872"}`,
			}

			switch {
			case tt.existing == nil:
				mock.ExpectQuery("SELECT (.+) FROM raw_posts WHERE platform = \\? AND source_post_id = \\?").
					WithArgs(input.Platform, input.SourcePostID).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectExec("INSERT INTO raw_posts").
					WithArgs(
						input.Platform,
						input.SourcePostID,
						input.Title,
						input.Content,
						input.ContentHash,
						input.AuthorName,
						input.PostURL,
						input.SourceCreatedAt,
						input.SourceEditedAt,
						sqlmock.AnyArg(),
						repository.ParseStatusPending,
						0,
						input.RawPayloadJSON,
					).
					WillReturnResult(sqlmock.NewResult(41, 1))
			case tt.want == repository.UpsertDispositionUnchanged:
				mockExistingPostLookup(mock, tt.existing)
			case tt.want == repository.UpsertDispositionUpdated:
				input.Content = "新内容"
				input.ContentHash = tt.updatedHash
				input.SourceEditedAt = editedAt.Add(1 * time.Hour)
				mockExistingPostLookup(mock, tt.existing)
				mock.ExpectExec("UPDATE raw_posts SET").
					WithArgs(
						input.Title,
						input.Content,
						input.ContentHash,
						input.AuthorName,
						input.PostURL,
						input.CompanyNameRaw,
						input.CompanyNameNorm,
						input.SourceCreatedAt,
						input.SourceEditedAt,
						sqlmock.AnyArg(),
						repository.ParseStatusPending,
						input.RawPayloadJSON,
						tt.existing.ID,
					).
					WillReturnResult(sqlmock.NewResult(0, 1))
			}

			result, err := repo.UpsertRawPost(context.Background(), input)
			if err != nil {
				t.Fatalf("UpsertRawPost: %v", err)
			}

			if result.Disposition != tt.want {
				t.Fatalf("expected disposition %q, got %q", tt.want, result.Disposition)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet sql expectations: %v", err)
			}
		})
	}
}

func TestScanRawPostAllowsNullCompanyFields(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	createdAt := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	editedAt := createdAt.Add(10 * time.Minute)
	lastCrawledAt := createdAt.Add(20 * time.Minute)

	rows := sqlmock.NewRows([]string{
		"id",
		"platform",
		"source_post_id",
		"title",
		"content",
		"content_hash",
		"author_name",
		"post_url",
		"company_name_raw",
		"company_name_norm",
		"source_created_at",
		"source_edited_at",
		"last_crawled_at",
		"parse_status",
		"parse_attempts",
		"raw_payload_json",
	}).AddRow(
		1,
		"nowcoder",
		"post-1",
		"阿里面经",
		"问了 Kafka",
		"hash-1",
		"alice",
		"https://www.nowcoder.com/discuss/post-1",
		nil,
		nil,
		createdAt,
		editedAt,
		lastCrawledAt,
		repository.ParseStatusPending,
		0,
		`{"id":"post-1"}`,
	)

	mock.ExpectQuery("SELECT (.+) FROM raw_posts WHERE platform = \\? AND source_post_id = \\?").
		WithArgs("nowcoder", "post-1").
		WillReturnRows(rows)

	repo := NewRepository(db)
	result, err := repo.findRawPostBySource(context.Background(), "nowcoder", "post-1")
	if err != nil {
		t.Fatalf("findRawPostBySource: %v", err)
	}

	if result.CompanyNameRaw != "" || result.CompanyNameNorm != "" {
		t.Fatalf("expected null company fields to scan as empty strings, got raw=%q norm=%q", result.CompanyNameRaw, result.CompanyNameNorm)
	}
}

func TestGetPostBySourceAllowsNullCompanyFields(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	createdAt := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	editedAt := createdAt.Add(10 * time.Minute)
	lastCrawledAt := createdAt.Add(20 * time.Minute)

	rows := sqlmock.NewRows([]string{
		"id",
		"platform",
		"source_post_id",
		"title",
		"content",
		"content_hash",
		"author_name",
		"post_url",
		"company_name_raw",
		"company_name_norm",
		"source_created_at",
		"source_edited_at",
		"last_crawled_at",
		"parse_status",
		"parse_attempts",
		"raw_payload_json",
		"raw_post_id",
		"schema_version",
		"llm_provider",
		"llm_model",
		"company_name",
		"sentiment",
		"sentiment_reason",
		"key_events_json",
		"questions_json",
		"tags_json",
		"raw_json",
		"parsed_at",
	}).AddRow(
		1,
		"nowcoder",
		"post-1",
		"阿里面经",
		"问了 Kafka",
		"hash-1",
		"alice",
		"https://www.nowcoder.com/discuss/post-1",
		nil,
		nil,
		createdAt,
		editedAt,
		lastCrawledAt,
		repository.ParseStatusPending,
		0,
		`{"id":"post-1"}`,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	mock.ExpectQuery("SELECT (.+) FROM raw_posts rp").
		WithArgs("nowcoder", "post-1").
		WillReturnRows(rows)

	repo := NewRepository(db)
	result, err := repo.GetPostBySource(context.Background(), "nowcoder", "post-1")
	if err != nil {
		t.Fatalf("GetPostBySource: %v", err)
	}

	if result.RawPost.CompanyNameRaw != "" || result.RawPost.CompanyNameNorm != "" {
		t.Fatalf("expected null company fields to scan as empty strings, got raw=%q norm=%q", result.RawPost.CompanyNameRaw, result.RawPost.CompanyNameNorm)
	}
	if result.ParseResult != nil {
		t.Fatalf("expected parse result to remain nil when left join columns are null")
	}
}

func mockExistingPostLookup(mock sqlmock.Sqlmock, post *repository.RawPost) {
	rows := sqlmock.NewRows([]string{
		"id",
		"platform",
		"source_post_id",
		"title",
		"content",
		"content_hash",
		"author_name",
		"post_url",
		"company_name_raw",
		"company_name_norm",
		"source_created_at",
		"source_edited_at",
		"last_crawled_at",
		"parse_status",
		"parse_attempts",
		"raw_payload_json",
	}).AddRow(
		post.ID,
		post.Platform,
		post.SourcePostID,
		post.Title,
		post.Content,
		post.ContentHash,
		post.AuthorName,
		post.PostURL,
		post.CompanyNameRaw,
		post.CompanyNameNorm,
		post.SourceCreatedAt,
		post.SourceEditedAt,
		time.Now().UTC(),
		repository.ParseStatusParsed,
		1,
		`{"id":"861032210369871872"}`,
	)

	mock.ExpectQuery("SELECT (.+) FROM raw_posts WHERE platform = \\? AND source_post_id = \\?").
		WithArgs(post.Platform, post.SourcePostID).
		WillReturnRows(rows)
}
