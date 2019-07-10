package main

import (
	"github.com/mattermost/mattermost-server/model"
)

// GetAllUsers returns all users
func (p *Plugin) GetAllUsers() (map[string]*model.User, *model.AppError) {
	allUsers := make(map[string]*model.User)
	teamUsers, err := p.getTeamUsers()
	if err != nil {
		return nil, err
	}
	for _, users := range teamUsers {
		for _, user := range users {
			allUsers[user.Id] = user
		}
	}
	return allUsers, nil
}

// GetAllChannels returns all channels
func (p *Plugin) GetAllChannels() (map[string]*model.Channel, *model.AppError) {
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
	return allChannels, nil
}

// GetAllPublicChannelsForUser returns all public channels for user
func (p *Plugin) GetAllPublicChannelsForUser(userID string) (map[string]*model.Channel, *model.AppError) {
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
	return allPublicChannels, nil
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
