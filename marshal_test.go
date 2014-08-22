package api2go_test

import (
	"github.com/univedo/api2go"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type SimplePost struct {
	Title, Text string
}

var _ = Describe("Marshalling", func() {
	Context("When marshaling simple objects", func() {
		var (
			firstPost, secondPost SimplePost
		)

		BeforeEach(func() {
			firstPost = SimplePost{Title: "First Post", Text: "Lipsum"}
			secondPost = SimplePost{Title: "Second Post", Text: "Getting more advanced!"}
		})

		It("marshals single object", func() {
			i, err := api2go.Marshal(firstPost)
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"simple_posts": []interface{}{
					firstPost,
				},
			}))
		})

		It("marshals collections object", func() {
			i, err := api2go.Marshal([]SimplePost{firstPost, secondPost})
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"simple_posts": []SimplePost{
					firstPost,
					secondPost,
				},
			}))
		})

		It("marshals empty collections", func() {
			i, err := api2go.Marshal([]SimplePost{})
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"simple_posts": []SimplePost{},
			}))
		})

		It("panics when passing interface{} slices", func() {
			Expect(func() {
				api2go.Marshal([]interface{}{})
			}).To(Panic())
		})

		It("marshals to JSON", func() {
			json, err := api2go.MarshalToJSON([]SimplePost{firstPost})
			Expect(err).To(BeNil())
			Expect(json).To(Equal([]byte(`{"simple_posts":[{"Title":"First Post","Text":"Lipsum"}]}`)))
		})
	})
})
