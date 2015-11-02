package jsonapi

import (
	"bytes"
	"encoding/json"
	"errors"
)

// Document represents a JSONAPI document like specified: http://jsonapi.org
type Document struct {
	Links    *Links         `json:"links,omitempty"`
	Data     *DataContainer `json:"data"`
	Included []Data         `json:"included,omitempty"`
}

// DataContainer is needed to either keep "data" contents as array or object.
type DataContainer struct {
	DataObject *Data
	DataArray  []Data
}

// UnmarshalJSON implements Unmarshaler because we have to detect if payload is an object
// or an array
func (c *DataContainer) UnmarshalJSON(payload []byte) error {
	if bytes.HasPrefix(payload, []byte("{")) {
		// payload is an object
		return json.Unmarshal(payload, &c.DataObject)
	}

	if bytes.HasPrefix(payload, []byte("[")) {
		// payload is an array
		return json.Unmarshal(payload, &c.DataArray)
	}

	return errors.New("Invalid json for data array/object")
}

// MarshalJSON either Marshals an array or object of data
func (c *DataContainer) MarshalJSON() ([]byte, error) {
	if c.DataArray != nil {
		return json.Marshal(c.DataArray)
	}
	return json.Marshal(c.DataObject)
}

// Links is general links struct for top level and relationships
type Links struct {
	Self     string `json:"self,omitempty"`
	Related  string `json:"related,omitempty"`
	First    string `json:"first,omitempty"`
	Previous string `json:"previous,omitempty"`
	Next     string `json:"next,omitempty"`
	Last     string `json:"last,omitempty"`
}

// Data for top level and included data
type Data struct {
	Type          string                  `json:"type"`
	ID            string                  `json:"id"`
	Attributes    json.RawMessage         `json:"attributes"`
	Relationships map[string]Relationship `json:"relationships,omitempty"`
	Links         *Links                  `json:"links,omitempty"`
}

// Relationship contains reference IDs to the related structs
type Relationship struct {
	Links *Links                     `json:"links,omitempty"`
	Data  *RelationshipDataContainer `json:"data,omitempty"`
}

// RelationshipDataContainer is needed to either keep relationship "data" contents as array or object.
type RelationshipDataContainer struct {
	DataObject *RelationshipData
	DataArray  []RelationshipData
}

// UnmarshalJSON implements Unmarshaler and also detects array/object type
func (c *RelationshipDataContainer) UnmarshalJSON(payload []byte) error {
	if bytes.HasPrefix(payload, []byte("{")) {
		// payload is an object
		return json.Unmarshal(payload, &c.DataObject)
	}

	if bytes.HasPrefix(payload, []byte("[")) {
		// payload is an array
		return json.Unmarshal(payload, &c.DataArray)
	}

	return errors.New("Invalid json for relationship data array/object")
}

// MarshalJSON either Marshals an array or object of relationship data
func (c *RelationshipDataContainer) MarshalJSON() ([]byte, error) {
	if c.DataArray != nil {
		return json.Marshal(c.DataArray)
	}
	return json.Marshal(c.DataObject)
}

// RelationshipData represents one specific reference ID
type RelationshipData struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}
