package jsonapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type marshalingContext struct {
	root           map[string]interface{}
	rootName       string
	isSingleStruct bool
	prefix         string
}

func makeContext(rootName string, isSingleStruct bool, prefix string) *marshalingContext {
	ctx := &marshalingContext{}
	ctx.rootName = rootName
	ctx.root = map[string]interface{}{}
	ctx.root[rootName] = []interface{}{}
	ctx.isSingleStruct = isSingleStruct
	ctx.prefix = prefix
	return ctx
}

// Marshal takes a struct (or slice of structs) and marshals them to a json encodable interface{} value
func Marshal(data interface{}) (interface{}, error) {
	return marshal(data, "")
}

// Identifier interface is necessary to give an element
// a unique ID. This interface must be implemented for
// marshal and unmarshal in order to let them store
// elements
type Identifier interface {
	GetID() string
	SetID(string) error
}

// ReferenceID todo later
type ReferenceID struct {
	ID   string
	Type string
	Name string
}

// Relations TODO blub
type Relations interface {
	GetReferencedIDs() []ReferenceID
	SetReferencedIDs([]ReferenceID) error
}

// MarshalPrefix2 does the same as Marshal but adds a prefix to generated URLs
func MarshalPrefix2(data Identifier, prefix string) (interface{}, error) {
	return marshal(data, prefix)
}

// MarshalPrefix does the same as Marshal but adds a prefix to generated URLs
func MarshalPrefix(data interface{}, prefix string) (interface{}, error) {
	return marshal(data, prefix)
}

// Marshal2 is the new shit
func Marshal2(data Identifier) (interface{}, error) {
	return marshal2(data, "")
}

// MarshalSlice marshals a slice TODO
func MarshalSlice(data interface{}) (interface{}, error) {
	result := make(map[string]interface{})

	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Slice {
		return result, errors.New("data must be a slice")
	}

	var dataElements []map[string]interface{}

	for i := 0; i < val.Len(); i++ {
		k := val.Index(i).Interface()
		element, ok := k.(Identifier)
		if !ok {
			return result, errors.New("all elements within the slice must implement api2go.Identifier")
		}

		content, err := marshalData(element)
		if err != nil {
			return result, err
		}

		dataElements = append(dataElements, content)
	}

	result["data"] = dataElements

	return result, nil
}

func marshalData(data Identifier) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	element, ok := data.(Identifier)
	if !ok {
		return result, errors.New("data must implement api2go.Identifier interface")
	}

	id := element.GetID()
	content := getStructFields(data)
	for k, v := range content {
		result[k] = v
	}

	// its important that the id from the interface
	// gets added afterwards, otherwise an ID field
	// could conflict with the actual marshalling
	result["id"] = id
	result["type"] = getStructType(data)

	// optional relationship interface for struct
	relationer, ok := data.(Relations)
	if ok {
		result["links"] = getStructLinks(relationer)
	}

	return result, nil
}

// getStructLinks returns the link struct with ids
func getStructLinks(relationer Relations) map[string]interface{} {
	referencedIDs := relationer.GetReferencedIDs()
	sortedResults := make(map[string][]ReferenceID)
	links := make(map[string]interface{})

	for _, referenceID := range referencedIDs {
		sortedResults[referenceID.Type] = append(sortedResults[referenceID.Type], referenceID)
	}

	for referenceType, referenceIDs := range sortedResults {
		switch len(referencedIDs) {
		case 0:
			continue
		case 1:
			links[referenceIDs[0].Name] = map[string]interface{}{
				"id":   referenceIDs[0].ID,
				"type": referenceType,
			}
		default:
			// multiple elements in links
			var ids []string

			for _, referenceID := range referenceIDs {
				ids = append(ids, referenceID.ID)
			}

			links[referenceIDs[0].Name] = map[string]interface{}{
				"ids":  ids,
				"type": referenceType,
			}
		}
	}
	return links
}

func marshal2(data Identifier, prefix string) (interface{}, error) {
	result := make(map[string]map[string]interface{})
	contentData, err := marshalData(data)
	if err != nil {
		return result, err
	}

	result["data"] = contentData
	return result, nil
}

