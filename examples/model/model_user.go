package model

import (
	"errors"

	"github.com/manyminds/api2go/jsonapi"
)

// User is a generic database user
type User struct {
	ID  string `json:"-"`
	LID string `json:"-"`
	//rename the username field to user-name.
	Username      string       `json:"user-name"`
	PasswordHash  string       `json:"-"`
	Chocolates    []*Chocolate `json:"-"`
	ChocolatesIDs []string     `json:"-"`
	exists        bool
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (u User) GetID() jsonapi.Identifier {
	return jsonapi.Identifier{ID: u.ID, LID: u.LID}
}

// SetID to satisfy jsonapi.UnmarshalIdentifier interface
func (u *User) SetID(ID jsonapi.Identifier) error {
	u.ID = ID.ID
	u.LID = ID.LID
	return nil
}

// GetReferences to satisfy the jsonapi.MarshalReferences interface
func (u User) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{
			Type: "chocolates",
			Name: "sweets",
		},
	}
}

// GetReferencedIDs to satisfy the jsonapi.MarshalLinkedRelations interface
func (u User) GetReferencedIDs() []jsonapi.ReferenceID {
	result := []jsonapi.ReferenceID{}
	for _, chocolateID := range u.ChocolatesIDs {
		result = append(result, jsonapi.ReferenceID{
			ID:   chocolateID,
			Type: "chocolates",
			Name: "sweets",
		})
	}

	return result
}

// GetReferencedStructs to satisfy the jsonapi.MarhsalIncludedRelations interface
func (u User) GetReferencedStructs() []jsonapi.MarshalIdentifier {
	result := []jsonapi.MarshalIdentifier{}
	for key := range u.Chocolates {
		result = append(result, u.Chocolates[key])
	}

	return result
}

// SetToManyReferenceIDs sets the sweets reference IDs and satisfies the jsonapi.UnmarshalToManyRelations interface
func (u *User) SetToManyReferenceIDs(name string, IDs []jsonapi.Identifier) error {
	if name == "sweets" {
		u.ChocolatesIDs = make([]string, 0, len(IDs))
		for _, id := range IDs {
			u.ChocolatesIDs = append(u.ChocolatesIDs, id.ID)
		}
		return nil
	}

	return errors.New("There is no to-many relationship with the name " + name)
}

// AddToManyIDs adds some new sweets that a users loves so much
func (u *User) AddToManyIDs(name string, IDs []string) error {
	if name == "sweets" {
		u.ChocolatesIDs = append(u.ChocolatesIDs, IDs...)
		return nil
	}

	return errors.New("There is no to-many relationship with the name " + name)
}

// DeleteToManyIDs removes some sweets from a users because they made him very sick
func (u *User) DeleteToManyIDs(name string, IDs []string) error {
	if name == "sweets" {
		for _, ID := range IDs {
			for pos, oldID := range u.ChocolatesIDs {
				if ID == oldID {
					// match, this ID must be removed
					u.ChocolatesIDs = append(u.ChocolatesIDs[:pos], u.ChocolatesIDs[pos+1:]...)
				}
			}
		}
	}

	return errors.New("There is no to-many relationship with the name " + name)
}
