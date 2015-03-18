package jsonapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
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

// MarshalIdentifier interface is necessary to give an element
// a unique ID. This interface must be implemented for
// marshal and unmarshal in order to let them store
// elements
type MarshalIdentifier interface {
	GetID() string
}

// ReferenceID todo later
type ReferenceID struct {
	ID   string
	Type string // Todo: Must be removed, is redundant because it's already in `Reference` struct
	Name string
}

// Reference information about possible references of a struct
type Reference struct {
	Type string
	Name string
}

// MarshalReferences must be implemented if the struct to be serialized has relations. This must be done
// because jsonapi needs information about relations even if many to many relations or many to one relations
// are empty
type MarshalReferences interface {
	GetReferences() []Reference
}

// MarshalLinkedRelations must be implemented if there are references and the reference IDs should be included
type MarshalLinkedRelations interface {
	MarshalReferences
	GetReferencedIDs() []ReferenceID
}

// MarshalIncludedRelations must be implemented if referenced structs should be included
type MarshalIncludedRelations interface {
	MarshalReferences
	GetReferencedStructs() []MarshalIdentifier
}

// MarshalPrefix does the same as Marshal but adds a prefix to generated URLs
func MarshalPrefix(data interface{}, prefix string) (interface{}, error) {
	return nil, errors.New("Will never be implemented, must be moved into API layer")
}

// Marshal is the new shit
func Marshal(data interface{}) (map[string]interface{}, error) {
	if data == nil {
		return map[string]interface{}{}, errors.New("nil cannot be marshalled")
	}

	switch reflect.TypeOf(data).Kind() {
	case reflect.Slice:
		return marshalSlice(data)
	case reflect.Struct:
		return marshalStruct(data.(MarshalIdentifier), "")
	default:
		return map[string]interface{}{}, errors.New("Marshal only accepts slice, struct or ptr types")
	}
}

// marshalSlice marshals a slice TODO
func marshalSlice(data interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Slice {
		return result, errors.New("data must be a slice")
	}

	dataElements := []map[string]interface{}{}
	var referencedStructs []MarshalIdentifier

	for i := 0; i < val.Len(); i++ {
		k := val.Index(i).Interface()
		element, ok := k.(MarshalIdentifier)
		if !ok {
			return result, errors.New("all elements within the slice must implement api2go.MarshalIdentifier")
		}

		content, err := marshalData(element)
		if err != nil {
			return result, err
		}

		dataElements = append(dataElements, content)

		included, ok := k.(MarshalIncludedRelations)
		if ok {
			referencedStructs = append(referencedStructs, included.GetReferencedStructs()...)
		}
	}

	includedElements, err := reduceDuplicates(referencedStructs, marshalData)
	if err != nil {
		return result, err
	}

	//data key is always present
	result["data"] = dataElements

	//included elements is only present when included
	if includedElements != nil {
		result["linked"] = includedElements
	}

	return result, nil
}

// reduceDuplicates eliminates duplicate MarshalIdentifier from input and calls `method` on every unique MarshalIdentifier
func reduceDuplicates(input []MarshalIdentifier, method func(MarshalIdentifier) (map[string]interface{}, error)) ([]map[string]interface{}, error) {
	var (
		alreadyIncluded  = make(map[string]map[string]bool)
		includedElements []map[string]interface{}
	)

	for _, referencedStruct := range input {
		if referencedStruct == nil {
			continue
		}

		structType := getStructType(referencedStruct)
		if alreadyIncluded[structType] == nil {
			alreadyIncluded[structType] = make(map[string]bool)
		}

		if !alreadyIncluded[structType][referencedStruct.GetID()] {
			marshalled, err := method(referencedStruct)
			if err != nil {
				return includedElements, err
			}

			includedElements = append(includedElements, marshalled)
			alreadyIncluded[structType][referencedStruct.GetID()] = true
		}
	}

	return includedElements, nil
}

func marshalData(element MarshalIdentifier) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	refValue := reflect.ValueOf(element)
	if refValue.Kind() == reflect.Ptr && refValue.IsNil() {
		return result, errors.New("MarshalIdentifier must not be nil")
	}

	fmt.Printf("\n%#v", element)

	id := element.GetID()
	content := getStructFields(element)
	for k, v := range content {
		result[k] = v
	}

	// its important that the id from the interface
	// gets added afterwards, otherwise an ID field
	// could conflict with the actual marshalling
	result["id"] = id
	result["type"] = getStructType(element)

	// optional relationship interface for struct
	references, ok := element.(MarshalLinkedRelations)
	if ok {
		result["links"] = getStructLinks(references)
	}

	return result, nil
}

// getStructLinks returns the link struct with ids
func getStructLinks(relationer MarshalLinkedRelations) map[string]interface{} {
	referencedIDs := relationer.GetReferencedIDs()
	sortedResults := make(map[string][]ReferenceID)
	links := make(map[string]interface{})

	for _, referenceID := range referencedIDs {
		sortedResults[referenceID.Type] = append(sortedResults[referenceID.Type], referenceID)
	}

	for referenceType, referenceIDs := range sortedResults {
		switch len(sortedResults[referenceType]) {
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

func getIncludedStructs(included MarshalIncludedRelations) ([]map[string]interface{}, error) {
	var result = make([]map[string]interface{}, 0)
	includedStructs := included.GetReferencedStructs()

	for key := range includedStructs {
		marshalled, err := marshalData(includedStructs[key])
		if err != nil {
			return result, err
		}

		result = append(result, marshalled)
	}

	return result, nil
}

func marshalStruct(data MarshalIdentifier, prefix string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	contentData, err := marshalData(data)
	if err != nil {
		return result, err
	}

	result["data"] = contentData

	included, ok := data.(MarshalIncludedRelations)
	if ok {
		linked, err := getIncludedStructs(included)
		if err != nil {
			return result, err
		}

		result["linked"] = linked
	}

	return result, nil
}

func getStructType(data MarshalIdentifier) string {
	reflectType := reflect.TypeOf(data)
	if reflectType.Kind() == reflect.Ptr {
		return Pluralize(Jsonify(reflectType.Elem().Name()))
	}

	return Pluralize(Jsonify(reflectType.Name()))
}

func getStructFields(data MarshalIdentifier) map[string]interface{} {
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

// MarshalToJSON marshals a struct to json
func MarshalToJSON(val interface{}) ([]byte, error) {
	result, err := Marshal(val)
	if err != nil {
		return []byte{}, err
	}

	return json.Marshal(result)
}

// MarshalToJSONPrefix does the same as MarshalToJSON but adds a prefix to generated URLs
func MarshalToJSONPrefix(val interface{}, prefix string) ([]byte, error) {
	//TODO must either be implemented with prefix or removed
	result, err := Marshal(val)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
