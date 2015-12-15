package jsonapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
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
// If IsNotLoaded is set to true, the `data` field will be omitted and only the `links` object will be
// generated. You should do this if there are some references, but you do not want to load them.
// Otherwise, if IsNotLoaded is false and GetReferencedIDs() returns no IDs for this reference name, an
// empty `data` field will be added which means that there are no references.
type Reference struct {
	Type        string
	Name        string
	IsNotLoaded bool
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

// MarshalWithURLs can be used to include the generation of `related` and `self` links
func MarshalWithURLs(data interface{}, information ServerInformation) ([]byte, error) {
	document, err := MarshalToStruct(data, information)
	if err != nil {
		return []byte(""), err
	}

	return json.Marshal(document)
}

// Marshal thats the input from `data` which can be a struct, a slice, or a pointer of it.
// Any struct in `data`or data itself, must at least implement the `MarshalIdentifier` interface.
// If so, it will generate a map[string]interface{} matching the jsonapi specification.
func Marshal(data interface{}) ([]byte, error) {
	document, err := MarshalToStruct(data, serverInformationNil)
	if err != nil {
		return []byte(""), err
	}
	return json.Marshal(document)
}

// MarshalToStruct marshals an api2go compatible struct into a jsonapi Document structure which then can be
// marshaled to JSON. You only need this method if you want to extract or extend parts of the document.
// You should directly use Marshal to get a []byte with JSON in it.
func MarshalToStruct(data interface{}, information ServerInformation) (Document, error) {
	if data == nil {
		return Document{}, nil
	}

	switch reflect.TypeOf(data).Kind() {
	case reflect.Slice:
		return marshalSlice(data, information)
	case reflect.Struct, reflect.Ptr:
		return marshalStruct(data.(MarshalIdentifier), information)
	default:
		return Document{}, errors.New("Marshal only accepts slice, struct or ptr types")
	}
}

func marshalSlice(data interface{}, information ServerInformation) (Document, error) {
	result := Document{}

	val := reflect.ValueOf(data)
	dataElements := []Data{}
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

		dataElements = append(dataElements, *content)

		included, ok := k.(MarshalIncludedRelations)
		if ok {
			referencedStructs = append(referencedStructs, included.GetReferencedStructs()...)
		}
	}

	includedElements := reduceDuplicates(referencedStructs, information, marshalData)

	//data key is always present
	result.Data = &DataContainer{
		DataArray: dataElements,
	}
	if includedElements != nil && len(*includedElements) > 0 {
		result.Included = *includedElements
	}

	return result, nil
}

// reduceDuplicates eliminates duplicate MarshalIdentifier from input and calls `method` on every unique MarshalIdentifier
func reduceDuplicates(
	input []MarshalIdentifier,
	information ServerInformation,
	method func(MarshalIdentifier, ServerInformation) (*Data, error),
) *[]Data {
	alreadyIncluded := map[string]map[string]bool{}
	includedElements := []Data{}

	for _, referencedStruct := range input {
		structType := getStructType(referencedStruct)
		if alreadyIncluded[structType] == nil {
			alreadyIncluded[structType] = make(map[string]bool)
		}

		if !alreadyIncluded[structType][referencedStruct.GetID()] {
			marshalled, _ := method(referencedStruct, information)
			includedElements = append(includedElements, *marshalled)
			alreadyIncluded[structType][referencedStruct.GetID()] = true
		}
	}

	return &includedElements
}

func marshalData(element MarshalIdentifier, information ServerInformation) (*Data, error) {
	var err error
	result := &Data{}

	refValue := reflect.ValueOf(element)
	if refValue.Kind() == reflect.Ptr && refValue.IsNil() {
		return result, errors.New("MarshalIdentifier must not be nil")
	}

	attributes, err := json.Marshal(element)
	result.Attributes = attributes
	result.ID = element.GetID()
	result.Type = getStructType(element)

	// optional relationship interface for struct
	references, ok := element.(MarshalLinkedRelations)
	if ok {
		result.Relationships = *getStructRelationships(references, information)
	}

	return result, err
}

