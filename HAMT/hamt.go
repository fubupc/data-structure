package hamt

import (
	"fmt"
	"math/bits"
	"unsafe"
)

const (
	symbolWidth    = 5 // bits
	cardinality    = uint64(1) << symbolWidth
	hashSymbolMask = cardinality - 1
	maxHashBits    = 64
	basePtrMask    = 1 << 63
)

const (
	signBitMask = uint64(1) << 63
)

type Key int64
type Value int64

func (k Key) hash() uint64 {
	return uint64(k) ^ signBitMask
}

type Map struct {
	count     int
	root      *entry
	allocator *allocator
}

func NewMap() *Map {
	return &Map{count: 0, root: nil, allocator: nil}
}

func (m *Map) Count() int {
	return m.count
}

func (m *Map) Find(k *Key) *Value {
	if m.root == nil {
		return nil
	}

	curr := m.root
	hash := k.hash()
	shiftBits := uint(0)
	for {
		if curr.isLeaf() {
			kv := curr.asKVPair()
			if *(kv.key) == *k {
				return kv.val
			}
			return nil
		}

		if shiftBits >= maxHashBits {
			bucket := curr.asKVBucket()
			kv := bucket.find(k)
			if kv == nil {
				return nil
			}
			return kv.val
		}

		amt := curr.asAMTNode()
		symbol := hash & hashSymbolMask
		if !amt.contains(symbol) {
			return nil
		}
		curr = amt.base.entryAt(amt.indexFor(symbol))
		shiftBits += symbolWidth
		hash >>= symbolWidth
	}
}

func (m *Map) Add(k *Key, v *Value) {
	if m.root == nil {
		m.allocator = newAllocator()
		base := m.allocator.alloc(1)
		base.entryAt(0).asKVPair().set(k, v)
		m.root = base.ptr()
		m.count++
		return
	}

	curr := m.root
	hash := k.hash()
	shiftBits := uint(0)
	for {
		if curr.isLeaf() {
			old := curr.asKVPair()
			oldK, oldV := old.key, old.val

			// replace val if key already exists
			if *(oldK) == *k {
				old.setVal(v)
				return
			}

			m.count++

			oldHash := oldK.hash() >> shiftBits
			//fmt.Printf("  * old key=%d hash=%x\n", *oldK, oldHash)
			for {
				if shiftBits >= maxHashBits {
					m.to2KVBucket(curr, oldK, k, oldV, v)
					return
				}

				newSymbol := hash & hashSymbolMask
				oldSymbol := oldHash & hashSymbolMask

				//fmt.Printf("  * new=[%d %x] old=[%d %x]\n", newSymbol, hash, oldSymbol, oldHash)

				// no collision
				if newSymbol != oldSymbol {
					m.to2KVAMT(curr, oldSymbol, newSymbol, oldK, k, oldV, v)
					return
				}

				// collision happen
				curr = m.extendAMTChain(curr.asAMTNode(), newSymbol)
				shiftBits += symbolWidth
				hash >>= symbolWidth
				oldHash >>= symbolWidth
			}
		}

		// curr must be bucket if hash runs out
		if shiftBits >= maxHashBits {
			bucket := curr.asKVBucket()
			old := bucket.find(k)
			if old == nil {
				m.count++
				m.bucketAppendKV(curr.asKVBucket(), k, v)
			} else {
				old.val = v
			}
			return
		}

		// curr is an intermediate AMT node
		amt := curr.asAMTNode()
		symbol := hash & hashSymbolMask
		index := amt.indexFor(symbol)
		if !amt.contains(symbol) {
			m.count++
			m.amtAddKV(amt, symbol, index, k, v)
			return
		}
		curr = amt.base.entryAt(index)
		shiftBits += symbolWidth
		hash >>= symbolWidth
	}
}

func (m *Map) extendAMTChain(n *amtNode, symbol uint64) *entry {
	base := m.allocator.alloc(1)
	n.set(bitmap(0).set(symbol), base)
	return base.ptr()
}

func (m *Map) to2KVAMT(leaf *entry, symbol1, symbol2 uint64, k1, k2 *Key, v1, v2 *Value) {
	base := m.allocator.alloc(2)
	amt := leaf.asAMTNode()
	amt.set(bitmap(0).set(symbol1).set(symbol2), base)
	//fmt.Printf("  * chain amt: base=%x k1=%d v1=%d k2=%d v2=%d child=%d\n", base, *k1, *v1, *k2, *v2, amt.childNum())
	if symbol1 < symbol2 {
		base.entryAt(0).asKVPair().set(k1, v1)
		base.entryAt(1).asKVPair().set(k2, v2)
	} else {
		base.entryAt(0).asKVPair().set(k2, v2)
		base.entryAt(1).asKVPair().set(k1, v1)
	}
}

