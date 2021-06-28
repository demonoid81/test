package models

import "reflect"

//func prepareKey(
//	vDstField, vSrcField reflect.Value,
//	fieldType reflect.StructField,
//	columns, parent map[string]interface{},
//	) error {
//	// получим все ключи
//	dbKey := fieldType.Tag.Get("db")
//	relayKey := fieldType.Tag.Get("relay")
//	// Ключ автогенерации ключа
//	autoGen := fieldType.Tag.Get("auto")
//	autoGenKey := true
//	if autoGen == "false" {
//		autoGenKey = false
//	}
//	switch vDstField.Kind() {
//	case reflect.Struct:
//		// Если работает со времем
//		if vDstField.Type() == typeOfTime {
//			if (!vDstField.IsNil() && vSrcField.IsNil()) ||
//				(!vSrcField.IsNil() && vDstField.Interface().(time.Time) != vSrcField.Interface().(time.Time)) {
//				columns[dbKey] = vSrcField.Interface()
//				return nil
//			}
//		}
//		// все что имеет связи надо проверить
//		if relayKey != "" {
//			// Отклонируетм родителя
//			clonedParent := make(map[string]interface{})
//			for key, value := range parent {
//				clonedParent[key] = value
//			}
//			// Проверяем вложеную структуру
//			value, err := utils.Invoke(vDstField, "Mutation", ctx, db, app, "uuid", clonedParent)
//			if err != nil {
//				app.Logger.Error().Str("module", "models").Str("func", "SqlGenKeys").Err(err).Msg("Error mutation Invoke")
//				return err
//			}
//			if !value[2].IsNil() {
//				return errors.New("Ошибко")
//			}
//			// закроем колонки на всякий случай они нам не нужны
//			if !value[0].IsNil() {
//				rows := value[0].Interface().(*pglx.Rows)
//				rows.Close()
//			}
//			if !value[1].IsNil() && vSrcField.IsNil() {
//				columns[relayKey] = value[1].Interface()
//			}
//		}
//	default:
//		// если значения не равны то обновим его
//		if (!vDstField.IsNil() && vSrcField.IsNil()) ||
//			(!vSrcField.IsNil() && vDstField.Interface() != vSrcField.Interface()) {
//			columns[dbKey] = vDstField.Interface()
//		}
//	}
//	return nil
//}
//
//func compareStruct(src, dsc interface{}, columns, parent map[string]interface{}) error {
//	vDst := reflect.ValueOf(dsc)
//	if vDst.Kind() == reflect.Ptr {
//		vDst = reflect.ValueOf(dsc).Elem()
//	}
//	vSrc := reflect.ValueOf(src)
//	if vSrc.Kind() == reflect.Ptr {
//		vSrc = reflect.ValueOf(src).Elem()
//	}
//	for i := 0; i < vDst.NumField(); i++ {
//		if err := prepareKey(vDst.Field(i), vSrc.Field(i),
//			vDst.Type().Field(i),
//			columns, parent); err != nil {
//			return err
//		}
//	}
//	return nil
//}

func Compare(dsc interface{}, columns map[string]interface{}) bool {
	vDst := reflect.ValueOf(dsc)
	if vDst.Kind() == reflect.Ptr {
		vDst = reflect.ValueOf(dsc).Elem()
	}
	count := len(columns)
	for i := 0; i < vDst.NumField(); i++ {
		key := vDst.Type().Field(i).Tag.Get("db")
		if result, ok := columns[key]; ok {
			if reflect.DeepEqual(result, vDst.Field(i).Interface()) {
				count--
				if count == 0 {
					return true
				}
			}
		}
	}
	return false
}
