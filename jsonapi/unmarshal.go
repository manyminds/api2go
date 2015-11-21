package jsonapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
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

	if ctx.Data == nil {
		return errors.New(`Source JSON is empty and has no "atributes" payload object`)
	}

	if ctx.Data.DataObject != nil {
		return setDataIntoTarget(ctx.Data.DataObject, target)
	}

	if ctx.Data.DataArray != nil {
		if reflect.TypeOf(target).Elem().Kind() != reflect.Slice {
			return errors.New("Source JSON contained an array, but single record was expected")
		}
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
	if reflect.ValueOf(target).Kind() == reflect.Struct {
		target = reflect.ValueOf(target).Addr().Interface()
	}
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
	toOneError := fmt.Errorf("struct %s does not implement UnmarshalToOneRelations", reflect.TypeOf(target))
	toManyError := fmt.Errorf("struct %s does not implement UnmarshalToManyRelations", reflect.TypeOf(target))

	for name, rel := range relationships {
		// relationship is empty case
		if rel.Data == nil {
			if Pluralize(name) == name {
				castedToMany, ok := target.(UnmarshalToManyRelations)
				if !ok {
					return toManyError
				}

				castedToMany.SetToManyReferenceIDs(name, []string{})
				break
			}

			castedToOne, ok := target.(UnmarshalToOneRelations)
			if !ok {
				return toOneError
			}

			castedToOne.SetToOneReferenceID(name, "")
			break
		}

		// valid toOne case
		if rel.Data.DataObject != nil {
			castedToOne, ok := target.(UnmarshalToOneRelations)
			if !ok {
				return toOneError
			}
			err := castedToOne.SetToOneReferenceID(name, rel.Data.DataObject.ID)
			if err != nil {
				return err
			}
		}

		// valid toMany case
		if rel.Data.DataArray != nil {
			castedToMany, ok := target.(UnmarshalToManyRelations)
			if !ok {
				return toManyError
			}
			IDs := make([]string, len(rel.Data.DataArray))
			for index, relData := range rel.Data.DataArray {
				IDs[index] = relData.ID
			}
			err := castedToMany.SetToManyReferenceIDs(name, IDs)
			if err != nil {
				return err
			}
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
