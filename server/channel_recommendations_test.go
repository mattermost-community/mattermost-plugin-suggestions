package main

import (
	"strconv"
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMapToSlice(t *testing.T) {
	t.Run("empty map", func(t *testing.T) {
		res := mapToSlice(make(map[string]*model.Channel))
		assert.Equal(t, make([]*model.Channel, 0), res)
	})

	t.Run("filled map", func(t *testing.T) {
		m := make(map[string]*model.Channel)
		count := 2
		for i := 0; i < count; i++ {
			key := strconv.Itoa(i + 1)
			m[key] = &model.Channel{Id: key}

		}
		res := mapToSlice(m)
		assert.Equal(t, count, len(res))
		for i := 0; i < count; i++ {
			assert.True(t, "1" <= res[i].Id)
			assert.True(t, res[i].Id <= "2")
		}
		assert.NotEqual(t, res[0].Id, res[1].Id)
	})

}

func TestPreCalculateRecommendations(t *testing.T) {
	t.Run("getActivity error", func(t *testing.T) {
		plugin, helpers := getStoreErrorFuncPlugin("KVGetJSON", timestampKey, mock.Anything)
		defer helpers.AssertExpectations(t)
		err := plugin.preCalculateRecommendations()
		assert.NotNil(t, err)
	})

	t.Run("GetAllPublicChannelsForUser error", func(t *testing.T) {
		plugin, api := getUsersInTeamPlugin()
		channels := []*model.Channel{
			&model.Channel{Id: "channel1"},
		}
		api.On("GetChannelsForTeamForUser", mock.Anything, mock.Anything, mock.Anything).Return(channels, nil)
		postList, _ := createMockPostList()
		api.On("GetPostsSince", channels[0].Id, mock.Anything).Return(postList, nil)
		helpers := &plugintest.Helpers{}
		plugin.SetHelpers(helpers)
		helpers.On("KVGetJSON", timestampKey, mock.Anything).Return(true, nil)
		helpers.On("KVGetJSON", userChannelActivityKey, mock.Anything).Return(true, nil)
		helpers.On("KVSetJSON", mock.Anything, mock.Anything).Return(nil)

		api.On("GetTeamsForUser", mock.Anything).Return(nil, model.NewAppError("", "", nil, "", 404))

		err := plugin.preCalculateRecommendations()
		assert.NotNil(t, err)
	})

}
