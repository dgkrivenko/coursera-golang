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

func copySimple(data interface{}, out interface{}) error {
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
			val := reflect.New(res.Field(i).Type()).Interface()
			if err := i2s(sourceVal, val); err != nil {
				return err
			}
			res.Field(i).Set(reflect.ValueOf(val).Elem())
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
	return nil
}

func i2s(data interface{}, out interface{}) error {
	if err := validateParams(data, out); err != nil {
		return err
	}
	switch data.(type) {
	case []interface{}:
		arr := data.([]interface{})
		for _, item := range arr {
			t := reflect.New(reflect.TypeOf(out).Elem().Elem()).Interface()
			if err := copySimple(item, t); err != nil {
				return err
			}
			reflect.ValueOf(out).Elem().Set(reflect.Append(reflect.ValueOf(out).Elem(), reflect.ValueOf(t).Elem()))
		}
	case map[string]interface{}:
		if err := copySimple(data, out); err != nil {
			return err
		}
	}
	return nil
}