package pglxqb

import (
	"fmt"
	"io"
)

type part struct {
	pred interface{}
	args []interface{}
}

func newPart(pred interface{}, args ...interface{}) QB {
	return &part{pred, args}
}

func (p part) ToSql() (sql string, args []interface{}, err error) {
	switch pred := p.pred.(type) {
	case nil:
		// no-op
	case QB:
		sql, args, err = pred.ToSql()
	case string:
		sql = pred
		args = p.args
	default:
		err = fmt.Errorf("expected string or QB, not %T", pred)
	}
	return
}

func appendToSql(parts []QB, w io.Writer, sep string, args []interface{}) ([]interface{}, error) {
	for i, p := range parts {
		partSql, partArgs, err := p.ToSql()
		if err != nil {
			return nil, err
		} else if len(partSql) == 0 {
			continue
		}

		if i > 0 {
			_, err := io.WriteString(w, sep)
			if err != nil {
				return nil, err
			}
		}

		_, err = io.WriteString(w, partSql)
		if err != nil {
			return nil, err
		}
		args = append(args, partArgs...)
	}
	return args, nil
}
