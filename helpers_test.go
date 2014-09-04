package api2go

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StringHelpers", func() {
	It("dejsonifies", func() {
		Expect(dejsonify("Post")).To(Equal("Post"))
		Expect(dejsonify("post")).To(Equal("Post"))
		Expect(dejsonify("id")).To(Equal("ID"))
		Expect(dejsonify("")).To(Equal(""))
	})

	It("jsonifies", func() {
		Expect(jsonify("Post")).To(Equal("post"))
		Expect(jsonify("post")).To(Equal("post"))
		Expect(jsonify("ID")).To(Equal("id"))
		Expect(jsonify("")).To(Equal(""))
	})

	It("pluralizes", func() {
		Expect(pluralize("post")).To(Equal("posts"))
		Expect(pluralize("posts")).To(Equal("posts"))
		Expect(pluralize("category")).To(Equal("categories"))
	})

	It("singularizes", func() {
		Expect(singularize("posts")).To(Equal("post"))
		Expect(singularize("post")).To(Equal("post"))
		Expect(singularize("categories")).To(Equal("category"))
	})
})
