package api2go

import "encoding/json"

// JSONContentMarshaler uses the standard encoding/json package for
// decoding requests and encoding responses in JSON format.
type JSONContentMarshaler struct {
}

// Marshal marshals with default JSON
func (m JSONContentMarshaler) Marshal(i interface{}) ([]byte, error) {
	return json.Marshal(i)
}

// Unmarshal with default JSON
func (m JSONContentMarshaler) Unmarshal(data []byte, i interface{}) error {
	return json.Unmarshal(data, i)
}
