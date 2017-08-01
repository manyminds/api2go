package api2go

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Context", func() {
	var c *APIContext

	BeforeEach(func() {
		c = &APIContext{}
	})

	Context("Set", func() {
		It("sets key", func() {
			c.Set("test", 1)
			_, ok := c.keys["test"]
			Expect(ok).To(BeTrue())
		})
	})

	Context("Get", func() {
		BeforeEach(func() {
			c.Set("test", 2)
		})

		It("gets key", func() {
			key, ok := c.Get("test")
			Expect(ok).To(BeTrue())
			Expect(key.(int)).To(Equal(2))
		})

		It("not okay if key does not exist", func() {
			key, ok := c.Get("nope")
			Expect(ok).To(BeFalse())
			Expect(key).To(BeNil())
		})
	})

	Context("Reset", func() {
		BeforeEach(func() {
			c.Set("test", 3)
		})

		It("reset removes keys", func() {
			c.Reset()
			_, ok := c.Get("test")
			Expect(ok).To(BeFalse())
		})
	})

	Context("Not yet implemented", func() {
		It("Deadline", func() {
			deadline, ok := c.Deadline()
			Expect(deadline).To(Equal(time.Time{}))
			Expect(ok).To(Equal(false))
		})

		It("Done", func() {
			var chanel <-chan struct{}
			Expect(c.Done()).To(Equal(chanel))
		})

		It("Err", func() {
			Expect(c.Err()).To(BeNil())
		})
	})

	Context("Value", func() {
		It("Value returns a set value", func() {
			c.Set("foo", "bar")
			Expect(c.Value("foo")).To(Equal("bar"))
		})

		It("Returns nil if key was not a string", func() {
			Expect(c.Value(1337)).To(BeNil())
		})

	})

	Context("ContextQueryParams", func() {
		It("returns them if set", func() {
			queryParams := map[string][]string{
				"foo": {"bar"},
			}

			c.Set("QueryParams", queryParams)
			Expect(ContextQueryParams(c)).To(Equal(queryParams))
		})

		It("sets empty ones if not set", func() {
			Expect(ContextQueryParams(c)).To(Equal(map[string][]string{}))
		})
	})
})
