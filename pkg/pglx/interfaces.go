package pglx

import (
	"context"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

// ColScanner is an interface used by MapScan and SliceScan
type ColScanner interface {
	Columns() ([]string, error)
	Scan(dest ...interface{}) error
	Err() error
}

type Scannable interface {
	Scan(dest ...interface{}) (err error)
}

// Ext is a union interface which can bind, query, and exec, used by
// NamedQuery and NamedExec.
type Ext interface {
	Queryer
	Execer
}

// Queryer is an interface used by Get and Select
type Queryer interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryX(ctx context.Context, sql string, args ...interface{}) (*Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	QueryRowX(ctx context.Context, sql string, args ...interface{}) *Row
}

// Execer is an interface used by MustExec and LoadFile
type Execer interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

type rowsi interface {
	Close()
	Columns() ([]string, error)
	Err() error
	Next() bool
	Scan(...interface{}) error
}
