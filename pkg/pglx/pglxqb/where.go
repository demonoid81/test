package pglxqb

import (
	"fmt"
)

type wherePart part

func newWherePart(pred interface{}, args ...interface{}) QB {
	return &wherePart{pred: pred, args: args}
}

func (p wherePart) ToSql() (sql string, args []interface{}, err error) {
	switch pred := p.pred.(type) {
	case nil:
		// no-op
	case rawQB:
		return pred.toSqlRaw()
	case QB:
		return pred.ToSql()
	case map[string]interface{}:
		return Eq(pred).ToSql()
	case string:
		sql = pred
		args = p.args
	default:
		err = fmt.Errorf("expected string-keyed map or string, not %T", pred)
	}
	return
}
