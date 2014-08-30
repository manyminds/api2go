package api2go

import (
	"encoding/json"
	"errors"
	"reflect"
)

type unmarshalContext map[string]interface{}

// Unmarshal reads a JSONAPI map to a model struct
func Unmarshal(ctx unmarshalContext, values interface{}) error {
	// Check that target is a *[]Model
	ptrVal := reflect.ValueOf(values)
	if ptrVal.Kind() != reflect.Ptr || ptrVal.IsNil() {
		panic("You must pass a pointer to a []struct to Unmarshal()")
	}
	sliceType := reflect.TypeOf(values).Elem()
	sliceVal := ptrVal.Elem()
	if sliceType.Kind() != reflect.Slice {
		panic("You must pass a pointer to a []struct to Unmarshal()")
	}
	structType := sliceType.Elem()
	if structType.Kind() != reflect.Struct {
		panic("You must pass a pointer to a []struct to Unmarshal()")
	}

	// Read models slice
	rootName := pluralize(jsonify(structType.Name()))
	var modelsInterface interface{}
	if modelsInterface = ctx[rootName]; modelsInterface == nil {
		return errors.New("Expected root document to include a '" + rootName + "' key but it didn't.")
	}
	models, ok := modelsInterface.([]interface{})
	if !ok {
		return errors.New("Expected slice under key '" + rootName + "'")
	}

	// Read all the models
	for _, m := range models {
		attributes, ok := m.(map[string]interface{})
		if !ok {
			return errors.New("Expected an array of objects under key '" + rootName + "'")
		}

		val := reflect.New(structType).Elem()
		for k, v := range attributes {
			fieldName := dejsonify(k)
			field := val.FieldByName(fieldName)
			if !field.IsValid() {
				return errors.New("Expected struct " + structType.Name() + " to have field " + fieldName)
			}
			field.Set(reflect.ValueOf(v))
		}

		sliceVal.Set(reflect.Append(sliceVal, val))
	}

	return nil
}

// UnmarshalJSON reads a JSONAPI compatible JSON document to a model struct
func UnmarshalJSON(data []byte, values interface{}) error {
	var ctx unmarshalContext
	err := json.Unmarshal(data, &ctx)
	if err != nil {
		return err
	}
	return Unmarshal(ctx, values)
}
