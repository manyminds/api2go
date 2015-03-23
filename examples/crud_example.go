// examples.go show how to implement a basic crud for one data structure with the api2go server functionality
// to play with this example server you can for example run some of the following curl requests

// Create a new user:
// `curl -X POST http://localhost:31415/v0/users -d '{"data" : [{"type" : "users" , "username" : "marvin", "id" : "1"}]}'`
// List users:
// `curl -X GET http://localhost:31415/v0/users`
// Update:
// `curl -vX PUT http://localhost:31415/v0/users/1 -d '{ "data" : {"type" : "users", "username" : "better marvin", "id" : "1"}}'`
// Delete:
// `curl -vX DELETE http://localhost:31415/v0/users/2`
// FindMultiple (this only works if you've called create a bunch of times :)
// `curl -X GET http://localhost:31415/v0/users/3,4`
package main

import (
	"errors"
	"fmt"

	"github.com/univedo/api2go"
)

import "net/http"

//User is a generic database user
type User struct {
	ID           string
	Username     string
	PasswordHash string `json:"-"`
	exists       bool
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

// the user resource holds all users in the array
type userResource struct {
	users   map[string]User
	idCount int
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
func (s *userResource) Create(obj interface{}) (string, error) {
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

	s.users[id] = user

	return id, nil
}

// Delete to satisfy `api2go.DataSource` interface
func (s *userResource) Delete(id string) error {
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
func (s *userResource) Update(obj interface{}) error {
	user, ok := obj.(User)
	if !ok {
		return api2go.NewHTTPError(errors.New("Invalid instance given"), "Invalid instance given", http.StatusBadRequest)
	}

	s.users[user.GetID()] = user

	return nil
}

func main() {
	api := api2go.NewAPI("v0")
	users := make(map[string]User)
	api.AddResource(User{}, &userResource{users: users})

	fmt.Println("Listening on :31415")
	http.ListenAndServe(":31415", api.Handler())
}