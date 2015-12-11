package jsonapi

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Book struct {
	ID       string      `json:"-"`
	Author   *StupidUser `json:"-"`
	AuthorID string      `json:"-"`
	Pages    []Page      `json:"-"`
	PagesIDs []string    `json:"-"`
}

func (b Book) GetID() string {
	return b.ID
}

func (b *Book) SetID(ID string) error {
	b.ID = ID

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
		result = append(result, ReferenceID{ID: b.Author.GetID(), Name: "author", Type: "stupidUsers"})
	}
	for _, page := range b.Pages {
		result = append(result, ReferenceID{ID: page.GetID(), Name: "pages", Type: "pages"})
	}

	return result
}

func (b *Book) SetToOneReferenceID(name, ID string) error {
	if name == "author" {
		b.AuthorID = ID

		return nil
	}

	return errors.New("There is no to-one relationship with name " + name)
}

func (b *Book) SetToManyReferenceIDs(name string, IDs []string) error {
	if name == "pages" {
		b.PagesIDs = IDs

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

func (s StupidUser) GetID() string {
	return s.ID
}

type Page struct {
	ID      string `json:"-"`
	Content string `json:"content"`
}

func (p Page) GetID() string {
	return p.ID
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

	testResult := `
		{ "data" : 
			{ 
				"id" : "TheOneAndOnlyID",
				"attributes": {},
				"relationships" : 
				{ 
						"author" : 
						{
							"data": {
								"id" : "A Magical UserID",
								"type" : "stupidUsers"
							}
						},
						"pages" : 
						{ 
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
					"type" : "books"
			},
			"included" : 
				[ 
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
		}	
	`

	testRequest := `{
		"data":{
			"id":"TheOneAndOnlyID",
			"type":"books",
			"attributes": {},
			"relationships":{
				"author":{
					"data": {
						"id":"A Magical UserID",
						"type":"users"
					}
				},
				"pages":{
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
					]}
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
