package jsonapi

import (
	"bytes"
	"encoding/json"
	"errors"
)

var objectSuffix = []byte("{")
var arraySuffix = []byte("[")

// A Document represents a JSON API document as specified here: http://jsonapi.org.
type Document struct {
	Links    *Links                 `json:"links,omitempty"`
	Data     *DataContainer         `json:"data"`
	Included []Data                 `json:"included,omitempty"`
	Meta     map[string]interface{} `json:"meta,omitempty"`
}

// A DataContainer is used to marshal and unmarshal single objects and arrays
// of objects.
type DataContainer struct {
	DataObject *Data
	DataArray  []Data
}

// UnmarshalJSON unmarshals the JSON-encoded data to the DataObject field if the
// root element is an object or to the DataArray field for arrays.
func (c *DataContainer) UnmarshalJSON(payload []byte) error {
	if bytes.HasPrefix(payload, objectSuffix) {
		return json.Unmarshal(payload, &c.DataObject)
	}

	if bytes.HasPrefix(payload, arraySuffix) {
		return json.Unmarshal(payload, &c.DataArray)
	}

	return errors.New("expected a JSON encoded object or array")
}

// MarshalJSON returns the JSON encoding of the DataArray field or the DataObject
// field. It will return "null" if neither of them is set.
func (c *DataContainer) MarshalJSON() ([]byte, error) {
	if c.DataArray != nil {
		return json.Marshal(c.DataArray)
	}

	return json.Marshal(c.DataObject)
}

// CustomLink represents a custom link for return in the document.
type CustomLink struct {
	Href string                 `json:"href"`
	Meta map[string]interface{} `json:"meta,omitempty"`
}

// CustomLinks contains a map of CustomLink objects as given by an element.
type CustomLinks map[string]CustomLink

// Links is a general struct for document links and relationship links.
type Links struct {
	Self     string `json:"self,omitempty"`
	Related  string `json:"related,omitempty"`
	First    string `json:"first,omitempty"`
	Previous string `json:"prev,omitempty"`
	Next     string `json:"next,omitempty"`
	Last     string `json:"last,omitempty"`
}

// Data is a general struct for document data and included data.
type Data struct {
	Type          string                  `json:"type"`
	ID            string                  `json:"id"`
	Attributes    json.RawMessage         `json:"attributes"`
	Relationships map[string]Relationship `json:"relationships,omitempty"`
	Links         map[string]interface{}  `json:"links,omitempty"`
}

// Relationship contains reference IDs to the related structs
type Relationship struct {
	Links *Links                     `json:"links,omitempty"`
	Data  *RelationshipDataContainer `json:"data,omitempty"`
	Meta  map[string]interface{}     `json:"meta,omitempty"`
}

// A RelationshipDataContainer is used to marshal and unmarshal single relationship
// objects and arrays of relationship objects.
type RelationshipDataContainer struct {
	DataObject *RelationshipData
	DataArray  []RelationshipData
}

// UnmarshalJSON unmarshals the JSON-encoded data to the DataObject field if the
// root element is an object or to the DataArray field for arrays.
func (c *RelationshipDataContainer) UnmarshalJSON(payload []byte) error {
	if bytes.HasPrefix(payload, objectSuffix) {
		// payload is an object
		return json.Unmarshal(payload, &c.DataObject)
	}

	if bytes.HasPrefix(payload, arraySuffix) {
		// payload is an array
		return json.Unmarshal(payload, &c.DataArray)
	}

	return errors.New("Invalid json for relationship data array/object")
}

// MarshalJSON returns the JSON encoding of the DataArray field or the DataObject
// field. It will return "null" if neither of them is set.
func (c *RelationshipDataContainer) MarshalJSON() ([]byte, error) {
	if c.DataArray != nil {
		return json.Marshal(c.DataArray)
	}
	return json.Marshal(c.DataObject)
}

// RelationshipData represents one specific reference ID.
type RelationshipData struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}
