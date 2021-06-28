package pglx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/sphera-erp/sphera/pkg/pglx/reflectx"
	"reflect"
	"regexp"
	"strconv"
	"unicode"
)

// convertMapStringInterface attempts to convert v to map[string]interface{}.
// Unlike v.(map[string]interface{}), this function works on named types that
// are convertible to map[string]interface{} as well.
func convertMapStringInterface(v interface{}) (map[string]interface{}, bool) {
	var m map[string]interface{}
	mtype := reflect.TypeOf(m)
	t := reflect.TypeOf(v)
	if !t.ConvertibleTo(mtype) {
		return nil, false
	}
	return reflect.ValueOf(v).Convert(mtype).Interface().(map[string]interface{}), true

}

func bindAnyArgs(names []string, arg interface{}, m *reflectx.Mapper) ([]interface{}, error) {
	if maparg, ok := convertMapStringInterface(arg); ok {
		return bindMapArgs(names, maparg)
	}
	return bindArgs(names, arg, m)
}

// private interface to generate a list of interfaces from a given struct
// type, given a list of names to pull out of the struct.  Used by public
// BindStruct interface.
func bindArgs(names []string, arg interface{}, m *reflectx.Mapper) ([]interface{}, error) {
	arglist := make([]interface{}, 0, len(names))

	// grab the indirected value of arg
	v := reflect.ValueOf(arg)
	for v = reflect.ValueOf(arg); v.Kind() == reflect.Ptr; {
		v = v.Elem()
	}

	err := m.TraversalsByNameFunc(v.Type(), names, func(i int, t []int) error {
		if len(t) == 0 {
			return fmt.Errorf("could not find name %s in %#v", names[i], arg)
		}

		val := reflectx.FieldByIndexesReadOnly(v, t)
		arglist = append(arglist, val.Interface())

		return nil
	})

	return arglist, err
}

// like bindArgs, but for maps.
func bindMapArgs(names []string, arg map[string]interface{}) ([]interface{}, error) {
	arglist := make([]interface{}, 0, len(names))

	for _, name := range names {
		val, ok := arg[name]
		if !ok {
			return arglist, fmt.Errorf("could not find name %s in %#v", name, arg)
		}
		arglist = append(arglist, val)
	}
	return arglist, nil
}

// bindStruct binds a named parameter query with fields from a struct argument.
// The rules for binding field names to parameter names follow the same
// conventions as for StructScan, including obeying the `db` struct tags.
func bindStruct(query string, arg interface{}, m *reflectx.Mapper) (string, []interface{}, error) {
	bound, names, err := compileNamedQuery([]byte(query))
	if err != nil {
		return "", []interface{}{}, err
	}

	arglist, err := bindAnyArgs(names, arg, m)
	if err != nil {
		return "", []interface{}{}, err
	}

	return bound, arglist, nil
}

var valueBracketReg = regexp.MustCompile(`\([^(]*.[^(]\)\s*$`)

func fixBound(bound string, loop int) string {
	loc := valueBracketReg.FindStringIndex(bound)
	if len(loc) != 2 {
		return bound
	}
	var buffer bytes.Buffer

	buffer.WriteString(bound[0:loc[1]])
	for i := 0; i < loop-1; i++ {
		buffer.WriteString(",")
		buffer.WriteString(bound[loc[0]:loc[1]])
	}
	buffer.WriteString(bound[loc[1]:])
	return buffer.String()
}

// bindArray binds a named parameter query with fields from an array or slice of
// structs argument.
func bindArray(query string, arg interface{}, m *reflectx.Mapper) (string, []interface{}, error) {
	// we can rebind it at the end.
	bound, names, err := compileNamedQuery([]byte(query))
	if err != nil {
		return "", []interface{}{}, err
	}
	arrayValue := reflect.ValueOf(arg)
	arrayLen := arrayValue.Len()
	if arrayLen == 0 {
		return "", []interface{}{}, fmt.Errorf("length of array is 0: %#v", arg)
	}
	var arglist = make([]interface{}, 0, len(names)*arrayLen)
	for i := 0; i < arrayLen; i++ {
		elemArglist, err := bindAnyArgs(names, arrayValue.Index(i).Interface(), m)
		if err != nil {
			return "", []interface{}{}, err
		}
		arglist = append(arglist, elemArglist...)
	}
	if arrayLen > 1 {
		bound = fixBound(bound, arrayLen)
	}
	return bound, arglist, nil
}

