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
		Expect(underscorize("SimplePostComment")).To(Equal("simple_post_comment"))
		Expect(underscorize("simple_post")).To(Equal("simple_post"))
		Expect(underscorize("XML")).To(Equal("xml"))
		Expect(underscorize("XMLPost")).To(Equal("xml_post"))
		Expect(underscorize("XMLPostComment")).To(Equal("xml_post_comment"))
	})

	It("pluralizes", func() {
		Expect(pluralize("post")).To(Equal("posts"))
		Expect(pluralize("posts")).To(Equal("posts"))
		Expect(pluralize("category")).To(Equal("categories"))
	})
})
