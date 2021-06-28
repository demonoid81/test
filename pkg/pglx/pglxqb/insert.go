package pglxqb

import (
	"bytes"
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/sphera-erp/sphera/pkg/pglx"

	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"io"
	"sort"
	"strings"

	"github.com/lann/builder"
)

type insertData struct {
	PlaceholderFormat PlaceholderFormat
	RunWith           BaseRunner
	Prefixes          []QB
	StatementKeyword  string
	Options           []string
	Into              string
	Columns           []string
	Values            [][]interface{}
	ConflictColumns   []string
	UpsetClauses      []setClause
	Suffixes          []QB
	Select            *SelectBuilder
}

func (d *insertData) Exec(ctx context.Context) (pgconn.CommandTag, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	return ExecWith(ctx, d.RunWith, d)
}

func (d *insertData) Query(ctx context.Context) (pgx.Rows, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	return QueryWith(ctx, d.RunWith, d)
}

func (d *insertData) QueryX(ctx context.Context) (*pglx.Rows, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	return QueryXWith(ctx, d.RunWith, d)
}

func (d *insertData) QueryRow(ctx context.Context) RowScanner {
	if d.RunWith == nil {
		return &Row{err: RunnerNotSet}
	}
	queryRower, ok := d.RunWith.(QueryRower)
	if !ok {
		return &Row{err: RunnerNotQueryRunner}
	}
	return QueryRowWith(ctx, queryRower, d)
}

func (d *insertData) ToSql() (sqlStr string, args []interface{}, err error) {
	if len(d.Into) == 0 {
		err = errors.New("insert statements must specify a table")
		return
	}
	if len(d.Values) == 0 && d.Select == nil {
		err = errors.New("insert statements must have at least one set of values or select clause")
		return
	}

	sql := &bytes.Buffer{}

	if len(d.Prefixes) > 0 {
		args, err = appendToSql(d.Prefixes, sql, " ", args)
		if err != nil {
			return
		}

		sql.WriteString(" ")
	}

	if d.StatementKeyword == "" {
		sql.WriteString("INSERT ")
	} else {
		sql.WriteString(d.StatementKeyword)
		sql.WriteString(" ")
	}

	if len(d.Options) > 0 {
		sql.WriteString(strings.Join(d.Options, " "))
		sql.WriteString(" ")
	}

	sql.WriteString("INTO ")
	sql.WriteString(d.Into)
	sql.WriteString(" ")

	if len(d.Columns) > 0 {
		sql.WriteString("(")
		sql.WriteString(strings.Join(d.Columns, ","))
		sql.WriteString(") ")
	}

	if d.Select != nil {
		args, err = d.appendSelectToSQL(sql, args)
	} else {
		args, err = d.appendValuesToSQL(sql, args)
	}

	fmt.Println(d.ConflictColumns)
	if len(d.ConflictColumns) > 0 {
		sql.WriteString(" ON CONFLICT ")
		sql.WriteString("(")
		sql.WriteString(strings.Join(d.ConflictColumns, ","))
		sql.WriteString(") ")
		sql.WriteString("DO ")
		if len(d.UpsetClauses) > 0 {
			sql.WriteString("UPDATE ")
			sql.WriteString("SET ")
			setSqls := make([]string, len(d.UpsetClauses))
			for i, setClause := range d.UpsetClauses {
				var valSql string
				if vs, ok := setClause.value.(QB); ok {
					vsql, vargs, err := vs.ToSql()
					if err != nil {
						return "", nil, err
					}
					if _, ok := vs.(SelectBuilder); ok {
						valSql = fmt.Sprintf("(%s)", vsql)
					} else {
						valSql = vsql
					}
					args = append(args, vargs...)
				} else {
					valSql = "?"
					args = append(args, setClause.value)
				}
				setSqls[i] = fmt.Sprintf("%s = %s", setClause.column, valSql)
			}
			sql.WriteString(strings.Join(setSqls, ", "))
		} else {
			sql.WriteString("NOTHING ")
		}
	}

	if err != nil {
		return
	}

	if len(d.Suffixes) > 0 {
		sql.WriteString(" ")
		args, err = appendToSql(d.Suffixes, sql, " ", args)
		if err != nil {
			return
		}
	}

	sqlStr, err = d.PlaceholderFormat.ReplacePlaceholders(sql.String())
	return
}

func (d *insertData) appendValuesToSQL(w io.Writer, args []interface{}) ([]interface{}, error) {
	if len(d.Values) == 0 {
		return args, errors.New("values for insert statements are not set")
	}

	io.WriteString(w, "VALUES ")

	valuesStrings := make([]string, len(d.Values))
	for r, row := range d.Values {
		valueStrings := make([]string, len(row))
		for v, val := range row {
			if vs, ok := val.(QB); ok {
				vsql, vargs, err := vs.ToSql()
				if err != nil {
					return nil, err
				}
				valueStrings[v] = vsql
				args = append(args, vargs...)
			} else {
				valueStrings[v] = "?"
				args = append(args, val)
			}
		}
		valuesStrings[r] = fmt.Sprintf("(%s)", strings.Join(valueStrings, ","))
	}

	io.WriteString(w, strings.Join(valuesStrings, ","))

	return args, nil
}

func (d *insertData) appendSelectToSQL(w io.Writer, args []interface{}) ([]interface{}, error) {
	if d.Select == nil {
		return args, errors.New("select clause for insert statements are not set")
	}

	selectClause, sArgs, err := d.Select.ToSql()
	if err != nil {
		return args, err
	}

	io.WriteString(w, selectClause)
	args = append(args, sArgs...)

	return args, nil
}

// Builder

