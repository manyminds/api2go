package api2go

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unmarshal", func() {
	type SimplePost struct {
		ID          string
		Title, Text string
	}

	type Post struct {
		ID          int
		Title       string
		CommentsIDs []int
		LikesIDs    []string
	}

	Context("When unmarshaling simple objects", func() {
		singleJSON := []byte(`{"simplePosts":[{"id": "1", "title":"First Post","text":"Lipsum"}]}`)
		firstPost := SimplePost{ID: "1", Title: "First Post", Text: "Lipsum"}
		secondPost := SimplePost{ID: "2", Title: "Second Post", Text: "Foobar!"}
		singlePostMap := map[string]interface{}{
			"simplePosts": []interface{}{
				map[string]interface{}{
					"id":    "1",
					"title": firstPost.Title,
					"text":  firstPost.Text,
				},
			},
		}
		multiplePostMap := map[string]interface{}{
			"simplePosts": []interface{}{
				map[string]interface{}{
					"id":    "1",
					"title": firstPost.Title,
					"text":  firstPost.Text,
				},
				map[string]interface{}{
					"id":    "2",
					"title": secondPost.Title,
					"text":  secondPost.Text,
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
				"posts": []interface{}{
					map[string]interface{}{
						"id":    "1",
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
				"posts": []interface{}{
					map[string]interface{}{
						"id":    "1",
						"title": post.Title,
						"links": map[string]interface{}{
							"likes": []interface{}{"1"},
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
		}

		It("unmarshals author id", func() {
			post := BlogPost{ID: 1, Text: "Test", AuthorID: 1, Author: nil}
			postMap := map[string]interface{}{
				"blogPosts": []interface{}{
					map[string]interface{}{
						"id":   "1",
						"text": "Test",
						"links": map[string]interface{}{
							"author": "1",
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
				"blogPosts": []interface{}{
					map[string]interface{}{
						"id":    "1",
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
				"posts": []interface{}{
					map[string]interface{}{
						"id":    "1",
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
				"simplePosts": []interface{}{
					map[string]interface{}{
						"id":    "1",
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
				"simplePosts": []interface{}{
					map[string]interface{}{
						"title": "Nice Title",
					},
				},
			}
			var posts []SimplePost
			err := Unmarshal(postMap, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]SimplePost{post}))
		})
	})
})
