package qfmalloc

import (
	"unsafe"
)

// Allocator quick-fit memory allocator for HAMT entry
// NOTE: it never release memory!
type Allocator struct {
	entrySize uintptr
	pool      *pool
	freelists []freelist
}

func New(entrySize uintptr, maxEntryNum int) *Allocator {
	return &Allocator{entrySize: entrySize, pool: new(pool), freelists: make([]freelist, maxEntryNum)}
}

func (a *Allocator) Alloc(entryNum int) unsafe.Pointer {
	fl := a.freelist(entryNum)
	if fl.empty() {
		return a.pool.alloc(a.entrySize, entryNum).payload()
	}
	return fl.remove().payload()
}

func (a *Allocator) Free(p unsafe.Pointer) {
	b := blockOfPayload(p)
	fl := a.freelist(b.entryNum)
	fl.add(b)
}

func (a *Allocator) freelist(entryNum int) *freelist {
	return &a.freelists[entryNum-1]
}

type blockptr uintptr

func (bp blockptr) ptr() *block {
	return (*block)(unsafe.Pointer(bp))
}

type block struct {
	entryNum int
	next     blockptr
}

const (
	blockPayloadOffset = unsafe.Offsetof(block{}.next)
)

func (b *block) payload() unsafe.Pointer {
	return unsafe.Pointer(uintptr(unsafe.Pointer(b)) + blockPayloadOffset)
}

func blockOfPayload(p unsafe.Pointer) *block {
	return (*block)(unsafe.Pointer(uintptr(p) - blockPayloadOffset))
}

const (
	pageSize          = 4096
	pagePayloadOffset = unsafe.Offsetof(struct {
		next    unsafe.Pointer
		payload int
	}{}.payload)
	pagePayloadSize = pageSize - pagePayloadOffset
)

type page struct {
	next     *page
	_payload [pagePayloadSize]byte
}

// pool memory pool for entry allocation
// TODO: this pool never release memory
type pool struct {
	currPage *page
	free     blockptr
}

// TODO: entries to be allocated cannot fit in a single page?
func (po *pool) alloc(entrySize uintptr, entryNum int) *block {
	if po.free == blockptr(0) || po.free.ptr().entryNum < entryNum {
		p, payload := po.newPage(entrySize)
		p.next = po.currPage
		po.currPage = p
		po.free = payload
	}

	currBlk := po.free
	//currNum := currBlk.ptr().entryNum
	currBlk.ptr().entryNum = entryNum

	nextBlk := blockptr(uintptr(currBlk) + blockPayloadOffset + entrySize*uintptr(entryNum))
	pageEnd := uintptr(unsafe.Pointer(po.currPage)) + pageSize
	nextNum := int(pageEnd-uintptr(nextBlk)-blockPayloadOffset) / int(entrySize)

	//fmt.Printf("  - alloc: page=%p need=%d have=%d remain=%d curr=%p next=%p\n", po.currPage, entryNum, currNum, nextNum, currBlk.ptr(), nextBlk.ptr())

	// allocate new page when no enough space for another block
	if nextNum <= 0 {
		p, payload := po.newPage(entrySize)
		//fmt.Printf("    - alloc new page: %p\n", p)
		p.next = po.currPage
		po.currPage = p
		po.free = payload
		return currBlk.ptr()
	}

	nextBlk.ptr().entryNum = nextNum
	po.free = nextBlk
	return currBlk.ptr()
}

func (po *pool) newPage(entrySize uintptr) (*page, blockptr) {
	p := new(page)
	payload := blockptr(unsafe.Pointer(&p._payload))
	payload.ptr().entryNum = int(pagePayloadSize / entrySize)
	return p, payload
}

type freelist struct {
	_head blockptr
}

func (fl *freelist) empty() bool {
	return fl._head == blockptr(0)
}

func (fl freelist) add(b *block) {
	b.next = fl._head
	fl._head = blockptr(unsafe.Pointer(b))
}

func (fl *freelist) remove() *block {
	b := fl._head.ptr()
	fl._head = b.next
	return b
}
