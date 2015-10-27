package jsonapi

import (
	"database/sql"
	"time"

	"gopkg.in/guregu/null.v2/zero"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Marshalling", func() {
	Context("When marshaling simple objects", func() {
		var (
			firstPost, secondPost                     SimplePost
			firstUserMap, firstPostMap, secondPostMap map[string]interface{}
			created                                   time.Time
		)

		BeforeEach(func() {
			created, _ = time.Parse(time.RFC3339, "2014-11-10T16:30:48.823Z")
			firstPost = SimplePost{ID: "first", Title: "First Post", Text: "Lipsum", Created: created}
			firstPostMap = map[string]interface{}{
				"type": "simplePosts",
				"id":   "first",
				"attributes": map[string]interface{}{
					"title":       firstPost.Title,
					"text":        firstPost.Text,
					"size":        0,
					"create-date": created,
				},
			}

			secondPost = SimplePost{ID: "second", Title: "Second Post", Text: "Getting more advanced!", Created: created, Updated: created}
			secondPostMap = map[string]interface{}{
				"type": "simplePosts",
				"id":   "second",
				"attributes": map[string]interface{}{
					"title":        secondPost.Title,
					"text":         secondPost.Text,
					"size":         0,
					"create-date":  created,
					"updated-date": created,
				},
			}

			firstUserMap = map[string]interface{}{
				"type": "users",
				"id":   "100",
				"attributes": map[string]interface{}{
					"name": "Nino",
				},
			}
		})

		It("marshals single object without relationships", func() {
			user := User{ID: 100, Name: "Nino", Password: "babymaus"}
			i, err := Marshal(user)
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"data": firstUserMap,
			}))
		})

		It("marshals single object without relationships as pointer", func() {
			user := User{ID: 100, Name: "Nino", Password: "babymaus"}
			i, err := Marshal(&user)
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"data": firstUserMap,
			}))
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
					"type":       "magics",
					"id":         "This should be visible",
					"attributes": map[string]interface{}{},
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
				"data": []map[string]interface{}{
					firstPostMap,
					secondPostMap,
				},
			}))
		})

		It("marshals empty collections", func() {
			i, err := Marshal([]SimplePost{})
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"data": []map[string]interface{}{},
			}))
		})

		It("marshals slices of interface with one struct", func() {
			i, err := Marshal([]interface{}{firstPost})
			Expect(err).ToNot(HaveOccurred())
			Expect(i).To(Equal(map[string]interface{}{
				"data": []map[string]interface{}{
					firstPostMap,
				},
			}))
		})

		It("marshals slices of interface with structs", func() {
			i, err := Marshal([]interface{}{firstPost, secondPost, User{ID: 1337, Name: "Nino", Password: "God"}})
			Expect(err).ToNot(HaveOccurred())
			Expect(i).To(Equal(map[string]interface{}{
				"data": []map[string]interface{}{
					firstPostMap,
					secondPostMap,
					{
						"id":   "1337",
						"type": "users",
						"attributes": map[string]interface{}{
							"name": "Nino",
						},
					},
				},
			}))
		})

		It("returns an error when passing an empty string", func() {
			_, err := Marshal("")
			Expect(err).To(HaveOccurred())
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

			i, err := MarshalWithURLs(posts, CompleteServerInformation{})
			Expect(err).To(BeNil())

			expected := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id": "1",
						"relationships": map[string]map[string]interface{}{
							"comments": {
								"links": map[string]string{
									"self":    completePrefix + "/posts/1/relationships/comments",
									"related": completePrefix + "/posts/1/comments",
								},
								"data": []map[string]interface{}{
									{
										"id":   "1",
										"type": "comments",
									},
									{
										"id":   "2",
										"type": "comments",
									},
								},
							},
							"author": {
								"links": map[string]string{
									"self":    completePrefix + "/posts/1/relationships/author",
									"related": completePrefix + "/posts/1/author",
								},
								"data": map[string]interface{}{
									"id":   "1",
									"type": "users",
								},
							},
						},
						"type": "posts",
						"attributes": map[string]interface{}{
							"title": "Foobar",
						},
					},
					{
						"id": "2",
						"relationships": map[string]map[string]interface{}{
							"comments": {
								"links": map[string]string{
									"self":    completePrefix + "/posts/2/relationships/comments",
									"related": completePrefix + "/posts/2/comments",
								},
								"data": []map[string]interface{}{
									{
										"id":   "1",
										"type": "comments",
									},
									{
										"id":   "2",
										"type": "comments",
									},
								},
							},
							"author": {
								"links": map[string]string{
									"self":    completePrefix + "/posts/2/relationships/author",
									"related": completePrefix + "/posts/2/author",
								},
								"data": map[string]interface{}{
									"id":   "1",
									"type": "users",
								},
							},
						},
						"type": "posts",
						"attributes": map[string]interface{}{
							"title": "Foobarbarbar",
						},
					},
				},
				"included": []map[string]interface{}{
					{
						"id":   "1",
						"type": "users",
						"attributes": map[string]interface{}{
							"name": "Test Author",
						},
					},
					{
						"id":   "1",
						"type": "comments",
						"attributes": map[string]interface{}{
							"text": "First!",
						},
					},
					{
						"id":   "2",
						"type": "comments",
						"attributes": map[string]interface{}{
							"text": "Second!",
						},
					},
				},
			}

			Expect(i).To(Equal(expected))
		})

		It("adds IDs", func() {
			post := Post{ID: 1, Comments: []Comment{}, CommentsIDs: []int{1}}
			i, err := MarshalWithURLs(post, CompleteServerInformation{})
			expected := map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "1",
					"type": "posts",
					"attributes": map[string]interface{}{
						"title": "",
					},
					"relationships": map[string]map[string]interface{}{
						"comments": {
							"links": map[string]string{
								"self":    completePrefix + "/posts/1/relationships/comments",
								"related": completePrefix + "/posts/1/comments",
							},
							"data": []map[string]interface{}{
								{
									"id":   "1",
									"type": "comments",
								},
							},
						},
						"author": {
							"data": nil,
							"links": map[string]string{
								"self":    completePrefix + "/posts/1/relationships/author",
								"related": completePrefix + "/posts/1/author",
							},
						},
					},
				},
			}
			Expect(err).To(BeNil())
			Expect(i).To(Equal(expected))
		})

		It("prefers nested structs when given both, structs and IDs", func() {
			comment := Comment{ID: 1}
			author := User{ID: 1, Name: "Tester"}
			post := Post{ID: 1, Comments: []Comment{comment}, CommentsIDs: []int{2}, Author: &author, AuthorID: sql.NullInt64{Int64: 1337}}
			i, err := Marshal(post)
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "1",
					"type": "posts",
					"attributes": map[string]interface{}{
						"title": "",
					},
					"relationships": map[string]map[string]interface{}{
						"comments": {
							"data": []map[string]interface{}{
								{
									"id":   "1",
									"type": "comments",
								},
							},
						},
						"author": {
							"data": map[string]interface{}{
								"id":   "1",
								"type": "users",
							},
						},
					},
				},
				"included": []map[string]interface{}{
					{
						"id":   "1",
						"type": "users",
						"attributes": map[string]interface{}{
							"name": "Tester",
						},
					},
					{
						"id":   "1",
						"type": "comments",
						"attributes": map[string]interface{}{
							"text": "",
						},
					},
				},
			}))
		})

		It("uses ID field if MarshalLinkedRelations is implemented", func() {
			anotherPost := AnotherPost{ID: 1, AuthorID: 1}
			i, err := Marshal(anotherPost)
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"data": map[string]interface{}{
					"id":         "1",
					"type":       "anotherPosts",
					"attributes": map[string]interface{}{},
					"relationships": map[string]map[string]interface{}{
						"author": {
							"data": map[string]interface{}{
								"id":   "1",
								"type": "users",
							},
						},
					},
				},
			}))
		})
	})

	Context("when marshalling with relations that were not loaded", func() {
		It("skips data field for not loaded relations", func() {
			post := Post{ID: 123, Title: "Test", CommentsEmpty: true, AuthorEmpty: true}

			// this only makes sense with MarshalWithURLs. Otherwise, the jsonapi spec would be
			// violated, because you at least need a data, links, or meta field
			i, err := MarshalWithURLs(post, CompleteServerInformation{})
			Expect(err).ToNot(HaveOccurred())
			Expect(i).To(Equal(map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "123",
					"type": "posts",
					"attributes": map[string]interface{}{
						"title": "Test",
					},
					"relationships": map[string]map[string]interface{}{
						"author": {
							"links": map[string]string{
								"self":    "http://my.domain/v1/posts/123/relationships/author",
								"related": "http://my.domain/v1/posts/123/author",
							},
						},
						"comments": {
							"links": map[string]string{
								"self":    "http://my.domain/v1/posts/123/relationships/comments",
								"related": "http://my.domain/v1/posts/123/comments",
							},
						},
					},
				},
			}))
		})
		It("skips data field for not loaded author relation", func() {
			post := Post{ID: 123, Title: "Test", AuthorEmpty: true}

			// this only makes sense with MarshalWithURLs. Otherwise, the jsonapi spec would be
			// violated, because you at least need a data, links, or meta field
			i, err := MarshalWithURLs(post, CompleteServerInformation{})
			Expect(err).ToNot(HaveOccurred())
			Expect(i).To(Equal(map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "123",
					"type": "posts",
					"attributes": map[string]interface{}{
						"title": "Test",
					},
					"relationships": map[string]map[string]interface{}{
						"author": {
							"links": map[string]string{
								"self":    "http://my.domain/v1/posts/123/relationships/author",
								"related": "http://my.domain/v1/posts/123/author",
							},
						},
						"comments": {
							"data": []interface{}{},
							"links": map[string]string{
								"self":    "http://my.domain/v1/posts/123/relationships/comments",
								"related": "http://my.domain/v1/posts/123/comments",
							},
						},
					},
				},
			}))
		})
		It("skips data field for not loaded comments", func() {
			post := Post{ID: 123, Title: "Test", CommentsEmpty: true}

			// this only makes sense with MarshalWithURLs. Otherwise, the jsonapi spec would be
			// violated, because you at least need a data, links, or meta field
			i, err := MarshalWithURLs(post, CompleteServerInformation{})
			Expect(err).ToNot(HaveOccurred())
			Expect(i).To(Equal(map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "123",
					"type": "posts",
					"attributes": map[string]interface{}{
						"title": "Test",
					},
					"relationships": map[string]map[string]interface{}{
						"author": {
							"data": nil,
							"links": map[string]string{
								"self":    "http://my.domain/v1/posts/123/relationships/author",
								"related": "http://my.domain/v1/posts/123/author",
							},
						},
						"comments": {
							"links": map[string]string{
								"self":    "http://my.domain/v1/posts/123/relationships/comments",
								"related": "http://my.domain/v1/posts/123/comments",
							},
						},
					},
				},
			}))
		})
	})

	Context("when marshalling zero value types", func() {
		theFloat := zero.NewFloat(2.3, true)
		post := ZeroPost{ID: "1", Title: "test", Value: theFloat}
		pointerPost := ZeroPostPointer{ID: "1", Title: "test", Value: &theFloat}

		It("correctly unmarshals driver values", func() {
			postMap := map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "1",
					"type": "zeroPosts",
					"attributes": map[string]interface{}{
						"title": "test",
						"value": theFloat,
					},
				},
			}

			marshalled, err := Marshal(post)

			Expect(err).To(BeNil())
			Expect(marshalled).To(Equal(postMap))
		})

		It("correctly unmarshals into json", func() {
			expectedJSON := []byte(`
				{"data":{
					"id":"1",
					"type":"zeroPosts",
					"attributes": {
						"title":"test",
						"value":2.3
					}
				}}
			`)

			json, err := MarshalToJSON(post)
			Expect(err).To(BeNil())
			Expect(json).To(MatchJSON(expectedJSON))
		})

		It("correctly unmarshals driver values with pointer", func() {
			postMap := map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "1",
					"type": "zeroPostPointers",
					"attributes": map[string]interface{}{
						"title": "test",
						"value": &theFloat,
					},
				},
			}

			marshalled, err := Marshal(pointerPost)

			Expect(err).To(BeNil())
			Expect(marshalled).To(BeEquivalentTo(postMap))
		})

		It("correctly unmarshals with pointer into json", func() {
			expectedJSON := []byte(`
				{"data":{
					"id":"1",
					"type":"zeroPostPointers",
					"attributes": {
						"title":"test",
						"value":2.3
					}
				}}
			`)

			json, err := MarshalToJSON(pointerPost)
			Expect(err).To(BeNil())
			Expect(json).To(MatchJSON(expectedJSON))
		})
	})

	Context("When marshalling objects linking to other instances of the same type", func() {
		question1 := Question{ID: "1", Text: "Does this test work?"}
		question1Duplicate := Question{ID: "1", Text: "Does this test work?"}
		question2 := Question{ID: "2", Text: "Will it ever work?", InspiringQuestionID: sql.NullString{String: "1", Valid: true}, InspiringQuestion: &question1}
		question3 := Question{ID: "3", Text: "It works now", InspiringQuestionID: sql.NullString{String: "1", Valid: true}, InspiringQuestion: &question1Duplicate}

		It("Correctly marshalls question2 and sets question1 into included", func() {
			expected := map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "2",
					"type": "questions",
					"attributes": map[string]interface{}{
						"text": "Will it ever work?",
					},
					"relationships": map[string]map[string]interface{}{
						"inspiringQuestion": {
							"data": map[string]interface{}{
								"id":   "1",
								"type": "questions",
							},
						},
					},
				},
				"included": []map[string]interface{}{
					{
						"id":   "1",
						"type": "questions",
						"attributes": map[string]interface{}{
							"text": "Does this test work?",
						},
						"relationships": map[string]map[string]interface{}{
							"inspiringQuestion": {
								"data": nil,
							},
						},
					},
				},
			}

			marshalled, err := Marshal(question2)
			Expect(err).To(BeNil())
			Expect(marshalled).To(Equal(expected))
		})

		It("Does not marshall same dependencies multiple times", func() {
			expected := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":   "3",
						"type": "questions",
						"attributes": map[string]interface{}{
							"text": "It works now",
						},
						"relationships": map[string]map[string]interface{}{
							"inspiringQuestion": {
								"data": map[string]interface{}{
									"id":   "1",
									"type": "questions",
								},
							},
						},
					},
					{
						"id":   "2",
						"type": "questions",
						"attributes": map[string]interface{}{
							"text": "Will it ever work?",
						},
						"relationships": map[string]map[string]interface{}{
							"inspiringQuestion": {
								"data": map[string]interface{}{
									"id":   "1",
									"type": "questions",
								},
							},
						},
					},
				},
				"included": []map[string]interface{}{
					{
						"id":   "1",
						"type": "questions",
						"attributes": map[string]interface{}{
							"text": "Does this test work?",
						},
						"relationships": map[string]map[string]interface{}{
							"inspiringQuestion": {
								"data": nil,
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
		It("Marshalls the slice field correctly", func() {
			expected := map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "1234",
					"type": "identities",
					"attributes": map[string]interface{}{
						"scopes": []string{
							"user_global",
						},
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
					"id":   "magicalUnicorn",
					"type": "unicorns",
					"attributes": map[string]interface{}{
						"scopes": []string{
							"user_global",
						},
						"unicornID": int64(1234),
					},
				},
			}

			marshalled, err := Marshal(Unicorn{1234, []string{"user_global"}})
			Expect(err).To(BeNil())
			Expect(marshalled).To(Equal(expected))
		})
	})

	Context("Test getStructTypes method", func() {
		comment := Comment{ID: 100, Text: "some text"}
		It("should work with normal value", func() {
			result := getStructType(comment)
			Expect(result).To(Equal("comments"))
		})

		It("should work with pointer to value", func() {
			result := getStructType(&comment)
			Expect(result).To(Equal("comments"))
		})

		It("checks for EntityNamer interface", func() {
			result := getStructType(RenamedComment{"something"})
			Expect(result).To(Equal("renamed-comments"))
		})
	})

	Context("test getStructFields method", func() {
		comment := Comment{ID: 100, Text: "some text"}
		expected := map[string]interface{}{"text": "some text"}
		It("should work with normal value", func() {
			result := getStructFields(comment)
			Expect(result).To(Equal(expected))
		})

		It("should work with pointer to value", func() {
			result := getStructFields(&comment)
			Expect(result).To(Equal(expected))
		})
	})

	Context("test getStructLinks", func() {
		var (
			post    Post
			comment Comment
			author  User
		)

		BeforeEach(func() {
			comment = Comment{ID: 1}
			author = User{ID: 1, Name: "Tester"}
			post = Post{ID: 1, Comments: []Comment{comment}, Author: &author}
		})

		It("Generates to-one relationships correctly", func() {
			links := getStructRelationships(post, serverInformationNil)
			Expect(links["author"]).To(Equal(map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "1",
					"type": "users",
				},
			}))
		})

		It("Generates to-many relationships correctly", func() {
			links := getStructRelationships(post, serverInformationNil)
			Expect(links["comments"]).To(Equal(map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":   "1",
						"type": "comments",
					},
				},
			}))
		})

		It("Generates self/related URLs with baseURL and prefix correctly", func() {
			links := getStructRelationships(post, CompleteServerInformation{})
			Expect(links["author"]).To(Equal(map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "1",
					"type": "users",
				},
				"links": map[string]string{
					"self":    "http://my.domain/v1/posts/1/relationships/author",
					"related": "http://my.domain/v1/posts/1/author",
				},
			}))
		})

		It("Generates self/related URLs with baseURL correctly", func() {
			links := getStructRelationships(post, BaseURLServerInformation{})
			Expect(links["author"]).To(Equal(map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "1",
					"type": "users",
				},
				"links": map[string]string{
					"self":    "http://my.domain/posts/1/relationships/author",
					"related": "http://my.domain/posts/1/author",
				},
			}))
		})

		It("Generates self/related URLs with prefix correctly", func() {
			links := getStructRelationships(post, PrefixServerInformation{})
			Expect(links["author"]).To(Equal(map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "1",
					"type": "users",
				},
				"links": map[string]string{
					"self":    "/v1/posts/1/relationships/author",
					"related": "/v1/posts/1/author",
				},
			}))
		})
	})

	Context("test reduceDuplicates", func() {
		input := []MarshalIdentifier{
			User{ID: 314, Name: "User314"},
			Comment{ID: 314},
			Comment{ID: 1},
			User{ID: 1, Name: "User1"},
			User{ID: 2, Name: "User2"},
			Comment{ID: 1},
			Comment{ID: 314},
			User{ID: 2, Name: "User2Kopie"},
		}

		expected := []map[string]interface{}{
			{"id": 314, "name": "User314", "type": "users"},
			{"text": "", "id": 314, "type": "comments"},
			{"text": "", "id": 1, "type": "comments"},
			{"name": "User1", "id": 1, "type": "users"},
			{"name": "User2", "id": 2, "type": "users"},
		}

		dummyFunc := func(m MarshalIdentifier, i ServerInformation) (map[string]interface{}, error) {
			return map[string]interface{}{"blub": m}, nil
		}

		It("should work with default marshalData", func() {
			actual, err := reduceDuplicates(input, serverInformationNil, marshalData)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(actual)).To(Equal(len(expected)))
		})

		It("should work with dummy marshalData", func() {
			actual, err := reduceDuplicates(input, serverInformationNil, dummyFunc)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(actual)).To(Equal(len(expected)))
		})
	})

	// In order to use the SQL Null-Types the Marshal/Unmarshal interfaces for these types must be implemented.
	// The library "gopkg.in/guregu/null.v2/zero" can be used for that.
	Context("SQL Null-Types", func() {
		var nullPost SQLNullPost

		It("correctly marshalls String, Int64, Float64, Bool and Time", func() {
			nullPost = SQLNullPost{
				ID:     "theID",
				Title:  zero.StringFrom("Test"),
				Likes:  zero.IntFrom(666),
				Rating: zero.FloatFrom(66.66),
				IsCool: zero.BoolFrom(true),
			}
			result, err := MarshalToJSON(nullPost)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(MatchJSON(`
				{
					"data": {
						"id": "theID",
						"type": "sqlNullPosts",
						"attributes": {
							"title": "Test",
							"likes": 666,
							"rating": 66.66,
							"isCool": true,
							"today": "0001-01-01T00:00:00Z"
						}
					}
				}
			`))
		})

		It("correctly marshalls Null String, Int64, Float64, Bool and Time", func() {
			nullPost = SQLNullPost{
				ID:     "theID",
				Title:  zero.StringFromPtr(nil),
				Likes:  zero.IntFromPtr(nil),
				Rating: zero.FloatFromPtr(nil),
				IsCool: zero.BoolFromPtr(nil),
				Today:  zero.TimeFromPtr(nil),
			}
			result, err := MarshalToJSON(nullPost)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(MatchJSON(`
				{
					"data": {
						"id": "theID",
						"type": "sqlNullPosts",
						"attributes": {
							"title": "",
							"likes": 0,
							"rating": 0,
							"isCool": false,
							"today": "0001-01-01T00:00:00Z"
						}
					}
				}
			`))
		})

	})
})
