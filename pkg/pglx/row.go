package pglx

import (
	"errors"
	"reflect"
)

func (r *Row) Scan(dest ...interface{}) error {
	//rows := (*pgx.)
	err := r.row.Scan(dest...)
	if err != nil {
		return err
	}
	return nil
}

func (r *Row) scanAny(dest interface{}) error {
	if r.row == nil {
		return errors.New("sql: no rows in result set")
	}

	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr {
		return errors.New("must pass a pointer, not a value, to StructScan destination")
	}
	if v.IsNil() {
		return errors.New("nil pointer passed to StructScan destination")
	}

	return r.Scan(dest)
}

// StructScan a single Row into dest.
func (r *Row) StructScan(dest interface{}) error {
	return r.scanAny(dest)
}
