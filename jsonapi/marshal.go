package jsonapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// MarshalIdentifier interface is necessary to give an element
// a unique ID. This interface must be implemented for
// marshal and unmarshal in order to let them store
// elements
type MarshalIdentifier interface {
	GetID() string
}

// ReferenceID contains all necessary information in order
// to reference another struct in jsonapi
type ReferenceID struct {
	ID   string
	Type string
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
	MarshalIdentifier
	GetReferencedIDs() []ReferenceID
}

// MarshalIncludedRelations must be implemented if referenced structs should be included
type MarshalIncludedRelations interface {
	MarshalReferences
	MarshalIdentifier
	GetReferencedStructs() []MarshalIdentifier
}

// ServerInformation can be passed to MarshalWithURLs to generate the `self` and `related` urls inside `links`
type ServerInformation interface {
	GetBaseURL() string
	GetPrefix() string
}

var serverInformationNil ServerInformation

// MarshalToJSON marshals a struct to json
// it works like `Marshal` but returns json instead
func MarshalToJSON(val interface{}) ([]byte, error) {
	result, err := Marshal(val)
	if err != nil {
		return []byte{}, err
	}

	return json.Marshal(result)
}

// MarshalToJSONWithURLs marshals a struct to json with URLs in `links`
func MarshalToJSONWithURLs(val interface{}, information ServerInformation) ([]byte, error) {
	result, err := MarshalWithURLs(val, information)
	if err != nil {
		return []byte{}, err
	}

	return json.Marshal(result)
}

// MarshalWithURLs can be used to include the generation of `related` and `self` links
func MarshalWithURLs(data interface{}, information ServerInformation) (map[string]interface{}, error) {
	return marshal(data, information)
}

// Marshal thats the input from `data` which can be a struct, a slice, or a pointer of it.
// Any struct in `data`or data itself, must at least implement the `MarshalIdentifier` interface.
// If so, it will generate a map[string]interface{} matching the jsonapi specification.
func Marshal(data interface{}) (map[string]interface{}, error) {
	return marshal(data, serverInformationNil)
}

func marshal(data interface{}, information ServerInformation) (map[string]interface{}, error) {
	if data == nil {
		return map[string]interface{}{}, errors.New("nil cannot be marshalled")
	}

	switch reflect.TypeOf(data).Kind() {
	case reflect.Slice:
		return marshalSlice(data, information)
	case reflect.Struct, reflect.Ptr:
		return marshalStruct(data.(MarshalIdentifier), information)
	default:
		return map[string]interface{}{}, errors.New("Marshal only accepts slice, struct or ptr types")
	}
}

func marshalSlice(data interface{}, information ServerInformation) (map[string]interface{}, error) {
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

		content, err := marshalData(element, information)
		if err != nil {
			return result, err
		}

		dataElements = append(dataElements, content)

		included, ok := k.(MarshalIncludedRelations)
		if ok {
			referencedStructs = append(referencedStructs, included.GetReferencedStructs()...)
		}
	}

	includedElements, err := reduceDuplicates(referencedStructs, information, marshalData)
	if err != nil {
		return result, err
	}

	//data key is always present
	result["data"] = dataElements
	if includedElements != nil && len(includedElements) > 0 {
		result["included"] = includedElements
	}

	return result, nil
}

// reduceDuplicates eliminates duplicate MarshalIdentifier from input and calls `method` on every unique MarshalIdentifier
func reduceDuplicates(
	input []MarshalIdentifier,
	information ServerInformation,
	method func(MarshalIdentifier, ServerInformation) (map[string]interface{}, error),
) (
	[]map[string]interface{},
	error,
) {
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
			marshalled, err := method(referencedStruct, information)
			if err != nil {
				return includedElements, err
			}

			includedElements = append(includedElements, marshalled)
			alreadyIncluded[structType][referencedStruct.GetID()] = true
		}
	}

	return includedElements, nil
}

func marshalData(element MarshalIdentifier, information ServerInformation) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	refValue := reflect.ValueOf(element)
	if refValue.Kind() == reflect.Ptr && refValue.IsNil() {
		return result, errors.New("MarshalIdentifier must not be nil")
	}

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
		result["links"] = getStructLinks(references, information)
	}

	return result, nil
}

