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
	Ints  []int
	Map   map[int]string
	PStr  *string
	PStr2 *string
}

func genInts() []int {
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
		Use(genInts).For("Ints"),
		Use(genMap).For("Map"),
		Use(genPStr).For("PStr"),
	)
)

var _ = Describe("CrashTest", func() {
	It("should factory S", func() {
		s := f.MustCreate().(*S)
		Ω(s.Ints).Should(Equal(genInts()))
		Ω(s.Map).Should(Equal(genMap()))
		Ω(*s.PStr).Should(Equal(*genPStr()))
		Ω(*s.PStr2).Should(Equal(str))
	})
})
