package utils

import (
	"fmt"
	"reflect"
	"strings"
)

func PrepareSuffix(columns interface{}) string {
	v := reflect.ValueOf(columns)
	switch v.Kind() {
	case reflect.String:
		return fmt.Sprintf("RETURNING %s", v.Interface())
	case reflect.Slice:
		var vCol []string
		for i := 0; i < v.Len(); i++ {
			vCol = append(vCol, v.Index(i).Interface().(string))
		}
		if v.Len() > 0 {
			return fmt.Sprintf("RETURNING (%s)", strings.Join(vCol, ","))
		}
		return "RETURNING *"
	}
	return "RETURNING *"
}
