package factory_test

import (
	. "github.com/kolach/go-factory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	str = "foo"
)

type S struct {
	Slice []int
	Map   map[int]string
	PStr  *string
	PStr2 *string
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

var (
	f = NewFactory(
		S{PStr2: &str},
		Use(genSlice).For("Slice"),
		Use(genMap).For("Map"),
		Use(genPStr).For("PStr"),
	)
)

var _ = Describe("CrashTest", func() {
	It("should factory S", func() {
		s := f.MustCreate().(*S)
		Ω(s.Slice).Should(Equal(genSlice()))
		Ω(s.Map).Should(Equal(genMap()))
		Ω(*s.PStr).Should(Equal(*genPStr()))
		Ω(*s.PStr2).Should(Equal(str))
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
