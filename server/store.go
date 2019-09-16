package main

const (
	timestampKey           = "timestamp"
	userChannelActivityKey = "userChannelActivity"
)

type recommendedChannel struct {
	ChannelID string  // identifier
	Score     float64 // score
}

// initStore method is for initializing the KVStore.
func (p *Plugin) initStore() error {
	err := p.saveTimestamp(-1)
	if err != nil {
		return err
	}
	return p.saveUserChannelActivity(make(userChannelActivity))
}

// saveUserRecommendations saves user recommendations in the KVStore.
func (p *Plugin) saveUserRecommendations(userID string, channels []*recommendedChannel) error {
	return p.Helpers.KVSetJSON(userID, channels)
}

// retreiveUserRecomendations gets user recommendations from the KVStore.
func (p *Plugin) retreiveUserRecomendations(userID string) ([]*recommendedChannel, error) {
	recommendations := make([]*recommendedChannel, 0)
	_, err := p.Helpers.KVGetJSON(userID, &recommendations)
	return recommendations, err
}

// saveTimestamp saves timestamp in the KVStore.
// All posts until this timestamp should already be analyzed.
func (p *Plugin) saveTimestamp(time int64) error {
	return p.Helpers.KVSetJSON(timestampKey, time)
}

// retreiveTimestamp gets timestamp from KVStore.
func (p *Plugin) retreiveTimestamp() (int64, error) {
	var time int64
	_, err := p.Helpers.KVGetJSON(timestampKey, &time)
	return time, err
}

// saveUserChannelActivity saves user-channel activity in the KVStore.
func (p *Plugin) saveUserChannelActivity(activity userChannelActivity) error {
	return p.Helpers.KVSetJSON(userChannelActivityKey, activity)
}

// retreiveUserChannelActivity gets user-channel activity from the KVStore.
func (p *Plugin) retreiveUserChannelActivity() (userChannelActivity, error) {
	var act userChannelActivity
	_, err := p.Helpers.KVGetJSON(userChannelActivityKey, &act)
	return act, err
}
