package main

import (
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

const (
	trigger                = "suggest"
	channelAction          = "channels"
	addRandomChannelAction = "add"
	resetAction            = "reset"
	computeAction          = "compute"

	displayName          = "Suggestions"
	desc                 = "Mattermost Suggestions Plugin"
	noNewChannelsText    = "No new channels for you."
	addRandomChannelText = "Channel was successfully added."
	resetText            = "Recommendations were cleared."
	computeText          = "Recomendations were computed."
)

const commandHelp = `
* |/suggest channels| - Suggests relevant channels for the user
* |/suggest reset| - Resets suggestions. For testing only.
* |/suggest compute| - Computes suggestions. For testing only.
`

func getCommand() *model.Command {
	return &model.Command{
		Trigger:          trigger,
		DisplayName:      displayName,
		Description:      desc,
		AutoComplete:     true,
		AutoCompleteDesc: "Available commands: channels, help",
		AutoCompleteHint: "[command]",
	}
}

func (p *Plugin) postCommandResponse(args *model.CommandArgs, text string) {
	post := &model.Post{
		UserId:    p.botUserID,
		ChannelId: args.ChannelId,
		Message:   text,
	}
	_ = p.API.SendEphemeralPost(args.UserId, post)
}

func (p *Plugin) helpResponse(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	text := "###### " + desc + " - Slash Command Help\n" + strings.Replace(commandHelp, "|", "`", -1)
	p.postCommandResponse(args, text)
	return &model.CommandResponse{}, nil
}

func appError(message string, err error) *model.AppError {
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}
	return model.NewAppError("Suggestions Plugin", message, nil, errorMessage, http.StatusBadRequest)
}

func (p *Plugin) suggestChannelResponse(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	recommendations, err := p.retreiveUserRecomendations(args.UserId)
	if err != nil {
		return nil, appError("Can't retreive user recommendations.", err)
	}
	channels, appErr := p.getChannelListFromRecommendations(recommendations)
	if appErr != nil {
		return nil, appErr
	}
	if len(channels) == 0 {
		p.postCommandResponse(args, noNewChannelsText)
		return &model.CommandResponse{}, nil
	}
	text := "Channels we recommend\n"
	for _, channel := range channels {
		text += " * ~" + channel.Name + " - " + channel.Purpose + "\n"
	}
	p.postCommandResponse(args, text)
	return &model.CommandResponse{}, nil
}

func (p *Plugin) reset(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	p.saveUserRecommendations(args.UserId, make([]*recommendedChannel, 0))
	p.saveUserChannelActivity(make(userChannelActivity))
	p.saveTimestamp(-1)
	p.postCommandResponse(args, resetText)
	return &model.CommandResponse{}, nil
}

func (p *Plugin) compute(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	err := p.preCalculateRecommendations()
	if err != nil {
		return nil, err
	}
	p.postCommandResponse(args, computeText)
	return &model.CommandResponse{}, nil
}

// ExecuteCommand executes a command that has been previously registered via the RegisterCommand API.
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	split := strings.Fields(args.Command)
	if len(split) == 0 {
		return nil, nil
	}
	command := split[0]
	action := ""
	if len(split) > 1 {
		action = split[1]
	}
	if command != "/"+trigger {
		return &model.CommandResponse{}, nil
	}
	switch action {
	case "":
		return p.helpResponse(args)
	case "help":
		return p.helpResponse(args)
	case channelAction:
		return p.suggestChannelResponse(args)
	case resetAction:
		return p.reset(args)
	case computeAction:
		return p.compute(args)
	}

	return &model.CommandResponse{}, nil
}
