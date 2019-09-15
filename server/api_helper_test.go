package main

import (
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAllUsers(t *testing.T) {
	t.Run("getTeamUsers error", func(t *testing.T) {
		plugin, api := getErrorFuncPlugin("GetTeams")
		defer api.AssertExpectations(t)
		_, err := plugin.GetAllUsers()
		assert.NotNil(t, err)
	})

	t.Run("No error", func(t *testing.T) {
		correctUsers := map[string]*model.User{
			"userID1": &model.User{Id: "userID1"},
			"userID2": &model.User{Id: "userID2"},
		}
		plugin, api := getUsersInTeamPlugin()
		defer api.AssertExpectations(t)
		users, err := plugin.GetAllUsers()
		assert.Nil(t, err)
		assert.Equal(t, correctUsers, users)
	})
}

func TestGetAllChannels(t *testing.T) {
	t.Run("getTeamUsers error", func(t *testing.T) {
		plugin, api := getErrorFuncPlugin("GetTeams")
		defer api.AssertExpectations(t)
		_, err := plugin.GetAllChannels()
		assert.NotNil(t, err)
	})

	t.Run("GetChannelsForTeamForUser error", func(t *testing.T) {
		plugin, api := getUsersInTeamPlugin()
		api.On("GetChannelsForTeamForUser", mock.Anything, mock.Anything, mock.Anything).Return(nil, model.NewAppError("", "", nil, "", 404))
		defer api.AssertExpectations(t)
		_, err := plugin.GetAllChannels()
		assert.NotNil(t, err)
	})

	t.Run("No error", func(t *testing.T) {
		plugin, api := getUsersInTeamPlugin()
		channels := []*model.Channel{
			&model.Channel{Id: "Id1"},
			&model.Channel{Id: "Id2"},
		}
		correctChannels := map[string]*model.Channel{
			"Id1": channels[0],
			"Id2": channels[1],
		}
		api.On("GetChannelsForTeamForUser", mock.Anything, mock.Anything, mock.Anything).Return(channels, nil)
		defer api.AssertExpectations(t)
		res, err := plugin.GetAllChannels()
		assert.Nil(t, err)
		assert.Equal(t, correctChannels, res)
	})
}

func TestGetTeamUsers(t *testing.T) {
	t.Run("GetTeams error", func(t *testing.T) {
		plugin, api := getErrorFuncPlugin("GetTeams")
		defer api.AssertExpectations(t)
		_, err := plugin.getTeamUsers()
		assert.NotNil(t, err)
	})

	t.Run("GetUsersInTeam error", func(t *testing.T) {
		api := &plugintest.API{}
		plugin := Plugin{}
		teams := make([]*model.Team, 1)
		teams[0] = &model.Team{Id: "teamID"}
		api.On("GetTeams").Return(teams, nil)
		api.On("GetUsersInTeam", mock.Anything, mock.Anything, mock.Anything).Return(nil, model.NewAppError("", "", nil, "", 404))
		defer api.AssertExpectations(t)
		plugin.SetAPI(api)
		_, err := plugin.getTeamUsers()
		assert.NotNil(t, err)
	})

	t.Run("No error", func(t *testing.T) {
		correctUsers := map[string][]*model.User{
			"teamID": []*model.User{
				&model.User{Id: "userID1"},
				&model.User{Id: "userID2"},
			},
		}
		plugin, api := getUsersInTeamPlugin()
		defer api.AssertExpectations(t)
		users, err := plugin.getTeamUsers()
		assert.Nil(t, err)
		assert.Equal(t, correctUsers, users)
	})
}

func TestGetAllPublicChannelsForUser(t *testing.T) {
	t.Run("GetTeamsForUser error", func(t *testing.T) {
		plugin, api := getErrorFuncPlugin("GetTeamsForUser", mock.Anything)
		defer api.AssertExpectations(t)
		_, err := plugin.GetAllPublicChannelsForUser("")
		assert.NotNil(t, err)
	})

	t.Run("empty teams no error", func(t *testing.T) {
		api := &plugintest.API{}
		plugin := &Plugin{}
		api.On("GetTeamsForUser", mock.Anything).Return(make([]*model.Team, 0), (*model.AppError)(nil))
		plugin.SetAPI(api)
		defer api.AssertExpectations(t)
		res, err := plugin.GetAllPublicChannelsForUser("")
		assert.Nil(t, err)
		assert.Equal(t, 0, len(res))
	})

	t.Run("GetPublicChannelsForTeam error", func(t *testing.T) {
		plugin, api := getErrorFuncPlugin("GetPublicChannelsForTeam", mock.Anything, mock.Anything, mock.Anything)
		teams := make([]*model.Team, 1)
		teams[0] = &model.Team{Id: "teamId"}
		api.On("GetTeamsForUser", mock.Anything).Return(teams, (*model.AppError)(nil))
		defer api.AssertExpectations(t)
		_, err := plugin.GetAllPublicChannelsForUser("")
		assert.NotNil(t, err)
	})

	t.Run("no error", func(t *testing.T) {
		api := &plugintest.API{}
		plugin := &Plugin{}
		teams := make([]*model.Team, 2)
		teams[0] = &model.Team{Id: "teamId1"}
		teams[1] = &model.Team{Id: "teamId2"}
		api.On("GetTeamsForUser", mock.Anything).Return(teams, (*model.AppError)(nil))
		teamChannels := make([]*model.Channel, 2)
		teamChannels[0] = &model.Channel{Id: "channelId1"}
		teamChannels[1] = &model.Channel{Id: "channelId2"}
		api.On("GetPublicChannelsForTeam", mock.Anything, 0, mock.Anything).Return(teamChannels, (*model.AppError)(nil))
		api.On("GetPublicChannelsForTeam", mock.Anything, 1, mock.Anything).Return(make([]*model.Channel, 0), (*model.AppError)(nil))
		plugin.SetAPI(api)
		defer api.AssertExpectations(t)

		correctChannels := map[string]*model.Channel{
			"channelId1": teamChannels[0],
			"channelId2": teamChannels[1],
		}
		res, err := plugin.GetAllPublicChannelsForUser("")
		assert.Nil(t, err)
		assert.Equal(t, correctChannels, res)
	})
}

func getUsersInTeamPlugin() (*Plugin, *plugintest.API) {
	api := &plugintest.API{}
	plugin := Plugin{}
	teams := make([]*model.Team, 1)
	teams[0] = &model.Team{Id: "teamID"}
	api.On("GetTeams").Return(teams, nil)
	users1 := make([]*model.User, 1)
	users1[0] = &model.User{Id: "userID1"}
	users2 := make([]*model.User, 1)
	users2[0] = &model.User{Id: "userID2"}

	api.On("GetUsersInTeam", mock.Anything, 0, mock.Anything).Return(users1, nil)
	api.On("GetUsersInTeam", mock.Anything, 1, mock.Anything).Return(users2, nil)
	api.On("GetUsersInTeam", mock.Anything, 2, mock.Anything).Return(make([]*model.User, 0), nil)
	plugin.SetAPI(api)
	return &plugin, api
}

func getErrorFuncPlugin(funcName string, args ...interface{}) (*Plugin, *plugintest.API) {
	api := &plugintest.API{}
	plugin := &Plugin{}
	api.On(funcName, args...).Return(nil, model.NewAppError("", "", nil, "", 404))
	plugin.SetAPI(api)
	return plugin, api
}
