package models

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/utils"
	"github.com/sphera-erp/sphera/pkg/pglx"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/sphera-erp/sphera/pkg/strcase"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

var typeOfUUID = reflect.TypeOf(uuid.Nil)
var typeOfTime = reflect.TypeOf(time.Time{})

func uniqueUUID(intSlice []*uuid.UUID) []*uuid.UUID {
	keys := make(map[uuid.UUID]bool)
	var list []*uuid.UUID
	for _, entry := range intSlice {
		if _, value := keys[*entry]; !value {
			keys[*entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func prepareKey(
	ctx context.Context,
	app *app.App,
	db pglxqb.BaseRunner,
	parentField, field reflect.Value,
	fieldType reflect.StructField,
	key, relayKey string,
	autoGenKey bool,
	columns, parent map[string]interface{}) error {
	var err error
	switch field.Kind() {
	case reflect.Ptr:
		if !field.IsNil() {
			return prepareKey(ctx, app, db, parentField, field.Elem(), fieldType, key, relayKey, autoGenKey, columns, parent)
		}
	case reflect.Struct:
		if relayKey != "" {
			//var vFieldUUID reflect.Value
			//var link reflect.Value
			//if parentField.IsValid() {
			//	link = parentField.FieldByName(fieldType.Tag.Get("link"))
			//	for i := 0; i < field.NumField(); i++ {
			//		if field.Type().Field(i).Tag.Get("json") == "uuid" {
			//			vFieldUUID = field.Field(i)
			//		}
			//	}
			//}
			//if link.IsValid() && vFieldUUID.IsValid() && !link.IsNil() && !vFieldUUID.IsNil() && reflect.DeepEqual(link.Interface(),  vFieldUUID.Interface()) {
			//	fmt.Println("Changed.LINK_UUID")
			//	columns[relayKey] = vFieldUUID.Interface()
			//} else {
			mutationColumns := make(map[string]interface{})
			for key, value := range parent {
				mutationColumns[key] = value
			}
			value, err := utils.Invoke(field, "Mutation", ctx, db, app, "uuid", mutationColumns)
			if err != nil {
				app.Logger.Error().Str("module", "models").Str("func", "SqlGenKeys").Err(err).Msg("Error mutation Invoke")
				return err
			}
			// todo Исправить
			if !value[2].IsNil() {
				//fmt.Println(value[2].Interface().(*gqlerror.Error))
				return value[2].Interface().(*gqlerror.Error)
			}
			// закроем колонки на всякий случай они нам не нужны
			if !value[0].IsNil() {
				fmt.Println("close rows")
				rows := value[0].Interface().(*pglx.Rows)
				rows.Close()
			}
			if !value[1].IsNil() {
				columns[relayKey] = value[1].Interface()
			}
			//}
		} else {
			if field.Type() == typeOfTime {
				columns[key] = field.Interface()
			}
		}
	case reflect.Slice, reflect.Array:
		if field.Len() > 0 {
			if field.Type() == typeOfUUID {
				if !autoGenKey {
					columns[key] = field.Interface().(uuid.UUID)
				}
			} else if relayKey != "" {
				slice := make(map[string]interface{})

				for i := 0; i < field.Len(); i++ {
					err = prepareKey(ctx, app, db, reflect.Value{}, field.Index(i), reflect.StructField{}, "", fmt.Sprintf("slice%d", i), false, slice, parent)
					if err != nil {
						app.Logger.Error().Str("module", "models").Str("func", "SqlGenKeys").Err(err).Msg("Error prepare slice key")
						return err
					}
				}
				var arrayUUID []*uuid.UUID
				for _, value := range slice {
					arrayUUID = append(arrayUUID, value.(*uuid.UUID))
				}
				columns[relayKey] = uniqueUUID(arrayUUID)
			}
		}
	case reflect.Map:
		return nil
	default:
		if relayKey == "" {
			columns[key] = field.Interface()
		}
	}
	return nil
}

func SqlGenKeys(ctx context.Context, app *app.App, db pglxqb.BaseRunner, i interface{}, columns, parent map[string]interface{}) (map[string]interface{}, error) {
	var err error
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if !v.IsValid() {
		return nil, nil
	}
	for i := 0; i < v.NumField(); i++ {
		//get key for struct tag
		key := v.Type().Field(i).Tag.Get("db")
		relayKey := v.Type().Field(i).Tag.Get("relay")
		autoGen := v.Type().Field(i).Tag.Get("auto")
		autoGenKey := true
		if autoGen == "false" {
			autoGenKey = false
		}
		err = prepareKey(ctx, app, db, v, v.Field(i), v.Type().Field(i), key, relayKey, autoGenKey, columns, parent)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "SqlGenKeys").Err(err).Msg("Error mutation Contact")
			return nil, err
		}
	}
	return columns, nil
}

func structToDBTable(field reflect.Value) string {
	switch field.Interface().(type) {
	case MedicalBook:
		return "medical_books"
	case Organization:
		return "organizations"
	case JobType:
		return "job_types"
	case Person:
		return "persons"
	case Address:
		return "addresses"
	case Region:
		return "regions"
	case Area:
		return "areas"
	case City:
		return "cities"
	case CityDistrict:
		return "city_districts"
	case Settlement:
		return "settlements"
	case JobTemplate:
		return "job_templates"
	case Course:
		return "courses"
	case Job:
		return "jobs"
	}
	return ""
}

func (p *Person) InnerSelect(ctx context.Context, app *app.App, db pglxqb.BaseRunner, table string, object interface{}) ([]uuid.UUID, error) {
	logger := app.Logger.Error().Str("module", "persons").Str("func", "Persons")
	sql := pglxqb.Select(fmt.Sprintf("%s.uuid", table)).From(table)
	result, sql, err := SqlGenSelectKeys(object, sql, table, 1)
	if err != nil {
		logger.Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	if len(result) > 0 {
		sql = sql.Where(pglxqb.Eq(result))
	}
	// }
	rows, err := sql.RunWith(db).Query(ctx)
	if err != nil {
		logger.Err(err).Msg("Error select object")
		return nil, gqlerror.Errorf("Error select object")
	}
	var arrayUuid []uuid.UUID
	for rows.Next() {
		var u uuid.UUID
		if err := rows.Scan(&u); err != nil {
			logger.Err(err).Msg("Error scan response to uuid")
			return nil, gqlerror.Errorf("Error scan response to uuid")
		}

		arrayUuid = append(arrayUuid, u)
	}
	return arrayUuid, nil
}

func prepareSelectKey(field reflect.Value, fieldType reflect.StructField, parent string, iterator int, sql pglxqb.SelectBuilder, columns map[string]interface{}) (map[string]interface{}, pglxqb.SelectBuilder, error) {
	var err error
	switch field.Kind() {
	case reflect.Ptr:
		if !field.IsNil() {
			return prepareSelectKey(field.Elem(), fieldType, parent, iterator, sql, columns)
		}
	case reflect.Struct:
		relay := structToDBTable(field)
		key := utils.RandStringBytesMaskImpr(8)
		sql = sql.LeftJoin(
			fmt.Sprintf("%s as %s_%s on %s_%s.uuid=%s.%s",
				relay, relay, key, relay, key, parent, fieldType.Tag.Get("relay")))
		fmt.Println(sql.ToSql())
		var res map[string]interface{}
		res, sql, err = SqlGenSelectKeys(
			field.Interface(),
			sql,
			fmt.Sprintf("%s_%s", relay, key),
			iterator+1)
		if err != nil {
			return nil, pglxqb.SelectBuilder{}, err
		}
		columns = mergeKeys(columns, res)
	case reflect.Slice, reflect.Array:
		if field.Type() == typeOfUUID {
			key := fieldType.Tag.Get("db")
			argName := fmt.Sprintf("%s.%s", parent, key)
			columns[argName] = field.Interface()
		} else {
			// obj := reflect.New(field.Type().Elem().Elem())
			// relay := structToDBTable(obj)
			// sql = sql.LeftJoin(
			// 	fmt.Sprintf("%s as %s%d on %s%d.uuid=%s.%s",
			// 		relay, relay, iterator, relay, iterator, parent, fieldType.Tag.Get("relay")))
			// for j := 0; j < field.Len(); j++ {
			// 	vField := field.Index(j)
			// 	value, err := InnerSelect(ctx, db, app, relay, vField.Interface())
			// }
			// нам надо сгенерировать запросы и получить результат
			// Этого не может быть
			fmt.Println("Type: Slice ", fieldType.Tag.Get("json"), fieldType.Tag.Get("db"), field.Interface())
		}
	case reflect.Map:
		return columns, sql, nil
	default:
		key := fieldType.Tag.Get("db")
		argName := fmt.Sprintf("%s.%s", parent, key)
		switch field.Interface().(type) {
		case CourseType:
			if field.Interface().(CourseType).String() != "" {
				columns[argName] = field.Interface()
			}
		default:
			columns[argName] = field.Interface()
		}
	}
	return columns, sql, nil
}

func SqlGenSelectKeys(i interface{}, sql pglxqb.SelectBuilder, parent string, iterator int) (map[string]interface{}, pglxqb.SelectBuilder, error) {
	var err error
	columns := make(map[string]interface{})
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if !v.IsValid() {
		return nil, sql, nil
	}
	for i := 0; i < v.NumField(); i++ {
		//get key for struct tag
		columns, sql, err = prepareSelectKey(v.Field(i), v.Type().Field(i), parent, iterator, sql, columns)
		if err != nil {
			return nil, pglxqb.SelectBuilder{}, err
		}
	}
	return columns, sql, nil
}

func restoreStructReflect(
	ctx context.Context,
	app *app.App,
	db pglxqb.BaseRunner,
	restoreStruct reflect.Value,
	field reflect.Value,
	fieldType reflect.StructField) error {
	switch field.Kind() {
	case reflect.Ptr:
		if !field.IsNil() {
			return restoreStructReflect(ctx, app, db, restoreStruct, field.Elem(), fieldType)
		}
	case reflect.Struct:
		// получим линк
		// тут у нас проеб, у нас есть старая ссылка и новая, нужно оставить новую а старую убрать
		if fieldType.Tag.Get("link") != "" {
			link := restoreStruct.FieldByName(fieldType.Tag.Get("link"))
			if link.IsNil() {
				return nil
			}
			// получили структуру
			value, err := utils.Invoke(field, "GetByUUID", ctx, app, db, link.Interface())
			// восстановили её
			if err != nil {
				fmt.Println(err)
			}
			object, errReflect := value[0], value[1]
			if !errReflect.IsNil() {
				return errReflect.Interface().(error)
			}
			if object.Kind() == reflect.Ptr {
				object = object.Elem()
			}
			for i, n := 0, field.NumField(); i < n; i++ {
				switch field.Field(i).Interface().(type) {
				case *uuid.UUID:
					if field.Field(i).IsNil() && !object.Field(i).IsNil() {
						field.Field(i).Set(object.Field(i))
					}
					//field.Field(i).Set(object.Field(i))
				}
			}
		}
	case reflect.Slice:
		if fieldType.Tag.Get("link") != "" {
			fmt.Println(fieldType.Tag.Get("link"))
			link := restoreStruct.FieldByName(fieldType.Tag.Get("link"))
			if link.IsNil() {
				return nil
			}

			for i := 0; i < link.Len(); i++ {
				linkUUID := link.Index(i)
				if linkUUID.Kind() == reflect.Ptr {
					linkUUID = link.Index(i).Elem()
				}
				updated := false
				for j := 0; j < field.Len(); j++ {
					vField := field.Index(j)
					if vField.Kind() == reflect.Ptr {
						vField = field.Index(j).Elem()
					}
					vFieldUUID := vField.FieldByName("UUID")
					if vFieldUUID.Interface() == linkUUID.Interface() {
						// запросим обьект
						value, err := utils.Invoke(vField, "GetByUUID", ctx, app, db, linkUUID.Interface())
						// востановили её
						if err != nil {
							fmt.Println(err)
						}
						object, errReflect := value[0], value[1]
						if !errReflect.IsNil() {
							return errReflect.Interface().(error)
						}
						if object.Kind() == reflect.Ptr {
							object = object.Elem()
						}
						updated = true
						if err := utils.Merge(object, vField); err != nil {
							return nil
						}
						break
					}
				}
				if !updated {
					// создадим пустой обьект с UUID
					obj := reflect.New(field.Type().Elem().Elem())
					if obj.Kind() == reflect.Ptr {
						obj.Elem().FieldByName("UUID").Set(link.Index(i))
					} else {
						obj.FieldByName("UUID").Set(link.Index(i))
					}
					field.Set(reflect.Append(field, obj))
				}
			}
		}
	}
	return nil
}

func SetField(obj reflect.Value, name string, value reflect.Value) error {
	if value.IsNil() {
		return fmt.Errorf("Nil Value field: %s in map", name)
	}
	fieldVal := obj.FieldByName(strcase.ToCamel(name))
	if !fieldVal.IsValid() {
		fmt.Println("notValid")
		return fmt.Errorf("No such field: %s in obj", name)
	}
	if !fieldVal.CanSet() {
		fmt.Println("NotCanSet")
		return fmt.Errorf("Cannot set %s field value", name)
	}
	val := value
	if fieldVal.Type() != val.Elem().Type() {
		if value.Elem().Kind() == reflect.Map {
			// if field value is struct
			if fieldVal.Kind() == reflect.Struct {
				return FillStruct(val.Elem(), fieldVal.Addr())
			}
			// if field value is a pointer to struct
			if fieldVal.Kind() == reflect.Ptr && fieldVal.Type().Elem().Kind() == reflect.Struct {
				if fieldVal.IsNil() {
					fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
				}
				// fmt.Printf("recursive: %v %v\n", m,fieldVal.Interface())
				return FillStruct(val.Elem(), fieldVal)
			}
		}
		return fmt.Errorf("Provided value type didn't match obj field type")
	}
	fieldVal.Set(val.Elem())
	return nil

}

func FillStruct(m reflect.Value, s reflect.Value) error {
	l := s
	if s.Kind() == reflect.Ptr {
		l = s.Elem()
	}
	for _, v := range m.MapKeys() {
		k := m.MapIndex(v)
		fmt.Println(v, k)
		_ = SetField(l, v.Interface().(string), k)
		//if err != nil {
		//	return err
		//}
	}
	return nil
}

func parseRequestedFields(
	ctx context.Context,
	app *app.App,
	db pglxqb.BaseRunner,
	fields []graphql.CollectedField,
	field interface{}) error {
	vField := reflect.ValueOf(field)
	if vField.Kind() == reflect.Ptr {
		vField = reflect.ValueOf(field).Elem()
	}
	for _, column := range fields {
		if len(column.Selections) > 0 {
			// получил имя колонки
			columnName := column.Name
			for i := 0; i < vField.NumField(); i++ {
				key := vField.Type().Field(i).Tag.Get("json")
				// fmt.Println(columnName, " - ", key)
				if columnName == key {
					// получим линку на структуру в базе

					link := vField.FieldByName(vField.Type().Field(i).Tag.Get("link"))
					// fmt.Println(link)
					// работает только с линками
					if !link.IsNil() {
						// fmt.Println("link", link, " king: ", link.Kind())
						// а как со слайсами )))
						switch link.Kind() {
						case reflect.Slice:
							// пока такой костыль, других не знаю так получить зависимости
							obj := reflect.New(vField.Field(i).Type().Elem().Elem())
							value, err := utils.Invoke(obj, "GetParsedObjectsByUUID", ctx, app, db, link.Interface(), column)
							if err != nil {
								app.Logger.Error().Str("module", "models").Str("func", "SqlGenKeys").Err(err).Msg("Error mutation ContactType")
								return err
							}
							object, errReflect := value[0], value[1]
							if !errReflect.IsNil() {
								return errReflect.Interface().(error)
							}
							vField.Field(i).Set(object)
						case reflect.Map:
							fmt.Println(vField.Field(i).Type().Elem())
							object := reflect.New(vField.Field(i).Type().Elem())
							err := FillStruct(link, object)
							if err != nil {
								app.Logger.Error().Str("module", "models").Str("func", "SqlGenKeys").Err(err).Msg("Error mutation ContactType")
								return err
							}
							fmt.Println(object)
							vField.Field(i).Set(object)
						default:
							value, err := utils.Invoke(vField.Field(i), "GetParsedObjectByUUID", ctx, app, db, link.Interface(), column)
							if err != nil {
								app.Logger.Error().Str("module", "models").Str("func", "SqlGenKeys").Err(err).Msg("Error mutation ContactType")
								return err
							}
							object, errReflect := value[0], value[1]
							if !errReflect.IsNil() {
								return errReflect.Interface().(error)
							}
							vField.Field(i).Set(object)
						}
					}
				}
			}
		}
	}
	return nil
}

type m = map[string]interface{}

func mergeKeys(left, right m) m {
	for key, rightVal := range right {
		if leftVal, present := left[key]; present {
			//then we don't want to replace it - recurse
			left[key] = mergeKeys(leftVal.(m), rightVal.(m))
		} else {
			// key not in left so we can just shove it in
			left[key] = rightVal
		}
	}
	return left
}
