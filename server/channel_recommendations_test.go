package main

import (
	"encoding/json"
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

func TestIsChannelOk(t *testing.T) {
	t.Run("GetPostsForChannel error", func(t *testing.T) {
		plugin, api := getErrorFuncPlugin("GetPostsForChannel", mock.Anything, mock.Anything, mock.Anything)
		defer api.AssertExpectations(t)
		ok := plugin.isChannelOk("")
		assert.False(t, ok)
	})
	t.Run("empty channel", func(t *testing.T) {
		api := &plugintest.API{}
		plugin := &Plugin{}
		posts := new(model.PostList)
		posts.MakeNonNil()
		api.On("GetPostsForChannel", mock.Anything, mock.Anything, mock.Anything).Return(posts, (*model.AppError)(nil))
		plugin.SetAPI(api)
		defer api.AssertExpectations(t)
		ok := plugin.isChannelOk("")
		assert.False(t, ok)
	})
	t.Run("no error", func(t *testing.T) {
		api := &plugintest.API{}
		plugin := &Plugin{}
		posts0, _ := createMockPostList()
		api.On("GetPostsForChannel", mock.Anything, mock.Anything, mock.Anything).Return(posts0, nil)
		plugin.SetAPI(api)
		defer api.AssertExpectations(t)
		ok := plugin.isChannelOk("")
		assert.True(t, ok)
	})

}

func TestPreCalculateRecommendations(t *testing.T) {
	t.Run("getActivity error", func(t *testing.T) {
		plugin, api := getErrorFuncPlugin("KVGet", mock.Anything)
		defer api.AssertExpectations(t)
		plugin.preCalculateRecommendations()
	})

	t.Run("GetAllPublicChannelsForUser error", func(t *testing.T) {
		plugin, api := getUsersInTeamPlugin()
		channels := []*model.Channel{
			&model.Channel{Id: "channel1"},
		}
		api.On("GetChannelsForTeamForUser", mock.Anything, mock.Anything, mock.Anything).Return(channels, nil)
		postList, _ := createMockPostList()
		api.On("GetPostsSince", channels[0].Id, mock.Anything).Return(postList, nil)
		api.On("KVGet", timestampKey).Return([]byte(`0`), nil)
		um := make(userChannelActivity)
		um["user10"] = map[string]int64{"chan": 100}
		j, _ := json.Marshal(um)
		api.On("KVGet", userChannelActivityKey).Return(j, nil)
		api.On("KVSet", mock.Anything, mock.Anything).Return(nil)
		api.On("GetTeamsForUser", mock.Anything).Return(nil, model.NewAppError("", "", nil, "", 404))

		plugin.preCalculateRecommendations()
	})

}
