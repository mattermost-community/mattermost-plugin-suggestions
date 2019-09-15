package ml

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetParams(t *testing.T) {
	assert := assert.New(t)
	t.Run("wrong types in params", func(t *testing.T) {
		params := make(map[string]interface{})
		params["similarity"] = 1
		params["k"] = "bla"
		knn := new(SimpleKNN)
		knn.SetParams(params)
		assert.Equal(defaultK, knn.k)
		f1 := reflect.ValueOf(cosineSimilarity)
		f2 := reflect.ValueOf(knn.similarity)
		assert.Equal(f1.Pointer(), f2.Pointer())
	})
	t.Run("nil params", func(t *testing.T) {
		knn := new(SimpleKNN)
		knn.SetParams(nil)
		assert.Equal(defaultK, knn.k)
		f1 := reflect.ValueOf(cosineSimilarity)
		f2 := reflect.ValueOf(knn.similarity)
		assert.Equal(f1.Pointer(), f2.Pointer())
	})

	t.Run("good params", func(t *testing.T) {
		params := make(map[string]interface{})
		params["similarity"] = (funcSimilarity)(cosineSimilarity)
		params["k"] = 17
		knn := new(SimpleKNN)
		knn.SetParams(params)
		assert.Equal(17, knn.k)
		f1 := reflect.ValueOf(cosineSimilarity)
		f2 := reflect.ValueOf(knn.similarity)
		assert.Equal(f1.Pointer(), f2.Pointer())
	})

}

func getUserChannelActivity() map[string]map[string]int64 {
	m := make(map[string]map[string]int64)
	m["user1"] = make(map[string]int64)
	m["user1"]["chan1"] = 1
	m["user1"]["chan2"] = 1
	m["user1"]["chan3"] = 1
	m["user2"] = make(map[string]int64)
	m["user2"]["chan2"] = 1
	m["user2"]["chan3"] = 1
	m["user2"]["chan4"] = 1
	return m
}

func TestComputeActivityMatrix(t *testing.T) {
	assert := assert.New(t)
	m := getUserChannelActivity()
	knn := new(SimpleKNN)
	knn.SetParams(make(map[string]interface{}))
	knn.computeActivityMatrix(m)
	assert.Equal(4, len(knn.activityMatrix))
	for i := 0; i < 4; i++ {
		assert.Equal(2, len(knn.activityMatrix[i]))
	}
	for i := 0; i < 2; i++ {
		zeros := 0
		ones := 0
		for j := 0; j < 4; j++ {
			if knn.activityMatrix[j][i] == 0 {
				zeros++
			} else if knn.activityMatrix[j][i] == 1 {
				ones++
			} else {
				assert.Fail("in activity matrix should be only 0s and 1s")
			}
		}
		assert.Equal(1, zeros)
		assert.Equal(3, ones)
	}
}

func TestComputeSimilarityMatrix(t *testing.T) {
	assert := assert.New(t)
	m := getUserChannelActivity()
	knn := new(SimpleKNN)
	knn.SetParams(make(map[string]interface{}))
	knn.computeActivityMatrix(m)
	knn.computeSimilarityMatrix()
	assert.Equal(4, len(knn.channelSimilarityMatrix))
	for i := 0; i < 4; i++ {
		assert.Equal(4, len(knn.channelSimilarityMatrix[i]))
		assert.Equal(1.0, knn.channelSimilarityMatrix[i][i])
	}
	epsilon := 0.0001

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			sim := knn.channelSimilarityMatrix[i][j]
			assert.True(sim == 0.0 || sim == 1.0 || sim-0.707106 < epsilon)
		}
	}
}

