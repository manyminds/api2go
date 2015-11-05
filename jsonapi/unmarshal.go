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

var (
	errInterface         = errors.New("target must implement UnmarshalIdentifier interface")
	errAttributesMissing = errors.New("missing mandatory attributes object")
)

// Unmarshal reads a jsonapi compatible JSON as []byte
// target must at least implement the `UnmarshalIdentifier` interface.
func Unmarshal(data []byte, target interface{}) error {
	ctx := &Document{}
	err := json.Unmarshal(data, ctx)
	if err != nil {
		return err
	}

	if ctx.Data.DataObject != nil {
		return setDataIntoTarget(ctx.Data.DataObject, target)
	}

	if ctx.Data.DataArray != nil {
		targetType := reflect.TypeOf(target).Elem().Elem()
		targetPointer := reflect.ValueOf(target)
		targetValue := targetPointer.Elem()

		for _, record := range ctx.Data.DataArray {
			// check if there already is an entry with the same id in target slice, otherwise
			// create a new target and append
			var targetRecord, emptyValue reflect.Value
			for i := 0; i < targetValue.Len(); i++ {
				marshalCasted, ok := targetValue.Index(i).Interface().(MarshalIdentifier)
				if !ok {
					return errors.New("existing structs must implement interface MarshalIdentifier")
				}
				if record.ID == marshalCasted.GetID() {
					targetRecord = targetValue.Index(i).Addr()
					break
				}
			}

			if targetRecord == emptyValue || targetRecord.IsNil() {
				targetRecord = reflect.New(targetType)
				err := setDataIntoTarget(&record, targetRecord.Interface())
				if err != nil {
					return err
				}
				targetValue = reflect.Append(targetValue, targetRecord.Elem())
			} else {
				err := setDataIntoTarget(&record, targetRecord.Interface())
				if err != nil {
					return err
				}
			}
		}

		targetPointer.Elem().Set(targetValue)
	}

	return nil
}

func setDataIntoTarget(data *Data, target interface{}) error {
	castedTarget, ok := target.(UnmarshalIdentifier)
	if !ok {
		return errInterface
	}

	err := checkType(data.Type, castedTarget)
	if err != nil {
		return err
	}

	if data.Attributes == nil {
		return errAttributesMissing
	}

	err = json.Unmarshal(data.Attributes, castedTarget)
	if err != nil {
		return err
	}
	castedTarget.SetID(data.ID)
	return setRelationshipIDs(data.Relationships, castedTarget)
}

// extracts all found relationships and set's them via SetToOneReferenceID or SetToManyReferenceIDs
func setRelationshipIDs(relationships map[string]Relationship, target UnmarshalIdentifier) error {
	for key, rel := range relationships {
		if rel.Data.DataObject != nil {
			castedToOne, ok := target.(UnmarshalToOneRelations)
			if !ok {
				return errors.New("struct <name> does not implement UnmarshalToOneRelations")
			}
			castedToOne.SetToOneReferenceID(key, rel.Data.DataObject.ID)
		}

		if rel.Data.DataArray != nil {
			castedToMany, ok := target.(UnmarshalToManyRelations)
			if !ok {
				return errors.New("struct <name> does not implement UnmarshalToManyRelations")
			}
			IDs := make([]string, len(rel.Data.DataArray))
			for index, relData := range rel.Data.DataArray {
				IDs[index] = relData.ID
			}
			castedToMany.SetToManyReferenceIDs(key, IDs)
		}
	}

	return nil
}

func checkType(incomingType string, target UnmarshalIdentifier) error {
	actualType := getStructType(target)
	if incomingType != actualType {
		return fmt.Errorf("Type %s in JSON does not match target struct type %s", incomingType, actualType)
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
	return Unmarshal([]byte("{}"), target)
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
				var expectedType string
				structType, ok := v.(string)
				if !ok {
					return errors.New("type must be string")
				}

				entityName, ok := val.Interface().(EntityNamer)
				if ok {
					expectedType = entityName.GetName()
				} else {
					expectedType = Pluralize(Jsonify(targetStructType.Name()))
				}
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
					fieldType, found := val.Type().FieldByName(fieldName)
					if !found {
						//check if there is any field tag with the given name available
						field, found = getFieldByTagName(val, fieldName)
					}

					if !found || fieldType.Tag.Get("jsonapi") == "-" {
						return fmt.Errorf("invalid key \"%s\" in json. Cannot be assigned to target struct \"%s\"", key, targetStructType.Name())
					}

					value := reflect.ValueOf(attributeValue)

					if !field.CanInterface() {
						return fmt.Errorf("field not exported. Expected field with name %s to exist", fieldName)
					}

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
					} else {
						// Set the field to its zero value if its type is from the zero or null package.
						if extType := field.Type(); extType != nil && extType.Kind().String() == "struct" {
							fieldSource := extType.Field(0).Name
							if strings.HasPrefix(fieldSource, "Null") || strings.HasPrefix(fieldSource, "Zero") || strings.HasPrefix(fieldSource, "Time") {
								field.Set(reflect.Zero(field.Type()))
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

// check if there is any field tag with the given name available
func getFieldByTagName(val reflect.Value, fieldName string) (field reflect.Value, found bool) {
	for x := 0; x < val.NumField(); x++ {
		tfield := val.Type().Field(x)
		if tfield.Tag.Get("jsonapi") == "-" {
			continue
		}

		// check if there is an embedded struct which needs to be searched
		if val.Field(x).CanAddr() && val.Field(x).Addr().CanInterface() {
			_, isEmbedded := val.Field(x).Addr().Interface().(UnmarshalIdentifier)
			if isEmbedded {
				field, found = getFieldByTagName(val.Field(x), fieldName)
				if found {
					return
				}
			}
		}

		// try to find the field
		name := GetTagValueByName(tfield, "name")
		if strings.ToLower(name) == strings.ToLower(fieldName) {
			field = val.Field(x)
			found = true
			return
		}
	}

	return
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

	// check for Unmarshaler interface first, after that try to guess the right type
	target, ok := field.Addr().Interface().(json.Unmarshaler)
	if ok {
		marshaledValue, err := json.Marshal(value.Interface())
		if err != nil {
			return err
		}

		err = target.UnmarshalJSON(marshaledValue)
		if err != nil {
			return err
		}
	} else {
		switch field.Type().Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			field.SetInt(int64(value.Float()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			field.SetUint(uint64(value.Float()))
		case reflect.Slice:
			// we always get a []interface{] from json and now need to cast to the right slice type
			if value.Type() == reflect.TypeOf([]interface{}{}) {
				targetSlice := reflect.MakeSlice(field.Type(), 0, value.Len())
				sliceData := value.Interface().([]interface{})

				// only iterate over the array if it's not empty
				if value.Len() > 0 {
					targetType := reflect.TypeOf(sliceData[0])
					for _, entry := range sliceData {
						casted := reflect.ValueOf(entry).Convert(targetType)
						targetSlice = reflect.Append(targetSlice, casted)
					}
				}

				field.Set(targetSlice)
			} else {
				// we have the correct type, hm this is only for tests that use direct type at the moment.. we have to refactor the unmarshalling
				// anyways..
				field.Set(value)
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
