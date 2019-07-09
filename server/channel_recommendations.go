package main

import (
	"fmt"
	"sort"

	"github.com/iomodo/mattermost-plugin-suggestions/server/ml"
	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
)

func mapToSlice(m map[string]*model.Channel) []*model.Channel {
	channels := make([]*model.Channel, len(m))
	index := 0
	for _, channel := range m {
		channels[index] = channel
		index++
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

func (p *Plugin) getChannelListFromRecommendations(recommendations []*recommendedChannel) []*model.Channel {
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})
	channels := make([]*model.Channel, 0)
	for _, rec := range recommendations {
		channel, err := p.API.GetChannel(rec.ChannelID)
		if err != nil {
			mlog.Error(fmt.Sprintf("Can't get channel - %v, err is %v", rec.ChannelID, err.Error()))
			continue
		}
		channels = append(channels, channel)
	}
	if len(channels) > 5 {
		channels = channels[:5]
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
				pred, err := knn.Predict(userID, channel.Id)

				if err != nil {
					// unknown user or unknown channel
					continue
				}
				if pred != 0 {
					recommendedChannels = append(recommendedChannels, &recommendedChannel{
						ChannelID: channel.Id,
						Score:     pred,
					})
				}
			}
		}
		p.saveUserRecommendations(userID, recommendedChannels)
	}

}
