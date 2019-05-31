package vEB

import (
	"fmt"
)

type Tree struct {
	root  *node
	count int
}

func NewTree() *Tree {
	return &Tree{}
}

func (t *Tree) Find(x uint64) bool {
	return find(t.root, x, 64)
}

func (t *Tree) Insert(x uint64) {
	t.root = insert(t.root, x, 64)
}

func (t *Tree) Successor(x uint64) (uint64, bool) {
	return successor(t.root, x, 64)
}

type node struct {
	min      uint64 // NOT stored recursively
	max      uint64 // NOT stored recursively
	summary  *node
	clusters map[uint64]*node // TODO: use perfect hash
}

func newNode(x uint64) *node {
	return &node{min: x, max: x}
}

func find(n *node, x uint64, bits uint8) bool {
	if n == nil {
		return false
	}
	if x == n.min || x == n.max {
		return true
	}
	if x < n.min || x > n.max {
		return false
	}
	c, i := split(x, bits)
	return find(n.clusters[c], i, bits/2)
}

func successor(n *node, x uint64, bits uint8) (uint64, bool) {
	if n == nil {
		return 0, false
	}
	if x >= n.max {
		return 0, false
	}

	if x < n.min {
		return n.min, true
	}

	// successor must be found from now on
	c, i := split(x, bits)
	cluster := n.clusters[c]

	if cluster != nil && i < cluster.max {
		i, _ = successor(cluster, i, bits/2)
		return concat(c, i, bits), true
	}

	c, found := successor(n.summary, c, bits/2)
	if !found {
		return n.max, true
	}
	return concat(c, n.clusters[c].min, bits), true
}

func insert(n *node, x uint64, bits uint8) *node {
	if n == nil {
		return newNode(x)
	}
	if x < n.min {
		// swap because min is not stored recursively
		x, n.min = n.min, x
	}
	if x > n.max {
		// swap because min is not stored recursively
		x, n.max = n.max, x
	}
	if n.min == x || n.max == x {
		return n
	}

	// lazy allocation
	if n.clusters == nil {
		n.clusters = make(map[uint64]*node)
	}

	c, i := split(x, bits)
	cluster := n.clusters[c]
	if cluster == nil {
		n.summary = insert(n.summary, c, bits/2)
	}
	n.clusters[c] = insert(cluster, i, bits/2)
	return n
}

func split(x uint64, bits uint8) (uint64, uint64) {
	return x >> (bits / 2), x & (1<<(bits/2) - 1)
}

func concat(c, i uint64, bits uint8) uint64 {
	return c<<(bits/2) | i
}

func debugNode(n *node) string {
	if n == nil {
		return "<nil>"
	}

	var clusters []uint64
	for c := range n.clusters {
		clusters = append(clusters, c)
	}
	return fmt.Sprintf("(min=%d max=%d c=%v)", n.min, n.max, clusters)
}

func debugTree(t *Tree) string {
	out := "------- vEB Tree -------"

	if t.root == nil {
		out += "\n<nil>\n"
		return out
	}

	nodeQueue := []*node{t.root}
	depthQueue := []int{0}
	prevDepth := -1
	for len(nodeQueue) > 0 {
		node, depth := nodeQueue[0], depthQueue[0]
		nodeQueue, depthQueue = nodeQueue[1:], depthQueue[1:]

		if depth != prevDepth {
			out += fmt.Sprintf("\n%d:", depth)
		}
		prevDepth = depth
		out += " " + debugNode(node) + " "

		for c := range node.clusters {
			nodeQueue = append(nodeQueue, node.clusters[c])
			depthQueue = append(depthQueue, depth+1)
		}
	}
	return out
}
