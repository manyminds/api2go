package api2go

import (
	"encoding/json"
	"reflect"
)

type marshalingContext struct {
	wrapper map[string]interface{}
}

func makeContext() *marshalingContext {
	ctx := &marshalingContext{}
	ctx.wrapper = make(map[string]interface{})
	return ctx
}

// Marshal takes a struct and marshals it to a json encodable interface{} value
func Marshal(val interface{}) (interface{}, error) {
	ctx := makeContext()

	if reflect.TypeOf(val).Kind() == reflect.Slice {
		// Using Elem() here to get the slice's element type
		rootKeyName := Pluralize(Underscorize(reflect.TypeOf(val).Elem().Name()))
		// Panic if empty string, i.e. passed []interface{}
		if rootKeyName == "" {
			panic("You passed a slice of interfaces []interface{}{...} to Marshal. We cannot determine key names from that. Use []YourObjectName{...} instead.")
		}
		// We already have a slice, so just assign it
		ctx.wrapper[rootKeyName] = val
	} else {
		rootKeyName := Pluralize(Underscorize(reflect.TypeOf(val).Name()))
		// We need to put single objects into a slice
		ctx.wrapper[rootKeyName] = []interface{}{val}
	}

	return ctx.wrapper, nil
}

// MarshalToJSON takes a struct and marshals it to JSONAPI compliant JSON
func MarshalToJSON(val interface{}) ([]byte, error) {
	result, err := Marshal(val)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
