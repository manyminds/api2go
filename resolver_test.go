package api2go

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resolver test", func() {
	Context("basic function of callback resolver", func() {
		It("works", func() {
			callback := func(r http.Request) string {
				if r.Header.Get("lol") != "" {
					return "funny"
				}

				return "unfunny"
			}

			resolver := NewCallbackResolver(callback)
			Expect(resolver.GetBaseURL()).To(Equal("unfunny"))
			req, err := http.NewRequest("GET", "/v1/posts", nil)
			req.Header.Set("lol", "lol")
			Expect(err).To(BeNil())
			requestResolver, ok := resolver.(RequestAwareURLResolver)
			Expect(ok).To(Equal(true), "does not implement interface")
			Expect(requestResolver.GetBaseURL()).To(Equal("unfunny"))
			requestResolver.SetRequest(*req)
			Expect(requestResolver.GetBaseURL()).To(Equal("funny"))
		})
	})
})
