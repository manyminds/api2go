package jsonapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// UnmarshalIdentifier interface to set ID when unmarshalling
type UnmarshalIdentifier interface {
	SetID(string) error
}

// UnmarshalLinkedRelations same as MarshalLinkedRelations for unmarshaling
type UnmarshalLinkedRelations interface {
	SetReferencedIDs([]ReferenceID) error
}

// Unmarshal reads a JSONAPI map to a model struct
// target must at least implement the `UnmarshalIdentifier` interface.
func Unmarshal(input map[string]interface{}, target interface{}) error {
	// Check that target is a *[]Model
	ptrVal := reflect.ValueOf(target)
	if ptrVal.Kind() != reflect.Ptr || ptrVal.IsNil() {
		return errors.New("You must pass a pointer to a []UnmarshalIdentifier to Unmarshal()")
	}
	sliceType := reflect.TypeOf(target).Elem()
	sliceVal := ptrVal.Elem()
	if sliceType.Kind() != reflect.Slice {
		return errors.New("You must pass a pointer to a []UnmarshalIdentifier to Unmarshal()")
	}
	structType := sliceType.Elem()
	if structType.Kind() != reflect.Struct {
		return errors.New("You must pass a pointer to a []UnmarshalIdentifier to Unmarshal()")
	}

	// Copy the value, then write into the new variable.
	// Later Set() the actual value of the pointee.
	val := sliceVal
	err := UnmarshalInto(input, structType, &val)
	if err != nil {
		return err
	}
	sliceVal.Set(val)
	return nil
}

// UnmarshalFromJSON reads a JSONAPI compatible JSON document to a model struct
func UnmarshalFromJSON(data []byte, target interface{}) error {
	var ctx map[string]interface{}
	err := json.Unmarshal(data, &ctx)
	if err != nil {
		return err
	}
	return Unmarshal(ctx, target)
}

// UnmarshalInto reads input params for one struct from `input` and marshals it into `targetSliceVal`
func UnmarshalInto(input map[string]interface{}, targetStructType reflect.Type, targetSliceVal *reflect.Value) error {
	// Read models slice
	var modelsInterface interface{}

	if modelsInterface = input["data"]; modelsInterface == nil {
		return errors.New("expected root document to include a data key but it didn't")
	}

	models, ok := modelsInterface.([]interface{})
	if !ok {
		models = []interface{}{modelsInterface}
	}

	// Read all the models
	for _, m := range models {
		attributes, ok := m.(map[string]interface{})
		if !ok {
			return errors.New("expected an array of objects under key data")
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
			for i := 0; i < targetSliceVal.Len(); i++ {
				obj := targetSliceVal.Index(i)
				existingObj, ok := obj.Interface().(MarshalIdentifier)
				if !ok {
					return errors.New("existing structs must implement interface MarshalIdentifier")
				}
				otherID := existingObj.GetID()
				if otherID == id {
					val = obj
					isNew = false
					break
				}
			}
		}
		// If the struct wasn't already there for updating, make a new one
		if !val.IsValid() {
			val = reflect.New(targetStructType).Elem()
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
				var i reflect.Value
				if val.CanAddr() {
					i = val.Addr()
				}
				targetStruct, ok := i.Interface().(UnmarshalIdentifier)
				if !ok {
					return errors.New("All target structs must implement UnmarshalIdentifier interface")
				}

				// Allow conversion of string id to int
				id, ok = v.(string)
				if !ok {
					return errors.New("expected id to be of type string")
				}

				targetStruct.SetID(id)

			case "type":
				structType, ok := v.(string)
				if !ok {
					return errors.New("type must be string")
				}

				expectedType := Pluralize(Jsonify(targetStructType.Name()))
				if structType != expectedType {
					return fmt.Errorf("type %s does not match expected type %s of target struct", structType, expectedType)
				}
				// do not unmarshal the `type` field

			default:
				fieldName := Dejsonify(k)
				field := val.FieldByName(fieldName)
				if !field.IsValid() {
					//check if there is any field tag with the given name available
					for x := 0; x < val.NumField(); x++ {
						tfield := val.Type().Field(x)
						if tag := tfield.Tag.Get("jsonapi"); tag != "" {
							//check key value pairs. Currently only name is supported.
							if strings.Contains(tag, "name") && strings.Contains(tag, "=") {
								entry := strings.Split(tag, "=")
								if len(entry) == 2 && entry[0] == "name" {
									if entry[1] == strings.ToLower(fieldName) {
										field = val.Field(x)
									}
								} else {
									return errors.New("tag must be formatted correctly for field " + fieldName)
								}
							}
						}
					}

					if !field.IsValid() {
						return errors.New("expected struct " + targetStructType.Name() + " to have field " + fieldName)
					}
				}

				value := reflect.ValueOf(v)

				if value.IsValid() {
					plainValue := reflect.ValueOf(v)

					switch field.Interface().(type) {
					case time.Time:
						t, err := time.Parse(time.RFC3339, plainValue.String())
						if err != nil {
							return errors.New("expected RFC3339 time string, got '" + plainValue.String() + "'")
						}

						field.Set(reflect.ValueOf(t))
					default:
						if field.CanAddr() {
							switch field.Addr().Interface().(type) {
							default:
								err := setFieldValue(&field, plainValue)
								if err != nil {
									return fmt.Errorf("Could not set field '%s'. %s", fieldName, err.Error())
								}

							}
						} else {
							err := setFieldValue(&field, plainValue)
							if err != nil {
								return fmt.Errorf("Could not set field '%s'. %s", fieldName, err.Error())
							}
						}
					}
				}
			}
		}

		if isNew {
			*targetSliceVal = reflect.Append(*targetSliceVal, val)
		}
	}

	return nil
}

