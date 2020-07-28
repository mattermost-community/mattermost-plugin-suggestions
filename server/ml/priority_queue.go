package ml

import (
	"container/heap"
)

// An Item is something we manage in a max heap.
type Item struct {
	value    interface{}
	priority float64
}

// A MaxHeapK implements heap.Interface.
type MaxHeapK struct {
	elements []*Item
	k        int
}

// NewMaxHeapK returns a brand new max heap with capacity k.
func NewMaxHeapK(k int) *MaxHeapK {
	mh := new(MaxHeapK)
	mh.k = k
	mh.elements = make([]*Item, 0, k)
	return mh
}

// Len returns number of elements in a heap
func (pq MaxHeapK) Len() int { return len(pq.elements) }

// Less implements comparison function for a heap
func (pq MaxHeapK) Less(i, j int) bool {
	return pq.elements[i].priority < pq.elements[j].priority
}

// Swap swaps two elements of a heap
func (pq MaxHeapK) Swap(i, j int) {
	pq.elements[i], pq.elements[j] = pq.elements[j], pq.elements[i]
}

// Push method pushes an item into the PQ.
func (pq *MaxHeapK) Push(x interface{}) {
	pq.elements = append(pq.elements, x.(*Item))
}

// Add method adds element into the PQ.
func (pq *MaxHeapK) Add(x interface{}) {
	heap.Push(pq, x)
	if pq.Len() > pq.k {
		heap.Pop(pq)
	}
}

// Pop method pops item of hightest priority from the PQ.
func (pq *MaxHeapK) Pop() interface{} {
	old := pq.elements
	n := len(old)
	item := old[n-1]
	pq.elements = old[0 : n-1]
	return item
}
