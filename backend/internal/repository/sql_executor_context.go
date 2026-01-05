package repository

import (
	"context"
	"database/sql"

	dbent "github.com/Wei-Shaw/sub2api/ent"
)

func sqlExecutorFromContext(ctx context.Context, db *sql.DB) sqlExecutor {
	if tx := dbent.TxFromContext(ctx); tx != nil {
		if exec, ok := tx.Client().Driver().(sqlExecutor); ok {
			return exec
		}
	}
	return db
}

