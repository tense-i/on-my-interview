package mysql

import (
	"context"
	"database/sql"
	_ "embed"
	"strings"
)

//go:embed migrations/001_init.sql
var initSchema string

func ApplyMigrations(ctx context.Context, db *sql.DB) error {
	statements := strings.SplitSeq(initSchema, ";")
	for statement := range statements {
		sqlText := strings.TrimSpace(statement)
		if sqlText == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, sqlText); err != nil {
			return err
		}
	}
	return nil
}
