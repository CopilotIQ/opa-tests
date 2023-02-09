package internals_test

import (
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"time"
)

type Policy = map[string]interface{}
type PoliciesList struct {
	Result []Policy
}

var _ = Describe("OPA Server", func() {

	When("the tests start", func() {
		It("should be running", func() {
			Expect(OpaContainer.Address).To(HavePrefix("localhost:"))
		})
		It("is healthy", func() {
			Eventually(func() int {
				resp, err := http.Get(fmt.Sprintf("http://%s/health", OpaContainer.Address))
				if err != nil {
					return http.StatusInternalServerError
				}
				return resp.StatusCode
			}, 5*time.Second).Should(Equal(http.StatusOK))
		})
		It("has the correct bundle", func() {
			Eventually(func() bool {
				res, err := OpaContainer.GetEndpoint("/v1/policies")
				if err != nil {
					return false
				}
				var bundledPolicies PoliciesList
				decoder := json.NewDecoder(res.Body)
				defer res.Body.Close()
				if err = decoder.Decode(&bundledPolicies); err == nil {
					for _, policy := range bundledPolicies.Result {
						Expect(testPolicies).To(ContainElement(policy["id"]))
					}
					return len(bundledPolicies.Result) == len(testPolicies)
				}
				return false
			}, 10*time.Second).Should(BeTrue())
		})
	})
})