// bindMap binds a named parameter query with a map of arguments.
func bindMap(query string, args map[string]interface{}) (string, []interface{}, error) {
	bound, names, err := compileNamedQuery([]byte(query))
	if err != nil {
		return "", []interface{}{}, err
	}

	arglist, err := bindMapArgs(names, args)
	return bound, arglist, err
}

// -- Compilation of Named Queries

// Allow digits and letters in bind params;  additionally runes are
// checked against underscores, meaning that bind params can have be
// alphanumeric with underscores.  Mind the difference between unicode
// digits and numbers, where '5' is a digit but '五' is not.
var allowedBindRunes = []*unicode.RangeTable{unicode.Letter, unicode.Digit}

// FIXME: this function isn't safe for unicode named params, as a failing test
// can testify.  This is not a regression but a failure of the original code
// as well.  It should be modified to range over runes in a string rather than
// bytes, even though this is less convenient and slower.  Hopefully the
// addition of the prepared NamedStmt (which will only do this once) will make
// up for the slightly slower ad-hoc NamedExec/NamedQuery.

// compile a NamedQuery into an unbound query (using the '?' bindvar) and
// a list of names.
func compileNamedQuery(qs []byte) (sql string, names []string, err error) {
	names = make([]string, 0, 10)
	rebound := make([]byte, 0, len(qs))

	inName := false
	last := len(qs) - 1
	currentVar := 1
	name := make([]byte, 0, 10)

	for i, b := range qs {
		// a ':' while we're in a name is an error
		if b == ':' {
			// if this is the second ':' in a '::' escape sequence, append a ':'
			if inName && i > 0 && qs[i-1] == ':' {
				rebound = append(rebound, ':')
				inName = false
				continue
			} else if inName {
				err = errors.New("unexpected `:` while reading named param at " + strconv.Itoa(i))
				return sql, names, err
			}
			inName = true
			name = []byte{}
		} else if inName && i > 0 && b == '=' && len(name) == 0 {
			rebound = append(rebound, ':', '=')
			inName = false
			continue
			// if we're in a name, and this is an allowed character, continue
		} else if inName && (unicode.IsOneOf(allowedBindRunes, rune(b)) || b == '_' || b == '.') && i != last {
			// append the byte to the name if we are in a name and not on the last byte
			name = append(name, b)
			// if we're in a name and it's not an allowed character, the name is done
		} else if inName {
			inName = false
			// if this is the final byte of the string and it is part of the name, then
			// make sure to add it to the name
			if i == last && unicode.IsOneOf(allowedBindRunes, rune(b)) {
				name = append(name, b)
			}
			// add the string representation to the names list
			names = append(names, string(name))
			// add a proper bindvar for the bindType
			rebound = append(rebound, '$')
			for _, b := range strconv.Itoa(currentVar) {
				rebound = append(rebound, byte(b))
			}
			currentVar++

			// add this byte to string unless it was not part of the name
			if i != last {
				rebound = append(rebound, b)
			} else if !unicode.IsOneOf(allowedBindRunes, rune(b)) {
				rebound = append(rebound, b)
			}
		} else {
			// this is a normal byte and should just go onto the rebound query
			rebound = append(rebound, b)
		}
	}
	return string(rebound), names, err
}

func bindNamedMapper(sql string, arg interface{}, m *reflectx.Mapper) (string, []interface{}, error) {
	t := reflect.TypeOf(arg)
	k := t.Kind()
	switch {
	case k == reflect.Map && t.Key().Kind() == reflect.String:
		m, ok := convertMapStringInterface(arg)
		if !ok {
			return "", nil, fmt.Errorf("pqlx.bindNamedMapper: unsupported map type: %T", arg)
		}
		return bindMap(sql, m)
	case k == reflect.Array || k == reflect.Slice:
		return bindArray(sql, arg, m)
	default:
		return bindStruct(sql, arg, m)
	}
}

// NamedQuery binds a named query and then runs Query on the result using the
// provided Ext (pglx.Tx, pglx.Db).  It works with both structs and with
// map[string]interface{} types.
func NamedQuery(ctx context.Context, e Ext, sql string, arg interface{}) (*Rows, error) {
	q, args, err := bindNamedMapper(sql, arg, mapperFor(e))
	if err != nil {
		return nil, err
	}
	return e.QueryX(ctx, q, args...)
}

// NamedExec uses BindStruct to get a query executable by the driver and
// then runs Exec on the result.  Returns an error from the binding
// or the query execution itself.
func NamedExec(ctx context.Context, e Ext, sql string, arg interface{}) (pgconn.CommandTag, error) {
	q, args, err := bindNamedMapper(sql, arg, mapperFor(e))
	if err != nil {
		return nil, err
	}
	return e.Exec(ctx, q, args...)
}
