package model

import "github.com/manyminds/api2go/jsonapi"

// Chocolate is the chocolate that a user consumes in order to get fat and happy
type Chocolate struct {
	ID    string `json:"-"`
	LID   string `json:"-"`
	Name  string `json:"name"`
	Taste string `json:"taste"`
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (c Chocolate) GetID() jsonapi.Identifier {
	return jsonapi.Identifier{ID: c.ID, LID: c.LID}
}

// SetID to satisfy jsonapi.UnmarshalIdentifier interface
func (c *Chocolate) SetID(ID jsonapi.Identifier) error {
	c.ID = ID.ID
	c.LID = ID.LID
	return nil
}
