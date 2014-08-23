package api2go

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StringHelpers", func() {
	It("underscorizes", func() {
		Expect(underscorize("Post")).To(Equal("post"))
		Expect(underscorize("post")).To(Equal("post"))
		Expect(underscorize("SimplePost")).To(Equal("simple_post"))
		Expect(underscorize("simple_post")).To(Equal("simple_post"))
		Expect(underscorize("SimplePostComment")).To(Equal("simple_post_comment"))
		Expect(underscorize("XML")).To(Equal("xml"))
		Expect(underscorize("XMLPost")).To(Equal("xml_post"))
		Expect(underscorize("XMLPostComment")).To(Equal("xml_post_comment"))
	})

	It("camelizes", func() {
		Expect(camelize("post")).To(Equal("Post"))
		Expect(camelize("post")).To(Equal("Post"))
		Expect(camelize("simple_post")).To(Equal("SimplePost"))
		Expect(camelize("simple_post")).To(Equal("SimplePost"))
		Expect(camelize("simple_post_comment")).To(Equal("SimplePostComment"))
		Expect(camelize("xml")).To(Equal("XML"))
		Expect(camelize("xml_post")).To(Equal("XMLPost"))
		Expect(camelize("xml_post_comment")).To(Equal("XMLPostComment"))
		Expect(camelize("id")).To(Equal("ID"))
		Expect(camelize("json")).To(Equal("JSON"))
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
