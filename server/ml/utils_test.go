package ml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCosineSimilarity(t *testing.T) {
	epsilon := 0.0001
	assert := assert.New(t)
	t.Run("same vectors", func(t *testing.T) {
		a := []float64{1, 2, 3}
		b := []float64{1, 2, 3}
		assert.Equal(1.0, cosineSimilarity(a, b))
	})

	t.Run("scaled vectors", func(t *testing.T) {
		a := []float64{1, 2, 3}
		b := []float64{2, 4, 6}
		assert.Equal(1.0, cosineSimilarity(a, b))
	})

	t.Run("random vectors", func(t *testing.T) {
		a := []float64{1, 2, 3}
		b := []float64{2, 3, 4}
		assert.InDelta(0.99258, cosineSimilarity(a, b), epsilon)
	})
}

func TestIndexUsers(t *testing.T) {
	assert := assert.New(t)
	m := make(map[string]map[string]int64)
	m["user1"] = nil
	m["user2"] = nil
	indexedMap := indexUsers(m)
	assert.Equal(2, len(indexedMap))
}

func TestIndexChannels(t *testing.T) {
	assert := assert.New(t)
	m := make(map[string]map[string]int64)
	m["user1"] = make(map[string]int64)
	m["user1"]["chan1"] = 0
	m["user1"]["chan2"] = 0
	m["user1"]["chan3"] = 0
	m["user2"] = make(map[string]int64)
	m["user2"]["chan2"] = 0
	m["user2"]["chan3"] = 0
	m["user2"]["chan4"] = 0
	indexedMap := indexChannels(m)
	assert.Equal(4, len(indexedMap))
}