// setFieldValue in a json object, there is only the number type, which defaults to float64. This method convertes float64 to the value
// of the underlying struct field, for example uint64, or int32 etc...
// If the field type is not one of the integers, it just sets the value
func setFieldValue(field *reflect.Value, value reflect.Value) (err error) {
	// catch all invalid types and return an error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Value '%v' had wrong type", value.Interface())
		}
	}()

	switch field.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		field.SetInt(int64(value.Float()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		field.SetUint(uint64(value.Float()))
	default:
		field.Set(value)
	}

	return nil
}

func unmarshalLinks(val reflect.Value, linksMap map[string]interface{}) error {
	referenceIDs := []ReferenceID{}

	for linkName, links := range linksMap {
		links, ok := links.(map[string]interface{})
		if !ok {
			return fmt.Errorf("link field for %s has invalid format, must be map[string]interface{}", linkName)
		}
		_, ok = links["linkage"]
		if !ok {
			return fmt.Errorf("Missing linkage field for %s", linkName)
		}
		if !ok {
			return fmt.Errorf("type field for %s links must be a string", linkName)
		}

		hasOne, ok := links["linkage"].(map[string]interface{})
		if ok {
			hasOneID, ok := hasOne["id"].(string)
			if !ok {
				return fmt.Errorf("linkage object must have a field id for %s", linkName)
			}
			hasOneType, ok := hasOne["type"].(string)
			if !ok {
				return fmt.Errorf("linkage object must have a field type for %s", linkName)
			}

			referenceIDs = append(referenceIDs, ReferenceID{ID: hasOneID, Name: linkName, Type: hasOneType})
		} else {
			hasMany, ok := links["linkage"].([]interface{})
			if !ok {
				fmt.Printf("%#v", links["linkage"])
				return fmt.Errorf("invalid linkage object or array, must be an object with \"id\" and \"type\" field for %s", linkName)
			}
			for _, entry := range hasMany {
				linkage, ok := entry.(map[string]interface{})
				if !ok {
					return fmt.Errorf("entry in linkage array must be an object for %s", linkName)
				}
				linkageID, ok := linkage["id"].(string)
				if !ok {
					return fmt.Errorf("all linkage objects must have a field id for %s", linkName)
				}
				linkageType, ok := linkage["type"].(string)
				if !ok {
					return fmt.Errorf("all linkage objects must have a field type for %s", linkName)
				}

				referenceIDs = append(referenceIDs, ReferenceID{ID: linkageID, Name: linkName, Type: linkageType})
			}
		}
	}

	if val.CanAddr() {
		val = val.Addr()
	}

	target, ok := val.Interface().(UnmarshalLinkedRelations)
	if !ok {
		return errors.New("target struct must implement interface UnmarshalLinkedRelations")
	}
	target.SetReferencedIDs(referenceIDs)

	return nil
}
