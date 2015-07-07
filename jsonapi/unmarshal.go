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

// UnmarshalToOneRelations must be implemented to unmarshal to-one relations
type UnmarshalToOneRelations interface {
	SetToOneReferenceID(name, ID string) error
}

// UnmarshalToManyRelations must be implemented to unmarshal to-many relations
type UnmarshalToManyRelations interface {
	SetToManyReferenceIDs(name string, IDs []string) error
}

// The EditToManyRelations interface can be optionally implemented to add and delete to-many
// relationships on a already unmarshalled struct. These methods are used by our API for the to-many
// relationship update routes.
/*
There are 3 HTTP Methods to edit to-many relations:

	PATCH /v1/posts/1/comments
	Content-Type: application/vnd.api+json
	Accept: application/vnd.api+json

	{
	  "data": [
		{ "type": "comments", "id": "2" },
		{ "type": "comments", "id": "3" }
	  ]
	}

this replaces all of the comments that belong to post with ID 1 and the SetToManyReferenceIDs method
will be called

	POST /v1/posts/1/comments
	Content-Type: application/vnd.api+json
	Accept: application/vnd.api+json

	{
	  "data": [
		{ "type": "comments", "id": "123" }
	  ]
	}

adds a new comment to the post with ID 1. The AddToManyIDs methid will be called.

	DELETE /v1/posts/1/comments
	Content-Type: application/vnd.api+json
	Accept: application/vnd.api+json

	{
	  "data": [
		{ "type": "comments", "id": "12" },
		{ "type": "comments", "id": "13" }
	  ]
	}

deletes comments that belong to post with ID 1. The DeleteToManyIDs method will be called.
*/
type EditToManyRelations interface {
	AddToManyIDs(name string, IDs []string) error
	DeleteToManyIDs(name string, IDs []string) error
}

// Unmarshal reads a JSONAPI map to a model struct
// target must at least implement the `UnmarshalIdentifier` interface.
func Unmarshal(input map[string]interface{}, target interface{}) error {
	var (
		structType reflect.Type
		sliceVal   reflect.Value
		isStruct   bool
	)

	typeError := errors.New("You must pass a pointer to a UnmarshalIdentifier or slice of it to Unmarshal()")

	// Check that target is a *[]Model
	ptrVal := reflect.ValueOf(target)
	if ptrVal.Kind() != reflect.Ptr || ptrVal.IsNil() {
		return typeError
	}
	targetType := reflect.TypeOf(target).Elem()

	if targetType.Kind() != reflect.Slice {
		// check for a struct or pointer to struct which are also allowed to unmarshal into
		if targetType.Kind() == reflect.Struct {
			structType = targetType
		} else if targetType.Kind() == reflect.Ptr {
			structType = targetType.Elem()
		} else {
			return typeError
		}
		sliceVal = reflect.New(reflect.SliceOf(targetType)).Elem()
		isStruct = true
	} else {
		sliceVal = ptrVal.Elem()
		structType = targetType.Elem()
		if structType.Kind() == reflect.Ptr {
			structType = structType.Elem()
		}
	}

	if structType.Kind() != reflect.Struct {
		return typeError
	}

	// Copy the value, then write into the new variable.
	// Later Set() the actual value of the pointee.
	val := sliceVal
	err := UnmarshalInto(input, structType, &val)
	if err != nil {
		return err
	}

	// if target is a struct, the first unmarshalled entry of a slice of its type will be set into it
	if isStruct {
		ptrVal.Elem().Set(val.Index(0))
	} else {
		sliceVal.Set(val)
	}
	return nil
}

// UnmarshalFromJSON reads a JSONAPI compatible JSON document to a model struct
// target must be a struct or a slice of it
func UnmarshalFromJSON(data []byte, target interface{}) error {
	var ctx map[string]interface{}
	err := json.Unmarshal(data, &ctx)
	if err != nil {
		return err
	}
	return Unmarshal(ctx, target)
}

