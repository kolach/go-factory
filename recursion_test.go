package factory_test

import (
	"bytes"
	"encoding/json"
	"runtime"
	"sync"

	randomdata "github.com/Pallinder/go-randomdata"
	. "github.com/kolach/go-factory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Node struct {
	Parent   *Node   `json:"-"` // parent is excluded to avoid recursive calls
	Children []*Node `json:"children"`
	Name     string  `json:"name"`
}

// String for debug output
func (n *Node) String() string {
	b, _ := json.Marshal(n)
	var buf bytes.Buffer
	json.Indent(&buf, b, "", "\t")
	return string(buf.Bytes())
}

var _ = Describe("RecursionTest", func() {
	var (
		factory = NewFactory(
			Node{},
			Use(randomdata.FirstName, randomdata.RandomGender).For("Name"),
			Use(func(ctx Ctx) (interface{}, error) {
				self := ctx.Factory

				if self.CallDepth() > randomdata.Number(2, 4) {
					// exit recursion if factory call depth is greater than [2, 4)
					return nil, nil
				}

				node := ctx.Instance.(*Node)    // current node that's being created
				size := randomdata.Number(1, 5) // number of children to make
				kids := make([]*Node, size)     // slice to store children nodes

				for i := 0; i < size; i++ {
					kids[i] = &Node{Parent: node}
					if err := self.SetFields(kids[i]); err != nil {
						return nil, err
					}
				}
				return kids, nil
			}).For("Children"),
		)
	)

	It("should create valid tree", func() {
		root := factory.MustCreate().(*Node)

		Ω(root).ShouldNot(BeNil())
		Ω(root.Parent).Should(BeNil())
		Ω(len(root.Children)).Should(BeNumerically(">", 0))

		child0 := root.Children[0]

		Ω(child0.Parent).Should(Equal(root))
		Ω(len(child0.Children)).Should(BeNumerically(">", 0))
		Ω(child0.Children[0].Parent).Should(Equal(child0))

		// fmt.Println(root)
	})

	It("should increment call depth on each recursive call", func() {
		callDepths := []int{}
		factory.MustCreate(
			Use(func(ctx Ctx) (interface{}, error) {
				self := ctx.Factory
				callDepths = append(callDepths, self.CallDepth())
				if self.CallDepth() > 4 {
					return nil, nil
				}
				kids := []*Node{&Node{}}
				if err := self.SetFields(kids[0]); err != nil {
					return nil, err
				}
				return kids, nil
			}).For("Children"),
		)

		Ω(callDepths).Should(Equal([]int{1, 2, 3, 4, 5}))
	})

	It("should be OK to use factory concurrently", func() {
		numCPU := runtime.NumCPU()
		if numCPU == 1 {
			Skip("Num CPU is 0, skipping")
		}
		var wg sync.WaitGroup
		for i := 0; i < numCPU; i++ {
			wg.Add(1)
			go func() {
				for i := 0; i < 10; i++ {
					var node Node
					err := factory.SetFields(&node)
					Ω(err).Should(BeNil())
				}
				wg.Done()
			}()
		}
		wg.Wait()
	})
})
