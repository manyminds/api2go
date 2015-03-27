// examples.go show how to implement a basic crud for one data structure with the api2go server functionality
// to play with this example server you can for example run some of the following curl requests

// Create a new user:
// `curl -X POST http://localhost:31415/v0/users -d '{"data" : [{"type" : "users" , "username" : "marvin"}]}'`
// List users:
// `curl -X GET http://localhost:31415/v0/users`
// Update:
// `curl -vX PUT http://localhost:31415/v0/users/1 -d '{ "data" : {"type" : "users", "username" : "better marvin", "id" : "1"}}'`
// Delete:
// `curl -vX DELETE http://localhost:31415/v0/users/2`
// FindMultiple (this only works if you've called create a bunch of times :)
// `curl -X GET http://localhost:31415/v0/users/3,4`
// Create a chocolate with the name sweet
// `curl -X POST http://localhost:31415/v0/chocolates -d '{"data" : [{"type" : "chocolates" , "name" : "Ritter Sport", "taste": "Very Good"}]}'`
// Link the sweet
// `curl -X POST http://localhost:31415/v0/users -d '{"data" : [{"type" : "users" , "username" : "marvin", "links": {"sweets": {"type": "chocolates", "ids": ["1"]}}}]}'`
package main

import (
	"errors"
	"fmt"

	"github.com/univedo/api2go"
	"github.com/univedo/api2go/jsonapi"
)

import "net/http"

