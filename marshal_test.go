package api2go

import (
	"database/sql"
	"encoding/json"
	"errors"

	"gopkg.in/guregu/null.v2/zero"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Magic struct {
	ID MagicID
}

type MagicID string

func (m MagicID) String() string {
	return "This should be visible"
}

var _ = Describe("Marshalling", func() {
	type SimplePost struct {
		Title, Text string
	}

	type Comment struct {
		ID   int
		Text string
	}

	type User struct {
		ID       int
		Name     string
		Password string `json:"-"`
	}

	type Post struct {
		ID          int
		Title       string
		Comments    []Comment
		CommentsIDs []int
		Author      *User
		AuthorID    sql.NullInt64
	}

	Context("When marshaling simple objects", func() {
		var (
			firstPost, secondPost       SimplePost
			firstPostMap, secondPostMap map[string]interface{}
		)

		BeforeEach(func() {
			firstPost = SimplePost{Title: "First Post", Text: "Lipsum"}
			firstPostMap = map[string]interface{}{
				"type":  "simplePosts",
				"title": firstPost.Title,
				"text":  firstPost.Text,
			}
			secondPost = SimplePost{Title: "Second Post", Text: "Getting more advanced!"}
			secondPostMap = map[string]interface{}{
				"type":  "simplePosts",
				"title": secondPost.Title,
				"text":  secondPost.Text,
			}
		})

		It("marshals single object", func() {
			i, err := Marshal(firstPost)
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"data": firstPostMap,
			}))
		})

		It("should prefer fmt.Stringer().String() over string contents", func() {
			m := Magic{}
			m.ID = "This should be only internal"

			expected := map[string]interface{}{
				"data": map[string]interface{}{
					"type": "magics",
					"id":   "This should be visible",
				},
			}

			v, e := Marshal(m)
			Expect(e).ToNot(HaveOccurred())
			Expect(v).To(Equal(expected))
		})

		It("marshal nil value", func() {
			_, err := Marshal(nil)
			Expect(err).To(HaveOccurred())
		})

		It("marshals collections object", func() {
			i, err := Marshal([]SimplePost{firstPost, secondPost})
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"data": []interface{}{
					firstPostMap,
					secondPostMap,
				},
			}))
		})

		It("marshals empty collections", func() {
			i, err := Marshal([]SimplePost{})
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"data": []interface{}{},
			}))
		})

		It("returns an error when passing interface{} slices", func() {
			_, err := Marshal([]interface{}{})
			Expect(err).To(HaveOccurred())
		})

		It("returns an error when passing an empty string", func() {
			_, err := Marshal("")
			Expect(err).To(HaveOccurred())
		})

		It("marshals to JSON", func() {
			j, err := MarshalToJSON([]SimplePost{firstPost})
			Expect(err).To(BeNil())
			var m map[string]interface{}
			Expect(json.Unmarshal(j, &m)).To(BeNil())
			Expect(m).To(Equal(map[string]interface{}{
				"data": []interface{}{
					firstPostMap,
				},
			}))
		})

		Context("when converting IDs to string", func() {
			It("leaves string", func() {
				type StringID struct{ ID string }
				i, err := Marshal(StringID{ID: "1"})
				Expect(err).To(BeNil())
				Expect(i).To(Equal(map[string]interface{}{
					"data": map[string]interface{}{
						"type": "stringIDs",
						"id":   "1",
					},
				}))
			})

			It("converts ints", func() {
				type IntID struct{ ID int }
				i, err := Marshal(IntID{ID: 1})
				Expect(err).To(BeNil())
				Expect(i).To(Equal(map[string]interface{}{
					"data": map[string]interface{}{
						"type": "intIDs",
						"id":   "1",
					},
				}))
			})

			It("converts uints", func() {
				type UintID struct{ ID uint }
				i, err := Marshal(UintID{ID: 1})
				Expect(err).To(BeNil())
				Expect(i).To(Equal(map[string]interface{}{
					"data": map[string]interface{}{
						"type": "uintIDs",
						"id":   "1",
					},
				}))
			})
		})
	})

	Context("When marshaling compound objects", func() {
		It("marshals nested objects", func() {
			comment1 := Comment{ID: 1, Text: "First!"}
			comment2 := Comment{ID: 2, Text: "Second!"}
			author := User{ID: 1, Name: "Test Author"}
			post1 := Post{ID: 1, Title: "Foobar", Comments: []Comment{comment1, comment2}, Author: &author}
			post2 := Post{ID: 2, Title: "Foobarbarbar", Comments: []Comment{comment1, comment2}, Author: &author}

			posts := []Post{post1, post2}

			i, err := Marshal(posts)
			Expect(err).To(BeNil())

			expected := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id": "1",
						"links": map[string]interface{}{
							"comments": map[string]interface{}{
								// "self":     "/posts/1/links/comments",
								"resource": "/posts/1/comments",
								"ids":      []interface{}{"1", "2"},
								"type":     "comments",
							},
							"author": map[string]interface{}{
								// "self":     "/posts/1/links/author",
								"resource": "/posts/1/author",
								"id":       "1",
								"type":     "users",
							},
						},
						"title": "Foobar",
						"type":  "posts",
					},
					map[string]interface{}{
						"id": "2",
						"links": map[string]interface{}{
							"comments": map[string]interface{}{
								// "self":     "/posts/2/links/comments",
								"resource": "/posts/2/comments",
								"ids":      []interface{}{"1", "2"},
								"type":     "comments",
							},
							"author": map[string]interface{}{
								// "self":     "/posts/2/links/author",
								"resource": "/posts/2/author",
								"id":       "1",
								"type":     "users",
							},
						},
						"title": "Foobarbarbar",
						"type":  "posts",
					},
				},
				"linked": []interface{}{
					map[string]interface{}{
						"id":   "1",
						"text": "First!",
						"type": "comments",
					},
					map[string]interface{}{
						"id":   "2",
						"text": "Second!",
						"type": "comments",
					},
					map[string]interface{}{
						"id":   "1",
						"name": "Test Author",
						"type": "users",
					},
				},
			}

			Expect(i).To(Equal(expected))
		})

		It("adds IDs", func() {
			post := Post{ID: 1, Comments: []Comment{}, CommentsIDs: []int{1}}
			i, err := Marshal(post)
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"data": map[string]interface{}{
					"id":    "1",
					"type":  "posts",
					"title": "",
					"links": map[string]interface{}{
						"comments": map[string]interface{}{
							"ids":      []interface{}{"1"},
							"type":     "comments",
							"resource": "/posts/1/comments",
						},
						"author": map[string]interface{}{
							"type":     "users",
							"resource": "/posts/1/author",
						},
					},
				},
			}))
		})

		It("marshal correctly with prefix", func() {
			post := Post{ID: 1, Comments: []Comment{}, CommentsIDs: []int{1}}
			i, err := MarshalPrefix(post, "/v1")
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"data": map[string]interface{}{
					"id":    "1",
					"type":  "posts",
					"title": "",
					"links": map[string]interface{}{
						"comments": map[string]interface{}{
							"ids":      []interface{}{"1"},
							"type":     "comments",
							"resource": "/v1/posts/1/comments",
						},
						"author": map[string]interface{}{
							"type":     "users",
							"resource": "/v1/posts/1/author",
						},
					},
				},
			}))
		})

		It("prefers nested structs when given both, structs and IDs", func() {
			comment := Comment{ID: 1}
			author := User{ID: 1, Name: "Tester"}
			post := Post{ID: 1, Comments: []Comment{comment}, CommentsIDs: []int{2}, Author: &author, AuthorID: sql.NullInt64{Int64: 1337}}
			i, err := Marshal(post)
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"data": map[string]interface{}{
					"id":    "1",
					"type":  "posts",
					"title": "",
					"links": map[string]interface{}{
						"comments": map[string]interface{}{
							"ids":      []interface{}{"1"},
							"type":     "comments",
							"resource": "/posts/1/comments",
						},
						"author": map[string]interface{}{
							"id":       "1",
							"type":     "users",
							"resource": "/posts/1/author",
						},
					},
				},
				"linked": []interface{}{
					map[string]interface{}{
						"id":   "1",
						"type": "comments",
						"text": "",
					},
					map[string]interface{}{
						"id":   "1",
						"type": "users",
						"name": "Tester",
					},
				},
			}))
		})

		It("uses ID field if single relation struct is nil", func() {
			type AnotherPost struct {
				ID       int
				AuthorID int
				Author   *User
			}

			anotherPost := AnotherPost{ID: 1, AuthorID: 1}
			i, err := Marshal(anotherPost)
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "1",
					"type": "anotherPosts",
					"links": map[string]interface{}{
						"author": map[string]interface{}{
							"id":       "1",
							"type":     "users",
							"resource": "/anotherPosts/1/author",
						},
					},
				},
			}))
		})

		It("uses ID field for the sql.NullInt64 type", func() {
			type SqlTypesPost struct {
				ID       int
				AuthorID sql.NullInt64
				Author   *User
			}

			anotherPost := SqlTypesPost{ID: 1, AuthorID: sql.NullInt64{1, true}}
			i, err := Marshal(anotherPost)
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "1",
					"type": "sqlTypesPosts",
					"links": map[string]interface{}{
						"author": map[string]interface{}{
							"id":       "1",
							"type":     "users",
							"resource": "/sqlTypesPosts/1/author",
						},
					},
				},
			}))
		})

		It("uses ID field for the sql.NullString type", func() {
			type SqlTypesPost struct {
				ID       int
				AuthorID sql.NullString
				Author   *User
			}

			anotherPost := SqlTypesPost{ID: 1, AuthorID: sql.NullString{"1", true}}
			i, err := Marshal(anotherPost)
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "1",
					"type": "sqlTypesPosts",
					"links": map[string]interface{}{
						"author": map[string]interface{}{
							"id":       "1",
							"type":     "users",
							"resource": "/sqlTypesPosts/1/author",
						},
					},
				},
			}))
		})

		It("returns an error if ID field but no struct field is in struct", func() {
			type WrongStruct struct {
				ID       int
				AuthorID int
			}

			wrongStruct := WrongStruct{ID: 1, AuthorID: 1}
			_, err := Marshal(wrongStruct)
			Expect(err).To(Equal(errors.New("expected struct to have field Author")))
		})
	})

	Context("when marshalling zero value types", func() {
		type ZeroPost struct {
			ID    string
			Title string
			Value zero.Float
		}

		type ZeroPostPointer struct {
			ID    string
			Title string
			Value *zero.Float
		}

		theFloat := zero.NewFloat(2.3, true)
		post := ZeroPost{ID: "1", Title: "test", Value: theFloat}
		pointerPost := ZeroPostPointer{ID: "1", Title: "test", Value: &theFloat}

		It("correctly unmarshals driver values", func() {
			postMap := map[string]interface{}{
				"data": map[string]interface{}{
					"id":    "1",
					"type":  "zeroPosts",
					"title": "test",
					"value": theFloat,
				},
			}

			marshalled, err := Marshal(post)

			Expect(err).To(BeNil())
			Expect(marshalled).To(Equal(postMap))
		})

		It("correctly unmarshals into json", func() {
			expectedJSON := []byte(`{"data":{"id":"1","type":"zeroPosts","title":"test","value":2.3}}`)

			json, err := MarshalToJSON(post)
			Expect(err).To(BeNil())
			Expect(json).To(MatchJSON(expectedJSON))
		})

		It("correctly unmarshals driver values with pointer", func() {
			postMap := map[string]interface{}{
				"data": map[string]interface{}{
					"id":    "1",
					"type":  "zeroPostPointers",
					"title": "test",
					"value": &theFloat,
				},
			}

			marshalled, err := Marshal(pointerPost)

			Expect(err).To(BeNil())
			Expect(marshalled).To(BeEquivalentTo(postMap))
		})

		It("correctly unmarshals with pointer into json", func() {
			expectedJSON := []byte(`{"data":{"id":"1","type":"zeroPostPointers","title":"test","value":2.3}}`)

			json, err := MarshalToJSON(pointerPost)
			Expect(err).To(BeNil())
			Expect(json).To(MatchJSON(expectedJSON))
		})
	})

	Context("When marshalling objects linking to other instances of the same type", func() {
		type Question struct {
			ID                  string
			Text                string
			InspiringQuestionID sql.NullString
			InspiringQuestion   *Question
		}

		question1 := Question{ID: "1", Text: "Does this test work?"}
		question1Duplicate := Question{ID: "1", Text: "Does this test work?"}
		question2 := Question{ID: "2", Text: "Will it ever work?", InspiringQuestionID: sql.NullString{"1", true}, InspiringQuestion: &question1}
		question3 := Question{ID: "3", Text: "It works now", InspiringQuestionID: sql.NullString{"1", true}, InspiringQuestion: &question1Duplicate}

		It("Correctly marshalls question2 and sets question1 into linked", func() {
			expected := map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "2",
					"type": "questions",
					"text": "Will it ever work?",
					"links": map[string]interface{}{
						"inspiringQuestion": map[string]interface{}{
							"id":       "1",
							"type":     "questions",
							"resource": "/questions/2/inspiringQuestion",
						},
					},
				},
				"linked": []interface{}{
					map[string]interface{}{
						"id":   "1",
						"type": "questions",
						"text": "Does this test work?",
						"links": map[string]interface{}{
							"inspiringQuestion": map[string]interface{}{
								"type":     "questions",
								"resource": "/questions/1/inspiringQuestion",
							},
						},
					},
				},
			}

			marshalled, err := Marshal(question2)
			Expect(err).To(BeNil())
			Expect(marshalled).To(BeEquivalentTo(expected))
		})

		It("Does not marshall same dependencies multiple times", func() {
			expected := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":   "3",
						"type": "questions",
						"text": "It works now",
						"links": map[string]interface{}{
							"inspiringQuestion": map[string]interface{}{
								"id":       "1",
								"type":     "questions",
								"resource": "/questions/3/inspiringQuestion",
							},
						},
					},
					map[string]interface{}{
						"id":   "2",
						"type": "questions",
						"text": "Will it ever work?",
						"links": map[string]interface{}{
							"inspiringQuestion": map[string]interface{}{
								"id":       "1",
								"type":     "questions",
								"resource": "/questions/2/inspiringQuestion",
							},
						},
					},
				},
				"linked": []interface{}{
					map[string]interface{}{
						"id":   "1",
						"type": "questions",
						"text": "Does this test work?",
						"links": map[string]interface{}{
							"inspiringQuestion": map[string]interface{}{
								"type":     "questions",
								"resource": "/questions/1/inspiringQuestion",
							},
						},
					},
				},
			}

			marshalled, err := Marshal([]Question{question3, question2})
			Expect(err).To(BeNil())
			Expect(marshalled).To(BeEquivalentTo(expected))
		})
	})

	Context("Slice fields", func() {
		type Identity struct {
			ID     int64    `json:"user_id"`
			Scopes []string `json:"scopes"`
		}

		type Unicorn struct {
			UnicornId int64    `json:"unicorn_id"` //Annotations are ignored
			Scopes    []string `json:"scopes"`
		}

		It("Marshalls the slice field correctly", func() {
			expected := map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "1234",
					"type": "identities",
					"scopes": []string{
						"user_global",
					},
				},
			}

			marshalled, err := Marshal(Identity{1234, []string{"user_global"}})
			Expect(err).To(BeNil())
			Expect(marshalled).To(BeEquivalentTo(expected))
		})

		It("Marshalls correctly without an ID field", func() {
			expected := map[string]interface{}{
				"data": map[string]interface{}{
					"unicornId": int64(1234), // this must not be unicornID or unicorn_id, because that is the convention for a link to another struct...
					"type":      "unicorns",
					"scopes": []string{
						"user_global",
					},
				},
			}

			marshalled, err := Marshal(Unicorn{1234, []string{"user_global"}})
			Expect(err).To(BeNil())
			Expect(marshalled).To(BeEquivalentTo(expected))
		})
	})
})