// chainBucketWith2KV convert entry to a kvBucket and append new key/val to bucket
func (m *Map) to2KVBucket(e *entry, k1, k2 *Key, v1, v2 *Value) {
	base := m.allocator.alloc(2)
	base.entryAt(0).asKVPair().set(k1, v1)
	base.entryAt(1).asKVPair().set(k2, v2)
	e.asKVBucket().set(2, base)
}

// bucketAppendKV reallocate bigger bucket to make room for new key/val
func (m *Map) bucketAppendKV(b *kvBucket, k *Key, v *Value) {
	oldBase := b.base
	newBase := m.allocator.alloc(int(b.count + 1))
	copyEntryList(newBase, oldBase, 0, 0, int(b.count))
	newBase.entryAt(int(b.count)).asKVPair().set(k, v)
	b.set(b.count+1, newBase)
	m.allocator.dealloc(oldBase)
}

// amtAddKV reallocate bigger sub-trie to make room for new k/v pair
func (m *Map) amtAddKV(n *amtNode, symbol uint64, index int, k *Key, v *Value) {
	oldBase := n.base
	newChildNum := n.childNum() + 1
	newBase := m.allocator.alloc(newChildNum)
	copyEntryList(newBase, oldBase, 0, 0, index)
	newBase.entryAt(index).asKVPair().set(k, v)
	copyEntryList(newBase, oldBase, index+1, index, newChildNum-index-1)
	n.set(n.bitmap.set(symbol), newBase)
	m.allocator.dealloc(oldBase)
}

func copyEntryList(dstBase, srcBase baseptr, dstStartIdx, srcStartIdx int, count int) {
	for i := 0; i < count; i++ {
		dst := dstBase.entryAt(dstStartIdx + i)
		src := srcBase.entryAt(srcStartIdx + i)
		*dst = *src
	}
}

// entry is a union type of `amtNode` and `kvPair` and `kvBucket`
// NOTE: type information is lost at runtime so we need some metadata to distinguish objects of the 3 types.
// `amtNode` is identified by setting most-significant-bit of `baseValPairs`.
// `kvBucket` can be identified by shift bits.
type entry struct {
	// mapKeyCnt stores `amtNode.bitmap` or `kvPair.key` or `kvBucket.count`
	mapKeyCnt uint64
	// baseVal stores `amtNode.base` or `kvPair.val` or `kvBucket.base`
	baseVal uint64
}

const (
	entrySize = unsafe.Sizeof(entry{})
)

// baseptr base pointer of entry list (with most-significant-bit set to 1)
type baseptr uintptr

func isBasePtr(x uint64) bool {
	return x&basePtrMask != 0
}

func toBasePtr(e *entry) baseptr {
	return baseptr(uintptr(unsafe.Pointer(e)) | basePtrMask)
}

// entryAt get entry address at specified index of list
// PS: entryAt(0) = ptr()
func (bp baseptr) entryAt(index int) *entry {
	return (*entry)(unsafe.Pointer(uintptr(bp^basePtrMask) + entrySize*uintptr(index)))
}

// ptr get real address of entry list
func (bp baseptr) ptr() *entry {
	return (*entry)(unsafe.Pointer(bp ^ basePtrMask))
}

// amtNode Array-Mapped-Trie node.
type amtNode struct {
	bitmap bitmap
	// base pointer of sub-trie
	base baseptr
}

// kvPair key/value pair
type kvPair struct {
	key *Key
	val *Value
}

// kvBucket store multi key/value pairs with conflict hash
type kvBucket struct {
	count int64
	base  baseptr // base pointer to kv-pair list  (with most-significant-bit as 1)
}

type bitmap uint64

func (m bitmap) countBelow(symbol uint64) int {
	return bits.OnesCount64(uint64(m) & (1<<symbol - 1))
}

func (m bitmap) isSet(symbol uint64) bool {
	return uint64(m)&(1<<symbol) != 0
}

func (m bitmap) set(symbol uint64) bitmap {
	return bitmap(uint64(m) | (1 << symbol))
}

func (m bitmap) reset() bitmap {
	return bitmap(0)
}

// amtNode cast entry to AMT
func (e *entry) asAMTNode() *amtNode {
	return (*amtNode)(unsafe.Pointer(e))
}

// kvPair cast entry to kvPair
func (e *entry) asKVPair() *kvPair {
	return (*kvPair)(unsafe.Pointer(e))
}

func (e *entry) copyFrom(e2 *entry) {
	e.mapKeyCnt = e2.mapKeyCnt
	e.baseVal = e2.baseVal
}