//User is a generic database user
type User struct {
	ID            string
	Username      string
	PasswordHash  string      `json:"-"`
	Chocolates    []Chocolate `json:"-"`
	ChocolatesIDs []string    `json:"-"`
	exists        bool
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (u User) GetID() string {
	return u.ID
}

// SetID to satisfy jsonapi.UnmarshalIdentifier interface
func (u *User) SetID(id string) error {
	u.ID = id
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
	for _, chocolate := range u.Chocolates {
		result = append(result, jsonapi.ReferenceID{
			ID:   chocolate.ID,
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

// SetReferencedIDs to satisfy the jsonapi.UnmarshalLinkedRelations interface
func (u *User) SetReferencedIDs(references []jsonapi.ReferenceID) error {
	for _, reference := range references {
		if reference.Name == "sweets" {
			u.ChocolatesIDs = append(u.ChocolatesIDs, reference.ID)
		}
	}

	return nil
}

// Chocolate is the chocolate that a user consumes in order to get fat and happy
type Chocolate struct {
	ID    string
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

// ChocolateStorage stores all of the tasty chocolate, needs to be injected into
// User and Chocolate Resource. In the real world, you would use a database for that.
type ChocolateStorage struct {
	chocolates map[string]Chocolate
	idCount    int
}

// GetAll of the chocolate
func (s ChocolateStorage) GetAll() []Chocolate {
	result := []Chocolate{}
	for key := range s.chocolates {
		result = append(result, s.chocolates[key])
	}

	return result
}

// GetOne tasty chocolate
func (s ChocolateStorage) GetOne(id string) (Chocolate, error) {
	choc, ok := s.chocolates[id]
	if ok {
		return choc, nil
	}

	return Chocolate{}, fmt.Errorf("Chocolate for id %s not found", id)
}

// Insert a fresh one
func (s *ChocolateStorage) Insert(c Chocolate) string {
	id := fmt.Sprintf("%d", s.idCount)
	c.ID = id
	s.chocolates[id] = c
	s.idCount++
	return id
}

// Delete one :(
func (s *ChocolateStorage) Delete(id string) error {
	_, exists := s.chocolates[id]
	if !exists {
		return fmt.Errorf("Chocolate with id %s does not exist", id)
	}
	delete(s.chocolates, id)

	return nil
}

// Update updates an existing chocolate
func (s *ChocolateStorage) Update(c Chocolate) error {
	_, exists := s.chocolates[c.ID]
	if !exists {
		return fmt.Errorf("Chocolate with id %s does not exist", c.ID)
	}
	s.chocolates[c.ID] = c

	return nil
}

// the user resource holds all users in the array
type userResource struct {
	chocStorage *ChocolateStorage
	users       map[string]User
	idCount     int
}

// FindAll to satisfy api2go data source interface
func (s *userResource) FindAll(r api2go.Request) (interface{}, error) {
	var users []User

	for _, value := range s.users {
		users = append(users, value)
	}

	return users, nil
}

// FindOne to satisfy `api2go.DataSource` interface
// this method should return the user with the given ID, otherwise an error
func (s *userResource) FindOne(ID string, r api2go.Request) (interface{}, error) {
	if user, ok := s.users[ID]; ok {
		return user, nil
	}

	return nil, api2go.NewHTTPError(errors.New("Not Found"), "Not Found", http.StatusNotFound)
}

// FindMultiple to satifiy `api2go.DataSource` interface
func (s *userResource) FindMultiple(IDs []string, r api2go.Request) (interface{}, error) {
	var users []User

	for _, id := range IDs {
		user, err := s.FindOne(id, r)
		if err != nil {
			return nil, err
		}

		if typedUser, ok := user.(User); ok {
			users = append(users, typedUser)
		}
	}

	return users, nil
}

// Create method to satisfy `api2go.DataSource` interface
func (s *userResource) Create(obj interface{}, r api2go.Request) (string, error) {
	user, ok := obj.(User)
	if !ok {
		return "", api2go.NewHTTPError(errors.New("Invalid instance given"), "Invalid instance given", http.StatusBadRequest)
	}

	if _, ok := s.users[user.GetID()]; ok {
		return "", api2go.NewHTTPError(errors.New("User exists"), "User exists", http.StatusConflict)
	}

	s.idCount++
	id := fmt.Sprintf("%d", s.idCount)
	user.SetID(id)

	// check references and get embedded objects
	for _, chocID := range user.ChocolatesIDs {
		choc, err := s.chocStorage.GetOne(chocID)
		if err != nil {
			return "", err
		}

		user.Chocolates = append(user.Chocolates, choc)
	}

	s.users[id] = user

	return id, nil
}

// Delete to satisfy `api2go.DataSource` interface
func (s *userResource) Delete(id string, r api2go.Request) error {
	obj, err := s.FindOne(id, api2go.Request{})
	if err != nil {
		return err
	}

	user, ok := obj.(User)
	if !ok {
		return errors.New("Invalid instance given")
	}

	delete(s.users, user.GetID())

	return nil
}

//Update stores all changes on the user
func (s *userResource) Update(obj interface{}, r api2go.Request) error {
	user, ok := obj.(User)
	if !ok {
		return api2go.NewHTTPError(errors.New("Invalid instance given"), "Invalid instance given", http.StatusBadRequest)
	}

	// check references and get embedded objects
	for _, chocID := range user.ChocolatesIDs {
		choc, err := s.chocStorage.GetOne(chocID)
		if err != nil {
			return err
		}

		user.Chocolates = append(user.Chocolates, choc)
	}

	s.users[user.GetID()] = user

	return nil
}

type chocolateResource struct {
	storage *ChocolateStorage
}

func (c *chocolateResource) FindAll(r api2go.Request) (interface{}, error) {
	return c.storage.GetAll(), nil
}

func (c *chocolateResource) FindOne(ID string, r api2go.Request) (interface{}, error) {
	return c.storage.GetOne(ID)
}

func (c *chocolateResource) FindMultiple(IDs []string, r api2go.Request) (interface{}, error) {
	var chocolates []Chocolate

	for _, id := range IDs {
		choc, err := c.FindOne(id, r)
		if err != nil {
			return nil, err
		}

		if typedChoc, ok := choc.(Chocolate); ok {
			chocolates = append(chocolates, typedChoc)
		}
	}

	return chocolates, nil
}

func (c *chocolateResource) Create(obj interface{}, r api2go.Request) (string, error) {
	choc, ok := obj.(Chocolate)
	if !ok {
		return "", api2go.NewHTTPError(errors.New("Invalid instance given"), "Invalid instance given", http.StatusBadRequest)
	}

	return c.storage.Insert(choc), nil
}

func (c *chocolateResource) Delete(id string, r api2go.Request) error {
	return c.storage.Delete(id)
}

func (c *chocolateResource) Update(obj interface{}, r api2go.Request) error {
	choc, ok := obj.(Chocolate)
	if !ok {
		return api2go.NewHTTPError(errors.New("Invalid instance given"), "Invalid instance given", http.StatusBadRequest)
	}

	return c.storage.Update(choc)
}

func main() {
	api := api2go.NewAPI("v0")
	users := make(map[string]User)
	chocStorage := ChocolateStorage{chocolates: make(map[string]Chocolate), idCount: 1}
	api.AddResource(User{}, &userResource{users: users, chocStorage: &chocStorage})
	api.AddResource(Chocolate{}, &chocolateResource{storage: &chocStorage})

	fmt.Println("Listening on :31415")
	http.ListenAndServe(":31415", api.Handler())
}
