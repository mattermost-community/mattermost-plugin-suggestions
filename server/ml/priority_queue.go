package ml

import (
	"container/heap"
)

// An Item is something we manage in a max heap.
type Item struct {
	value    interface{}
	priority float64
}

// A MaxHeapK implements heap.Interface and holds k Items with highest priority
type MaxHeapK struct {
	elem []*Item
	k    int
}

// NewMaxHeapK returns a brand new max heap with capacity k
func NewMaxHeapK(k int) *MaxHeapK {
	mh := new(MaxHeapK)
	mh.k = k
	mh.elem = make([]*Item, 0, k)
	return mh
}

func (pq MaxHeapK) Len() int { return len(pq.elem) }

func (pq MaxHeapK) Less(i, j int) bool {
	return pq.elem[i].priority < pq.elem[j].priority
}

func (pq MaxHeapK) Swap(i, j int) {
	pq.elem[i], pq.elem[j] = pq.elem[j], pq.elem[i]
}

// Push method pushes an item into the PQ
func (pq *MaxHeapK) Push(x interface{}) {
	pq.elem = append(pq.elem, x.(*Item))
}

// Add method adds element into the PQ
func (pq *MaxHeapK) Add(x interface{}) {
	heap.Push(pq, x)
	if pq.Len() > pq.k {
		heap.Pop(pq)
	}
}

// Pop method pops item of hightest priority from the PQ
func (pq *MaxHeapK) Pop() interface{} {
	old := pq.elem
	n := len(old)
	item := old[n-1]
	pq.elem = old[0 : n-1]
	return item
}
