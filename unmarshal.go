package api2go

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"
)

type unmarshalContext map[string]interface{}

// Unmarshal reads a JSONAPI map to a model struct
func Unmarshal(ctx unmarshalContext, values interface{}) error {
	// Check that target is a *[]Model
	ptrVal := reflect.ValueOf(values)
	if ptrVal.Kind() != reflect.Ptr || ptrVal.IsNil() {
		return errors.New("You must pass a pointer to a []struct to Unmarshal()")
	}
	sliceType := reflect.TypeOf(values).Elem()
	sliceVal := ptrVal.Elem()
	if sliceType.Kind() != reflect.Slice {
		return errors.New("You must pass a pointer to a []struct to Unmarshal()")
	}
	structType := sliceType.Elem()
	if structType.Kind() != reflect.Struct {
		return errors.New("You must pass a pointer to a []struct to Unmarshal()")
	}

	// Copy the value, then write into the new variable.
	// Later Set() the actual value of the pointee.
	val := sliceVal
	err := unmarshalInto(ctx, structType, &val)
	if err != nil {
		return err
	}
	sliceVal.Set(val)
	return nil
}

// fillSqlScanner extracts the value of into the field of the target struct
func fillSqlScanner(structField interface{}, value interface{}) (sql.Scanner, error) {
	newTarget := reflect.TypeOf(structField)

	intf := reflect.New(newTarget.Elem()).Interface()

	intf2, ok := intf.(sql.Scanner)
	if !ok {
		return nil, fmt.Errorf("could not type cast into sql.Scanner: %#v", structField)
	}

	intf2.Scan(value)

	return intf2, nil
}

