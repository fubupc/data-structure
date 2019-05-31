package vEB

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitConcat(t *testing.T) {
	scenarios := []struct {
		x    uint64
		bits uint8
		c    uint64
		i    uint64
	}{
		{x: 0, bits: 2, c: 0, i: 0},
		{x: 1, bits: 2, c: 0, i: 1},
		{x: 2, bits: 2, c: 1, i: 0},
		{x: 3, bits: 2, c: 1, i: 1},
		{x: 4, bits: 2, c: 2, i: 0}, // x out of bound can be represented by bits

		{x: 0, bits: 4, c: 0, i: 0},
		{x: 1, bits: 4, c: 0, i: 1},
		{x: 2, bits: 4, c: 0, i: 2},
		{x: 3, bits: 4, c: 0, i: 3},
		{x: 4, bits: 4, c: 1, i: 0},
		{x: 5, bits: 4, c: 1, i: 1},
		{x: 6, bits: 4, c: 1, i: 2},
		{x: 7, bits: 4, c: 1, i: 3},
		{x: 8, bits: 4, c: 2, i: 0}, // x out of bound can be represented by bits
	}

	for _, s := range scenarios {
		c, i := split(s.x, s.bits)
		assert.Equal(t, s.c, c, "input: x=%d bits=%d, result: c=%d i=%d", s.x, s.bits, c, i)
		assert.Equal(t, s.i, i, "input: x=%d bits=%d, result: c=%d i=%d", s.x, s.bits, c, i)
	}
}

func TestTree_Insert(t *testing.T) {
	tree := NewTree()

	for x := uint64(0); x < 10000; x++ {
		tree.Insert(x)
	}

	for x := uint64(0); x < 10000; x++ {
		assert.True(t, tree.Find(x))
	}

	for x := uint64(10000); x < 20000; x++ {
		assert.False(t, tree.Find(x))
	}
}

func TestTree_Successor(t *testing.T) {
	tree := NewTree()

	for x := uint64(0); x < 10000; x++ {
		tree.Insert(x)
	}

	for x := uint64(0); x < 9999; x++ {
		s, found := tree.Successor(x)
		assert.True(t, found)
		assert.Equal(t, x+1, s)
	}

	for x := uint64(9999); x < 20000; x++ {
		_, found := tree.Successor(x)
		assert.False(t, found)
	}
}
