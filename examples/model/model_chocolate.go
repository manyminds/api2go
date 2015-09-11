package model

// Chocolate is the chocolate that a user consumes in order to get fat and happy
type Chocolate struct {
	ID    string `jsonapi:"-"`
	Name  string
	Taste string
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (c Chocolate) GetID() string {
	return c.ID
}

// SetID to satisfy jsonapi.UnmarshalIdentifier interface
func (c *Chocolate) SetID(id string) error {
	c.ID = id
	return nil
}