// setFieldValue in a json object, there is only the number type, which defaults to float64. This method convertes float64 to the value
// of the underlying struct field, for example uint64, or int32 etc...
// If the field type is not one of the integers, it just sets the value
func setFieldValue(field *reflect.Value, value reflect.Value) {
	switch field.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		field.SetInt(int64(value.Float()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		field.SetUint(uint64(value.Float()))
	default:
		field.Set(value)
	}
}

func unmarshalInto(ctx unmarshalContext, structType reflect.Type, sliceVal *reflect.Value) error {
	// Read models slice
	var modelsInterface interface{}
	rootName := pluralize(jsonify(structType.Name()))

	if modelsInterface = ctx[rootName]; modelsInterface == nil {
		rootName = "data"
		if modelsInterface = ctx[rootName]; modelsInterface == nil {
			return errors.New("expected root document to include a '" + rootName + "' key but it didn't.")
		}
	}

	models, ok := modelsInterface.([]interface{})
	if !ok {
		models = []interface{}{modelsInterface}
	}

	// Read all the models
	for _, m := range models {
		attributes, ok := m.(map[string]interface{})
		if !ok {
			return errors.New("expected an array of objects under key '" + rootName + "'")
		}

		var val reflect.Value
		isNew := true
		id := ""

		if v := attributes["id"]; v != nil {
			id, ok = v.(string)
			if !ok {
				return errors.New("id must be a string")
			}

			// If we have an ID, check if there's already an object with that ID in the slice
			// TODO This is O(n^2), make it O(n)
			for i := 0; i < sliceVal.Len(); i++ {
				obj := sliceVal.Index(i)
				otherID, err := idFromObject(obj)
				if err != nil {
					return err
				}
				if otherID == id {
					val = obj
					isNew = false
					break
				}
			}
		}
		// If the struct wasn't already there for updating, make a new one
		if !val.IsValid() {
			val = reflect.New(structType).Elem()
		}

		for k, v := range attributes {
			switch k {
			case "links":
				linksMap, ok := v.(map[string]interface{})
				if !ok {
					return errors.New("expected links to be an object")
				}
				if err := unmarshalLinks(val, linksMap); err != nil {
					return err
				}

			case "id":
				// Allow conversion of string id to int
				id, ok = v.(string)
				if !ok {
					return errors.New("expected id to be of type string")
				}
				if err := setObjectID(val, id); err != nil {
					return err
				}

			case "type":
				// do not unmarshal the `type` field

			default:
				fieldName := dejsonify(k)
				field := val.FieldByName(fieldName)
				if !field.IsValid() {
					return errors.New("expected struct " + structType.Name() + " to have field " + fieldName)
				}
				value := reflect.ValueOf(v)

				if value.IsValid() {
					plainValue := reflect.ValueOf(v)

					switch element := field.Interface().(type) {
					case time.Time:
						t, err := time.Parse(time.RFC3339, plainValue.String())
						if err != nil {
							return errors.New("expected RFC3339 time string, got '" + plainValue.String() + "'")
						}

						field.Set(reflect.ValueOf(t))
					case sql.Scanner:
						scanner, err := fillSqlScanner(element, plainValue.Interface())
						if err != nil {
							return err
						}

						field.Set(reflect.ValueOf(scanner))
					default:
						if field.CanAddr() {
							switch element := field.Addr().Interface().(type) {
							case sql.Scanner:
								scanner, err := fillSqlScanner(element, plainValue.Interface())
								if err != nil {
									return err
								}

								field.Set(reflect.ValueOf(scanner).Elem())
							default:
								setFieldValue(&field, plainValue)
							}
						} else {
							setFieldValue(&field, plainValue)
						}
					}
				}
			}
		}

		if isNew {
			*sliceVal = reflect.Append(*sliceVal, val)
		}
	}

	return nil
}

func unmarshalLinks(val reflect.Value, linksMap map[string]interface{}) error {
	for linkName, linkObj := range linksMap {
		switch links := linkObj.(type) {
		case []interface{}:
			// Has-many
			// Check for field named 'FoobarsIDs' for key 'foobars'
			structFieldName := dejsonify(linkName) + "IDs"
			sliceField := val.FieldByName(structFieldName)
			if !sliceField.IsValid() || sliceField.Kind() != reflect.Slice {
				return errors.New("expected struct to have a " + structFieldName + " slice")
			}

			sliceField.Set(reflect.MakeSlice(sliceField.Type(), len(links), len(links)))
			for i, idInterface := range links {
				if err := setIDValue(sliceField.Index(i), idInterface); err != nil {
					return err
				}
			}

		case string:
			// Belongs-to or has-one
			// Check for field named 'FoobarID' for key 'foobar'
			structFieldName := dejsonify(linkName) + "ID"
			field := val.FieldByName(structFieldName)
			if err := setIDValue(field, links); err != nil {
				return err
			}

		case map[string]interface{}:
			// Belongs-to or has-one
			// Check for field named 'FooID' for key 'foo' if the type is 'foobar'
			if links["id"] != nil {
				id := links["id"].(string)
				structFieldName := dejsonify(linkName) + "ID"
				field := val.FieldByName(structFieldName)
				if err := setIDValue(field, id); err != nil {
					return err
				}

				continue
			}

			// Has-many
			// Check for field named 'FoosIDs' for key 'foos' if the type is 'foobars'
			if links["ids"] != nil {
				ids := links["ids"].([]interface{})

				structFieldName := dejsonify(linkName) + "IDs"
				sliceField := val.FieldByName(structFieldName)
				if !sliceField.IsValid() || sliceField.Kind() != reflect.Slice {
					return errors.New("expected struct to have a " + structFieldName + " slice")
				}

				sliceField.Set(reflect.MakeSlice(sliceField.Type(), len(ids), len(ids)))
				for i, idInterface := range ids {
					if err := setIDValue(sliceField.Index(i), idInterface); err != nil {
						return err
					}
				}

				continue
			}

			return errors.New("Invalid object in links object")
		default:
			return errors.New("expected string, array or an object with field id(s) in links object")
		}
	}
	return nil
}

// UnmarshalFromJSON reads a JSONAPI compatible JSON document to a model struct
func UnmarshalFromJSON(data []byte, values interface{}) error {
	var ctx unmarshalContext
	err := json.Unmarshal(data, &ctx)
	if err != nil {
		return err
	}
	return Unmarshal(ctx, values)
}
