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
				"type":        "simplePosts",
				"id":          "first",
				"title":       firstPost.Title,
				"text":        firstPost.Text,
				"size":        0,
				"create-date": created,
			}

			secondPost = SimplePost{ID: "second", Title: "Second Post", Text: "Getting more advanced!", Created: created}
			secondPostMap = map[string]interface{}{
				"type":        "simplePosts",
				"id":          "second",
				"title":       secondPost.Title,
				"text":        secondPost.Text,
				"size":        0,
				"create-date": created,
			}

			firstUserMap = map[string]interface{}{
				"type": "users",
				"id":   "100",
				"name": "Nino",
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
					map[string]interface{}{
						"id":   "1337",
						"name": "Nino",
						"type": "users",
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
					map[string]interface{}{
						"id": "1",
						"links": map[string]map[string]interface{}{
							"comments": map[string]interface{}{
								"self":    completePrefix + "/posts/1/links/comments",
								"related": completePrefix + "/posts/1/comments",
								"linkage": []map[string]interface{}{
									map[string]interface{}{
										"id":   "1",
										"type": "comments",
									},
									map[string]interface{}{
										"id":   "2",
										"type": "comments",
									},
								},
							},
							"author": map[string]interface{}{
								"self":    completePrefix + "/posts/1/links/author",
								"related": completePrefix + "/posts/1/author",
								"linkage": map[string]interface{}{
									"id":   "1",
									"type": "users",
								},
							},
						},
						"title": "Foobar",
						"type":  "posts",
					},
					map[string]interface{}{
						"id": "2",
						"links": map[string]map[string]interface{}{
							"comments": map[string]interface{}{
								"self":    completePrefix + "/posts/2/links/comments",
								"related": completePrefix + "/posts/2/comments",
								"linkage": []map[string]interface{}{
									map[string]interface{}{
										"id":   "1",
										"type": "comments",
									},
									map[string]interface{}{
										"id":   "2",
										"type": "comments",
									},
								},
							},
							"author": map[string]interface{}{
								"self":    completePrefix + "/posts/2/links/author",
								"related": completePrefix + "/posts/2/author",
								"linkage": map[string]interface{}{
									"id":   "1",
									"type": "users",
								},
							},
						},
						"title": "Foobarbarbar",
						"type":  "posts",
					},
				},
				"included": []map[string]interface{}{
					map[string]interface{}{
						"id":   "1",
						"name": "Test Author",
						"type": "users",
					},
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
				},
			}

			Expect(i).To(Equal(expected))
		})

		It("adds IDs", func() {
			post := Post{ID: 1, Comments: []Comment{}, CommentsIDs: []int{1}}
			i, err := MarshalWithURLs(post, CompleteServerInformation{})
			expected := map[string]interface{}{
				"data": map[string]interface{}{
					"id":    "1",
					"type":  "posts",
					"title": "",
					"links": map[string]map[string]interface{}{
						"comments": map[string]interface{}{
							"self":    completePrefix + "/posts/1/links/comments",
							"related": completePrefix + "/posts/1/comments",
							"linkage": []map[string]interface{}{
								map[string]interface{}{
									"id":   "1",
									"type": "comments",
								},
							},
						},
						"author": map[string]interface{}{
							"linkage": nil,
							"self":    completePrefix + "/posts/1/links/author",
							"related": completePrefix + "/posts/1/author",
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
					"id":    "1",
					"type":  "posts",
					"title": "",
					"links": map[string]map[string]interface{}{
						"comments": map[string]interface{}{
							"linkage": []map[string]interface{}{
								map[string]interface{}{
									"id":   "1",
									"type": "comments",
								},
							},
						},
						"author": map[string]interface{}{
							"linkage": map[string]interface{}{
								"id":   "1",
								"type": "users",
							},
						},
					},
				},
				"included": []map[string]interface{}{
					map[string]interface{}{
						"id":   "1",
						"type": "users",
						"name": "Tester",
					},
					map[string]interface{}{
						"id":   "1",
						"type": "comments",
						"text": "",
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
					"id":   "1",
					"type": "anotherPosts",
					"links": map[string]map[string]interface{}{
						"author": map[string]interface{}{
							"linkage": map[string]interface{}{
								"id":   "1",
								"type": "users",
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
		question1 := Question{ID: "1", Text: "Does this test work?"}
		question1Duplicate := Question{ID: "1", Text: "Does this test work?"}
		question2 := Question{ID: "2", Text: "Will it ever work?", InspiringQuestionID: sql.NullString{String: "1", Valid: true}, InspiringQuestion: &question1}
		question3 := Question{ID: "3", Text: "It works now", InspiringQuestionID: sql.NullString{String: "1", Valid: true}, InspiringQuestion: &question1Duplicate}

		It("Correctly marshalls question2 and sets question1 into included", func() {
			expected := map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "2",
					"type": "questions",
					"text": "Will it ever work?",
					"links": map[string]map[string]interface{}{
						"inspiringQuestion": map[string]interface{}{
							"linkage": map[string]interface{}{
								"id":   "1",
								"type": "questions",
							},
						},
					},
				},
				"included": []map[string]interface{}{
					map[string]interface{}{
						"id":   "1",
						"type": "questions",
						"text": "Does this test work?",
						"links": map[string]map[string]interface{}{
							"inspiringQuestion": map[string]interface{}{
								"linkage": nil,
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
					map[string]interface{}{
						"id":   "3",
						"type": "questions",
						"text": "It works now",
						"links": map[string]map[string]interface{}{
							"inspiringQuestion": map[string]interface{}{
								"linkage": map[string]interface{}{
									"id":   "1",
									"type": "questions",
								},
							},
						},
					},
					map[string]interface{}{
						"id":   "2",
						"type": "questions",
						"text": "Will it ever work?",
						"links": map[string]map[string]interface{}{
							"inspiringQuestion": map[string]interface{}{
								"linkage": map[string]interface{}{
									"id":   "1",
									"type": "questions",
								},
							},
						},
					},
				},
				"included": []map[string]interface{}{
					map[string]interface{}{
						"id":   "1",
						"type": "questions",
						"text": "Does this test work?",
						"links": map[string]map[string]interface{}{
							"inspiringQuestion": map[string]interface{}{
								"linkage": nil,
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
					"id":        "magicalUnicorn",
					"unicornID": int64(1234),
					"type":      "unicorns",
					"scopes": []string{
						"user_global",
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

		It("Generates to-one links correctly", func() {
			links := getStructLinks(post, serverInformationNil)
			Expect(links["author"]).To(Equal(map[string]interface{}{
				"linkage": map[string]interface{}{
					"id":   "1",
					"type": "users",
				},
			}))
		})

		It("Generates to-many links correctly", func() {
			links := getStructLinks(post, serverInformationNil)
			Expect(links["comments"]).To(Equal(map[string]interface{}{
				"linkage": []map[string]interface{}{
					map[string]interface{}{
						"id":   "1",
						"type": "comments",
					},
				},
			}))
		})

		It("Generates self/related URLs with baseURL and prefix correctly", func() {
			links := getStructLinks(post, CompleteServerInformation{})
			Expect(links["author"]).To(Equal(map[string]interface{}{
				"linkage": map[string]interface{}{
					"id":   "1",
					"type": "users",
				},
				"self":    "http://my.domain/v1/posts/1/links/author",
				"related": "http://my.domain/v1/posts/1/author",
			}))
		})

		It("Generates self/related URLs with baseURL correctly", func() {
			links := getStructLinks(post, BaseURLServerInformation{})
			Expect(links["author"]).To(Equal(map[string]interface{}{
				"linkage": map[string]interface{}{
					"id":   "1",
					"type": "users",
				},
				"self":    "http://my.domain/posts/1/links/author",
				"related": "http://my.domain/posts/1/author",
			}))
		})

		It("Generates self/related URLs with prefix correctly", func() {
			links := getStructLinks(post, PrefixServerInformation{})
			Expect(links["author"]).To(Equal(map[string]interface{}{
				"linkage": map[string]interface{}{
					"id":   "1",
					"type": "users",
				},
				"self":    "/v1/posts/1/links/author",
				"related": "/v1/posts/1/author",
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
		var nullPost SqlNullPost

		BeforeEach(func() {
			nullPost = SqlNullPost{
				ID:     "theID",
				Title:  zero.StringFrom("Test"),
				Likes:  zero.IntFrom(666),
				Rating: zero.FloatFrom(66.66),
				IsCool: zero.BoolFrom(true),
			}
		})

		It("correctly marshalls String, Int64, Float64 and Bool", func() {
			result, err := MarshalToJSON(nullPost)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(MatchJSON(`
				{
					"data": {
						"id": "theID",
						"title": "Test",
						"likes": 666,
						"rating": 66.66,
						"isCool": true,
						"type": "sqlNullPosts"
					}
				}
			`))
		})
	})
})
