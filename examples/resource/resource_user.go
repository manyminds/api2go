package resource

import (
	"errors"
	"net/http"
	"sort"
	"strconv"

	"github.com/manyminds/api2go"
	"github.com/manyminds/api2go/examples/model"
	"github.com/manyminds/api2go/examples/storage"
)

// UserResource for api2go routes
type UserResource struct {
	ChocStorage *storage.ChocolateStorage
	UserStorage *storage.UserStorage
}

// FindAll to satisfy api2go data source interface
func (s UserResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	var result []model.User
	users := s.UserStorage.GetAll()

	for _, user := range users {
		// get all sweets for the user
		user.Chocolates = []*model.Chocolate{}
		for _, chocolateID := range user.ChocolatesIDs {
			choc, err := s.ChocStorage.GetOne(chocolateID)
			if err != nil {
				return &Response{}, err
			}
			user.Chocolates = append(user.Chocolates, &choc)
		}
		result = append(result, *user)
	}

	return &Response{Res: result}, nil
}

// PaginatedFindAll can be used to load users in chunks
func (s UserResource) PaginatedFindAll(r api2go.Request) (uint, api2go.Responder, error) {
	var (
		result                      []model.User
		number, size, offset, limit string
		keys                        []int
	)
	users := s.UserStorage.GetAll()

	for k := range users {
		i, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			return 0, &Response{}, err
		}

		keys = append(keys, int(i))
	}
	sort.Ints(keys)

	numberQuery, ok := r.QueryParams["page[number]"]
	if ok {
		number = numberQuery[0]
	}
	sizeQuery, ok := r.QueryParams["page[size]"]
	if ok {
		size = sizeQuery[0]
	}
	offsetQuery, ok := r.QueryParams["page[offset]"]
	if ok {
		offset = offsetQuery[0]
	}
	limitQuery, ok := r.QueryParams["page[limit]"]
	if ok {
		limit = limitQuery[0]
	}

	if size != "" {
		sizeI, err := strconv.ParseUint(size, 10, 64)
		if err != nil {
			return 0, &Response{}, err
		}

		numberI, err := strconv.ParseUint(number, 10, 64)
		if err != nil {
			return 0, &Response{}, err
		}

		start := sizeI * (numberI - 1)
		for i := start; i < start+sizeI; i++ {
			if i >= uint64(len(users)) {
				break
			}
			result = append(result, *users[strconv.FormatInt(int64(keys[i]), 10)])
		}
	} else {
		limitI, err := strconv.ParseUint(limit, 10, 64)
		if err != nil {
			return 0, &Response{}, err
		}

		offsetI, err := strconv.ParseUint(offset, 10, 64)
		if err != nil {
			return 0, &Response{}, err
		}

		for i := offsetI; i < offsetI+limitI; i++ {
			if i >= uint64(len(users)) {
				break
			}
			result = append(result, *users[strconv.FormatInt(int64(keys[i]), 10)])
		}
	}

	return uint(len(users)), &Response{Res: result}, nil
}

// FindOne to satisfy `api2go.DataSource` interface
// this method should return the user with the given ID, otherwise an error
func (s UserResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	user, err := s.UserStorage.GetOne(ID)
	if err != nil {
		return &Response{}, api2go.NewHTTPError(err, err.Error(), http.StatusNotFound)
	}

	user.Chocolates = []*model.Chocolate{}
	for _, chocolateID := range user.ChocolatesIDs {
		choc, err := s.ChocStorage.GetOne(chocolateID)
		if err != nil {
			return &Response{}, err
		}
		user.Chocolates = append(user.Chocolates, &choc)
	}
	return &Response{Res: user}, nil
}

// Create method to satisfy `api2go.DataSource` interface
func (s UserResource) Create(obj interface{}, r api2go.Request) (api2go.Responder, error) {
	user, ok := obj.(model.User)
	if !ok {
		return &Response{}, api2go.NewHTTPError(errors.New("Invalid instance given"), "Invalid instance given", http.StatusBadRequest)
	}

	id := s.UserStorage.Insert(user)
	user.ID = id

	return &Response{Res: user, Code: http.StatusCreated}, nil
}

// Delete to satisfy `api2go.DataSource` interface
func (s UserResource) Delete(id string, r api2go.Request) (api2go.Responder, error) {
	err := s.UserStorage.Delete(id)
	return &Response{Code: http.StatusNoContent}, err
}

//Update stores all changes on the user
func (s UserResource) Update(obj interface{}, r api2go.Request) (api2go.Responder, error) {
	user, ok := obj.(model.User)
	if !ok {
		return &Response{}, api2go.NewHTTPError(errors.New("Invalid instance given"), "Invalid instance given", http.StatusBadRequest)
	}

	err := s.UserStorage.Update(user)
	return &Response{Res: user, Code: http.StatusNoContent}, err
}
