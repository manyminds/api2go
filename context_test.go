package api2go

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Context", func() {
	Context("Set", func() {
		c := &APIContext{}

		It("sets key", func() {
			c.Set("test", 1)
			_, ok := c.keys["test"]
			Expect(ok).To(BeTrue())
		})
	})

	Context("Get", func() {
		c := &APIContext{}
		c.Set("test", 2)

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
		c := &APIContext{}
		c.Set("test", 3)

		It("reset removes keys", func() {
			c.Reset()
			_, ok := c.Get("test")
			Expect(ok).To(BeFalse())
		})
	})
})
