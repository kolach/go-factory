package factory_test

import (
	. "github.com/kolach/go-factory"
	. "github.com/kolach/gomega-matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	str = "foo"
)

type Color int

const (
	Black Color = iota
	White
	Red
)

type S struct {
	Slice []int
	Map   map[int]string
	PStr  *string
	PStr2 *string
	Color Color
}

func genSlice() []int {
	return []int{1, 2, 3}
}

func genMap() map[int]string {
	return map[int]string{
		1: "foo",
		2: "bar",
	}
}

func genPStr() *string {
	str := "foo"
	return &str
}

var _ = Describe("CrashTest", func() {
	var (
		f = NewFactory(
			S{PStr2: &str},
			Use(genSlice).For("Slice"),
			Use(genMap).For("Map"),
			Use(genPStr).For("PStr"),
			Use(Black, White).For("Color"),
		)
	)

	It("should create S", func() {
		i, err := f.Create()
		Ω(err).Should(BeNil())

		s, ok := i.(*S)
		Ω(ok).Should(BeTrue())
		Ω(s.Slice).Should(Equal(genSlice()))
		Ω(s.Map).Should(Equal(genMap()))
		Ω(*s.PStr).Should(Equal(*genPStr()))
		Ω(*s.PStr2).Should(Equal(str))
		Ω(s.Color).Should(BelongTo(White, Black))
	})

	Context("enums", func() {
		var s S

		It("should work with custom generator functions", func() {
			err := f.SetFields(&s, Use(func() Color { return Red }).For("Color"))
			Ω(err).Should(BeNil())
			Ω(s.Color).Should(Equal(Red))
		})

		It("should work with canonical generator functions", func() {
			err := f.SetFields(&s, Use(func(ctx Ctx) (interface{}, error) { return Red, nil }).For("Color"))
			Ω(err).Should(BeNil())
			Ω(s.Color).Should(Equal(Red))
		})
	})

	It("should set nil to pointer fields", func() {
		s := S{Map: map[int]string{1: "foo"}}
		err := f.SetFields(
			&s,
			Use(nil).For("PStr"),
			Use(nil).For("Slice"),
			Use(nil).For("Map"), // check Map is reset
		)
		Ω(err).Should(BeNil())
		Ω(s.PStr).Should(BeNil())
		Ω(s.Slice).Should(BeNil())
		Ω(s.Map).Should(BeNil())
	})
})
