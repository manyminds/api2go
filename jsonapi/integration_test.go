package jsonapi

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Book struct {
	ID        string      `json:"-"`
	LID       string      `json:"-"`
	Author    *StupidUser `json:"-"`
	AuthorID  string      `json:"-"`
	AuthorLID string      `json:"-"`
	Pages     []Page      `json:"-"`
	PagesIDs  []string    `json:"-"`
}

func (b Book) GetID() Identifier {
	return Identifier{ID: b.ID, LID: b.LID}
}

func (b *Book) SetID(ID Identifier) error {
	b.ID = ID.ID
	b.LID = ID.LID
	return nil
}

func (b Book) GetReferences() []Reference {
	return []Reference{
		{
			Type: "stupidUsers",
			Name: "author",
		},
		{
			Type: "pages",
			Name: "pages",
		},
	}
}

func (b Book) GetReferencedIDs() []ReferenceID {
	result := []ReferenceID{}

	if b.Author != nil {
		id := b.Author.GetID()
		result = append(result, ReferenceID{
			ID:   id.ID,
			LID:  id.LID,
			Name: "author",
			Type: "stupidUsers",
		})
	}

	for _, page := range b.Pages {
		id := page.GetID()
		result = append(result, ReferenceID{
			ID:   id.ID,
			LID:  id.LID,
			Name: "pages",
			Type: "pages",
		})
	}

	return result
}

func (b *Book) SetToOneReferenceID(name string, ID *Identifier) error {
	if name == "author" {
		b.AuthorID = ID.ID
		b.AuthorLID = ID.LID

		return nil
	}

	return errors.New("There is no to-one relationship with name " + name)
}

func (b *Book) SetToManyReferenceIDs(name string, IDs []Identifier) error {
	if name == "pages" {
		b.PagesIDs = make([]string, 0, len(IDs))
		for _, id := range IDs {
			b.PagesIDs = append(b.PagesIDs, id.ID)
		}

		return nil
	}

	return errors.New("There is no to-many relationship with name " + name)
}

func (b Book) GetReferencedStructs() []MarshalIdentifier {
	result := []MarshalIdentifier{}

	if b.Author != nil {
		result = append(result, *b.Author)
	}

	for key := range b.Pages {
		result = append(result, b.Pages[key])
	}

	return result
}

type StupidUser struct {
	ID   string `json:"-"`
	Name string `json:"name"`
}

func (s StupidUser) GetID() Identifier {
	return Identifier{ID: s.ID, LID: ""}
}

type Page struct {
	ID      string `json:"-"`
	Content string `json:"content"`
}

func (p Page) GetID() Identifier {
	return Identifier{ID: p.ID, LID: ""}
}

var _ = Describe("Test for the public api of this package", func() {
	author := StupidUser{
		ID:   "A Magical UserID",
		Name: "Terry Pratchett",
	}

	pages := []Page{
		{ID: "Page 1", Content: "First Page"},
		{ID: "Page 2", Content: "Second Page"},
		{ID: "Page 3", Content: "Final page"},
	}

	testBook := Book{
		ID:     "TheOneAndOnlyID",
		Author: &author,
		Pages:  pages,
	}

	testResult := `{
		"data": {
			"id": "TheOneAndOnlyID",
			"attributes": {},
			"relationships": {
				"author": {
					"data": {
						"id" : "A Magical UserID",
						"type" : "stupidUsers"
					}
				},
				"pages": {
					"data": [
						{
							"id": "Page 1",
							"type": "pages"
						},
						{
							"id": "Page 2",
							"type": "pages"
						},
						{
							"id": "Page 3",
							"type": "pages"
						}
					]
				}
			},
			"type": "books"
		},
		"included": [
			{
				"id" : "A Magical UserID",
				"attributes": {
					"name" : "Terry Pratchett"
				},
				"type" : "stupidUsers"
			},
			{
				"attributes": {
					"content" : "First Page"
				},
				"id" : "Page 1",
				"type" : "pages"
			},
			{
				"attributes": {
					"content" : "Second Page"
				},
				"id" : "Page 2",
				"type" : "pages"
			},
			{
				"attributes": {
					"content" : "Final page"
				},
				"id" : "Page 3",
				"type" : "pages"
			}
		]
	}`

	testRequest := `{
		"data": {
			"id": "TheOneAndOnlyID",
			"type": "books",
			"attributes": {},
			"relationships": {
				"author": {
					"data": {
						"id":"A Magical UserID",
						"type":"users"
					}
				},
				"pages": {
					"data": [
						{
							"id": "Page 1",
							"type": "pages"
						},
						{
							"id": "Page 2",
							"type": "pages"
						},
						{
							"id": "Page 3",
							"type": "pages"
						}
					]
				}
			}
		}
	}`

	Context("Marshal and Unmarshal data", func() {
		It("Should be marshalled correctly", func() {
			marshalResult, err := Marshal(testBook)
			Expect(err).ToNot(HaveOccurred())
			Expect(marshalResult).To(MatchJSON(testResult))
		})

		It("Should be unmarshalled correctly", func() {
			result := &Book{}
			expected := Book{
				ID:       "TheOneAndOnlyID",
				AuthorID: "A Magical UserID",
				PagesIDs: []string{"Page 1", "Page 2", "Page 3"},
			}
			err := Unmarshal([]byte(testRequest), result)
			Expect(err).ToNot(HaveOccurred())
			Expect(*result).To(Equal(expected))
		})
	})
})
