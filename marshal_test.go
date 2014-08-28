package api2go

import (
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Marshalling", func() {
	type SimplePost struct {
		Title, Text string
	}

	type Comment struct {
		ID   int
		Text string
	}

	type Post struct {
		ID       int
		Title    string
		Comments []Comment
	}

	Context("When marshaling simple objects", func() {
		var (
			firstPost, secondPost       SimplePost
			firstPostMap, secondPostMap map[string]interface{}
		)

		BeforeEach(func() {
			firstPost = SimplePost{Title: "First Post", Text: "Lipsum"}
			firstPostMap = map[string]interface{}{
				"title": firstPost.Title,
				"text":  firstPost.Text,
			}
			secondPost = SimplePost{Title: "Second Post", Text: "Getting more advanced!"}
			secondPostMap = map[string]interface{}{
				"title": secondPost.Title,
				"text":  secondPost.Text,
			}
		})

		It("marshals single object", func() {
			i, err := Marshal(firstPost)
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"simple_posts": []interface{}{
					firstPostMap,
				},
			}))
		})

		It("marshals collections object", func() {
			i, err := Marshal([]SimplePost{firstPost, secondPost})
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"simple_posts": []interface{}{
					firstPostMap,
					secondPostMap,
				},
			}))
		})

		It("marshals empty collections", func() {
			i, err := Marshal([]SimplePost{})
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"simple_posts": []interface{}{},
			}))
		})

		It("panics when passing interface{} slices", func() {
			Expect(func() {
				Marshal([]interface{}{})
			}).To(Panic())
		})

		It("marshals to JSON", func() {
			j, err := MarshalToJSON([]SimplePost{firstPost})
			Expect(err).To(BeNil())
			var m map[string]interface{}
			Expect(json.Unmarshal(j, &m)).To(BeNil())
			Expect(m).To(Equal(map[string]interface{}{
				"simple_posts": []interface{}{
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
					"string_ids": []interface{}{
						map[string]interface{}{
							"id": "1",
						},
					},
				}))
			})

			It("converts ints", func() {
				type IntID struct{ ID int }
				i, err := Marshal(IntID{ID: 1})
				Expect(err).To(BeNil())
				Expect(i).To(Equal(map[string]interface{}{
					"int_ids": []interface{}{
						map[string]interface{}{
							"id": "1",
						},
					},
				}))
			})

			It("converts uints", func() {
				type UintID struct{ ID uint }
				i, err := Marshal(UintID{ID: 1})
				Expect(err).To(BeNil())
				Expect(i).To(Equal(map[string]interface{}{
					"uint_ids": []interface{}{
						map[string]interface{}{
							"id": "1",
						},
					},
				}))
			})
		})
	})

	Context("When marshaling compound objects", func() {
		var (
			posts              []Post
			comment1, comment2 Comment
		)

		BeforeEach(func() {
			comment1 = Comment{ID: 1, Text: "First!"}
			comment2 = Comment{ID: 2, Text: "Second!"}
			post1 := Post{ID: 1, Title: "Foobar", Comments: []Comment{comment1, comment2}}
			post2 := Post{ID: 2, Title: "Foobarbarbar", Comments: []Comment{comment1, comment2}}

			posts = append(posts, post1, post2)
		})

		It("marshals objects", func() {
			i, err := Marshal(posts)
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"posts": []interface{}{
					map[string]interface{}{
						"id":    "1",
						"title": "Foobar",
						"links": map[string][]interface{}{
							"comments": []interface{}{"1", "2"},
						},
					},
					map[string]interface{}{
						"id":    "2",
						"title": "Foobarbarbar",
						"links": map[string][]interface{}{
							"comments": []interface{}{"1", "2"},
						},
					},
				},
				"linked": map[string][]interface{}{
					"comments": []interface{}{
						map[string]interface{}{
							"id":   "1",
							"text": "First!",
						},
						map[string]interface{}{
							"id":   "2",
							"text": "Second!",
						},
					},
				},
			}))
		})
	})
})
