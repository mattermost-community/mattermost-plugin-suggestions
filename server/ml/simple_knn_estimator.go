package ml

import (
	"container/heap"
	"errors"
)

const defaultK = 10

// SimpleKNN struct
type SimpleKNN struct {
	params                  map[string]interface{}
	channelSimilarityMatrix [][]float64
	activityMatrix          [][]float64
	userIndexes             map[string]int
	channelIndexes          map[string]int
	similarity              funcSimilarity
	k                       int
}

// BaseEstimator determines interface for all estimators for user-channel suggestions
type BaseEstimator interface {
	SetParams(params map[string]interface{})
	Predict(userID, channelID string) (float64, error)
	Fit(activities map[string]map[string]int64)
}

// NewSimpleKNN returns Simple KNN Estimator
func NewSimpleKNN(params map[string]interface{}) BaseEstimator {
	simpleKNN := new(SimpleKNN)
	simpleKNN.SetParams(params)
	return simpleKNN
}

// SetParams sets parameters for KNN estimator
func (knn *SimpleKNN) SetParams(params map[string]interface{}) {
	if val, exist := params["similarity"]; exist {
		switch val.(type) {
		case funcSimilarity:
			knn.similarity = val.(funcSimilarity)
		default:
			knn.similarity = cosineSimilarity
		}
	} else {
		knn.similarity = cosineSimilarity
	}
	if val, exist := params["k"]; exist {
		switch val.(type) {
		case int:
			knn.k = val.(int)
		default:
			knn.k = defaultK
		}
	} else {
		knn.k = defaultK
	}
}

func (knn *SimpleKNN) computeActivityMatrix(activities map[string]map[string]int64) {
	knn.userIndexes = indexUsers(activities)
	knn.channelIndexes = indexChannels(activities)
	knn.activityMatrix = make([][]float64, len(knn.channelIndexes))
	for i := 0; i < len(knn.channelIndexes); i++ {
		knn.activityMatrix[i] = make([]float64, len(knn.userIndexes))
	}

	for user, m := range activities {
		for channel, activity := range m {
			uIndex := knn.userIndexes[user]
			chIndex := knn.channelIndexes[channel]
			knn.activityMatrix[chIndex][uIndex] = float64(activity)
		}
	}

}

func (knn *SimpleKNN) createSimilarityMatrix(channelCount int) {
	knn.channelSimilarityMatrix = make([][]float64, channelCount)
	for i := 0; i < channelCount; i++ {
		knn.channelSimilarityMatrix[i] = make([]float64, channelCount)
	}
}

func (knn *SimpleKNN) computeSimilarityMatrix() {
	channelCount := len(knn.activityMatrix)
	knn.createSimilarityMatrix(channelCount)
	for i := 0; i < channelCount; i++ {
		for j := 0; j < channelCount; j++ {
			knn.channelSimilarityMatrix[i][j] = knn.similarity(knn.activityMatrix[i], knn.activityMatrix[j])
		}
	}
}

// Fit the KNN estimator
func (knn *SimpleKNN) Fit(activities map[string]map[string]int64) {
	knn.computeActivityMatrix(activities)
	knn.computeSimilarityMatrix()
}

// assumes len(channel) >= knn.k
func (knn *SimpleKNN) getNeighbors(channel int) []int {
	chanVector := knn.channelSimilarityMatrix[channel]

	pq := NewMaxHeapK(knn.k)
	for i := 0; i < len(chanVector); i++ {
		if i != channel {
			pq.Add(&Item{
				value:    i,
				priority: chanVector[i],
			})
		}
	}
	neighbors := make([]int, knn.k)
	for i := 0; i < knn.k; i++ {
		index := heap.Pop(pq).(*Item).value.(int)
		neighbors[i] = index
	}
	return neighbors
}

// Predict the activity of channel channelID for userID
func (knn *SimpleKNN) Predict(userID, channelID string) (float64, error) {
	channel, exists := knn.channelIndexes[channelID]
	if !exists {
		return 0, errors.New("unknown channelID: " + channelID)
	}
	user, exists := knn.userIndexes[userID]
	if !exists {
		return 0, errors.New("unknown userID" + userID)
	}

	if len(knn.channelSimilarityMatrix) < knn.k {
		return 0, nil // TODO
	}
	neighbors := knn.getNeighbors(channel)

	score := 0.0
	sum := 0.0
	for i := 0; i < len(neighbors); i++ {
		score += knn.channelSimilarityMatrix[channel][neighbors[i]] * knn.activityMatrix[neighbors[i]][user]
		sum += knn.channelSimilarityMatrix[channel][neighbors[i]]
	}
	if sum != 0 {
		score = score / sum
	}
	return score, nil
}
