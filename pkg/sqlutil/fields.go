package sqlutil

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx/reflectx"
)

const noInsertTag = ",noi"

// GetFields возвращает список полей структуры для SQL-запросов
func GetFields(entity interface{}) ([]string, error) {
	mapper := reflectx.NewMapperFunc("db", strings.ToLower)
	fields := make([]string, 0)

	structMap := mapper.TypeMap(reflect.TypeOf(entity))
	for _, fieldInfo := range structMap.Index {
		if fieldInfo.Name == "-" { // Пропускаем поля с тегом "-"
			continue
		}

		tag := fieldInfo.Field.Tag.Get("db")
		if strings.Contains(tag, noInsertTag) {
			continue
		}

		fields = append(fields, fieldInfo.Name)
	}

	if len(fields) == 0 {
		return nil, fmt.Errorf("no fields found for entity")
	}

	return fields, nil
}

// GetValues возвращает список значений полей структуры для SQL-запросов
func GetValues(entity interface{}) ([]interface{}, error) {
	mapper := reflectx.NewMapperFunc("db", strings.ToLower)
	values := make([]interface{}, 0)

	val := reflect.ValueOf(entity)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	structMap := mapper.TypeMap(val.Type())
	for _, fieldInfo := range structMap.Index {
		if fieldInfo.Name == "-" { // Пропускаем поля с тегом "-"
			continue
		}

		tag := fieldInfo.Field.Tag.Get("db")
		if strings.Contains(tag, noInsertTag) {
			continue
		}

		field := reflectx.FieldByIndexes(val, fieldInfo.Index)
		values = append(values, field.Interface())
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("no values found for entity")
	}

	return values, nil
}
