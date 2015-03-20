package jsonapi

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Book struct {
	ID       string
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

func (b *Book) SetReferencedIDs(IDs []ReferenceID) error {
	for _, reference := range IDs {
		switch reference.Name {
		case "author":
			b.AuthorID = reference.ID
		case "pages":
			b.PagesIDs = append(b.PagesIDs, reference.ID)
		}
	}

	return nil
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
	ID   string
	Name string
}

func (s StupidUser) GetID() string {
	return s.ID
}

type Page struct {
	ID      string
	Content string
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
		Page{ID: "Page 1", Content: "First Page"},
		Page{ID: "Page 2", Content: "Second Page"},
		Page{ID: "Page 3", Content: "Final page"},
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
				"links" : 
				{ 
						"author" : 
						{
							"id" : "A Magical UserID",
							"type" : "stupidUsers"
						},
						"pages" : 
						{ 
							"ids" : [ "Page 1","Page 2","Page 3"],
							"type" : "pages"
						}
				},
					"type" : "books"
			},
			"linked" : 
				[ 
					{ "id" : "A Magical UserID",
						"name" : "Terry Pratchett",
						"type" : "stupidUsers"
					},
					{ "content" : "First Page",
						"id" : "Page 1",
						"type" : "pages"
					},
					{ "content" : "Second Page",
						"id" : "Page 2",
						"type" : "pages"
					},
					{ "content" : "Final page",
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
			"links":{
				"author":{
					"id":"A Magical UserID",
					"type":"users"
				},
				"pages":{
						"ids":["Page 1","Page 2","Page 3"],
						"type": "pages"
					}
				}
			},
			"linked":
				[
					{"content":"First Page","id":"Page 1","type":"pages"},
					{"content":"Second Page","id":"Page 2","type":"pages"},
					{"content":"Final page","id":"Page 3","type":"pages"},
					{"id":"A Magical UserID","name":"Terry Pratchett","type":"users"}
				]
		}`

	Context("Marshal and Unmarshal data", func() {
		It("Should be marshalled correctly", func() {
			marshalResult, err := MarshalToJSON(testBook)
			Expect(err).ToNot(HaveOccurred())
			Expect(marshalResult).To(MatchJSON(testResult))
		})

		It("Should be unmarshalled correctly", func() {
			result := &[]Book{}
			expected := []Book{
				Book{
					ID:       "TheOneAndOnlyID",
					AuthorID: "A Magical UserID",
					PagesIDs: []string{"Page 1", "Page 2", "Page 3"},
				},
			}
			err := UnmarshalFromJSON([]byte(testRequest), result)
			Expect(err).ToNot(HaveOccurred())
			Expect(*result).To(Equal(expected))
		})
	})
})
