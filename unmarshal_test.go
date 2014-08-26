package api2go

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unmarshal", func() {
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

	Context("When unmarshaling simple objects", func() {
		singleJSON := []byte(`{"simple_posts":[{"title":"First Post","text":"Lipsum"}]}`)
		firstPost := SimplePost{Title: "First Post", Text: "Lipsum"}
		secondPost := SimplePost{Title: "Second Post", Text: "Foobar!"}
		singlePostMap := map[string]interface{}{
			"simple_posts": []interface{}{
				map[string]interface{}{
					"title": firstPost.Title,
					"text":  firstPost.Text,
				},
			},
		}
		multiplePostMap := map[string]interface{}{
			"simple_posts": []interface{}{
				map[string]interface{}{
					"title": firstPost.Title,
					"text":  firstPost.Text,
				},
				map[string]interface{}{
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

		It("panics on invalid params", func() {
			Expect(func() {
				Unmarshal(singlePostMap, nil)
			}).To(Panic())
			Expect(func() {
				Unmarshal(singlePostMap, 42)
			}).To(Panic())
			Expect(func() {
				Unmarshal(singlePostMap, []interface{}{})
			}).To(Panic())
			Expect(func() {
				Unmarshal(singlePostMap, &[]interface{}{})
			}).To(Panic())
		})

		It("errors on empty maps", func() {
			var posts []SimplePost
			err := Unmarshal(map[string]interface{}{}, &posts)
			Expect(err).ToNot(BeNil())
		})

		It("errors on non-array root", func() {
			var posts []SimplePost
			err := Unmarshal(map[string]interface{}{
				"simple_posts": 42,
			}, &posts)
			Expect(err).ToNot(BeNil())
		})

		It("errors on non-documents", func() {
			var posts []SimplePost
			err := Unmarshal(map[string]interface{}{
				"simple_posts": []interface{}{42},
			}, &posts)
			Expect(err).ToNot(BeNil())
		})

		It("errors with wrong keys", func() {
			var posts []SimplePost
			err := Unmarshal(map[string]interface{}{
				"simple_posts": []interface{}{
					map[string]interface{}{
						"foobar": 42,
					},
				},
			}, &posts)
			Expect(err).ToNot(BeNil())
		})

		It("unmarshals JSON", func() {
			var posts []SimplePost
			err := UnmarshalJSON(singleJSON, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]SimplePost{firstPost}))
		})
	})
})
