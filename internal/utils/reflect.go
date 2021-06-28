package utils

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"math/rand"
	"reflect"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandStringBytesMaskImpr(n int) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func SliceRemoveItem(slice reflect.Value, i int) {
	v := slice
	if slice.Kind() == reflect.Ptr {
		v = slice.Elem()
	}
	v.Set(reflect.AppendSlice(v.Slice(0, i), v.Slice(i+1, v.Len())))
}

// Invoke - firstResult, err := invoke(AnyStructInterface, MethodName, Params...)
func Invoke(any reflect.Value, name string, args ...interface{}) ([]reflect.Value, error) {
	var method reflect.Value
	if any.Kind() == reflect.Ptr {
		method = any.MethodByName(name)
	} else {
		method = any.Addr().MethodByName(name)
	}
	methodType := method.Type()
	numIn := methodType.NumIn()
	if numIn > len(args) {
		return nil, fmt.Errorf("Method %s must have minimum %d params. Have %d", name, numIn, len(args))
	}
	if numIn != len(args) && !methodType.IsVariadic() {
		return nil, fmt.Errorf("Method %s must have %d params. Have %d", name, numIn, len(args))
	}
	in := make([]reflect.Value, len(args))
	for i := 0; i < len(args); i++ {
		var inType reflect.Type
		if methodType.IsVariadic() && i >= numIn-1 {
			inType = methodType.In(numIn - 1).Elem()
		} else {
			inType = methodType.In(i)
		}
		argValue := reflect.ValueOf(args[i])
		if !argValue.IsValid() {
			return nil, fmt.Errorf("Method %s. Param[%d] must be %s. Have %s", name, i, inType, argValue.String())
		}
		argType := argValue.Type()
		if argType.ConvertibleTo(inType) {
			in[i] = argValue.Convert(inType)
		} else {
			return nil, fmt.Errorf("Method %s. Param[%d] must be %s. Have %s", name, i, inType, argType)
		}
	}
	return method.Call(in), nil
}

func Merge(dst, src reflect.Value) error {
	vDst := dst
	if vDst.Kind() == reflect.Ptr {
		vDst = vDst.Elem()
	}
	vSrc := src
	// We check if vSrc is a pointer to dereference it.
	if vSrc.Kind() == reflect.Ptr {
		vSrc = vSrc.Elem()
	}
	if vDst.Type() != vSrc.Type() {
		return errors.New("src and dst must be of same type")
	}
	for i := 0; i < vDst.NumField(); i++ {
		switch vDst.Field(i).Kind() {
		case reflect.Struct:
			fmt.Println("This is struct")
		default:
			if !vSrc.Field(i).IsNil() {
				vDst.Field(i).Set(vSrc.Field(i))
			}
		}
	}
	vSrc.Set(vDst)
	return nil
}

// Востановим все ссылки и заместим их в Исходнике
func RestoreUUID(dst, src interface{}) {
	var vDst, vSrc reflect.Value
	vDst = reflect.ValueOf(dst)
	if vDst.Kind() == reflect.Ptr {
		vDst = vDst.Elem()
	}
	vSrc = reflect.ValueOf(src)
	if vSrc.Kind() == reflect.Ptr {
		vSrc = vSrc.Elem()
	}
	for i, n := 0, vDst.NumField(); i < n; i++ {
		switch vDst.Field(i).Interface().(type) {
		case *uuid.UUID, []*uuid.UUID:
			if vDst.Field(i).IsNil() && !vSrc.Field(i).IsNil() {
				vDst.Field(i).Set(vSrc.Field(i))
			}
		}
	}
}

// сверяет поля структуры и поля запроса, уберает лишние
func ClearSQLFields(v interface{}, columns map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	vField := reflect.ValueOf(v)
	if vField.Kind() == reflect.Ptr {
		vField = reflect.ValueOf(v).Elem()
	}
	for key, value := range columns {
		for i := 0; i < vField.NumField(); i++ {
			vKey := vField.Type().Field(i).Tag.Get("db")
			if vKey == key {
				result[key] = value
			}
		}
	}
	return result
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	default:
		return v.IsNil()
	}
}

// сверяет поля структуры и поля запроса, уберает лишние
func CountFillFields(v interface{}) int {
	result := 0
	vField := reflect.ValueOf(v)
	if vField.Kind() == reflect.Ptr {
		vField = reflect.ValueOf(v).Elem()
	}
	for i := 0; i < vField.NumField(); i++ {
		if !isEmptyValue(vField.Field(i)) {
			result++
		}
	}
	return result
}

