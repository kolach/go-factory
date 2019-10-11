package factory_test

import (
	randomdata "github.com/Pallinder/go-randomdata"
	"github.com/kolach/go-factory"
	. "github.com/kolach/gomega-matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Person struct {
	FirstName string
	LastName  string
	Email     string
	Age       int
	Married   bool
}

var _ = Describe("Builder", func() {
	It("should build factory", func() {
		f := factory.NewBuilder(
			User{},
		).Use("John").For(
			"FirstName",
		).And("Smith", "Doe", "Milner").For(
			"LastName",
		).And("mail@hotmail.com").For(
			"Email",
		).And(randomdata.Number, 20, 50).For(
			"Age",
		).And(true, false).For(
			"Married",
		).Build()

		u := f.MustCreate().(*User)

		Ω(u.LastName).Should(BelongTo("Smith", "Doe", "Milner"))
		Ω(u.FirstName).Should(Equal("John"))
		Ω(u.Email).Should(Equal("mail@hotmail.com"))
		Ω(u.Age).Should(And(BeNumerically(">=", 20), BeNumerically("<", 50)))
		Ω(u.Married).Should(BelongTo(true, false))
	})
})
