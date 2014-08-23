package api2go

import (
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type SimplePost struct {
	Title, Text string
}

type Post struct {
	ID       int
	Title    string
	Comments []Comment
}

type Comment struct {
	ID   int
	Text string
}

var _ = Describe("Marshalling", func() {
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
	})

	Context("When marshaling compound objects", func() {
		var (
			post               Post
			comment1, comment2 Comment
		)

		BeforeEach(func() {
			comment1 = Comment{ID: 1, Text: "First!"}
			comment2 = Comment{ID: 2, Text: "Second!"}
			post = Post{ID: 1, Title: "Foobar", Comments: []Comment{comment1, comment2}}
		})

		It("marshals objects", func() {
			i, err := Marshal(post)
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"posts": []interface{}{
					map[string]interface{}{
						"id":    1,
						"title": "Foobar",
						"links": map[string][]interface{}{
							"comments": []interface{}{1, 2},
						},
					},
				},
				"linked": map[string][]interface{}{
					"comments": []interface{}{
						map[string]interface{}{
							"id":   1,
							"text": "First!",
						},
						map[string]interface{}{
							"id":   2,
							"text": "Second!",
						},
					},
				},
			}))
		})
	})
})
