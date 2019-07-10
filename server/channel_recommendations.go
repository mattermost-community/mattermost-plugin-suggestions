package main

import (
	"sort"

	"github.com/mattermost/mattermost-plugin-suggestions/server/ml"
	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
)

const numberOfRecommendedChannels = 5

func mapToSlice(m map[string]*model.Channel) []*model.Channel {
	channels := make([]*model.Channel, 0, len(m))
	for _, channel := range m {
		channels = append(channels, channel)
	}
	return channels
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (p *Plugin) getChannelListFromRecommendations(recommendations []*recommendedChannel) ([]*model.Channel, *model.AppError) {
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})
	n := min(numberOfRecommendedChannels, len(recommendations))
	channels := make([]*model.Channel, 0, n)
	for i := 0; i < n; i++ {
		channel, err := p.API.GetChannel(recommendations[i].ChannelID)
		if err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	return channels, nil
}

func (p *Plugin) preCalculateRecommendations() *model.AppError {
	mlog.Info("preCalculateRecommendations")

	userActivity, err := p.getActivity()
	if err != nil {
		return appError("Can't get user activity.", err)
	}
	params := map[string]interface{}{"k": 10}
	knn := ml.NewSimpleKNN(params)
	knn.Fit(userActivity)
	for userID := range userActivity {
		recommendedChannels := make([]*recommendedChannel, 0)
		channels, appErr := p.GetAllPublicChannelsForUser(userID)

		if appErr != nil {
			return appErr
		}
		for _, channel := range channels {
			if _, ok := userActivity[userID][channel.Id]; !ok {
				score, err := knn.Predict(userID, channel.Id)

				if err != nil {
					// unknown user or unknown channel
					continue
				}
				if score != 0 {
					recommendedChannels = append(recommendedChannels, &recommendedChannel{
						ChannelID: channel.Id,
						Score:     score,
					})
				}
			}
		}
		p.saveUserRecommendations(userID, recommendedChannels)
	}
	return nil
}