// kvBucket cast entry to kvBucket
func (e *entry) asKVBucket() *kvBucket {
	return (*kvBucket)(unsafe.Pointer(e))
}

// isLeaf check if underlying type of entry is leaf
func (e *entry) isLeaf() bool {
	return !isBasePtr(e.baseVal)
}

func (n *amtNode) set(m bitmap, base baseptr) {
	n.bitmap = m
	n.base = base
}

func (n *amtNode) childNum() int {
	return n.bitmap.countBelow(cardinality)
}

func (n *amtNode) indexFor(symbol uint64) int {
	return n.bitmap.countBelow(symbol)
}

func (n *amtNode) contains(symbol uint64) bool {
	return n.bitmap.isSet(symbol)
}

func (kv *kvPair) set(k *Key, v *Value) {
	kv.key = k
	kv.val = v
}

func (kv *kvPair) setVal(v *Value) {
	kv.val = v
}

func (b *kvBucket) set(count int64, base baseptr) {
	b.count = count
	b.base = base
}

// find linear search key in bucket
func (b *kvBucket) find(k *Key) *kvPair {
	base := b.base
	cnt := int(b.count)
	for i := 0; i < cnt; i++ {
		kv := base.entryAt(i).asKVPair()
		if *kv.key == *k {
			return kv
		}
	}
	return nil
}

type allocator struct {
	ptrs []unsafe.Pointer
}

func newAllocator() *allocator {
	return &allocator{}
}

func (alloc *allocator) alloc(size int) baseptr {
	var head *entry
	switch size {
	case 1:
		head = &new([1]entry)[0]
	case 2:
		head = &new([2]entry)[0]
	case 3:
		head = &new([3]entry)[0]
	case 4:
		head = &new([4]entry)[0]
	case 5:
		head = &new([5]entry)[0]
	case 6:
		head = &new([6]entry)[0]
	case 7:
		head = &new([7]entry)[0]
	case 8:
		head = &new([8]entry)[0]
	case 9:
		head = &new([9]entry)[0]
	case 10:
		head = &new([10]entry)[0]
	case 11:
		head = &new([11]entry)[0]
	case 12:
		head = &new([12]entry)[0]
	case 13:
		head = &new([13]entry)[0]
	case 14:
		head = &new([14]entry)[0]
	case 15:
		head = &new([15]entry)[0]
	case 16:
		head = &new([16]entry)[0]
	case 17:
		head = &new([17]entry)[0]
	case 18:
		head = &new([18]entry)[0]
	case 19:
		head = &new([19]entry)[0]
	case 20:
		head = &new([20]entry)[0]
	case 21:
		head = &new([21]entry)[0]
	case 22:
		head = &new([22]entry)[0]
	case 23:
		head = &new([23]entry)[0]
	case 24:
		head = &new([24]entry)[0]
	case 25:
		head = &new([25]entry)[0]
	case 26:
		head = &new([26]entry)[0]
	case 27:
		head = &new([27]entry)[0]
	case 28:
		head = &new([28]entry)[0]
	case 29:
		head = &new([29]entry)[0]
	case 30:
		head = &new([30]entry)[0]
	case 31:
		head = &new([31]entry)[0]
	case 32:
		head = &new([32]entry)[0]
	}
	alloc.ptrs = append(alloc.ptrs, unsafe.Pointer(head))
	return toBasePtr(head)
}

func (alloc *allocator) dealloc(bp baseptr) {
}

type debugItem struct {
	e     *entry
	depth int
}

func debugMap(m *Map) string {
	if m.root == nil {
		return "<nil>"
	}
	out := "=== hamt.Map ==="
	prevDepth := -1
	queue := []debugItem{{e: m.root, depth: 0}}
	for len(queue) > 0 {
		top, depth := queue[0].e, queue[0].depth
		queue = queue[1:]

		if prevDepth != depth {
			out += fmt.Sprintf("\n[L%d]", depth)
		}
		prevDepth = depth

		if top.isLeaf() {
			kv := top.asKVPair()
			out += fmt.Sprintf(" <leaf: %d %d>", *kv.key, *kv.val)
		} else if depth*symbolWidth >= maxHashBits {
			b := top.asKVBucket()
			out += fmt.Sprintf(" <bucket: %d %p>", b.count, b.base.ptr())
			for i := 0; i < int(b.count); i++ {
				child := b.base.entryAt(i)
				queue = append(queue, debugItem{e: child, depth: depth + 1})
			}
		} else {
			n := top.asAMTNode()
			out += fmt.Sprintf(" <amt: %d %p>", n.childNum(), n.base.ptr())
			for i := 0; i < n.childNum(); i++ {
				child := n.base.entryAt(i)
				queue = append(queue, debugItem{e: child, depth: depth + 1})
			}
		}
	}
	return out
}