// TODO: итератор множественой вложености как делать

func expr(table, expr, key string, value reflect.Value, iterator string, sql pglxqb.SelectBuilder) (pglxqb.SelectBuilder, pglxqb.QB) {
	field := fmt.Sprintf("%s_%s.%s", table, iterator, key)
	if iterator == "" {
		field = fmt.Sprintf("%s.%s", table, key)
	}
	switch expr {
	case "gt":
		return sql, pglxqb.Gt{field: value.Interface()}
	case "gte":
		return sql, pglxqb.GtOrEq{field: value.Interface()}
	case "lt":
		return sql, pglxqb.Lt{field: value.Interface()}
	case "lte":
		return sql, pglxqb.LtOrEq{field: value.Interface()}
	case "not":
		return sql, pglxqb.NotEq{field: value.Interface()}
	case "and": // todo: есть ли смысл в Этом операторе
		var andQb pglxqb.And
		for i := 0; i < value.Len(); i++ {
			var qb pglxqb.QB
			sql, qb = filterBuilder(table, sql, value.Index(i), iterator)
			andQb = append(andQb, qb)
		}
		return sql, andQb
	case "or":
		var orQb pglxqb.Or
		for i := 0; i < value.Len(); i++ {
			// тут у нас структура у нее тоже могут быть фильтры
			var qb pglxqb.QB
			sql, qb = filterBuilder(table, sql, value.Index(i), iterator)
			orQb = append(orQb, qb)
		}
		return sql, orQb
	case "between":
		return sql, pglxqb.Between{Field: field, X: value.Index(0).Interface(), Y: value.Index(1).Interface()}
	// todo это лишние но надо тогда разобраться с возвратом
	default:
		return sql, pglxqb.Eq{field: value.Interface()}
	}
}

func fieldFilter(field reflect.Value, fieldType reflect.StructField, table, iterator string, sql pglxqb.SelectBuilder) (pglxqb.SelectBuilder, pglxqb.QB) {
	// работаем с полями
	key := fieldType.Tag.Get("db")
	if key != "" {
		vField := field
		if field.Kind() == reflect.Ptr {
			vField = field.Elem()
		}
		if vField.Kind() == reflect.Bool {
			field := fmt.Sprintf("%s_%s.%s", table, iterator, key)
			if iterator == "" {
				field = fmt.Sprintf("%s.%s", table, key)
			}
			return sql, pglxqb.Eq{field: vField.Interface()}
		} else {
			for j := 0; j < vField.NumField(); j++ {
				if vField.Field(j).IsNil() {
					continue
				}
				return expr(table, vField.Type().Field(j).Tag.Get("json"), key, vField.Field(j), iterator, sql)
			}
		}

	}
	//работаем со структурами
	vTable := fieldType.Tag.Get("table")
	if vTable != "" {
		fmt.Println(table)
		randString := RandStringBytesMaskImpr(8)
		if iterator == "" {
			sql = sql.LeftJoin(
				fmt.Sprintf("%s as %s_%s on %s_%s.uuid=%s.%s",
					vTable, vTable, randString, vTable, randString, table, fieldType.Tag.Get("link")))
		} else {
			sql = sql.LeftJoin(
				fmt.Sprintf("%s as %s_%s on %s_%s.uuid=%s_%s.%s",
					vTable, vTable, randString, vTable, randString, table, iterator, fieldType.Tag.Get("link")))
		}
		return filterBuilder(vTable, sql, field, randString)
	}
	// Логические операции
	return expr(table, fieldType.Tag.Get("json"), "", field, iterator, sql)
}

func filterBuilder(table string, sql pglxqb.SelectBuilder, v reflect.Value, iterator string) (pglxqb.SelectBuilder, pglxqb.QB) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	var andQb pglxqb.And
	var qb pglxqb.QB
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).IsNil() {
			continue
		}
		sql, qb = fieldFilter(v.Field(i), v.Type().Field(i), table, iterator, sql)
		andQb = append(andQb, qb)
	}
	if len(andQb) > 1 {
		return sql, andQb
	}
	// вернем массив фильтров
	return sql, qb
}

func ReflectFilter(table string, sql pglxqb.SelectBuilder, filter interface{}) pglxqb.SelectBuilder {
	v := reflect.ValueOf(filter)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).IsNil() {
			continue
		}
		var qb pglxqb.QB
		sql, qb = fieldFilter(v.Field(i), v.Type().Field(i), table, "", sql)
		sql = sql.Where(qb)
	}
	return sql
}
