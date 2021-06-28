package pglx

import (
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"golang.org/x/net/context"
)

// Unsafe returns a version of Tx which will silently succeed to scan when
// columns in the SQL result have no fields in the destination struct.
func (tx *Tx) Unsafe() *Tx {
	return &Tx{Tx: tx.Tx, unsafe: true, Mapper: tx.Mapper}
}

// BindNamed binds a query within a transaction's bindvar type.
func (tx *Tx) BindNamed(sql string, arg interface{}) (string, []interface{}, error) {
	return bindNamedMapper(sql, arg, tx.Mapper)
}

// NamedQuery within a transaction.
// Any named placeholder parameters are replaced with fields from arg.
func (tx *Tx) NamedQuery(ctx context.Context, sql string, arg interface{}) (*Rows, error) {
	return NamedQuery(ctx, tx, sql, arg)
}

// NamedExec a named query within a transaction.
// Any named placeholder parameters are replaced with fields from arg.
func (tx *Tx) NamedExec(ctx context.Context, sql string, arg interface{}) (pgconn.CommandTag, error) {
	return NamedExec(ctx, tx, sql, arg)
}

// Select within a transaction.
// Any placeholder parameters are replaced with supplied args.
func (tx *Tx) Select(ctx context.Context, dest interface{}, sql string, args ...interface{}) error {
	return Select(ctx, tx, dest, sql, args...)
}

func (tx *Tx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	r, err := tx.Tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return r, err
}

// QueryX within a transaction.
// Any placeholder parameters are replaced with supplied args.
func (tx *Tx) QueryX(ctx context.Context, sql string, args ...interface{}) (*Rows, error) {
	r, err := tx.Tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{Rows: r, unsafe: tx.unsafe, Mapper: tx.Mapper}, err
}

func (tx *Tx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	row := tx.Tx.QueryRow(ctx, sql, args...)
	return row
}

// QueryRowx within a transaction.
// Any placeholder parameters are replaced with supplied args.
func (tx *Tx) QueryRowX(ctx context.Context, sql string, args ...interface{}) *Row {
	row := tx.Tx.QueryRow(ctx, sql, args...)
	return &Row{row: row, unsafe: tx.unsafe, Mapper: tx.Mapper}
}

// Get within a transaction.
// Any placeholder parameters are replaced with supplied args.
// An error is returned if the result set is empty.
func (tx *Tx) Get(ctx context.Context, dest interface{}, sql string, args ...interface{}) error {
	return Get(ctx, tx, dest, sql, args...)
}

// MustExec runs MustExec within a transaction.
// Any placeholder parameters are replaced with supplied args.
func (tx *Tx) MustExec(ctx context.Context, sql string, args ...interface{}) pgconn.CommandTag {
	return MustExec(ctx, tx, sql, args...)
}
