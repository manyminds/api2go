package jsonapi

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StringHelpers", func() {
	It("dejsonifies", func() {
		Expect(Dejsonify("Post")).To(Equal("Post"))
		Expect(Dejsonify("post")).To(Equal("Post"))
		Expect(Dejsonify("id")).To(Equal("ID"))
		Expect(Dejsonify("")).To(Equal(""))
	})

	It("jsonifies", func() {
		Expect(Jsonify("Post")).To(Equal("post"))
		Expect(Jsonify("post")).To(Equal("post"))
		Expect(Jsonify("ID")).To(Equal("id"))
		Expect(Jsonify("")).To(Equal(""))
	})

	It("Pluralizes", func() {
		Expect(Pluralize("post")).To(Equal("posts"))
		Expect(Pluralize("posts")).To(Equal("posts"))
		Expect(Pluralize("category")).To(Equal("categories"))
	})

	It("singularizes", func() {
		Expect(Singularize("posts")).To(Equal("post"))
		Expect(Singularize("post")).To(Equal("post"))
		Expect(Singularize("categories")).To(Equal("category"))
	})
})
