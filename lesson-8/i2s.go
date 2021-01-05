package main

import (
	"fmt"
	"reflect"
)

func validateParams(data interface{}, out interface{}) error {
	// Check out type
	outType := reflect.ValueOf(out).Type().Kind()
	if outType != reflect.Slice && outType != reflect.Ptr {
		return fmt.Errorf("out field is not a pointer type")
	}

	// Check source type and out type under pointer
	dataKind := reflect.TypeOf(data).Kind()
	outKind := reflect.ValueOf(out).Elem().Kind()
	if dataKind == reflect.Slice && outKind != reflect.Slice {
		return fmt.Errorf("source and result types do not match")
	}
	if dataKind == reflect.Map && outKind != reflect.Struct {
		return fmt.Errorf("source and result types do not match")
	}
	if dataKind != reflect.Map && dataKind != reflect.Slice {
		return fmt.Errorf("invalid source type")
	}
	if outKind != reflect.Struct && outKind != reflect.Slice {
		return fmt.Errorf("invalid result type")
	}
	return nil
}

func copySimple() error {
	return nil
}

func i2s(data interface{}, out interface{}) error {
	if err := validateParams(data, out); err != nil {
		return err
	}
	switch data.(type) {
	case []map[string]interface{}:
		slice := data.([]map[string]interface{})
		for _, item := range slice {
			fmt.Println(item)
		}
	case map[string]interface{}:
		source := data.(map[string]interface{})
		res := reflect.ValueOf(out).Elem()
		for i := 0; i < res.NumField(); i++ {
			sourceVal, exist:= source[res.Type().Field(i).Name]
			if !exist {
				return fmt.Errorf("field %+v not found", res.Type().Field(i).Name)
			}

			expectedType := res.Type().Field(i).Type.Kind()
			sourceType := reflect.TypeOf(sourceVal).Kind()

			if expectedType == reflect.Struct || expectedType == reflect.Slice {
				e := reflect.New(res.Field(i).Type()).Interface()
				if err := i2s(sourceVal, e); err != nil {
					return err
				}
				res.Field(i).Set(reflect.ValueOf(e).Elem())
				continue
			}

			if expectedType == reflect.Int && sourceType == reflect.Float64 {
				val := sourceVal.(float64)
				sourceVal = int(val)
				sourceType = reflect.Int
			}

			if expectedType != sourceType {
				return fmt.Errorf("invalid type for field %+v, got: %+v, expected: %+v",
					res.Type().Field(i).Name, sourceType, expectedType)
			}
			res.Field(i).Set(reflect.ValueOf(sourceVal))
		}
	}

	return nil
}

type Test struct {
	ID int
	Field string
}

func main() {
	var test interface{}
	test = map[string]interface{}{
		"ID":  100,
		"Field": "test",
	}
	t := Test{}
	if err := i2s(test, &t); err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println(t)
}
