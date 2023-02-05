package internals_test

import (
	"context"
	"github.com/CopilotIQ/opa-tests/gentests/internals"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	testBundle = "../../testdata/authz-0.6.26.tar.gz"
)

var (
	// A constant by all means
	testPolicies = []string{
		"copilotiq/tokens.rego",
		"copilotiq/common.rego",
		"copilotiq/users.rego",
	}
)

func TestInternals(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Internals Suite")
}

var OpaContainer *internals.OpaServer

var _ = BeforeSuite(func() {
	var err error
	OpaContainer, err = internals.NewOpaContainer(context.Background(), testBundle)
	Expect(err).ShouldNot(HaveOccurred())
})