func getStructType(data Identifier) string {
	return Pluralize(Jsonify(reflect.TypeOf(data).Elem().Name()))
}

func getStructFields(data Identifier) map[string]interface{} {
	result := make(map[string]interface{})
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	valType := val.Type()
	for i := 0; i < val.NumField(); i++ {
		tag := valType.Field(i).Tag.Get("json")
		if tag == "-" {
			continue
		}

		field := val.Field(i)
		keyName := Jsonify(valType.Field(i).Name)
		result[keyName] = field.Interface()
	}

	return result
}

func marshal(data interface{}, prefix string) (interface{}, error) {
	if data == nil || data == "" {
		return nil, errors.New("marshal only works with objects")
	}

	var ctx *marshalingContext

	if reflect.TypeOf(data).Kind() == reflect.Slice {
		// We were passed a slice
		// Using Elem() here to get the slice's element type
		rootName := Pluralize(Jsonify(reflect.TypeOf(data).Elem().Name()))

		// Error on empty string, i.e. passed []interface{}
		if rootName == "" {
			return nil, errors.New("you passed a slice of interfaces []interface{}{...} to Marshal. we cannot determine key names from that. Use []YourObjectName{...} instead")
		}
		ctx = makeContext("data", false, prefix)

		// Marshal all elements
		// We iterate using reflections to save copying the slice to a []interface{}
		sliceValue := reflect.ValueOf(data)
		for i := 0; i < sliceValue.Len(); i++ {
			if err := ctx.marshalRootStruct(sliceValue.Index(i)); err != nil {
				return nil, err
			}
		}
	} else {
		// We were passed a single object
		ctx = makeContext("data", true, prefix)

		// Marshal the value
		if err := ctx.marshalRootStruct(reflect.ValueOf(data)); err != nil {
			return nil, err
		}
	}

	return ctx.root, nil
}

// marshalRootStruct is a more convenient name for marshalling structs at root level
func (ctx *marshalingContext) marshalRootStruct(val reflect.Value) error {
	return ctx.marshalStruct(&val, false)
}

// marshalLinkedStruct is a more convenient name for marshalling structs that were linked to a root level struct
func (ctx *marshalingContext) marshalLinkedStruct(val reflect.Value) error {
	return ctx.marshalStruct(&val, true)
}

