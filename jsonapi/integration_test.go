package jsonapi

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Book struct {
	ID       string
	Author   *User
	AuthorID string
	Pages    []Page
	PagesIDs []string
}

type User struct {
	ID   string
	Name string
}

type Page struct {
	ID      string
	Content string
}

var _ = Describe("Test for the public api of this package", func() {
	author := User{
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
							"resource" : "/books/TheOneAndOnlyID/author",
							"type" : "users"
						},
						"pages" : 
						{ 
							"ids" : [ "Page 1","Page 2","Page 3"],
							"resource" : "/books/TheOneAndOnlyID/pages",
							"type" : "pages"
						}
				},
					"type" : "books"
			},
			"linked" : 
				[ 
					{ "id" : "A Magical UserID",
						"name" : "Terry Pratchett",
						"type" : "users"
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
						"ids":["Page 1","Page 2","Page 3"]
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
