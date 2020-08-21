package ml

import (
	"container/heap"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPQ(t *testing.T) {
	assert := assert.New(t)

	items := map[string]float64{
		"banana": 3, "apple": 5, "pear": 2, "orange": 4,
	}
	pq := NewMaxHeapK(5)
	heap.Init(pq)

	for value, priority := range items {
		heap.Push(pq, &Item{
			value:    value,
			priority: priority,
		})
	}

	correct := []string{"pear", "banana", "orange", "apple"}
	for i := 0; i < pq.Len(); i++ {
		item := heap.Pop(pq).(*Item).value
		assert.Equal(correct[i], item)
	}
}

func TestPQ2(t *testing.T) {
	assert := assert.New(t)
	chanVector := []float64{0.6902684899626333, 0.7592566023652966, 0.5707817929853929}

	pq := NewMaxHeapK(2)
	heap.Init(pq)

	for i := 0; i < len(chanVector); i++ {
		pq.Add(&Item{
			value:    i,
			priority: chanVector[i],
		})

	}

	correct := []int{0, 1}
	for i := 0; i < pq.Len(); i++ {
		item := heap.Pop(pq).(*Item).value
		assert.Equal(correct[i], item)
	}
}
