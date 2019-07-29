package memory_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	// "github.com/stretchr/testify/assert"
	"github.com/xn3cr0nx/bitgodine_code/internal/disjoint/memory"
)

var _ = Describe("Testing with Ginkgo", func() {
	var (
		arr []int
		set memory.DisjointSet
	)

	BeforeEach(func() {
		arr = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		set = memory.NewDisjointSet()
	})

	It("Should create a new set", func() {
		for k := range arr {
			set.MakeSet(k)
		}
		Expect(arr).To(HaveLen(int(set.Size())))
		Expect(set.Parent).To(HaveLen(len(arr)))
		Expect(set.Rank).To(HaveLen(len(arr)))
		test := []string{"1f2pY5RYyRwSe7uKNUDaJVpXLKBhgjfd6", "1f2pY5RYyRwSe7uKNUDaJVpXLKBhgjfd6"}
		set2 := memory.NewDisjointSet()
		for _, k := range test {
			set2.MakeSet(k)
		}
		Expect(set2.Size()).To(Equal(uint32(1)))
	})

	It("Should find the element of the array", func() {
		for _, k := range arr {
			set.MakeSet(k)
		}
		root, _ := set.Find(arr[1])
		Expect(int(root)).To(Equal(1))
	})

	It("Should union two elements of the array", func() {
		for _, k := range arr {
			set.MakeSet(k)
		}
		root, _ := set.Union(arr[1], arr[2])
		Expect(int(root)).To(Equal(arr[2]))
	})
})