// InsertBuilder builds SQL INSERT statements.
type InsertBuilder builder.Builder

func init() {
	builder.Register(InsertBuilder{}, insertData{})
}

// Format methods

// PlaceholderFormat sets PlaceholderFormat (e.g. Question or Dollar) for the
// query.
func (b InsertBuilder) PlaceholderFormat(f PlaceholderFormat) InsertBuilder {
	return builder.Set(b, "PlaceholderFormat", f).(InsertBuilder)
}

// Runner methods

// RunWith sets a Runner (like database/sql.DB) to be used with e.g. Exec.
func (b InsertBuilder) RunWith(runner BaseRunner) InsertBuilder {
	return setRunWith(b, runner).(InsertBuilder)
}

// Exec builds and Execs the query with the Runner set by RunWith.
func (b InsertBuilder) Exec(ctx context.Context) (pgconn.CommandTag, error) {
	data := builder.GetStruct(b).(insertData)
	return data.Exec(ctx)
}

// Query builds and Querys the query with the Runner set by RunWith.
func (b InsertBuilder) Query(ctx context.Context) (pgx.Rows, error) {
	data := builder.GetStruct(b).(insertData)
	return data.Query(ctx)
}

// Query builds and Querys the query with the Runner set by RunWith.
func (b InsertBuilder) QueryX(ctx context.Context) (*pglx.Rows, error) {
	data := builder.GetStruct(b).(insertData)
	return data.QueryX(ctx)
}

// QueryRow builds and QueryRows the query with the Runner set by RunWith.
func (b InsertBuilder) QueryRow(ctx context.Context) RowScanner {
	data := builder.GetStruct(b).(insertData)
	return data.QueryRow(ctx)
}

// Scan is a shortcut for QueryRow().Scan.
func (b InsertBuilder) Scan(ctx context.Context, dest ...interface{}) error {
	return b.QueryRow(ctx).Scan(dest...)
}

// SQL methods

// ToSql builds the query into a SQL string and bound args.
func (b InsertBuilder) ToSql() (string, []interface{}, error) {
	data := builder.GetStruct(b).(insertData)
	return data.ToSql()
}

// Prefix adds an expression to the beginning of the query
func (b InsertBuilder) Prefix(sql string, args ...interface{}) InsertBuilder {
	return b.PrefixExpr(Expr(sql, args...))
}

// PrefixExpr adds an expression to the very beginning of the query
func (b InsertBuilder) PrefixExpr(expr QB) InsertBuilder {
	return builder.Append(b, "Prefixes", expr).(InsertBuilder)
}

// Options adds keyword options before the INTO clause of the query.
func (b InsertBuilder) Options(options ...string) InsertBuilder {
	return builder.Extend(b, "Options", options).(InsertBuilder)
}

// Into sets the INTO clause of the query.
func (b InsertBuilder) Into(from string) InsertBuilder {
	return builder.Set(b, "Into", from).(InsertBuilder)
}

// Columns adds insert columns to the query.
func (b InsertBuilder) Columns(columns ...string) InsertBuilder {
	return builder.Extend(b, "Columns", columns).(InsertBuilder)
}

// Values adds a single row's values to the query.
func (b InsertBuilder) Values(values ...interface{}) InsertBuilder {
	return builder.Append(b, "Values", values).(InsertBuilder)
}

// Suffix adds an expression to the end of the query
func (b InsertBuilder) Suffix(sql string, args ...interface{}) InsertBuilder {
	return b.SuffixExpr(Expr(sql, args...))
}

// SuffixExpr adds an expression to the end of the query
func (b InsertBuilder) SuffixExpr(expr QB) InsertBuilder {
	return builder.Append(b, "Suffixes", expr).(InsertBuilder)
}

// Values adds a single row's values to the query.
func (b InsertBuilder) OnConflictNothing(columns ...string) InsertBuilder {
	return builder.Extend(b, "ConflictColumns", columns).(InsertBuilder)
}

func (b InsertBuilder) OnConflictUpdateMap(clauses map[string]interface{}, columns ...string) InsertBuilder {
	b = builder.Extend(b, "ConflictColumns", columns).(InsertBuilder)
	keys := make([]string, len(clauses))
	i := 0
	for key := range clauses {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	for _, key := range keys {
		val, _ := clauses[key]
		b = b.Upset(key, val)
	}
	return b
}

func (b InsertBuilder) Upset(column string, value interface{}) InsertBuilder {
	return builder.Append(b, "UpsetClauses", setClause{column: column, value: value}).(InsertBuilder)
}

// SetMap set columns and values for insert builder from a map of column name and value
// note that it will reset all previous columns and values was set if any
func (b InsertBuilder) SetMap(clauses map[string]interface{}) InsertBuilder {
	// Keep the columns in a consistent order by sorting the column key string.
	cols := make([]string, 0, len(clauses))
	for col := range clauses {
		cols = append(cols, col)
	}
	sort.Strings(cols)

	vals := make([]interface{}, 0, len(clauses))
	for _, col := range cols {
		vals = append(vals, clauses[col])
	}

	b = builder.Set(b, "Columns", cols).(InsertBuilder)
	b = builder.Set(b, "Values", [][]interface{}{vals}).(InsertBuilder)

	return b
}

// Select set Select clause for insert query
// If Values and Select are used, then Select has higher priority
func (b InsertBuilder) Select(sb SelectBuilder) InsertBuilder {
	return builder.Set(b, "Select", &sb).(InsertBuilder)
}

func (b InsertBuilder) statementKeyword(keyword string) InsertBuilder {
	return builder.Set(b, "StatementKeyword", keyword).(InsertBuilder)
}
