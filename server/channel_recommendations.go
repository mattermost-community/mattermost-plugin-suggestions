package main

import (
	"fmt"
	"sort"

	"github.com/iomodo/mattermost-plugin-suggestions-1/server/ml"
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

func (p *Plugin) isChannelOk(channelID string) bool {
	posts, err := p.API.GetPostsForChannel(channelID, 0, 1)

	if err != nil || len(posts.Order) == 0 {
		return false
	}
	return true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (p *Plugin) getChannelListFromRecommendations(recommendations []*recommendedChannel) []*model.Channel {
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})
	n := min(numberOfRecommendedChannels, len(recommendations))
	channels := make([]*model.Channel, 0, n)
	for i := 0; i < n; i++ {
		channel, err := p.API.GetChannel(recommendations[i].ChannelID)
		if err != nil {
			mlog.Error(fmt.Sprintf("Can't get channel - %v, err is %v", recommendations[i].ChannelID, err.Error()))
			continue
		}
		channels = append(channels, channel)
	}
	return channels
}

func (p *Plugin) preCalculateRecommendations() {
	mlog.Info("preCalculateRecommendations")

	userActivity, err := p.getActivity()
	if err != nil {
		mlog.Error("Can't get user activity. " + err.Error())
		return
	}
	params := map[string]interface{}{"k": 10}
	knn := ml.NewSimpleKNN(params)
	knn.Fit(userActivity)
	for userID := range userActivity {
		recommendedChannels := make([]*recommendedChannel, 0)
		channels, appErr := p.GetAllPublicChannelsForUser(userID)

		if appErr != nil {
			mlog.Error("Can't get public channels for user. " + appErr.Error())
			return
		}
		for _, channel := range channels {
			if _, ok := userActivity[userID][channel.Id]; !ok {
				if !p.isChannelOk(channel.Id) {
					continue
				}
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

}
