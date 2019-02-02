package disjoint

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeSet(t *testing.T) {
	arr := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	set := NewDisjointSet()
	for k := range arr {
		set.MakeSet(k)
	}
	assert.Equal(t, len(arr), int(set.Size()))
	assert.Equal(t, len(arr), len(set.Parent))
	assert.Equal(t, len(arr), len(set.Rank))

	test := []string{"1f2pY5RYyRwSe7uKNUDaJVpXLKBhgjfd6", "1f2pY5RYyRwSe7uKNUDaJVpXLKBhgjfd6"}
	set2 := NewDisjointSet()
	for _, k := range test {
		set2.MakeSet(k)
	}
	assert.Equal(t, 1, set2.Size())
}

func TestFind(t *testing.T) {
	arr := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	set := NewDisjointSet()
	for _, k := range arr {
		set.MakeSet(k)
	}
	root, _ := set.Find(arr[1])
	assert.Equal(t, 1, int(root))
}

func TestUnion(t *testing.T) {
	arr := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	set := NewDisjointSet()
	for _, k := range arr {
		set.MakeSet(k)
	}
	root, _ := set.Union(arr[1], arr[2])
	assert.Equal(t, arr[2], int(root))
}
