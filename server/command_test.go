package main

import (
	"strings"
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getPostPlugin(userID, channelID, text string) (*Plugin, *plugintest.API) {
	api := &plugintest.API{}
	plugin := &Plugin{}
	plugin.botUserID = "test"
	post := &model.Post{
		UserId:    plugin.botUserID,
		ChannelId: channelID,
		Message:   text,
	}
	api.On("SendEphemeralPost", userID, post).Return(nil)
	plugin.SetAPI(api)
	return plugin, api
}

func TestGetCommand(t *testing.T) {
	command := getCommand()
	assert.Equal(t, trigger, command.Trigger)
	assert.Equal(t, desc, command.Description)
	assert.True(t, command.AutoComplete)
	assert.Equal(t, displayName, command.DisplayName)
}

func TestPostCommandResponse(t *testing.T) {
	api := &plugintest.API{}
	plugin := &Plugin{}
	args := &model.CommandArgs{UserId: "user"}
	text := "text"
	api.On("SendEphemeralPost", args.UserId, mock.Anything).Return(nil)
	defer api.AssertExpectations(t)
	plugin.SetAPI(api)
	plugin.postCommandResponse(args, text)
}

func TestExecuteCommandTrivial(t *testing.T) {
	t.Run("empty command", func(t *testing.T) {
		plugin := Plugin{}
		args := &model.CommandArgs{
			Command: "",
		}
		resp, err := plugin.ExecuteCommand(nil, args)
		assert.Nil(t, err)
		assert.Nil(t, resp)
	})

	t.Run("random command", func(t *testing.T) {
		plugin := Plugin{}
		args := &model.CommandArgs{
			Command: "random",
		}
		resp, err := plugin.ExecuteCommand(nil, args)
		assert.Nil(t, err)
		assert.Equal(t, &model.CommandResponse{}, resp)
	})

	t.Run("suggest command, random action", func(t *testing.T) {
		plugin := Plugin{}
		args := &model.CommandArgs{
			Command: "/suggest random",
		}
		resp, err := plugin.ExecuteCommand(nil, args)
		assert.Nil(t, err)
		assert.Equal(t, &model.CommandResponse{}, resp)
	})

	t.Run("suggest command, empty action", func(t *testing.T) {
		user := "user"
		channel := "channel"
		text := "###### " + desc + " - Slash Command Help\n" + strings.Replace(commandHelp, "|", "`", -1)
		plugin, api := getPostPlugin(user, channel, text)
		defer api.AssertExpectations(t)
		args := &model.CommandArgs{
			UserId:    user,
			ChannelId: channel,
			Command:   "/suggest",
		}
		resp, err := plugin.ExecuteCommand(nil, args)
		assert.Nil(t, err)
		assert.Equal(t, &model.CommandResponse{}, resp)
	})

	t.Run("suggest command, help action", func(t *testing.T) {
		user := "user"
		channel := "channel"
		text := "###### " + desc + " - Slash Command Help\n" + strings.Replace(commandHelp, "|", "`", -1)
		plugin, api := getPostPlugin(user, channel, text)
		defer api.AssertExpectations(t)
		args := &model.CommandArgs{
			UserId:    user,
			ChannelId: channel,
			Command:   "/suggest help",
		}
		resp, err := plugin.ExecuteCommand(nil, args)
		assert.Nil(t, err)
		assert.Equal(t, &model.CommandResponse{}, resp)
	})
}

func TestExecuteCommandSuggestChannels(t *testing.T) {
	user := "user"
	channel := "channel"
	args := &model.CommandArgs{
		UserId:    user,
		ChannelId: channel,
		Command:   "/suggest channels",
	}
	t.Run("zero channels", func(t *testing.T) {
		text := noNewChannelsText
		plugin, api := getPostPlugin(user, channel, text)
		defer api.AssertExpectations(t)
		helpers := &plugintest.Helpers{}
		plugin.SetHelpers(helpers)
		helpers.On("KVGetJSON", mock.Anything, mock.Anything).Return(true, nil)
		defer helpers.AssertExpectations(t)
		resp, err := plugin.ExecuteCommand(nil, args)
		assert.Nil(t, err)
		assert.Equal(t, &model.CommandResponse{}, resp)
	})
	t.Run("GetChannel channel error", func(t *testing.T) {
		plugin := &Plugin{}
		api := &plugintest.API{}
		helpers := &plugintest.Helpers{}
		plugin.SetHelpers(helpers)
		helpers.On("KVGetJSON", mock.Anything, mock.Anything).Return(true, nil).Run(func(args mock.Arguments) {
			arg := args.Get(1).(*[]*recommendedChannel)
			*arg = make([]*recommendedChannel, 1)
			(*arg)[0] = &recommendedChannel{ChannelID: "chan", Score: 0.1}
		})
		defer helpers.AssertExpectations(t)
		api.On("GetChannel", mock.Anything).Return(nil, model.NewAppError("", "", nil, "", 404))
		plugin.SetAPI(api)
		defer api.AssertExpectations(t)
		_, err := plugin.ExecuteCommand(nil, args)
		assert.NotNil(t, err)
	})

	t.Run("retreive user recommendations error", func(t *testing.T) {
		plugin := &Plugin{}
		helpers := &plugintest.Helpers{}
		plugin.SetHelpers(helpers)
		helpers.On("KVGetJSON", mock.Anything, mock.Anything).Return(false, model.NewAppError("", "", nil, "", 404))
		defer helpers.AssertExpectations(t)
		_, err := plugin.ExecuteCommand(nil, args)
		assert.NotNil(t, err)
	})
	t.Run("no error", func(t *testing.T) {
		text := "Channels we recommend\n * ~highest - \n * ~CoolChannel - \n * ~CoolChannel - \n * ~CoolChannel - \n * ~CoolChannel - \n"
		plugin, api := getPostPlugin(user, channel, text)
		helpers := &plugintest.Helpers{}
		plugin.SetHelpers(helpers)
		helpers.On("KVGetJSON", user, mock.Anything).Return(true, nil).Run(func(args mock.Arguments) {
			channels := args.Get(1).(*[]*recommendedChannel)
			*channels = make([]*recommendedChannel, 6)
			(*channels)[0] = &recommendedChannel{ChannelID: "chan", Score: 0.1}
			(*channels)[1] = &recommendedChannel{ChannelID: "chan", Score: 0.2}
			(*channels)[2] = &recommendedChannel{ChannelID: "chan", Score: 0.3}
			(*channels)[3] = &recommendedChannel{ChannelID: "chan", Score: 0.4}
			(*channels)[4] = &recommendedChannel{ChannelID: "highest", Score: 0.5}
			(*channels)[5] = &recommendedChannel{ChannelID: "chan", Score: 0.24}
		})
		api.On("GetChannel", "highest").Return(&model.Channel{Name: "highest"}, (*model.AppError)(nil))
		api.On("GetChannel", mock.Anything).Return(&model.Channel{Name: "CoolChannel"}, (*model.AppError)(nil))
		defer api.AssertExpectations(t)
		defer helpers.AssertExpectations(t)
		resp, err := plugin.ExecuteCommand(nil, args)
		assert.Nil(t, err)
		assert.Equal(t, &model.CommandResponse{}, resp)
	})
}

func TestExecuteCommandReset(t *testing.T) {
	plugin, api := getPostPlugin("user", "channel", resetText)
	helpers := &plugintest.Helpers{}
	plugin.SetHelpers(helpers)
	helpers.On("KVSetJSON", mock.Anything, mock.Anything).Return((*model.AppError)(nil))
	defer api.AssertExpectations(t)
	defer helpers.AssertExpectations(t)
	args := &model.CommandArgs{
		UserId:    "user",
		ChannelId: "channel",
		Command:   "/suggest reset",
	}
	resp, err := plugin.ExecuteCommand(nil, args)
	assert.Nil(t, err)
	assert.Equal(t, &model.CommandResponse{}, resp)
}
