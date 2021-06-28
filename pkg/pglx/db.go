package pglx

import (
	"context"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/sphera-erp/sphera/pkg/pglx/reflectx"
)

func (db *DB) Close() {
	db.Pool.Close()
}

// MapperFunc sets a new mapper for this db using the default sqlx struct tag
// and the provided mapper function.
func (db *DB) MapperFunc(mf func(string) string) {
	db.Mapper = reflectx.NewMapperFunc("db", mf)
}

// BindNamed binds a query using the DB driver's bindvar type.
func (db *DB) BindNamed(sql string, arg interface{}) (string, []interface{}, error) {
	return bindNamedMapper(sql, arg, db.Mapper)
}

// NamedQuery using this DB.
// Any named placeholder parameters are replaced with fields from arg.
func (db *DB) NamedQuery(ctx context.Context, sql string, arg interface{}) (*Rows, error) {
	return NamedQuery(ctx, db, sql, arg)
}

// NamedExec using this DB.
// Any named placeholder parameters are replaced with fields from arg.
func (db *DB) NamedExec(ctx context.Context, sql string, arg interface{}) (pgconn.CommandTag, error) {
	return NamedExec(ctx, db, sql, arg)
}

// Select using this DB.
// Any placeholder parameters are replaced with supplied args.
func (db *DB) Select(ctx context.Context, dest interface{}, sql string, args ...interface{}) error {
	return Select(ctx, db, dest, sql, args...)
}

// Get using this DB.
// Any placeholder parameters are replaced with supplied args.
// An error is returned if the result set is empty.
func (db *DB) Get(ctx context.Context, dest interface{}, sql string, args ...interface{}) error {
	return Get(ctx, db, dest, sql, args...)
}

// MustBegin starts a transaction, and panics on error.  Returns an *sqlx.Tx instead
// of an *sql.Tx.
func (db *DB) MustBegin(ctx context.Context) *Tx {
	tx, err := db.BeginX(ctx)
	if err != nil {
		panic(err)
	}
	return tx
}

// Beginx begins a transaction and returns an *pglx.Tx instead of an *sql.Tx.
func (db *DB) BeginX(ctx context.Context) (*Tx, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &Tx{Tx: tx, unsafe: db.unsafe, Mapper: db.Mapper}, err
}

func (db *DB) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return db.Pool.Exec(ctx, sql, arguments...)
}

func (db *DB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	r, err := db.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Queryx queries the database and returns an *sqlx.Rows.
// Any placeholder parameters are replaced with supplied args.
func (db *DB) QueryX(ctx context.Context, sql string, args ...interface{}) (*Rows, error) {
	r, err := db.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{Rows: r, unsafe: db.unsafe, Mapper: db.Mapper}, err
}

func (db *DB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	r := db.Pool.QueryRow(ctx, sql, args...)
	return r
}

// QueryRowx queries the database and returns an *sqlx.Row.
// Any placeholder parameters are replaced with supplied args.
func (db *DB) QueryRowX(ctx context.Context, sql string, args ...interface{}) *Row {
	row := db.Pool.QueryRow(ctx, sql, args...)
	return &Row{row: row, unsafe: db.unsafe, Mapper: db.Mapper}
}

// MustExec (panic) runs MustExec using this database.
// Any placeholder parameters are replaced with supplied args.
func (db *DB) MustExec(ctx context.Context, sql string, args ...interface{}) pgconn.CommandTag {
	return MustExec(ctx, db, sql, args...)
}