// marshalStruct marshals a struct and places it in the context's root
func (ctx *marshalingContext) marshalStruct(val *reflect.Value, isLinked bool) error {
	result := map[string]interface{}{}
	linksMap := map[string]interface{}{}
	idFieldRegex := regexp.MustCompile("^.*ID$")

	valType := val.Type()
	name := Jsonify(Pluralize(valType.Name()))

	buildLinksMap := func(referenceIDs []interface{}, single bool, field reflect.Value, name, keyName string) map[string]interface{} {
		var resource string
		if ctx.prefix != "/" && ctx.prefix != "" {
			resource = fmt.Sprintf("%s%s/%s/%s", ctx.prefix, name, result["id"], keyName)
		} else {
			resource = fmt.Sprintf("/%s/%s/%s", name, result["id"], keyName)
		}

		result := make(map[string]interface{})
		result["type"] = Pluralize(Jsonify(field.Type().Elem().Name()))
		result["resource"] = resource

		if single {
			if referenceIDs[0] != "" {
				result["id"] = referenceIDs[0]
			}
		} else {
			result["ids"] = referenceIDs
		}

		return result
	}

	for i := 0; i < val.NumField(); i++ {
		tag := valType.Field(i).Tag.Get("json")
		if tag == "-" {
			continue
		}

		field := val.Field(i)
		keyName := Jsonify(valType.Field(i).Name)

		if field.Kind() == reflect.Slice {
			// A slice indicates nested objects.

			// First, check whether this is a slice of structs which we need to nest
			if field.Type().Elem().Kind() == reflect.Struct {
				ids := []interface{}{}
				for i := 0; i < field.Len(); i++ {
					id, err := idFromObject(field.Index(i))
					if err != nil {
						return err
					}
					ids = append(ids, id)

					if err := ctx.marshalLinkedStruct(field.Index(i)); err != nil {
						return err
					}
				}

				linksMap[keyName] = buildLinksMap(ids, false, field, name, keyName)
			} else if strings.HasSuffix(keyName, "IDs") {
				// Treat slices of non-struct type as lists of IDs if the suffix is IDs
				keyName = strings.TrimSuffix(keyName, "IDs")
				linksMapReflect := reflect.TypeOf(linksMap[keyName].(map[string]interface{})["ids"])

				// Don't overwrite any existing links, since they came from nested structs
				if linksMap[keyName] == nil || linksMapReflect.Kind() == reflect.Slice && len(linksMap[keyName].(map[string]interface{})["ids"].([]interface{})) == 0 {
					ids := []interface{}{}
					for i := 0; i < field.Len(); i++ {
						id, err := idFromValue(field.Index(i))
						if err != nil {
							return err
						}
						ids = append(ids, id)
					}

					structFieldName := Dejsonify(keyName)
					typeField := val.FieldByName(structFieldName)

					if typeField.IsValid() {
						linksMap[keyName] = buildLinksMap(ids, false, typeField, name, keyName)
					} else {
						return fmt.Errorf("expected struct to have field %s", structFieldName)
					}
				}
			} else {
				result[keyName] = field.Interface()

			}
		} else if keyName == "id" {
			// ID needs to be converted to string
			id, err := idFromValue(field)
			if err != nil {
				return err
			}
			result[keyName] = id
		} else if field.Type().Kind() == reflect.Ptr {
			if !field.IsNil() {
				id, err := idFromObject(field)
				if err == nil {
					linksMap[keyName] = buildLinksMap([]interface{}{id}, true, field, name, keyName)

					if err := ctx.marshalLinkedStruct(field.Elem()); err != nil {
						return err
					}
				} else {
					// the field is not a referenced struct, it is a normal property, so add it to the result
					result[keyName] = field.Interface()
				}
			}
		} else if idFieldRegex.MatchString(keyName) {
			keyNameWithoutID := strings.TrimSuffix(keyName, "ID")
			structFieldName := Dejsonify(keyNameWithoutID)
			// struct must be preferred, only use this field if struct ptr is nil
			structFieldValue := val.FieldByName(structFieldName)
			if !structFieldValue.IsValid() {
				return fmt.Errorf("expected struct to have field %s", structFieldName)
			}
			if structFieldValue.Kind() == reflect.Ptr && structFieldValue.IsNil() {
				id, err := idFromValue(field)
				if err != nil {
					return err
				}

				linksMap[keyNameWithoutID] = buildLinksMap([]interface{}{id}, true, structFieldValue, name, keyNameWithoutID)
			}
		} else {
			result[keyName] = field.Interface()
		}
	}

	// add object type
	result["type"] = Pluralize(Jsonify(valType.Name()))

	if len(linksMap) > 0 {
		result["links"] = linksMap
	}

	ctx.addValue(result, isLinked)
	return nil
}

// addValue adds an object to the context's root
// `name` should be the Pluralized and underscorized object type.
func (ctx *marshalingContext) addValue(val map[string]interface{}, isLinked bool) {
	if !isLinked {
		if ctx.isSingleStruct {
			ctx.root["data"] = val
		} else {
			// Root objects are placed directly into the root doc
			ctx.root["data"] = append(ctx.root["data"].([]interface{}), val)
		}
	} else {
		// Linked objects are placed in a slice under the `linked` key
		var linkedSlice []interface{}

		if ctx.root["linked"] == nil {
			linkedSlice = []interface{}{}
			ctx.root["linked"] = linkedSlice
		} else {
			linkedSlice = ctx.root["linked"].([]interface{})
		}

		// add to linked slice if not already present
		alreadyLinked := false
		for _, linked := range linkedSlice {
			m := reflect.ValueOf(linked).Interface().(map[string]interface{})
			if val["id"] == m["id"] && val["type"] == m["type"] {
				alreadyLinked = true
			}
		}

		if alreadyLinked == false {
			ctx.root["linked"] = append(linkedSlice, val)
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

// MarshalToJSONPrefix does the same as MarshalToJSON but adds a prefix to generated URLs
func MarshalToJSONPrefix(val interface{}, prefix string) ([]byte, error) {
	result, err := marshal(val, prefix)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
