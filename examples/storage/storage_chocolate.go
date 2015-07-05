package storage

import (
	"fmt"
	"sort"

	"github.com/manyminds/api2go/examples/model"
)

// sorting
type byID []model.Chocolate

func (c byID) Len() int {
	return len(c)
}

func (c byID) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c byID) Less(i, j int) bool {
	return c[i].GetID() < c[j].GetID()
}

// NewChocolateStorage initializes the storage
func NewChocolateStorage() *ChocolateStorage {
	return &ChocolateStorage{make(map[string]*model.Chocolate), 1}
}

// ChocolateStorage stores all of the tasty chocolate, needs to be injected into
// User and Chocolate Resource. In the real world, you would use a database for that.
type ChocolateStorage struct {
	chocolates map[string]*model.Chocolate
	idCount    int
}

// GetAll of the chocolate
func (s ChocolateStorage) GetAll() []model.Chocolate {
	result := []model.Chocolate{}
	for key := range s.chocolates {
		result = append(result, *s.chocolates[key])
	}

	sort.Sort(byID(result))
	return result
}

// GetOne tasty chocolate
func (s ChocolateStorage) GetOne(id string) (model.Chocolate, error) {
	choc, ok := s.chocolates[id]
	if ok {
		return *choc, nil
	}

	return model.Chocolate{}, fmt.Errorf("Chocolate for id %s not found", id)
}

// Insert a fresh one
func (s *ChocolateStorage) Insert(c model.Chocolate) string {
	id := fmt.Sprintf("%d", s.idCount)
	c.ID = id
	s.chocolates[id] = &c
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
func (s *ChocolateStorage) Update(c model.Chocolate) error {
	_, exists := s.chocolates[c.ID]
	if !exists {
		return fmt.Errorf("Chocolate with id %s does not exist", c.ID)
	}
	s.chocolates[c.ID] = &c

	return nil
}
