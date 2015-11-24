package jsonapi

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StringHelpers", func() {
	Context("json funcs", func() {
		It("Pluralizes", func() {
			Expect(Pluralize("post")).To(Equal("posts"))
			Expect(Pluralize("posts")).To(Equal("posts"))
			Expect(Pluralize("category")).To(Equal("categories"))
		})
	})
})
