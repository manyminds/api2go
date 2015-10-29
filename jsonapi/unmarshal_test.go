package jsonapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"gopkg.in/guregu/null.v2/zero"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unmarshal", func() {
	Context("When unmarshaling simple objects", func() {
		t, _ := time.Parse(time.RFC3339, "2014-11-10T16:30:48.823Z")
		singleJSON := []byte(`{"data":{"id": "1", "type": "simplePosts", "attributes": {"title":"First Post","text":"Lipsum", "Created": "2014-11-10T16:30:48.823Z"}}}`)
		firstPost := SimplePost{ID: "1", Title: "First Post", Text: "Lipsum", Created: t}
		secondPost := SimplePost{ID: "2", Title: "Second Post", Text: "Foobar!", Created: t, Updated: t}
		singlePostMap := map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "1",
				"type": "simplePosts",
				"attributes": map[string]interface{}{
					"title":       firstPost.Title,
					"text":        firstPost.Text,
					"create-date": "2014-11-10T16:30:48.823Z",
				},
			},
		}
		multiplePostMap := map[string]interface{}{
			"data": []interface{}{
				map[string]interface{}{
					"id":   "1",
					"type": "simplePosts",
					"attributes": map[string]interface{}{
						"title":        firstPost.Title,
						"text":         firstPost.Text,
						"create-date":  "2014-11-10T16:30:48.823Z",
						"updated-date": nil,
					},
				},
				map[string]interface{}{
					"id":   "2",
					"type": "simplePosts",
					"attributes": map[string]interface{}{
						"title":        secondPost.Title,
						"text":         secondPost.Text,
						"create-date":  "2014-11-10T16:30:48.823Z",
						"updated-date": "2014-11-10T16:30:48.823Z",
					},
				},
			},
		}

		It("unmarshals single objects into a slice", func() {
			var posts []SimplePost
			err := Unmarshal(singlePostMap, &posts)
			Expect(err).ToNot(HaveOccurred())
			Expect(posts).To(Equal([]SimplePost{firstPost}))
		})
		It("unmarshals single objects into a struct", func() {
			var post SimplePost
			err := Unmarshal(singlePostMap, &post)
			Expect(err).To(BeNil())
			Expect(post).To(Equal(firstPost))
		})

		It("does not unmarshal private fields", func() {
			failingPostMap := map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "1",
					"type": "simplePosts",
					"attributes": map[string]interface{}{
						"title":       firstPost.Title,
						"text":        firstPost.Text,
						"create-date": "2014-11-10T16:30:48.823Z",
						"top-secret":  "fish",
					},
				},
			}
			var post SimplePost
			err := Unmarshal(failingPostMap, &post)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("field not exported. Expected field with name Top-secret to exist"))
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
				"data": map[string]interface{}{
					"attributes": map[string]interface{}{
						"foobar": 42,
					},
				},
			}, &posts)
			Expect(err).To(HaveOccurred())
		})

		It("errors if a field is in the json that is ignored by the struct", func() {
			var post SimplePost
			err := Unmarshal(map[string]interface{}{
				"data": map[string]interface{}{
					"id": "1234",
					"attributes": map[string]interface{}{
						"title":    "something",
						"text":     "blubb",
						"internal": "1337",
					},
				},
			}, &post)
			Expect(err.Error()).To(Equal("invalid key \"internal\" in json. Cannot be assigned to target struct \"SimplePost\""))
		})

		It("errors with wrong type, expected int, got a string", func() {
			var posts []SimplePost
			err := Unmarshal(map[string]interface{}{
				"data": map[string]interface{}{
					"attributes": map[string]interface{}{
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
				"data": map[string]interface{}{
					"id":   "1",
					"type": "simplePosts",
					"attributes": map[string]interface{}{
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

		Context("slice fields", func() {
			It("unmarshal slice fields correctly", func() {
				sliceMap := map[string]interface{}{
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

				var identity Identity
				err := Unmarshal(sliceMap, &identity)
				Expect(err).ToNot(HaveOccurred())
				Expect(identity.Scopes).To(HaveLen(1))
				Expect(identity.Scopes[0]).To(Equal("user_global"))
			})

			It("unmarshal slice fields from json input", func() {
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
				err := UnmarshalFromJSON([]byte(input), &identity)
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
				err := UnmarshalFromJSON([]byte(input), &identity)
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
				err := UnmarshalFromJSON([]byte(input), &renamedPost)
				Expect(err).ToNot(HaveOccurred())
				Expect(renamedPost.Field).To(Equal("field content"))
				Expect(renamedPost.Other).To(Equal("other content"))
			})
		})
	})

	Context("when unmarshaling objects with relationships", func() {
		It("unmarshals into integer relationships", func() {
			post := Post{ID: 1, CommentsIDs: []int{1}}
			postMap := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":    "1",
						"type":  "posts",
						"title": post.Title,
						"relationships": map[string]interface{}{
							"comments": map[string]interface{}{
								"data": []interface{}{
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

		It("unmarshals aliased relationships", func() {
			post := Post{ID: 1, CommentsIDs: []int{1}}
			postMap := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":    "1",
						"type":  "posts",
						"title": post.Title,
						"relationships": map[string]interface{}{
							"comments": map[string]interface{}{
								"data": []interface{}{
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
						"id":   "1",
						"type": "posts",
						"attributes": map[string]interface{}{
							"title": "Test",
						},
						"relationships": map[string]interface{}{
							"author": map[string]interface{}{
								"data": map[string]interface{}{
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
						"id":   "3",
						"type": "posts",
						"attributes": map[string]interface{}{
							"title": "Test",
						},
						"relationships": map[string]interface{}{
							"author": map[string]interface{}{
								"data": map[string]interface{}{
									"id":   "1",
									"type": "users",
								},
							},
							"comments": map[string]interface{}{
								"data": []interface{}{
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
						"id":   "1",
						"type": "posts",
						"attributes": map[string]interface{}{
							"title": "Test",
						},
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
						"id":   "1",
						"type": "posts",
						"attributes": map[string]interface{}{
							"title": "New Title",
						},
					},
				},
			}
			posts := []Post{post}
			err := Unmarshal(postMap, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]Post{{ID: 1, Title: "New Title"}}))
		})
	})

	Context("when unmarshaling with null values", func() {
		It("adding a new entry", func() {
			post := SimplePost{ID: "1", Title: "Nice Title"}
			postMap := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":   "1",
						"type": "simplePosts",
						"attributes": map[string]interface{}{
							"title": "Nice Title",
							"text":  nil,
						},
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
						"type": "simplePosts",
						"attributes": map[string]interface{}{
							"title": "Nice Title",
						},
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

			err := UnmarshalFromJSON([]byte(json), &numberPosts)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(numberPosts)).To(Equal(1))
			Expect(numberPosts[0].UnsignedNumber).To(Equal(uint64(1337)))
		})
	})

	Context("SQL Null-Types", func() {
		var nullPosts []SQLNullPost
		var timeZero time.Time

		BeforeEach(func() {
			nullPosts = []SQLNullPost{}
			timeZero = time.Time{}
		})

		It("correctly unmarshals String, Int64, Float64 and Time", func() {
			err := UnmarshalFromJSON([]byte(fmt.Sprintf(`
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
				`, timeZero.Format(time.RFC3339))), &nullPosts)
			Expect(err).ToNot(HaveOccurred())
			Expect(nullPosts).To(HaveLen(1))
			Expect(nullPosts[0]).To(Equal(SQLNullPost{
				ID:     "theID",
				Title:  zero.StringFrom("Test"),
				Likes:  zero.IntFrom(666),
				Rating: zero.FloatFrom(66.66),
				IsCool: zero.BoolFrom(true),
				Today:  zero.TimeFrom(timeZero.UTC()),
			}))
		})

		It("correctly unmarshals Null String, Int64, Float64 and Time", func() {
			err := UnmarshalFromJSON([]byte(`
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
				`), &nullPosts)
			Expect(err).ToNot(HaveOccurred())
			Expect(nullPosts).To(HaveLen(1))
			Expect(nullPosts[0]).To(Equal(SQLNullPost{
				ID:     "theID",
				Title:  zero.StringFromPtr(nil),
				Likes:  zero.IntFromPtr(nil),
				Rating: zero.FloatFromPtr(nil),
				IsCool: zero.BoolFromPtr(nil),
				Today:  zero.TimeFromPtr(nil),
			}))
		})

		It("overwrites existing data with nulls when marshaling", func() {
			target := SQLNullPost{
				ID:     "newID",
				Title:  zero.StringFrom("TestTitle"),
				Likes:  zero.IntFrom(11),
				IsCool: zero.BoolFrom(true),
				Rating: zero.FloatFrom(4.5),
				Today:  zero.TimeFrom(time.Now().UTC())}
			nullPosts = append(nullPosts, target)
			Expect(nullPosts).To(HaveLen(1))
			ctx := map[string]interface{}{}
			err := json.Unmarshal([]byte(`
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
				`), &ctx)
			Expect(err).ToNot(HaveOccurred())
			// This follows the technique used in api.go
			updatingObjs := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(target)), 1, 1)
			updatingObjs.Index(0).Set(reflect.ValueOf(target))
			err = UnmarshalInto(ctx, reflect.TypeOf(target), &updatingObjs)
			Expect(err).ToNot(HaveOccurred())
			updatingObj := updatingObjs.Index(0).Interface()
			Expect(updatingObjs.Len()).To(Equal(1))
			Expect(updatingObj).To(Equal(SQLNullPost{
				ID:     "newID",
				Title:  zero.StringFromPtr(nil),
				Likes:  zero.IntFromPtr(nil),
				Rating: zero.FloatFromPtr(nil),
				IsCool: zero.BoolFromPtr(nil),
				Today:  zero.TimeFromPtr(nil),
			}))
		})

	})
})
