package main

import (
	"time"

	"github.com/mattermost/mattermost-server/model"
)

// GetAllChannels returns all channels
func (p *Plugin) GetAllChannels() ([]*model.Channel, *model.AppError) {
	allChannels := make(map[string]*model.Channel)
	teamUsers, err := p.getTeamUsers()
	if err != nil {
		return nil, err
	}
	for team, users := range teamUsers {
		for _, user := range users {
			channels, err := p.API.GetChannelsForTeamForUser(team, user.Id, true)
			if err != nil {
				return nil, err
			}
			for _, channel := range channels {
				allChannels[channel.Id] = channel
			}
		}
	}
	return convertMapToSlice(allChannels), nil
}

// GetAllPublicChannelsForUser returns all public channels for user
func (p *Plugin) GetAllPublicChannelsForUser(userID string) ([]*model.Channel, *model.AppError) {
	allPublicChannels := make(map[string]*model.Channel)
	teams, err := p.API.GetTeamsForUser(userID)

	if err != nil {
		return nil, err
	}
	for _, team := range teams {
		perPage := 100
		for page := 0; ; page++ {
			channelsForTeam, err := p.API.GetPublicChannelsForTeam(team.Id, page, perPage)
			if err != nil {
				return nil, err
			}
			if len(channelsForTeam) == 0 {
				break
			}
			for _, channel := range channelsForTeam {
				allPublicChannels[channel.Id] = channel
			}
		}
	}
	return convertMapToSlice(allPublicChannels), nil
}

// getTeamUsers returns slice of users for every team
func (p *Plugin) getTeamUsers() (map[string][]*model.User, *model.AppError) {
	teamUsers := make(map[string][]*model.User)
	teams, err := p.API.GetTeams()
	if err != nil {
		return nil, err
	}
	for _, team := range teams {
		page := 0
		perPage := 100
		for {
			users, err := p.API.GetUsersInTeam(team.Id, page, perPage)
			if err != nil {
				return nil, err
			}
			if len(users) == 0 {
				break
			}
			teamUsers[team.Id] = append(teamUsers[team.Id], users...)
			page++
		}
	}
	return teamUsers, nil
}

// RunOnSingleNode will run function f on only a single node in a HA environment
// Method behaves asynchronously on the other nodes and will return without waiting f()
func (p *Plugin) RunOnSingleNode(f func()) error {
	runOnSingleNodeKey := "RunOnSingleNode"
	expire := 10 * 60 * 1000 // 10 minutes
	var savedTime int64
	ok, err := p.Helpers.KVGetJSON(runOnSingleNodeKey, &savedTime)
	if err != nil {
		return err
	}
	if ok {
		if time.Now().Unix()-savedTime < int64(expire) {
			return nil
		}
		p.Helpers.KVSetJSON(runOnSingleNodeKey, nil)
	}
	timeNow := time.Now().Unix()
	ok, err = p.Helpers.KVCompareAndSetJSON(runOnSingleNodeKey, nil, timeNow)
	if err != nil {
		return err
	}
	if ok { // ok will be true only for the single node
		f()
		p.Helpers.KVCompareAndSetJSON("RunOnSingleNode", timeNow, nil)
	}
	return nil
}

func convertMapToSlice(channelsMap map[string]*model.Channel) []*model.Channel {
	channels := make([]*model.Channel, 0, len(channelsMap))
	for _, channel := range channelsMap {
		channels = append(channels, channel)
	}
	return channels
}
