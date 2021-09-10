package common_test

import (
	"github.com/gp42/aws-auth-operator/controllers/common"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("common", func() {
	var (
		mainString      string
		similarString   string
		unorderedString string
		incorrectString string
	)

	BeforeEach(func() {
		mainString = "foo\nbar"
		similarString = "foo\nbar"
		unorderedString = "bar\nfoo"
		incorrectString = "foo\nbaz"
	})

	Describe("Sha512SumFromStringSorted", func() {
		Context("lines with similar order", func() {
			It("sums must be equal", func() {
				Expect(common.Sha512SumFromStringSorted(mainString)).To(Equal(common.Sha512SumFromStringSorted(similarString)))
			})
		})
		Context("lines with different order", func() {
			It("sums must be equal", func() {
				Expect(common.Sha512SumFromStringSorted(mainString)).To(Equal(common.Sha512SumFromStringSorted(unorderedString)))
			})
		})
		Context("different lines", func() {
			It("sums must not be equal", func() {
				Expect(common.Sha512SumFromStringSorted(mainString)).ToNot(Equal(common.Sha512SumFromStringSorted(incorrectString)))
			})
		})
	})
})
