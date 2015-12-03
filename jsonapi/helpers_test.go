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

		Context("Jsonify", func() {
			It("handles empty strings", func() {
				Expect(Jsonify("")).To(Equal(""))
			})

			It("uses common initialisms", func() {
				Expect(Jsonify("RAM")).To(Equal("ram"))
			})
		})
	})
})