func TestFit(t *testing.T) {
	assert := assert.New(t)
	m := getUserChannelActivity()
	kn := NewSimpleKNN(nil)
	knn := kn.(*SimpleKNN)
	knn.Fit(m)

	assert.Equal(4, len(knn.activityMatrix))
	for i := 0; i < 4; i++ {
		assert.Equal(2, len(knn.activityMatrix[i]))
	}
	for i := 0; i < 2; i++ {
		zeros := 0
		ones := 0
		for j := 0; j < 4; j++ {
			if knn.activityMatrix[j][i] == 0 {
				zeros++
			} else if knn.activityMatrix[j][i] == 1 {
				ones++
			} else {
				assert.Fail("in activity matrix should be only 0s and 1s")
			}
		}
		assert.Equal(1, zeros)
		assert.Equal(3, ones)
	}

	assert.Equal(4, len(knn.channelSimilarityMatrix))
	for i := 0; i < 4; i++ {
		assert.Equal(4, len(knn.channelSimilarityMatrix[i]))
		assert.Equal(1.0, knn.channelSimilarityMatrix[i][i])
	}
	epsilon := 0.0001

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			sim := knn.channelSimilarityMatrix[i][j]
			assert.True(sim == 0.0 || sim == 1.0 || sim-0.707106 < epsilon)
		}
	}
}

func setMockSim(knn *SimpleKNN) {
	channelCount := 4
	knn.createSimilarityMatrix(channelCount)
	knn.channelSimilarityMatrix[0] = []float64{1, 0.5, 0.9, 0.1}
	knn.channelSimilarityMatrix[1] = []float64{0.5, 1, 0, 0.4}
	knn.channelSimilarityMatrix[2] = []float64{0.9, 0, 1, 0}
	knn.channelSimilarityMatrix[3] = []float64{0.1, 0.4, 0, 1}
}

func TestGetNeighbors(t *testing.T) {
	assert := assert.New(t)
	m := make(map[string]interface{})
	m["k"] = 2
	knn := new(SimpleKNN)
	knn.SetParams(m)
	setMockSim(knn)
	neighbors := knn.getNeighbors(1)
	correctNeighbors := []int{3, 0}
	assert.Equal(2, knn.k)
	assert.Equal(2, len(neighbors))
	for i := 0; i < knn.k; i++ {
		assert.Equal(correctNeighbors[i], neighbors[i])
	}
}

func getUserChannelActivity2() map[string]map[string]int64 {
	m := make(map[string]map[string]int64)
	m["user1"] = make(map[string]int64)
	m["user1"]["chan1"] = 1
	m["user1"]["chan2"] = 2
	m["user1"]["chan3"] = 3
	// m["user1"]["chan4"] = 1.76051
	m["user2"] = make(map[string]int64)
	// m["user2"]["chan1"] = 1.42595
	m["user2"]["chan2"] = 4
	m["user2"]["chan3"] = 2
	m["user2"]["chan4"] = 1
	m["user3"] = make(map[string]int64)
	m["user3"]["chan1"] = 3
	// m["user3"]["chan2"] = 4.10728
	m["user3"]["chan3"] = 2
	m["user3"]["chan4"] = 5
	return m
}

func TestPredict(t *testing.T) {
	assert := assert.New(t)
	m := getUserChannelActivity2()
	kn := NewSimpleKNN(nil)
	knn := kn.(*SimpleKNN)
	knn.k = 2
	knn.Fit(m)
	t.Run("userID error", func(t *testing.T) {
		_, err := knn.Predict("user7", "chan1")
		assert.NotNil(err)
	})

	t.Run("channelID error", func(t *testing.T) {
		_, err := knn.Predict("user1", "chan7")
		assert.NotNil(err)
	})

	epsilon := 0.001
	t.Run("no error", func(t *testing.T) {
		pred, err := knn.Predict("user1", "chan4")
		assert.Nil(err)
		assert.InDelta(1.76051, pred, epsilon)

		pred, err = knn.Predict("user2", "chan1")
		assert.Nil(err)
		assert.InDelta(1.42595, pred, epsilon)

		pred, err = knn.Predict("user3", "chan2")
		assert.Nil(err)
		assert.InDelta(2.56301, pred, epsilon)
	})
}
