package mysql

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func Open(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, nil
	}
	return sql.Open("mysql", dsn)
}
