package jsonapi

import "reflect"

// ConventionWrapper restores api2go compatibilty
// to the old version.
// simply let your struct be wrapped around
// ConventionWrapper and enjoy the old lazy
// way to use api2go.
// Deprecated will be removed in further
// releases. Its strongly advised to invest
// the time to implement api2gos interfaces.
type ConventionWrapper struct {
	Item interface{}
}

const (
	// NoIDFieldPresent will be returned by GetID if nothing could be found
	NoIDFieldPresent = "NoIDFieldPresent"
)

// GetID will search your struct via reflection
// and return the value within an field called
// ID, this field must be one of the types:
// int* uint* string
// will return -1 if the field was not found
func (c ConventionWrapper) GetID() string {
	val := reflect.ValueOf(c.Item)
	valType := val.Type()

	for i := 0; i < val.NumField(); i++ {
		tag := valType.Field(i).Tag.Get("json")
		if tag == "-" {
			continue
		}

		field := val.Field(i)
		keyName := Jsonify(valType.Field(i).Name)

		if keyName == "id" {
			id, _ := idFromValue(field)
			return id
		}
	}

	return NoIDFieldPresent
}

// GetReferences will dynamically search for TODO
func (c ConventionWrapper) GetReferences() []Reference {

	return []Reference{}
}

// GetReferencedIDs will dynamicaly search for TODO
func (c ConventionWrapper) GetReferencedIDs() []ReferenceID {

	return []ReferenceID{}
}

// GetReferencedStructs will dynamically search for TODO
func (c ConventionWrapper) GetReferencedStructs() []MarshalIdentifier {
	return []MarshalIdentifier{}
}
