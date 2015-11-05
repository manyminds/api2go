package jsonapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/guregu/null.v2/zero"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unmarshal", func() {
	Context("When unmarshaling simple objects", func() {
		t, _ := time.Parse(time.RFC3339, "2014-11-10T16:30:48.823Z")
		/*
		 *singleJSON := []byte(`{"data":{"id": "1", "type": "simplePosts", "attributes": {"title":"First Post","text":"Lipsum", "Created": "2014-11-10T16:30:48.823Z"}}}`)
		 */
		firstPost := SimplePost{ID: "1", Title: "First Post", Text: "Lipsum", Created: t}
		secondPost := SimplePost{ID: "2", Title: "Second Post", Text: "Foobar!", Created: t, Updated: t}

		singlePostJSON := []byte(`
		{
			"data": {
				"id": "1",
				"type": "simplePosts",
				"attributes": {
					"title": "First Post",
					"text": "Lipsum",
					"created-date": "2014-11-10T16:30:48.823Z"
				}
			}
		}
		`)

		multiplePostJSON := []byte(`
				{
					"data": [
					{
						"id": "1",
						"type": "simplePosts",
						"attributes": {
							"title": "First Post",
							"text": "Lipsum",
							"created-date": "2014-11-10T16:30:48.823Z"
						}
					},
					{
						"id": "2",
						"type": "simplePosts",
						"attributes": {
							"title": "Second Post",
							"text": "Foobar!",
							"created-date": "2014-11-10T16:30:48.823Z",
							"updated-date": "2014-11-10T16:30:48.823Z"
						}
					}
					]
				}
				`)

		It("unmarshals single object into a struct", func() {
			var post SimplePost
			err := Unmarshal(singlePostJSON, &post)
			Expect(err).ToNot(HaveOccurred())
			Expect(post).To(Equal(firstPost))
		})

		It("unmarshals multiple objects into a slice", func() {
			var posts []SimplePost
			err := Unmarshal(multiplePostJSON, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]SimplePost{firstPost, secondPost}))
		})

		It("errors on invalid param nil", func() {
			err := Unmarshal(singlePostJSON, nil)
			Expect(err).Should(HaveOccurred())
		})

		It("errors on invalid param map", func() {
			err := Unmarshal(singlePostJSON, []interface{}{})
			Expect(err).Should(HaveOccurred())
		})

		It("errors on invalid pointer", func() {
			err := Unmarshal(singlePostJSON, &[]interface{}{})
			Expect(err).Should(HaveOccurred())
		})

		It("errors on non-array root", func() {
			var post SimplePost
			err := Unmarshal([]byte(`
			{
				"data": 42
			}
			`), &post)
			Expect(err).Should(HaveOccurred())
		})

		It("errors on non-documents", func() {
			var post SimplePost
			err := Unmarshal([]byte(`
			{
				"data": {42}
			}
			`), &post)
			Expect(err).Should(HaveOccurred())
		})

		It("it ignores fields that can not be unmarshaled like the nomral json.Unmarshaler", func() {
			var post SimplePost
			err := Unmarshal([]byte(`
			{
				"data": {
					"attributes": {
						"title": "something",
						"text": "blubb",
						"internal": "1337"
					},
					"type": "simplePosts"
				}
			}
			`), &post)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("errors with wrong type, expected int, got a string", func() {
			var post SimplePost
			err := Unmarshal([]byte(`
			{
				"data": {
					"attributes": {
						"text": "Gopher",
						"size": "blubb"
					},
					"type": "simplePosts"
				}
			}
			`), &post)
			Expect(err).To(HaveOccurred())
			Expect(err).Should(BeAssignableToTypeOf(&json.UnmarshalTypeError{}))
			typeError := err.(*json.UnmarshalTypeError)
			Expect(typeError.Value).To(Equal("string"))
		})

		It("errors with invalid time format", func() {
			t, err := time.Parse(time.RFC3339, "2014-11-10T16:30:48.823Z")
			faultyPostMap := []byte(`
			{
				"data": {
					"attributes": {
						"title": "` + firstPost.Title + `",
						"text": "` + firstPost.Text + `!",
						"created-date": "` + t.Format(time.RFC1123) + `"
					},
					"type": "simplePosts"
				}
			}`)
			var post SimplePost
			err = Unmarshal(faultyPostMap, &post)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("parsing time"))
		})

		Context("slice fields", func() {
			It("unmarshal slice fields with single entry correctly", func() {
				sliceJSON := []byte(`
				{
					"data": {
						"id":   "1234",
						"type": "identities",
						"attributes": {
							"scopes": [
								"user_global"
								]
							}
						}
					}
				`)
				var identity Identity
				err := Unmarshal(sliceJSON, &identity)
				Expect(err).ToNot(HaveOccurred())
				Expect(identity.Scopes).To(HaveLen(1))
				Expect(identity.Scopes[0]).To(Equal("user_global"))
			})

			It("unmarshal slice fields with multiple entries", func() {
				input := `
				{
					"data": {
						"id": "1234",
						"type": "identities",
						"attributes": {
							"scopes": ["test", "1234"]
						}
					}
				}
				`

				var identity Identity
				err := Unmarshal([]byte(input), &identity)
				Expect(err).ToNot(HaveOccurred())
				Expect(identity.Scopes[0]).To(Equal("test"))
				Expect(identity.Scopes[1]).To(Equal("1234"))
			})

			It("unmarshal empty slice fields from json input", func() {
				input := `
				{
					"data": {
						"id": "1234",
						"type": "identities",
						"attributes": {
							"scopes": []
						}
					}
				}
				`

				var identity Identity
				err := Unmarshal([]byte(input), &identity)
				Expect(err).ToNot(HaveOccurred())
				Expect(identity.Scopes).To(Equal([]string{}))
			})

			It("unmarshals renamed fields", func() {
				input := `
				{
					"data": {
						"id": "1",
						"type": "renamedPostWithEmbeddings",
						"attributes": {
							"foo": "field content",
							"bar-bar": "other content",
							"another": "foo"
						}
					}
				}`

				var renamedPost RenamedPostWithEmbedding
				err := Unmarshal([]byte(input), &renamedPost)
				Expect(err).ToNot(HaveOccurred())
				Expect(renamedPost.Field).To(Equal("field content"))
				Expect(renamedPost.Other).To(Equal("other content"))
			})
		})
	})

	Context("when unmarshaling objects with relationships", func() {
		It("unmarshals to many relationship IDs", func() {
			expectedPost := Post{ID: 1, CommentsIDs: []int{1}}
			postJSON := []byte(`
			{
				"data": {
					"id": "1",
					"type": "posts",
					"attributes": {},
					"relationships": {
						"comments": {
							"data": [
							{
								"id":   "1",
								"type": "links"
							}]
						}
					}
				}
			}
			`)
			var post Post
			err := Unmarshal(postJSON, &post)
			Expect(err).To(BeNil())
			Expect(expectedPost).To(Equal(post))
		})

		It("unmarshals aliased relationships with array data payload", func() {
			post := Post{ID: 1, CommentsIDs: []int{1}}
			postJSON := []byte(`
			{
				"data": [{
					"id":    "1",
					"type":  "posts",
					"attributes": {"title": "` + post.Title + `"},
					"relationships": {
						"comments": {
							"data": [{
								"id":   "1",
								"type": "votes"
							}]
						}
					}
				}]
			}
			`)
			var posts []Post
			err := Unmarshal(postJSON, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]Post{post}))
		})
	})

	Context("when unmarshaling objects with relations", func() {
		It("unmarshal to-one and to-many relations", func() {
			expectedPost := Post{ID: 3, Title: "Test", AuthorID: sql.NullInt64{Valid: true, Int64: 1}, Author: nil, CommentsIDs: []int{1, 2}}
			postJSON := []byte(`
			{
				"data": {
					"id":   "3",
					"type": "posts",
					"attributes": {
						"title": "Test"
					},
					"relationships": {
						"author": {
							"data": {
								"id":   "1",
								"type": "users"
							}
						},
						"comments": {
							"data": [
							{
								"id":   "1",
								"type": "comments"
							},
							{
								"id":   "2",
								"type": "comments"
							}
							]
						}
					}
				}
			}
			`)
			var post Post
			err := Unmarshal(postJSON, &post)
			Expect(err).To(BeNil())
			Expect(post).To(Equal(expectedPost))
		})

		It("check if type field matches target struct", func() {
			postJSON := []byte(`
			{
				"data": {
					"id":    "1",
					"type":  "totallyWrongType",
					"attributes": {
						"title": "Test"
					}
				}
			}`)
			var post Post
			err := Unmarshal(postJSON, &post)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when unmarshaling into an existing slice", func() {
		It("overrides existing entries", func() {
			post := Post{ID: 1, Title: "Old Title"}
			postJSON := []byte(`
			{
				"data": [{
					"id":   "1",
					"type": "posts",
					"attributes": {
						"title": "New Title"
					}
				}]
			}`)
			posts := []Post{post}
			err := Unmarshal(postJSON, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]Post{{ID: 1, Title: "New Title"}}))
		})
	})

	Context("when unmarshaling with null values", func() {
		It("adding a new entry", func() {
			expectedPost := SimplePost{ID: "1", Title: "Nice Title"}
			postJSON := []byte(`
			{
				"data": {
					"id":   "1",
					"type": "simplePosts",
					"attributes": {
						"title": "Nice Title",
						"text":  null
					}
				}
			}`)
			var post SimplePost
			err := Unmarshal(postJSON, &post)
			Expect(err).To(BeNil())
			Expect(post).To(Equal(expectedPost))
		})
	})

	Context("when unmarshaling without id", func() {
		It("adding a new entry", func() {
			expectedPost := SimplePost{Title: "Nice Title"}
			postJSON := []byte(`
			{
				"data": {
					"type": "simplePosts",
					"attributes": {
						"title": "Nice Title"
					}
				}
			}`)
			var post SimplePost
			err := Unmarshal(postJSON, &post)
			Expect(err).To(BeNil())
			Expect(post).To(Equal(expectedPost))
		})
	})

	Context("when unmarshalling objects with numbers", func() {
		It("correctly converts number to int64", func() {
			json := `
			{
				"data": [
				{
					"id": "test",
					"type": "numberPosts",
					"attributes": {
						"title": "Blubb",
						"number": 1337
					}
				}
				]
			}
			`
			var numberPosts []NumberPost

			err := Unmarshal([]byte(json), &numberPosts)
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
					"type": "numberPosts",
					"attributes": {
						"title": "Blubb",
						"number": -1337
					}
				}
				]
			}
			`
			var numberPosts []NumberPost

			err := Unmarshal([]byte(json), &numberPosts)
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
					"type": "numberPosts",
					"attributes": {
						"title": "Blubb",
						"unsignedNumber": 1337
					}
				}
				]
			}
			`

			var numberPosts []NumberPost

			err := Unmarshal([]byte(json), &numberPosts)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(numberPosts)).To(Equal(1))
			Expect(numberPosts[0].UnsignedNumber).To(Equal(uint64(1337)))
		})
	})

	Context("SQL Null-Types", func() {
		var (
			nullPosts []SQLNullPost
			nullPost  SQLNullPost
			timeZero  time.Time
		)

		BeforeEach(func() {
			nullPosts = []SQLNullPost{}
			nullPost = SQLNullPost{}
			timeZero = time.Time{}
		})

		It("correctly unmarshals String, Int64, Float64 and Time", func() {
			err := Unmarshal([]byte(fmt.Sprintf(`
			{
				"data": {
					"id": "theID",
					"type": "sqlNullPosts",
					"attributes": {
						"title": "Test",
						"likes": 666,
						"rating": 66.66,
						"isCool": true,
						"today": "%v"
					}
				}
			}
			`, timeZero.Format(time.RFC3339))), &nullPost)
			Expect(err).ToNot(HaveOccurred())
			Expect(nullPost).To(Equal(SQLNullPost{
				ID:     "theID",
				Title:  zero.StringFrom("Test"),
				Likes:  zero.IntFrom(666),
				Rating: zero.FloatFrom(66.66),
				IsCool: zero.BoolFrom(true),
				Today:  zero.TimeFrom(timeZero.UTC()),
			}))
		})

		It("correctly unmarshals Null String, Int64, Float64 and Time", func() {
			err := Unmarshal([]byte(`
			{
				"data": {
					"id": "theID",
					"type": "sqlNullPosts",
					"attributes": {
						"title": null,
						"likes": null,
						"rating": null,
						"isCool": null,
						"today": null
					}
				}
			}
			`), &nullPost)
			Expect(err).ToNot(HaveOccurred())
			Expect(nullPost).To(Equal(SQLNullPost{
				ID:     "theID",
				Title:  zero.StringFromPtr(nil),
				Likes:  zero.IntFromPtr(nil),
				Rating: zero.FloatFromPtr(nil),
				IsCool: zero.BoolFromPtr(nil),
				Today:  zero.TimeFromPtr(nil),
			}))
		})

		// No it will not do this because of the implementation in zero library.
		It("sets existing zero value to invalid when unmarshaling null values", func() {
			target := SQLNullPost{
				ID:     "newID",
				Title:  zero.StringFrom("TestTitle"),
				Likes:  zero.IntFrom(11),
				IsCool: zero.BoolFrom(true),
				Rating: zero.FloatFrom(4.5),
				Today:  zero.TimeFrom(time.Now().UTC())}
			err := Unmarshal([]byte(`
			{
				"data": {
					"id": "newID",
					"type": "sqlNullPosts",
					"attributes": {
						"title": null,
						"likes": null,
						"rating": null,
						"isCool": null,
						"today": null
					}
				}
			}
			`), &target)
			Expect(err).ToNot(HaveOccurred())
			Expect(target.Title.Valid).To(Equal(false))
			Expect(target.Likes.Valid).To(Equal(false))
			Expect(target.Rating.Valid).To(Equal(false))
			Expect(target.IsCool.Valid).To(Equal(false))
			Expect(target.Today.Valid).To(Equal(false))
		})
	})
})
