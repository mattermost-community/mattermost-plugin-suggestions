package main

const (
	timestampKey           = "timestamp"
	userChannelActivityKey = "userChannelActivity"
	userDataKeyAddon       = "_data"
	userListKey            = "users"
)

type recommendedChannel struct {
	ChannelID string  // identifier
	Score     float64 // score
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
	exists, err := p.Helpers.KVGetJSON(timestampKey, &time)
	if !exists {
		time = -1
	}
	return time, err
}

// saveUserChannelActivity saves every users' channel activity in the KVStore.
func (p *Plugin) saveUserChannelActivity(activity userChannelActivity) error {
	users := make([]string, 0, len(activity))
	for user, channels := range activity {
		if err := p.Helpers.KVSetJSON(getUserDataKey(user), channels); err != nil {
			return err
		}
		users = append(users, user)
	}
	return p.Helpers.KVSetJSON(userListKey, users)
}

// retreiveUserChannelActivity gets every users' channel activity from the KVStore.
func (p *Plugin) retreiveUserChannelActivity() (userChannelActivity, error) {
	var users []string
	if _, err := p.Helpers.KVGetJSON(userListKey, &users); err != nil {
		return nil, err
	}

	act := make(userChannelActivity)
	for _, user := range users {
		var m map[string]int64
		if _, err := p.Helpers.KVGetJSON(getUserDataKey(user), &m); err != nil {
			return nil, err
		}
		act[user] = make(map[string]int64)
		for k, v := range m {
			act[user][k] = v
		}
	}

	return act, nil
}

func getUserDataKey(user string) string {
	return user + userDataKeyAddon
}
