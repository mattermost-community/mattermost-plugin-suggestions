package suggest

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
)

// GetAllPublicChannelsForUser returns all public channels for user
func (s *ServiceImpl) GetAllPublicChannelsForUser(userID, teamID string) ([]*model.Channel, error) {
	allPublicChannels := make(map[string]*model.Channel)
	teams, err := s.pluginAPI.Team.List(pluginapi.FilterTeamsByUser(userID))
	if err != nil {
		return nil, err
	}

	for _, team := range teams {
		perPage := 1000
		for page := 0; ; page++ {
			channelsForTeam, err := s.pluginAPI.Channel.ListPublicChannelsForTeam(team.Id, page, perPage)
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

func convertMapToSlice(channelsMap map[string]*model.Channel) []*model.Channel {
	channels := make([]*model.Channel, 0, len(channelsMap))
	for _, channel := range channelsMap {
		channels = append(channels, channel)
	}
	return channels
}
