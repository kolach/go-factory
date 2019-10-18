package factory_test

import (
	"bytes"
	"encoding/json"
	"fmt"

	randomdata "github.com/Pallinder/go-randomdata"
	. "github.com/kolach/go-factory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Color int

const (
	Red   = 1
	Black = 0
)

type Node struct {
	Parent *Node `json:"-"` // parent is excluded to avoid recursive calls
	Left   *Node `json:"left"`
	Right  *Node `json:"right"`
	Color  int   `json:"color"`
	Key    int   `json:"key"`
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
		treeFact = NewFactory(
			Node{},
			Use(randomdata.Number, 1, 100).For("Key"),
			Use(func(ctx Ctx) (interface{}, error) {
				if ctx.CallDepth() >= randomdata.Number(2, 5) {
					return nil, nil
				}
				var child Node
				err := ctx.Factory.SetFields(&child)
				if err != nil {
					return nil, err
				}
				child.Parent = ctx.Instance.(*Node)
				return &child, nil
			}).For("Left", "Right"),
			Use(func(ctx Ctx) (interface{}, error) {
				node := ctx.Instance.(*Node)
				if node.Parent == nil {
					// root node is always black
					return Black, nil
				}
				if node.Left == nil || node.Right == nil {
					return Black, nil
				}
				return Red, nil
			}).For("Color"),
		)
	)

	It("should create red-black tree", func() {
		root := treeFact.MustCreate().(*Node)

		Ω(root).ShouldNot(BeNil())
		Ω(root.Parent).Should(BeNil())
		Ω(root.Color).Should(Equal(Black))

		fmt.Println(root)
	})
})
