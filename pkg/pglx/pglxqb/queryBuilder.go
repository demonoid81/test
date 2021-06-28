// Package squirrel provides a fluent SQL generator.
//
// See https://github.com/Masterminds/squirrel for examples.
package pglxqb

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/sphera-erp/sphera/pkg/pglx"
	"strings"

	"github.com/lann/builder"
)

// QB is the interface that wraps the ToSql method.
//
// ToSql returns a SQL representation of the QB, along with a slice of args
// as passed to e.g. database/sql.Exec. It can also return an error.
type QB interface {
	ToSql() (string, []interface{}, error)
}

// rawQB is expected to do what QB does, but without finalizing placeholders.
// This is useful for nested queries.
type rawQB interface {
	toSqlRaw() (string, []interface{}, error)
}

// Execer is the interface that wraps the Exec method.
//
// Exec executes the given query as implemented by database/sql.Exec.
type Execer interface {
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
}

// Queryer is the interface that wraps the Query method.
//
// Query executes the given query as implemented by database/sql.Query.
type Queryer interface {
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	QueryX(ctx context.Context, query string, args ...interface{}) (*pglx.Rows, error)
}

// QueryRower is the interface that wraps the QueryRow method.
//
// QueryRow executes the given query as implemented by database/sql.QueryRow.
type QueryRower interface {
	QueryRow(ctx context.Context, query string, args ...interface{}) RowScanner
}

// BaseRunner groups the Execer and Queryer interfaces.
type BaseRunner interface {
	Execer
	Queryer
}

// Runner groups the Execer, Queryer, and QueryRower interfaces.
type Runner interface {
	Execer
	Queryer
	QueryRower
}

// WrapStdSql wraps a type implementing the standard SQL interface with methods that
// squirrel expects.
func WrapStdSql(stdSql StdSql) Runner {
	return &stdsqlRunner{stdSql}
}

// StdSql encompasses the standard methods of the *sql.DB type, and other types that
// wrap these methods.
type StdSql interface {
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryX(context.Context, string, ...interface{}) (*pglx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
}

type stdsqlRunner struct {
	StdSql
}

func (r *stdsqlRunner) QueryRow(ctx context.Context, query string, args ...interface{}) RowScanner {
	return r.StdSql.QueryRow(ctx, query, args...)
}

func setRunWith(b interface{}, runner BaseRunner) interface{} {
	switch r := runner.(type) {
	case StdSql:
		runner = WrapStdSql(r)
	}
	return builder.Set(b, "RunWith", runner)
}

// RunnerNotSet is returned by methods that need a Runner if it isn't set.
var RunnerNotSet = fmt.Errorf("cannot run; no Runner set (RunWith)")

// RunnerNotQueryRunner is returned by QueryRow if the RunWith value doesn't implement QueryRower.
var RunnerNotQueryRunner = fmt.Errorf("cannot QueryRow; Runner is not a QueryRower")

// ExecWith Execs the SQL returned by s with db.
func ExecWith(ctx context.Context, db Execer, s QB) (res pgconn.CommandTag, err error) {
	query, args, err := s.ToSql()
	if err != nil {
		return
	}
	return db.Exec(ctx, query, args...)
}

// QueryWith Querys the SQL returned by s with db.
func QueryWith(ctx context.Context, db Queryer, s QB) (rows pgx.Rows, err error) {
	query, args, err := s.ToSql()
	if err != nil {
		return
	}
	return db.Query(ctx, query, args...)
}

// QueryWith Querys the SQL returned by s with db.
func QueryXWith(ctx context.Context, db Queryer, s QB) (rows *pglx.Rows, err error) {
	query, args, err := s.ToSql()
	if err != nil {
		return
	}
	return db.QueryX(ctx, query, args...)
}

// QueryRowWith QueryRows the SQL returned by s with db.
func QueryRowWith(ctx context.Context, db QueryRower, s QB) RowScanner {
	query, args, err := s.ToSql()
	return &Row{RowScanner: db.QueryRow(ctx, query, args...), err: err}
}

// DebugSqlizer calls ToSql on s and shows the approximate SQL to be executed
//
// If ToSql returns an error, the result of this method will look like:
// "[ToSql error: %s]" or "[DebugSqlizer error: %s]"
//
// IMPORTANT: As its name suggests, this function should only be used for
// debugging. While the string result *might* be valid SQL, this function does
// not try very hard to ensure it. Additionally, executing the output of this
// function with any untrusted user input is certainly insecure.
func DebugSqlizer(s QB) string {
	sql, args, err := s.ToSql()
	if err != nil {
		return fmt.Sprintf("[ToSql error: %s]", err)
	}

	var placeholder string
	downCast, ok := s.(placeholderDebugger)
	if !ok {
		placeholder = "?"
	} else {
		placeholder = downCast.debugPlaceholder()
	}
	// TODO: dedupe this with placeholder.go
	buf := &bytes.Buffer{}
	i := 0
	for {
		p := strings.Index(sql, placeholder)
		if p == -1 {
			break
		}
		if len(sql[p:]) > 1 && sql[p:p+2] == "??" { // escape ?? => ?
			buf.WriteString(sql[:p])
			buf.WriteString("?")
			if len(sql[p:]) == 1 {
				break
			}
			sql = sql[p+2:]
		} else {
			if i+1 > len(args) {
				return fmt.Sprintf(
					"[DebugSqlizer error: too many placeholders in %#v for %d args]",
					sql, len(args))
			}
			buf.WriteString(sql[:p])
			fmt.Fprintf(buf, "'%v'", args[i])
			// advance our sql string "cursor" beyond the arg we placed
			sql = sql[p+1:]
			i++
		}
	}
	if i < len(args) {
		return fmt.Sprintf(
			"[DebugSqlizer error: not enough placeholders in %#v for %d args]",
			sql, len(args))
	}
	// "append" any remaning sql that won't need interpolating
	buf.WriteString(sql)
	return buf.String()
}
