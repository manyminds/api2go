package storage

import (
	"fmt"

	"github.com/manyminds/api2go/examples/model"
)

// NewUserStorage initializes the storage
func NewUserStorage() *UserStorage {
	return &UserStorage{make(map[string]*model.User), 1}
}

// UserStorage stores all users
type UserStorage struct {
	users   map[string]*model.User
	idCount int
}

// GetAll returns the user map (because we need the ID as key too)
func (s UserStorage) GetAll() map[string]*model.User {
	return s.users
}

// GetOne user
func (s UserStorage) GetOne(id string) (model.User, error) {
	user, ok := s.users[id]
	if ok {
		return *user, nil
	}

	return model.User{}, fmt.Errorf("User for id %s not found", id)
}

// Insert a user
func (s *UserStorage) Insert(c model.User) string {
	id := fmt.Sprintf("%d", s.idCount)
	c.ID = id
	s.users[id] = &c
	s.idCount++
	return id
}

// Delete one :(
func (s *UserStorage) Delete(id string) error {
	_, exists := s.users[id]
	if !exists {
		return fmt.Errorf("User with id %s does not exist", id)
	}
	delete(s.users, id)

	return nil
}

// Update a user
func (s *UserStorage) Update(c model.User) error {
	_, exists := s.users[c.ID]
	if !exists {
		return fmt.Errorf("User with id %s does not exist", c.ID)
	}
	s.users[c.ID] = &c

	return nil
}
