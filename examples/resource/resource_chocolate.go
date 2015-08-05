package resource

import (
	"errors"
	"net/http"

	"github.com/manyminds/api2go"
	"github.com/manyminds/api2go/examples/model"
	"github.com/manyminds/api2go/examples/storage"
)

// ChocolateResource for api2go routes
type ChocolateResource struct {
	ChocStorage *storage.ChocolateStorage
	UserStorage *storage.UserStorage
}

// FindAll chocolates
func (c ChocolateResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	usersID, ok := r.QueryParams["usersID"]
	sweets := c.ChocStorage.GetAll()
	if ok {
		// this means that we want to show all sweets of a user, this is the route
		// /v0/users/1/sweets
		userID := usersID[0]
		// filter out sweets with userID, in real world, you would just run a different database query
		filteredSweets := []model.Chocolate{}
		user, err := c.UserStorage.GetOne(userID)
		if err != nil {
			return &Response{}, err
		}
		for _, sweetID := range user.ChocolatesIDs {
			sweet, err := c.ChocStorage.GetOne(sweetID)
			if err != nil {
				return &Response{}, err
			}
			filteredSweets = append(filteredSweets, sweet)
		}

		return &Response{Res: filteredSweets}, nil
	}
	return &Response{Res: sweets}, nil
}

// FindOne choc
func (c ChocolateResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	res, err := c.ChocStorage.GetOne(ID)
	return &Response{Res: res}, err
}

// Create a new choc
func (c ChocolateResource) Create(obj interface{}, r api2go.Request) (api2go.Responder, error) {
	choc, ok := obj.(model.Chocolate)
	if !ok {
		return &Response{}, api2go.NewHTTPError(errors.New("Invalid instance given"), "Invalid instance given", http.StatusBadRequest)
	}

	id := c.ChocStorage.Insert(choc)
	choc.ID = id
	return &Response{Res: choc, Code: http.StatusCreated}, nil
}

// Delete a choc :(
func (c ChocolateResource) Delete(id string, r api2go.Request) (api2go.Responder, error) {
	err := c.ChocStorage.Delete(id)
	return &Response{Code: http.StatusOK}, err
}

// Update a choc
func (c ChocolateResource) Update(obj interface{}, r api2go.Request) (api2go.Responder, error) {
	choc, ok := obj.(model.Chocolate)
	if !ok {
		return &Response{}, api2go.NewHTTPError(errors.New("Invalid instance given"), "Invalid instance given", http.StatusBadRequest)
	}

	err := c.ChocStorage.Update(choc)
	return &Response{Res: choc, Code: http.StatusNoContent}, err
}
