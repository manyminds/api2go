package api2go

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestApi2go(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Api2go Suite")
}
