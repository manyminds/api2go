package api2go

import (
	"database/sql"
	"time"

	"gopkg.in/guregu/null.v2/zero"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unmarshal", func() {
	type SimplePost struct {
		ID          string
		Title, Text string
		Created     time.Time
	}

	type Post struct {
		ID          int
		Title       string
		CommentsIDs []int
		LikesIDs    []string
	}

	Context("When unmarshaling simple objects", func() {
		t, _ := time.Parse(time.RFC3339, "2014-11-10T16:30:48.823Z")
		singleJSON := []byte(`{"data":{"id": "1", "type": "simplePosts", "title":"First Post","text":"Lipsum", "Created": "2014-11-10T16:30:48.823Z"}}`)
		firstPost := SimplePost{ID: "1", Title: "First Post", Text: "Lipsum", Created: t}
		secondPost := SimplePost{ID: "2", Title: "Second Post", Text: "Foobar!", Created: t}
		singlePostMap := map[string]interface{}{
			"data": map[string]interface{}{
				"id":      "1",
				"type":    "simplePosts",
				"title":   firstPost.Title,
				"text":    firstPost.Text,
				"created": "2014-11-10T16:30:48.823Z",
			},
		}
		multiplePostMap := map[string]interface{}{
			"data": []interface{}{
				map[string]interface{}{
					"id":      "1",
					"type":    "simplePosts",
					"title":   firstPost.Title,
					"text":    firstPost.Text,
					"created": "2014-11-10T16:30:48.823Z",
				},
				map[string]interface{}{
					"id":      "2",
					"type":    "simplePosts",
					"title":   secondPost.Title,
					"text":    secondPost.Text,
					"created": "2014-11-10T16:30:48.823Z",
				},
			},
		}

		It("unmarshals single objects", func() {
			var posts []SimplePost
			err := Unmarshal(singlePostMap, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]SimplePost{firstPost}))
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
				"simplePosts": 42,
			}, &posts)
			Expect(err).ToNot(BeNil())
		})

		It("errors on non-documents", func() {
			var posts []SimplePost
			err := Unmarshal(map[string]interface{}{
				"simplePosts": []interface{}{42},
			}, &posts)
			Expect(err).ToNot(BeNil())
		})

		It("errors with wrong keys", func() {
			var posts []SimplePost
			err := Unmarshal(map[string]interface{}{
				"simplePosts": []interface{}{
					map[string]interface{}{
						"foobar": 42,
					},
				},
			}, &posts)
			Expect(err).ToNot(BeNil())
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
							"comments": []interface{}{"1"},
						},
					},
				},
			}
			var posts []Post
			err := Unmarshal(postMap, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]Post{post}))
		})

		It("unmarshals into string links", func() {
			post := Post{ID: 1, LikesIDs: []string{"1"}}
			postMap := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":    "1",
						"type":  "posts",
						"title": post.Title,
						"links": map[string]interface{}{
							"likes": map[string]interface{}{
								"ids":  []interface{}{"1"},
								"type": "likes",
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
			post := Post{ID: 1, LikesIDs: []string{"1"}}
			postMap := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":    "1",
						"type":  "posts",
						"title": post.Title,
						"links": map[string]interface{}{
							"likes": map[string]interface{}{
								"ids":  []interface{}{"1"},
								"type": "votes",
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

		It("unmarshals aliased links and normal links", func() {
			post := Post{ID: 1, LikesIDs: []string{"1"}, CommentsIDs: []int{2, 3}}
			postMap := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":    "1",
						"type":  "posts",
						"title": post.Title,
						"links": map[string]interface{}{
							"likes": map[string]interface{}{
								"ids":  []interface{}{"1"},
								"type": "votes",
							},
							"comments": []interface{}{"2", "3"},
							// "comments": map[string]interface{}{
							// 	"ids": []interface{}{"2", "3"},
							// 	"type": "comments",
							// },
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
			post := BlogPost{ID: 1, Text: "Test", AuthorID: 1, Author: nil}
			postMap := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":   "1",
						"type": "blogPosts",
						"text": "Test",
						"links": map[string]interface{}{
							"author": map[string]interface{}{
								"id":   "1",
								"type": "blogAuthors",
							},
						},
					},
				},
			}
			var posts []BlogPost
			err := Unmarshal(postMap, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]BlogPost{post}))
		})

		It("unmarshals aliased id", func() {
			post := BlogPost{ID: 1, Text: "Test", AuthorID: 1, Author: nil}
			postMap := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":   "1",
						"type": "bogPosts",
						"text": "Test",
						"links": map[string]interface{}{
							"author": map[string]interface{}{
								"id":   "1",
								"type": "user",
							},
						},
					},
				},
			}
			var posts []BlogPost
			err := Unmarshal(postMap, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]BlogPost{post}))
		})

		It("unmarshals aliased id and normal id", func() {
			post := BlogPost{ID: 3, Text: "Test", AuthorID: 1, Author: nil, ParentID: sql.NullInt64{Int64: 2, Valid: true}}
			postMap := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":   "3",
						"type": "blogPosts",
						"text": "Test",
						"links": map[string]interface{}{
							"author": map[string]interface{}{
								"id":   "1",
								"type": "user",
							},
							"parent": "2",
						},
					},
				},
			}
			var posts []BlogPost
			err := Unmarshal(postMap, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]BlogPost{post}))
		})

		It("unmarshal no linked content", func() {
			post := BlogPost{ID: 1, Text: "Test", AuthorID: 0, Author: nil}
			postMap := map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":    "1",
						"type":  "blogPosts",
						"text":  "Test",
						"links": map[string]interface{}{},
					},
				},
			}
			var posts []BlogPost
			err := Unmarshal(postMap, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]BlogPost{post}))
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

	Context("unmarshall all int datatypes", func() {
		It("Should work with uint", func() {
			type User struct {
				ID   uint
				Name string
			}

			var users []User
			userMap := map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{
						"id":   "1",
						"Name": "test"}}}

			err := Unmarshal(userMap, &users)
			Expect(err).To(BeNil())
		})

		It("Should work with uint8", func() {
			type User struct {
				ID   uint8
				Name string
			}

			var users []User
			userMap := map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{
						"id":   "1",
						"Name": "test"}}}

			err := Unmarshal(userMap, &users)
			Expect(err).To(BeNil())
		})

		It("Should work with uint16", func() {
			type User struct {
				ID   uint16
				Name string
			}

			var users []User
			userMap := map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{
						"id":   "1",
						"Name": "test"}}}

			err := Unmarshal(userMap, &users)
			Expect(err).To(BeNil())
		})

		It("Should work with uint32", func() {
			type User struct {
				ID   uint32
				Name string
			}

			var users []User
			userMap := map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{
						"id":   "1",
						"Name": "test"}}}

			err := Unmarshal(userMap, &users)
			Expect(err).To(BeNil())
		})

		It("Should work with uint64", func() {
			type User struct {
				ID   uint64
				Name string
			}

			var users []User
			userMap := map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{
						"id":   "1",
						"Name": "test"}}}

			err := Unmarshal(userMap, &users)
			Expect(err).To(BeNil())
		})

		It("Should work with int", func() {
			type User struct {
				ID   int
				Name string
			}

			var users []User
			userMap := map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{
						"id":   "1",
						"Name": "test"}}}

			err := Unmarshal(userMap, &users)
			Expect(err).To(BeNil())
		})

		It("Should work with int8", func() {
			type User struct {
				ID   int8
				Name string
			}

			var users []User
			userMap := map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{
						"id":   "1",
						"Name": "test"}}}

			err := Unmarshal(userMap, &users)
			Expect(err).To(BeNil())
		})

		It("Should work with int16", func() {
			type User struct {
				ID   int16
				Name string
			}

			var users []User
			userMap := map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{
						"id":   "1",
						"Name": "test"}}}

			err := Unmarshal(userMap, &users)
			Expect(err).To(BeNil())
		})

		It("Should work with int32", func() {
			type User struct {
				ID   int32
				Name string
			}

			var users []User
			userMap := map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{
						"id":   "1",
						"Name": "test"}}}

			err := Unmarshal(userMap, &users)
			Expect(err).To(BeNil())
		})

		It("Should work with int64", func() {
			type User struct {
				ID   int64
				Name string
			}

			var users []User
			userMap := map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{
						"id":   "1",
						"Name": "test"}}}

			err := Unmarshal(userMap, &users)
			Expect(err).To(BeNil())
		})

		It("Should work with sql.NullString with value", func() {
			type User struct {
				ID        int64
				Name      string
				ForeignID sql.NullString
			}

			var users []User
			userMap := map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{
						"id":   "1",
						"Name": "test",
						"links": map[string]interface{}{
							"foreign": "1337",
						}}}}

			err := Unmarshal(userMap, &users)
			Expect(err).To(BeNil())
			Expect(len(users)).To(Equal(1))
			Expect(users[0].ForeignID).To(Equal(sql.NullString{"1337", true}))
		})

		It("Should work with sql.NullInt64 with value", func() {
			type User struct {
				ID        int64
				Name      string
				ForeignID sql.NullInt64
			}

			var users []User
			userMap := map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{
						"id":   "1",
						"Name": "test",
						"links": map[string]interface{}{
							"foreign": "1337",
						}}}}

			err := Unmarshal(userMap, &users)
			Expect(err).To(BeNil())
			Expect(len(users)).To(Equal(1))
			Expect(users[0].ForeignID).To(Equal(sql.NullInt64{1337, true}))
		})
	})

	Context("when using zero values", func() {
		type ZeroPost struct {
			ID    string
			Title string
			Value *zero.Float
		}

		It("correctly unmarshals driver values", func() {
			postMap := map[string]interface{}{
				"zeroPosts": []interface{}{
					map[string]interface{}{
						"id":    "1",
						"title": "test",
						"value": 2.3,
					},
				},
			}

			var zeroPosts []ZeroPost

			err := Unmarshal(postMap, &zeroPosts)
			Expect(err).To(BeNil())
			Expect(len(zeroPosts)).To(Equal(1))
			Expect(*zeroPosts[0].Value).To(Equal(zero.NewFloat(2.3, true)))
		})

		type ZeroPostValue struct {
			ID    string
			Title string
			Value zero.Float
		}

		It("correctly unmarshals driver values", func() {
			postMap := map[string]interface{}{
				"zeroPostValues": []interface{}{
					map[string]interface{}{
						"id":    "1",
						"title": "test",
						"value": 2.3,
					},
				},
			}

			var zeroPosts []ZeroPostValue

			err := Unmarshal(postMap, &zeroPosts)
			Expect(err).To(BeNil())
			Expect(len(zeroPosts)).To(Equal(1))
			Expect(zeroPosts[0].Value).To(Equal(zero.NewFloat(2.3, true)))
		})
	})

	Context("when unmarshalling objects with numbers", func() {
		type NumberPost struct {
			ID             string
			Title          string
			Number         int64
			UnsignedNumber uint64
		}

		It("correctly converts number to int64", func() {
			json := `
				{
					"numberPosts": [
						{
							"id": "test",
							"title": "Blubb",
							"number": 1337
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
					"numberPosts": [
						{
							"id": "test",
							"title": "Blubb",
							"number": -1337
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
					"numberPosts": [
						{
							"id": "test",
							"title": "Blubb",
							"unsignedNumber": 1337
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
})
