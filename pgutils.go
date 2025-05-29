package pgutils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func sqlErr(err error, query string, args ...interface{}) error {
	return fmt.Errorf(`run query "%s" with args %+v: %w`, query, args, err)
}

// Exec does not returns result from query.
func Exec(ctx context.Context, db sqlx.ExecerContext, query string, args ...interface{}) (sql.Result, error) {
	res, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return res, sqlErr(err, query, args...)
	}

	return res, nil
}

// Get returns no more than one result from query.
func Get(ctx context.Context, db sqlx.QueryerContext, dest interface{}, query string, args ...interface{}) error {
	if err := sqlx.GetContext(ctx, db, dest, query, args...); err != nil {
		return sqlErr(err, query, args...)
	}

	return nil
}

// Select returns multiple results from query.
func Select(ctx context.Context, db sqlx.QueryerContext, dest interface{}, query string, args ...interface{}) error {
	if err := sqlx.SelectContext(ctx, db, dest, query, args...); err != nil {
		return sqlErr(err, query, args...)
	}

	return nil
}

type TxFunc func(tx *sqlx.Tx) error

type TxRunner interface {
	BeginTxx(context.Context, *sql.TxOptions) (*sqlx.Tx, error)
}

// RunTx wraps the function into a transaction.
// If f returns error, RunTx will roll back transaction.
func RunTx(ctx context.Context, db TxRunner, f TxFunc) (err error) {
	var tx *sqlx.Tx

	opts := &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	}

	tx, err = db.BeginTxx(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback())
		} else {
			err = tx.Commit()
		}
	}()

	return f(tx)
}

var ErrNoAffectedRows = errors.New("no affected rows")

// RequireAffected checks the result of the Exec function for the presence of at least an affected table row.
// Returns ErrNoAffectedRows.
//
//	err := pgutils.RequireAffected(pgutils.Exec(ctx, db, `INSERT ...`))
func RequireAffected(res sql.Result, err error) error {
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrNoAffectedRows
	}

	return nil
}