// getStructRelationships returns the relationships struct with ids
func getStructRelationships(relationer MarshalLinkedRelations, information ServerInformation) *map[string]Relationship {
	referencedIDs := relationer.GetReferencedIDs()
	sortedResults := map[string][]ReferenceID{}
	relationships := map[string]Relationship{}

	for _, referenceID := range referencedIDs {
		sortedResults[referenceID.Name] = append(sortedResults[referenceID.Name], referenceID)
	}

	references := relationer.GetReferences()

	// helper mad to check if all references are included to also include mepty ones
	notIncludedReferences := map[string]Reference{}
	for _, reference := range references {
		notIncludedReferences[reference.Name] = reference
	}

	for name, referenceIDs := range sortedResults {
		relationships[name] = Relationship{}
		// if referenceType is plural, we need to use an array for data, otherwise it's just an object
		container := RelationshipDataContainer{}
		if Pluralize(name) == name {
			// multiple elements in links
			container.DataArray = []RelationshipData{}
			for _, referenceID := range referenceIDs {
				container.DataArray = append(container.DataArray, RelationshipData{
					Type: referenceID.Type,
					ID:   referenceID.ID,
				})
			}
		} else {
			container.DataObject = &RelationshipData{
				Type: referenceIDs[0].Type,
				ID:   referenceIDs[0].ID,
			}
		}

		// set URLs if necessary
		links := getLinksForServerInformation(relationer, name, information)

		relationship := Relationship{
			Data:  &container,
			Links: links,
		}

		relationships[name] = relationship

		// this marks the reference as already included
		delete(notIncludedReferences, referenceIDs[0].Name)
	}

	// check for empty references
	for name, reference := range notIncludedReferences {
		container := RelationshipDataContainer{}
		// Plural empty relationships need an empty array and empty to-one need a null in the json
		if !reference.IsNotLoaded && Pluralize(name) == name {
			container.DataArray = []RelationshipData{}
		}

		links := getLinksForServerInformation(relationer, name, information)
		relationship := Relationship{
			Links: links,
		}

		// skip relationship data completely if IsNotLoaded is set
		if !reference.IsNotLoaded {
			relationship.Data = &container
		}

		relationships[name] = relationship
	}

	return &relationships
}

// helper method to generate URL fields for `links`
func getLinksForServerInformation(relationer MarshalLinkedRelations, name string, information ServerInformation) *Links {
	links := &Links{}

	if information != serverInformationNil {
		prefix := strings.Trim(information.GetBaseURL(), "/")
		namespace := strings.Trim(information.GetPrefix(), "/")
		structType := getStructType(relationer)

		if namespace != "" {
			prefix += "/" + namespace
		}

		links.Self = fmt.Sprintf("%s/%s/%s/relationships/%s", prefix, structType, relationer.GetID(), name)
		links.Related = fmt.Sprintf("%s/%s/%s/%s", prefix, structType, relationer.GetID(), name)

		return links
	}

	return nil
}

func getIncludedStructs(included MarshalIncludedRelations, information ServerInformation) (*[]Data, error) {
	result := []Data{}
	includedStructs := included.GetReferencedStructs()

	for key := range includedStructs {
		marshalled, err := marshalData(includedStructs[key], information)
		if err != nil {
			return &result, err
		}

		result = append(result, *marshalled)
	}

	return &result, nil
}

func marshalStruct(data MarshalIdentifier, information ServerInformation) (Document, error) {
	result := Document{}
	contentData, err := marshalData(data, information)
	if err != nil {
		return result, err
	}

	result.Data = &DataContainer{
		DataObject: contentData,
	}

	included, ok := data.(MarshalIncludedRelations)
	if ok {
		included, err := getIncludedStructs(included, information)
		if err != nil {
			return result, err
		}

		if len(*included) > 0 {
			result.Included = *included
		}
	}

	return result, nil
}

func getStructType(data interface{}) string {
	entityName, ok := data.(EntityNamer)
	if ok {
		return entityName.GetName()
	}

	reflectType := reflect.TypeOf(data)
	if reflectType.Kind() == reflect.Ptr {
		return Pluralize(Jsonify(reflectType.Elem().Name()))
	}

	return Pluralize(Jsonify(reflectType.Name()))
}
