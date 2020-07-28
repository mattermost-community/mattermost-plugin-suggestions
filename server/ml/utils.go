package ml

import "math"

type funcSimilarity func(a, b []float64) float64

// cosineSimilarity function
func cosineSimilarity(a, b []float64) float64 {
	ab := .0
	aa := .0
	bb := .0
	for i := 0; i < len(a); i++ {
		ab += a[i] * b[i]
		aa += a[i] * a[i]
		bb += b[i] * b[i]
	}
	return ab / math.Sqrt(aa*bb)
}

// indexUsers indexes all users in the userchannelActivity map
func indexUsers(userChannelActivity map[string]map[string]int64) map[string]int {
	dict := make(map[string]int)
	index := 0
	for k := range userChannelActivity {
		if _, ok := dict[k]; !ok {
			dict[k] = index
			index++
		}
	}
	return dict
}

// indexChannels indexes all channels in the userChannelActivity map
func indexChannels(userChannelActivity map[string]map[string]int64) map[string]int {
	dict := make(map[string]int)
	index := 0
	for _, v := range userChannelActivity {
		for channel := range v {
			if _, ok := dict[channel]; !ok {
				dict[channel] = index
				index++
			}
		}
	}
	return dict
}
