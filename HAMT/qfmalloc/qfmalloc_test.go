package qfmalloc

import (
	"math/rand"
	"testing"
	"unsafe"
)

func TestAllocator(t *testing.T) {
	allocator := New(16, 32)
	var intps []unsafe.Pointer
	for i := 0; i < 100000; i++ {
		entryNum := int(rand.Int63n(32)) + 1
		ptr := allocator.Alloc(entryNum)
		//fmt.Printf(">> alloc %p\n", ptr)
		intps = append(intps, ptr)
	}

	for i := 0; i < len(intps); i++ {
		ptr := intps[i]
		//fmt.Printf(">> free %d %p\n", i, unsafe.Pointer(ptr))
		allocator.Free(unsafe.Pointer(ptr))
	}
}
