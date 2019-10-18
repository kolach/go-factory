package factory_test

import (
	randomdata "github.com/Pallinder/go-randomdata"
	. "github.com/kolach/go-factory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Person struct {
	Name   string  `json:"name"`
	Father *Person `json:"father,omitempty"`
	Mother *Person `json:"mother,omitempty"`
}

var _ = Describe("CrashTest", func() {

	var (
		f = NewFactory(
			Person{},
			Use(randomdata.FullName, randomdata.RandomGender).For("Name"),
			Use(func(ctx Ctx) (interface{}, error) {
				if ctx.Factory.CallDepth() > 4 {
					return nil, nil
				}
				return ctx.Factory.Create(Use(randomdata.FullName, randomdata.Male).For("Name"))
			}).For("Father"),
			Use(func(ctx Ctx) (interface{}, error) {
				if ctx.Factory.CallDepth() > randomdata.Number(4, 8) {
					return nil, nil
				}
				return ctx.Factory.Create(Use(randomdata.FullName, randomdata.Female).For("Name"))
			}).For("Mother"),
		)
	)

	It("should create family tree", func() {
		p := f.MustCreate().(*Person)
		Ω(p).ShouldNot(BeNil())
		// b, _ := json.Marshal(p)
		// var prettyJSON bytes.Buffer
		// json.Indent(&prettyJSON, b, "", "\t")
		// fmt.Println(string(prettyJSON.Bytes()))
	})
})
