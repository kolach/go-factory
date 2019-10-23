package factory_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kolach/go-factory"
)

var _ = Describe("generators", func() {
	Describe("Seq", func() {
		It("should generate number in [0, max) interval", func() {
			seq := Seq(5)
			results := []int{}
			for i := 0; i < 7; i++ {
				results = append(results, seq())
			}
			Ω(results).To(HaveLen(7))
			Ω(results).To(Equal([]int{0, 1, 2, 3, 4, 0, 1}))
		})
	})
})