// UnmarshalInto reads input params for one struct from `input` and marshals it into `targetSliceVal`,
// which may be a slice of targetStructType or a slice of pointers to targetStructType.
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
		data, ok := m.(map[string]interface{})
		if !ok {
			return errors.New("expected an array of objects under key data")
		}

		var val reflect.Value
		isNew := true
		id := ""

		if v := data["id"]; v != nil {
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
					if obj.Type().Kind() == reflect.Struct {
						val = obj
					} else {
						val = obj.Elem()
					}
					isNew = false
					break
				}
			}
		}
		// If the struct wasn't already there for updating, make a new one
		if !val.IsValid() {
			val = reflect.New(targetStructType).Elem()
		}

		for k, v := range data {
			switch k {
			case "relationships":
				relationshipsMap, ok := v.(map[string]interface{})
				if !ok {
					return errors.New("expected relationships to be an object")
				}
				if err := unmarshalRelationships(val, relationshipsMap); err != nil {
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

			case "attributes":
				attributes, ok := v.(map[string]interface{})
				if !ok {
					return errors.New("expected attributes to be an object")
				}

				for key, attributeValue := range attributes {
					fieldName := Dejsonify(key)
					field := val.FieldByName(fieldName)
					if !field.IsValid() {
						//check if there is any field tag with the given name available
						for x := 0; x < val.NumField(); x++ {
							tfield := val.Type().Field(x)
							name := GetTagValueByName(tfield, "name")
							if name == strings.ToLower(fieldName) {
								field = val.Field(x)
							}
						}

						if !field.IsValid() {
							return errors.New("did not expect struct " + targetStructType.Name() + " to have field " + fieldName)
						}
					}

					value := reflect.ValueOf(attributeValue)

					if value.IsValid() {
						plainValue := reflect.ValueOf(attributeValue)

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
		}

		if isNew {
			if targetSliceVal.Type().Elem().Kind() == reflect.Struct {
				*targetSliceVal = reflect.Append(*targetSliceVal, val)
			} else {
				*targetSliceVal = reflect.Append(*targetSliceVal, val.Addr())
			}
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
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		field.SetUint(uint64(value.Float()))
	default:
		// try to set it with json.Unmarshaler interface, if that does not work, set value directly
		switch target := field.Addr().Interface().(type) {
		case json.Unmarshaler:
			marshaledValue, err := json.Marshal(value.Interface())
			if err != nil {
				return err
			}

			err = target.UnmarshalJSON(marshaledValue)
			if err != nil {
				return err
			}
		default:
			field.Set(value)
		}
	}

	return nil
}

// UnmarshalRelationshipsData is used by api2go.API to only unmarshal references inside a data object.
// The target interface must implement UnmarshalToOneRelations or UnmarshalToManyRelations interface.
// The linksMap is the content of the data object from the json
func UnmarshalRelationshipsData(target interface{}, name string, links interface{}) error {
	return processRelationshipsData(links, name, target)
}

func unmarshalRelationships(val reflect.Value, relationshipsMap map[string]interface{}) error {
	for relationshipName, relationships := range relationshipsMap {
		relationships, ok := relationships.(map[string]interface{})
		if !ok {
			return fmt.Errorf("link field for %s has invalid format, must be map[string]interface{}", relationshipName)
		}
		_, ok = relationships["data"]
		if !ok {
			return fmt.Errorf("Missing data field for %s", relationshipName)
		}

		if val.CanAddr() {
			val = val.Addr()
		}

		err := processRelationshipsData(relationships["data"], relationshipName, val.Interface())
		if err != nil {
			return err
		}
	}

	return nil
}

func processRelationshipsData(data interface{}, linkName string, target interface{}) error {
	hasOne, ok := data.(map[string]interface{})
	if ok {
		hasOneID, ok := hasOne["id"].(string)
		if !ok {
			return fmt.Errorf("data object must have a field id for %s", linkName)
		}

		target, ok := target.(UnmarshalToOneRelations)
		if !ok {
			return errors.New("target struct must implement interface UnmarshalToOneRelations")
		}

		target.SetToOneReferenceID(linkName, hasOneID)
	} else if data == nil {
		// this means that a to-one relationship must be deleted
		target, ok := target.(UnmarshalToOneRelations)
		if !ok {
			return errors.New("target struct must implement interface UnmarshalToOneRelations")
		}

		target.SetToOneReferenceID(linkName, "")
	} else {
		hasMany, ok := data.([]interface{})
		if !ok {
			return fmt.Errorf("invalid data object or array, must be an object with \"id\" and \"type\" field for %s", linkName)
		}

		target, ok := target.(UnmarshalToManyRelations)
		if !ok {
			return errors.New("target struct must implement interface UnmarshalToManyRelations")
		}

		hasManyIDs := []string{}

		for _, entry := range hasMany {
			data, ok := entry.(map[string]interface{})
			if !ok {
				return fmt.Errorf("entry in data array must be an object for %s", linkName)
			}
			dataID, ok := data["id"].(string)
			if !ok {
				return fmt.Errorf("all data objects must have a field id for %s", linkName)
			}

			hasManyIDs = append(hasManyIDs, dataID)
		}

		target.SetToManyReferenceIDs(linkName, hasManyIDs)
	}

	return nil
}
