package pglx

import (
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sphera-erp/sphera/pkg/pglx/reflectx"
)

// DB is
type DB struct {
	Pool   pgxpool.Pool
	unsafe bool
	Mapper *reflectx.Mapper
}

type Row struct {
	unsafe bool
	row    pgx.Row
	Mapper *reflectx.Mapper
}

// Tx is an pqlx wrapper around pgpool.Tx with extra functionality
type Tx struct {
	pgx.Tx
	unsafe bool
	Mapper *reflectx.Mapper
}

type Rows struct {
	pgx.Rows
	unsafe bool
	Mapper *reflectx.Mapper
	// these fields cache memory use for a rows during iteration w/ structScan
	started bool
	fields  [][]int
	values  []interface{}
}
