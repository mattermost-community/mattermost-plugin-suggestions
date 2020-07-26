package command

import "github.com/mattermost/mattermost-server/v5/model"

const noNewChannelsText = "No new channels for you."

func createChannelsAutocompleteData() *model.AutocompleteData {
	channels := model.NewAutocompleteData("channels", "", "Get relevant channel suggestions for you")
	return channels
}

func (c *Command) suggestChannelResponse() {
	channels, appErr := c.suggester.GetChannelRecommendations(c.args.UserId, c.args.TeamId)
	if appErr != nil {
		c.postCommandResponse("can't get recommendations")
	}
	if len(channels) == 0 {
		c.postCommandResponse(noNewChannelsText)
		return
	}
	text := "Channels we recommend\n"
	for _, channel := range channels {
		text += " * ~" + channel.Name
		if channel.Purpose != "" {
			text += " - " + channel.Purpose
		}
		text += "\n"
	}
	c.postCommandResponse(text)
}
