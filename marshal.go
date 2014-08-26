package api2go

import (
	"encoding/json"
	"reflect"
)

type marshalingContext struct {
	root     map[string]interface{}
	rootName string
}

func makeContext(rootName string) *marshalingContext {
	ctx := &marshalingContext{}
	ctx.rootName = rootName
	ctx.root = map[string]interface{}{}
	ctx.root[rootName] = []interface{}{}
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

	return ctx.root, nil
}

// marshalStruct marshals a struct and places it in the context's root
func (ctx *marshalingContext) marshalStruct(val reflect.Value) error {
	result := map[string]interface{}{}
	linksMap := map[string][]interface{}{}

	valType := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := underscorize(valType.Field(i).Name)

		if field.Kind() == reflect.Slice {
			ids := []interface{}{}

			// Treat nested objects
			for i := 0; i < field.Len(); i++ {
				if id := field.Index(i).FieldByName("ID"); id.IsValid() {
					ids = append(ids, id.Interface())
				} else {
					// BUG(lucas): Maybe just silently ignore them?
					panic("structs passed to Marshal need to contain ID fields")
				}

				if err := ctx.marshalStruct(field.Index(i)); err != nil {
					return err
				}
			}

			linksMap[fieldName] = ids
		} else {
			result[fieldName] = field.Interface()
		}
	}

	if len(linksMap) > 0 {
		result["links"] = linksMap
	}

	ctx.addValue(pluralize(underscorize(valType.Name())), result)
	return nil
}

// addValue adds an object to the context's root
// `name` should be the pluralized and underscorized object type.
func (ctx *marshalingContext) addValue(name string, val map[string]interface{}) {
	if name == ctx.rootName {
		// Root objects are placed directly into the root doc
		// BUG(lucas): If an object links to its own type, linked objects must be placed into the linked map.
		ctx.root[name] = append(ctx.root[name].([]interface{}), val)
	} else {
		// Linked objects are placed in a map under the `linked` key
		var linkedMap map[string][]interface{}
		if ctx.root["linked"] == nil {
			linkedMap = map[string][]interface{}{}
			ctx.root["linked"] = linkedMap
		} else {
			linkedMap = ctx.root["linked"].(map[string][]interface{})
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
