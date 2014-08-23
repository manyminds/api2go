package api2go

import (
	"encoding/json"
	"reflect"
)

type marshalingContext struct {
	wrapper  map[string]interface{}
	rootName string
}

func makeContext(rootName string) *marshalingContext {
	ctx := &marshalingContext{}
	ctx.rootName = rootName
	ctx.wrapper = map[string]interface{}{}
	ctx.wrapper[rootName] = []interface{}{}
	return ctx
}

// Marshal takes a struct (or slice of structs) and marshals them to a json encodable interface{} value
func Marshal(data interface{}) (interface{}, error) {
	var ctx *marshalingContext

	if reflect.TypeOf(data).Kind() == reflect.Slice {
		// We were passed a slice
		// Using Elem() here to get the slice's element type
		rootName := pluralize(underscorize(reflect.TypeOf(data).Elem().Name()))

		// Panic if empty string, i.e. passed []interface{}
		if rootName == "" {
			panic("You passed a slice of interfaces []interface{}{...} to Marshal. We cannot determine key names from that. Use []YourObjectName{...} instead.")
		}
		ctx = makeContext(rootName)

		// Marshal all elements
		// We iterate using reflections to save copying the slice to a []interface{}
		sliceValue := reflect.ValueOf(data)
		for i := 0; i < sliceValue.Len(); i++ {
			if err := ctx.marshalStruct(sliceValue.Index(i)); err != nil {
				return nil, err
			}
		}
	} else {
		// We were passed a single object
		rootName := pluralize(underscorize(reflect.TypeOf(data).Name()))
		ctx = makeContext(rootName)

		// Marshal the value
		if err := ctx.marshalStruct(reflect.ValueOf(data)); err != nil {
			return nil, err
		}
	}

	return ctx.wrapper, nil
}

// marshalStruct marshals a struct and places it in the context's wrapper
func (ctx *marshalingContext) marshalStruct(val reflect.Value) error {
	result := map[string]interface{}{}

	valType := val.Type()
	for i := 0; i < val.NumField(); i++ {
		result[underscorize(valType.Field(i).Name)] = val.Field(i).Interface()
	}

	ctx.addValue(pluralize(underscorize(valType.Name())), result)
	return nil
}

// addValue adds an object to the context's wrapper
// `name` should be the pluralized and underscorized object type.
func (ctx *marshalingContext) addValue(name string, val map[string]interface{}) {
	if name == ctx.rootName {
		// Root objects are placed directly into the root doc
		// BUG(lucas): If an object links to its own type, linked objects must be placed into the linked map.
		ctx.wrapper[name] = append(ctx.wrapper[name].([]interface{}), val)
	} else {
		// Linked objects are placed in a map under the `linked` key
		var linkedMap map[string][]interface{}
		if ctx.wrapper["linked"] == nil {
			linkedMap = map[string][]interface{}{}
			ctx.wrapper["linked"] = linkedMap
		} else {
			linkedMap = ctx.wrapper["linked"].(map[string][]interface{})
		}
		if s := linkedMap[name]; s != nil {
			linkedMap[name] = append(s, val)
		} else {
			linkedMap[name] = []interface{}{val}
		}
	}
}

// MarshalToJSON takes a struct and marshals it to JSONAPI compliant JSON
func MarshalToJSON(val interface{}) ([]byte, error) {
	result, err := Marshal(val)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
