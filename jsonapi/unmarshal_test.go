package jsonapi

import (
	"database/sql"
	"time"

	"gopkg.in/guregu/null.v2/zero"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unmarshal", func() {
	Context("When unmarshaling simple objects", func() {
		t, _ := time.Parse(time.RFC3339, "2014-11-10T16:30:48.823Z")
		singleJSON := []byte(`{"data":{"id": "1", "type": "simplePosts", "title":"First Post","text":"Lipsum", "Created": "2014-11-10T16:30:48.823Z"}}`)
		firstPost := SimplePost{ID: "1", Title: "First Post", Text: "Lipsum", Created: t}
		secondPost := SimplePost{ID: "2", Title: "Second Post", Text: "Foobar!", Created: t}
		singlePostMap := map[string]interface{}{
			"data": map[string]interface{}{
				"id":          "1",
				"type":        "simplePosts",
				"title":       firstPost.Title,
				"text":        firstPost.Text,
				"create-date": "2014-11-10T16:30:48.823Z",
			},
		}
		multiplePostMap := map[string]interface{}{
			"data": []interface{}{
				map[string]interface{}{
					"id":          "1",
					"type":        "simplePosts",
					"title":       firstPost.Title,
					"text":        firstPost.Text,
					"create-date": "2014-11-10T16:30:48.823Z",
				},
				map[string]interface{}{
					"id":          "2",
					"type":        "simplePosts",
					"title":       secondPost.Title,
					"text":        secondPost.Text,
					"create-date": "2014-11-10T16:30:48.823Z",
				},
			},
		}

		It("unmarshals single objects into a slice", func() {
			var posts []SimplePost
			err := Unmarshal(singlePostMap, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]SimplePost{firstPost}))
		})

		It("unmarshals single objects into a struct", func() {
			var post SimplePost
			err := Unmarshal(singlePostMap, &post)
			Expect(err).To(BeNil())
			Expect(post).To(Equal(firstPost))
		})

		It("unmarshals multiple objects", func() {
			var posts []SimplePost
			err := Unmarshal(multiplePostMap, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]SimplePost{firstPost, secondPost}))
		})

		It("errors on invalid param nil", func() {
			err := Unmarshal(singlePostMap, nil)
			Expect(err).Should(HaveOccurred())
		})

		It("errors on invalid param int", func() {
			err := Unmarshal(singlePostMap, nil)
			Expect(err).Should(HaveOccurred())
		})

		It("errors on invalid param map", func() {
			err := Unmarshal(singlePostMap, []interface{}{})
			Expect(err).Should(HaveOccurred())
		})

		It("errors on invalid pointer", func() {
			err := Unmarshal(singlePostMap, &[]interface{}{})
			Expect(err).Should(HaveOccurred())
		})

		It("errors on empty maps", func() {
			var posts []SimplePost
			err := Unmarshal(map[string]interface{}{}, &posts)
			Expect(err).ToNot(BeNil())
		})

		It("errors on non-array root", func() {
			var posts []SimplePost
			err := Unmarshal(map[string]interface{}{
				"data": 42,
			}, &posts)
			Expect(err).ToNot(BeNil())
		})

		It("errors on non-documents", func() {
			var posts []SimplePost
			err := Unmarshal(map[string]interface{}{
				"data": []interface{}{42},
			}, &posts)
			Expect(err).ToNot(BeNil())
		})

		It("errors with wrong keys", func() {
			var posts []SimplePost
			err := Unmarshal(map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"foobar": 42,
					},
				},
			}, &posts)
			Expect(err).ToNot(BeNil())
		})

		It("errors with wrong type, expected int, got a string", func() {
			var posts []SimplePost
			err := Unmarshal(map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"text": "Gopher",
						"size": "blubb",
					},
				},
			}, &posts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Could not set field 'Size'. Value 'blubb' had wrong type"))
		})

		It("errors with invalid time format", func() {
			t, err := time.Parse(time.RFC3339, "2014-11-10T16:30:48.823Z")
			faultyPostMap := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":      "1",
						"type":    "simplePosts",
						"title":   firstPost.Title,
						"text":    firstPost.Text,
						"created": t.Format(time.RFC1123Z),
					},
				},
			}
			var posts []SimplePost
			err = Unmarshal(faultyPostMap, &posts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("expected RFC3339 time string, got 'Mon, 10 Nov 2014 16:30:48 +0000'"))
		})

		It("unmarshals JSON", func() {
			var posts []SimplePost
			err := UnmarshalFromJSON(singleJSON, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]SimplePost{firstPost}))
		})
	})

	Context("when unmarshaling objects with links", func() {
		It("unmarshals into integer links", func() {
			post := Post{ID: 1, CommentsIDs: []int{1}}
			postMap := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":    "1",
						"type":  "posts",
						"title": post.Title,
						"links": map[string]interface{}{
							"comments": map[string]interface{}{
								"linkage": []interface{}{
									map[string]interface{}{
										"id":   "1",
										"type": "links",
									},
								},
							},
						},
					},
				},
			}
			var posts []Post
			err := Unmarshal(postMap, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]Post{post}))
		})

		It("unmarshals aliased links", func() {
			post := Post{ID: 1, CommentsIDs: []int{1}}
			postMap := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":    "1",
						"type":  "posts",
						"title": post.Title,
						"links": map[string]interface{}{
							"comments": map[string]interface{}{
								"linkage": []interface{}{
									map[string]interface{}{
										"id":   "1",
										"type": "votes",
									},
								},
							},
						},
					},
				},
			}
			var posts []Post
			err := Unmarshal(postMap, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]Post{post}))
		})
	})

	Context("when unmarshaling objects with single relation", func() {
		type BlogAuthor struct {
			ID   int
			Name string
		}

		type BlogPost struct {
			ID       int
			Text     string
			AuthorID int
			Author   *BlogAuthor
			ParentID sql.NullInt64
		}

		It("unmarshals author id", func() {
			post := Post{ID: 1, Title: "Test", AuthorID: sql.NullInt64{Valid: true, Int64: 1}, Author: nil}
			postMap := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":    "1",
						"type":  "posts",
						"title": "Test",
						"links": map[string]interface{}{
							"author": map[string]interface{}{
								"linkage": map[string]interface{}{
									"id":   "1",
									"type": "users",
								},
							},
						},
					},
				},
			}
			var posts []Post
			err := Unmarshal(postMap, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]Post{post}))
		})

		It("unmarshal to-one and to-many relations", func() {
			post := Post{ID: 3, Title: "Test", AuthorID: sql.NullInt64{Valid: true, Int64: 1}, Author: nil, CommentsIDs: []int{1, 2}}
			postMap := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":    "3",
						"type":  "posts",
						"title": "Test",
						"links": map[string]interface{}{
							"author": map[string]interface{}{
								"linkage": map[string]interface{}{
									"id":   "1",
									"type": "users",
								},
							},
							"comments": map[string]interface{}{
								"linkage": []interface{}{
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
						},
					},
				},
			}
			var posts []Post
			err := Unmarshal(postMap, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]Post{post}))
		})

		It("unmarshal no linked content", func() {
			post := Post{ID: 1, Title: "Test"}
			postMap := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":    "1",
						"type":  "posts",
						"title": "Test",
					},
				},
			}
			var posts []Post
			err := Unmarshal(postMap, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]Post{post}))
		})

		It("check if type field matches target struct", func() {
			postMap := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":    "1",
						"type":  "blogPosts",
						"title": "Test",
					},
				},
			}
			var posts []Post
			err := Unmarshal(postMap, &posts)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when unmarshaling into an existing slice", func() {
		It("updates existing entries", func() {
			post := Post{ID: 1, Title: "Old Title"}
			postMap := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":    "1",
						"type":  "posts",
						"title": "New Title",
					},
				},
			}
			posts := []Post{post}
			err := Unmarshal(postMap, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]Post{Post{ID: 1, Title: "New Title"}}))
		})
	})

	Context("when unmarshaling with null values", func() {
		It("adding a new entry", func() {
			post := SimplePost{ID: "1", Title: "Nice Title"}
			postMap := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":    "1",
						"type":  "simplePosts",
						"title": "Nice Title",
						"text":  nil,
					},
				},
			}
			var posts []SimplePost
			err := Unmarshal(postMap, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]SimplePost{post}))
		})
	})

	Context("when unmarshaling without id", func() {
		It("adding a new entry", func() {
			post := SimplePost{Title: "Nice Title"}
			postMap := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"title": "Nice Title",
						"type":  "simplePosts",
					},
				},
			}
			var posts []SimplePost
			err := Unmarshal(postMap, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]SimplePost{post}))
		})
	})

	Context("when unmarshalling objects with numbers", func() {
		It("correctly converts number to int64", func() {
			json := `
				{
					"data": [
						{
							"id": "test",
							"title": "Blubb",
							"number": 1337,
							"type": "numberPosts"
						}
					]
				}
			`

			var numberPosts []NumberPost

			err := UnmarshalFromJSON([]byte(json), &numberPosts)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(numberPosts)).To(Equal(1))
			Expect(numberPosts[0].Number).To(Equal(int64(1337)))
		})

		It("correctly converts negative number to int64", func() {
			json := `
				{
					"data": [
						{
							"id": "test",
							"title": "Blubb",
							"number": -1337,
							"type": "numberPosts"
						}
					]
				}
			`

			var numberPosts []NumberPost

			err := UnmarshalFromJSON([]byte(json), &numberPosts)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(numberPosts)).To(Equal(1))
			Expect(numberPosts[0].Number).To(Equal(int64(-1337)))
		})

		It("correctly converts number to uint64", func() {
			json := `
				{
					"data": [
						{
							"id": "test",
							"title": "Blubb",
							"unsignedNumber": 1337,
							"type": "numberPosts"
						}
					]
				}
			`

			var numberPosts []NumberPost

			err := UnmarshalFromJSON([]byte(json), &numberPosts)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(numberPosts)).To(Equal(1))
			Expect(numberPosts[0].UnsignedNumber).To(Equal(uint64(1337)))
		})
	})

	Context("SQL Null-Types", func() {
		var nullPosts []SqlNullPost

		BeforeEach(func() {
			nullPosts = []SqlNullPost{}
		})

		It("correctly unmarshal String, Int64 and Float64", func() {
			err := UnmarshalFromJSON([]byte(`
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
			`), &nullPosts)
			Expect(err).ToNot(HaveOccurred())
			Expect(nullPosts).To(HaveLen(1))
			Expect(nullPosts[0]).To(Equal(SqlNullPost{
				ID:     "theID",
				Title:  zero.StringFrom("Test"),
				Likes:  zero.IntFrom(666),
				Rating: zero.FloatFrom(66.66),
				IsCool: zero.BoolFrom(true),
			}))
		})
	})
})