// getStructLinks returns the link struct with ids
func getStructLinks(relationer MarshalLinkedRelations, information ServerInformation) map[string]map[string]interface{} {
	referencedIDs := relationer.GetReferencedIDs()
	sortedResults := make(map[string][]ReferenceID)
	links := make(map[string]map[string]interface{})

	for _, referenceID := range referencedIDs {
		sortedResults[referenceID.Type] = append(sortedResults[referenceID.Type], referenceID)
	}

	references := relationer.GetReferences()

	// helper mad to check if all references are included to also include mepty ones
	notIncludedReferences := map[string]Reference{}
	for _, reference := range references {
		notIncludedReferences[reference.Name] = reference
	}

	for referenceType, referenceIDs := range sortedResults {
		name := referenceIDs[0].Name
		links[name] = map[string]interface{}{}
		// if referenceType is plural, we need to use an array for linkage, otherwise it's just an object
		if Pluralize(name) == name {
			// multiple elements in links
			linkage := []map[string]interface{}{}

			for _, referenceID := range referenceIDs {
				linkage = append(linkage, map[string]interface{}{
					"type": referenceType,
					"id":   referenceID.ID,
				})
			}

			links[name]["linkage"] = linkage
		} else {
			links[name] = map[string]interface{}{
				"linkage": map[string]interface{}{
					"type": referenceType,
					"id":   referenceIDs[0].ID,
				},
			}
		}

		// set URLs if necessary
		for key, value := range getLinksForServerInformation(relationer, name, information) {
			links[name][key] = value
		}

		// this marks the reference as already included
		delete(notIncludedReferences, referenceIDs[0].Name)
	}

	// check for empty references
	for name := range notIncludedReferences {
		links[name] = map[string]interface{}{}
		// Plural empty relationships need an empty array and empty to-one need a null in the json
		if Pluralize(name) == name {
			links[name]["linkage"] = []interface{}{}
		} else {
			links[name]["linkage"] = nil
		}
		for key, value := range getLinksForServerInformation(relationer, name, information) {
			links[name][key] = value
		}
	}

	return links
}

// helper method to generate URL fields for `links`
func getLinksForServerInformation(relationer MarshalLinkedRelations, name string, information ServerInformation) map[string]string {
	links := map[string]string{}
	// generate links if necessary
	if information != serverInformationNil {
		prefix := ""
		baseURL := information.GetBaseURL()
		if baseURL != "" {
			prefix = baseURL
		}
		p := information.GetPrefix()
		if p != "" {
			prefix += "/" + p
		}

		if prefix != "" {
			links["self"] = fmt.Sprintf("%s/%s/%s/links/%s", prefix, getStructType(relationer), relationer.GetID(), name)
			links["related"] = fmt.Sprintf("%s/%s/%s/%s", prefix, getStructType(relationer), relationer.GetID(), name)
		} else {
			links["self"] = fmt.Sprintf("/%s/%s/links/%s", getStructType(relationer), relationer.GetID(), name)
			links["related"] = fmt.Sprintf("/%s/%s/%s", getStructType(relationer), relationer.GetID(), name)
		}
	}

	return links
}

func getIncludedStructs(included MarshalIncludedRelations, information ServerInformation) ([]map[string]interface{}, error) {
	var result = make([]map[string]interface{}, 0)
	includedStructs := included.GetReferencedStructs()

	for key := range includedStructs {
		marshalled, err := marshalData(includedStructs[key], information)
		if err != nil {
			return result, err
		}

		result = append(result, marshalled)
	}

	return result, nil
}

func marshalStruct(data MarshalIdentifier, information ServerInformation) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	contentData, err := marshalData(data, information)
	if err != nil {
		return result, err
	}

	result["data"] = contentData

	included, ok := data.(MarshalIncludedRelations)
	if ok {
		included, err := getIncludedStructs(included, information)
		if err != nil {
			return result, err
		}

		if len(included) > 0 {
			result["included"] = included
		}
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

		//skip private fields
		if !field.CanInterface() {
			continue
		}

		name := GetTagValueByName(valType.Field(i), "name")
		if name != "" {
			keyName = name
		}

		result[keyName] = field.Interface()
	}

	return result
}
