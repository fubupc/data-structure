package hamt

import (
	"unsafe"
)

type dummyAllocator struct {
	ptrs []unsafe.Pointer
}

func newAllocator() *dummyAllocator {
	return &dummyAllocator{}
}

// TODO: use freelist to manually manage memory
func (da *dummyAllocator) Alloc(entryNum int) unsafe.Pointer {
	block := make([]entry, entryNum)
	da.ptrs = append(da.ptrs, unsafe.Pointer(&block[0]))
	return unsafe.Pointer(&block[0])
}

func (da *dummyAllocator) Free(p unsafe.Pointer) {}
