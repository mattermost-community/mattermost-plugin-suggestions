package suggest

import "github.com/pkg/errors"

const (
	timestampKey     = "timestamp"
	userDataKeyAddon = "_data"
	userListKey      = "users"
)

// saveChannelRecommendations saves user recommendations in the KVStore.
func (s *ServiceImpl) saveChannelRecommendations(userID, teamID string, channels []*ChannelScore) error {
	key := getRecommendationKey(userID, teamID)
	_, err := s.pluginAPI.KV.Set(key, channels)
	return err
}

// retrieveChannelRecommendations gets user recommendations from the KVStore.
func (s *ServiceImpl) retrieveChannelRecommendations(userID, teamID string) ([]*ChannelScore, error) {
	key := getRecommendationKey(userID, teamID)
	recommendations := make([]*ChannelScore, 0)
	err := s.pluginAPI.KV.Get(key, &recommendations)
	if err != nil {
		return nil, errors.Wrap(err, "Can't retreive channel recommendations")
	}
	return recommendations, nil
}

func getRecommendationKey(userID, teamID string) string {
	return "s_" + userID + "_" + teamID[:10]
}
