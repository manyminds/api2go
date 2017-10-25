package routing_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRouting(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Routing Suite")
}
