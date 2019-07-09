package main

import (
	"encoding/json"
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestActivityUnion(t *testing.T) {
	t.Run("Add new activity", func(t *testing.T) {
		r1 := userChannelActivity{
			"user1": {
				"channel1": 100,
				"channel2": 200,
			},
		}
		r2 := userChannelActivity{
			"user2": {
				"channel1": 100,
				"channel2": 200,
			},
		}
		activityUnion(r1, r2)
		res := userChannelActivity{
			"user1": {
				"channel1": 100,
				"channel2": 200,
			},
			"user2": {
				"channel1": 100,
				"channel2": 200,
			},
		}
		assert.Equal(t, res, r1)
	})

	t.Run("Mix users activity", func(t *testing.T) {
		r1 := userChannelActivity{
			"user1": {
				"channel1": 100,
				"channel2": 200,
			},
			"user2": {
				"channel1": 100,
				"channel2": 200,
			},
		}
		r2 := userChannelActivity{
			"user2": {
				"channel1": 100,
				"channel2": 200,
			},
		}
		activityUnion(r1, r2)
		res := userChannelActivity{
			"user1": {
				"channel1": 100,
				"channel2": 200,
			},
			"user2": {
				"channel1": 200,
				"channel2": 400,
			},
		}
		assert.Equal(t, res, r1)
	})
}

func TestGetActivityFromPosts(t *testing.T) {
	postList, correctActivity := createMockPostList()
	activity := getActivityFromPosts(postList, "channel1")
	assert.Equal(t, correctActivity, activity)
}

func createMockPostList() (*model.PostList, userChannelActivity) {
	postList := model.NewPostList()
	channelID := "channel1"
	post1 := model.Post{Id: "post1", UserId: "user1", ChannelId: channelID}
	post2 := model.Post{Id: "post2", UserId: "user1", ChannelId: channelID}
	post3 := model.Post{Id: "post3", UserId: "user2", ChannelId: channelID}

	postList.AddPost(&post1)
	postList.AddOrder("post1")
	postList.AddPost(&post2)
	postList.AddOrder("post2")
	postList.AddPost(&post3)
	postList.AddOrder("post3")

	activities := make(userChannelActivity)
	activities["user1"] = map[string]int64{"channel1": 2}
	activities["user2"] = map[string]int64{"channel1": 1}
	return postList, activities
}
func TestGetActivityForChannelSince(t *testing.T) {
	t.Run("GetPostsSince error", func(t *testing.T) {
		plugin, api := getErrorFuncPlugin("GetPostsSince", mock.Anything, mock.Anything)
		defer api.AssertExpectations(t)
		_, err := plugin.getActivityForChannelSince("", 0)
		assert.NotNil(t, err)
	})

	t.Run("no error", func(t *testing.T) {
		api := &plugintest.API{}
		plugin := Plugin{}
		channelID := "channel1"
		postList, correctActivity := createMockPostList()
		api.On("GetPostsSince", channelID, mock.Anything).Return(postList, nil)
		plugin.SetAPI(api)
		activity, err := plugin.getActivityForChannelSince(channelID, 0)
		assert.Nil(t, err)
		assert.Equal(t, correctActivity, activity)
	})
}

func TestGetActivityForChannel(t *testing.T) {
	t.Run("GetPostsForChannel error", func(t *testing.T) {
		plugin, api := getErrorFuncPlugin("GetPostsForChannel", mock.Anything, mock.Anything, mock.Anything)
		defer api.AssertExpectations(t)
		_, err := plugin.getActivityForChannel("")
		assert.NotNil(t, err)
	})

	t.Run("no error", func(t *testing.T) {
		api := &plugintest.API{}
		plugin := Plugin{}
		channelID := "channel1"
		posts0, correctActivity := createMockPostList()
		posts1 := new(model.PostList)
		posts1.MakeNonNil()
		api.On("GetPostsForChannel", channelID, 0, mock.Anything).Return(posts0, nil)
		api.On("GetPostsForChannel", channelID, 1, mock.Anything).Return(posts1, nil)
		plugin.SetAPI(api)
		defer api.AssertExpectations(t)
		res, err := plugin.getActivityForChannel(channelID)
		assert.Nil(t, err)
		assert.Equal(t, correctActivity, res)
	})
}

func TestGetActivitySince(t *testing.T) {
	t.Run("GetAllChannels error", func(t *testing.T) {
		plugin, api := getErrorFuncPlugin("GetTeams")
		defer api.AssertExpectations(t)
		_, err := plugin.getActivitySince(0)
		assert.NotNil(t, err)
	})
	t.Run("getActivityForChannel error", func(t *testing.T) {
		plugin, api := getUsersInTeamPlugin()
		channels := []*model.Channel{
			&model.Channel{Id: "Id1"},
			&model.Channel{Id: "Id2"},
		}

		api.On("GetChannelsForTeamForUser", mock.Anything, mock.Anything, mock.Anything).Return(channels, nil)
		api.On("GetPostsForChannel", mock.Anything, mock.Anything, mock.Anything).Return(nil, model.NewAppError("", "", nil, "", 404))
		defer api.AssertExpectations(t)
		_, err := plugin.getActivitySince(-1)
		assert.NotNil(t, err)
	})

	t.Run("get all activities", func(t *testing.T) {
		plugin, api := getUsersInTeamPlugin()
		channels := []*model.Channel{
			&model.Channel{Id: "channel1"},
		}
		api.On("GetChannelsForTeamForUser", mock.Anything, mock.Anything, mock.Anything).Return(channels, nil)
		posts0, correctActivity := createMockPostList()
		posts1 := new(model.PostList)
		posts1.MakeNonNil()
		api.On("GetPostsForChannel", channels[0].Id, 0, mock.Anything).Return(posts0, nil)
		api.On("GetPostsForChannel", channels[0].Id, 1, mock.Anything).Return(posts1, nil)
		plugin.SetAPI(api)
		defer api.AssertExpectations(t)
		res, err := plugin.getActivitySince(-1)
		assert.Nil(t, err)
		assert.Equal(t, correctActivity, res)
	})

	t.Run("get activities since", func(t *testing.T) {
		plugin, api := getUsersInTeamPlugin()
		channels := []*model.Channel{
			&model.Channel{Id: "channel1"},
		}
		api.On("GetChannelsForTeamForUser", mock.Anything, mock.Anything, mock.Anything).Return(channels, nil)
		postList, correctActivity := createMockPostList()
		api.On("GetPostsSince", channels[0].Id, mock.Anything).Return(postList, nil)
		plugin.SetAPI(api)
		activity, err := plugin.getActivitySince(0)
		assert.Nil(t, err)
		assert.Equal(t, correctActivity, activity)
	})
}
func TestGetActivity(t *testing.T) {
	t.Run("retreiveTimestamp error", func(t *testing.T) {
		plugin, api := getErrorFuncPlugin("KVGet", timestampKey)
		defer api.AssertExpectations(t)
		_, err := plugin.getActivity()
		assert.NotNil(t, err)
	})
	t.Run("getActivitySince error", func(t *testing.T) {
		plugin, api := getErrorFuncPlugin("GetTeams")
		api.On("KVGet", timestampKey).Return([]byte(`0`), nil)
		defer api.AssertExpectations(t)
		_, err := plugin.getActivity()
		assert.NotNil(t, err)
	})
	t.Run("retreiveUserChannelActivity error", func(t *testing.T) {
		plugin, api := getUsersInTeamPlugin()
		channels := []*model.Channel{
			&model.Channel{Id: "channel1"},
		}
		api.On("GetChannelsForTeamForUser", mock.Anything, mock.Anything, mock.Anything).Return(channels, nil)
		postList, _ := createMockPostList()
		api.On("GetPostsSince", channels[0].Id, mock.Anything).Return(postList, nil)
		api.On("KVGet", timestampKey).Return([]byte(`0`), nil)
		api.On("KVGet", userChannelActivityKey).Return(nil, model.NewAppError("", "", nil, "", 404))
		defer api.AssertExpectations(t)
		_, err := plugin.getActivity()
		assert.NotNil(t, err)
	})

	t.Run("saveTimestamp error", func(t *testing.T) {
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
		api.On("KVSet", mock.Anything, mock.Anything).Return(model.NewAppError("", "", nil, "", 404))
		defer api.AssertExpectations(t)
		_, err := plugin.getActivity()
		assert.NotNil(t, err)
	})

	t.Run("no error", func(t *testing.T) {
		plugin, api := getUsersInTeamPlugin()
		channels := []*model.Channel{
			&model.Channel{Id: "channel1"},
		}
		api.On("GetChannelsForTeamForUser", mock.Anything, mock.Anything, mock.Anything).Return(channels, nil)
		postList, correctActivity := createMockPostList()
		api.On("GetPostsSince", channels[0].Id, mock.Anything).Return(postList, nil)
		api.On("KVGet", timestampKey).Return([]byte(`0`), nil)
		um := make(userChannelActivity)
		um["user10"] = map[string]int64{"chan": 100}
		j, _ := json.Marshal(um)
		api.On("KVGet", userChannelActivityKey).Return(j, nil)
		api.On("KVSet", mock.Anything, mock.Anything).Return(nil)
		activity, err := plugin.getActivity()
		assert.Nil(t, err)
		activityUnion(correctActivity, um)
		assert.Equal(t, correctActivity, activity)
	})

}

/*

func TestGetRankingSinceForTeam(t *testing.T) {
	assert := assert.New(t)
	t.Run("GetPublicChannelsForTeam error", func(t *testing.T) {
		api := &plugintest.API{}
		plugin := Plugin{}
		api.On("GetPublicChannelsForTeam", mock.Anything, mock.Anything, mock.Anything).Return(nil, model.NewAppError("", "", nil, "", 404))
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything)
		plugin.SetAPI(api)
		_, err := plugin.getRankingSinceForTeam("", 0)
		assert.NotNil(err)
	})

	t.Run("GetPublicChannelsForTeam 0 channels", func(t *testing.T) {
		api := &plugintest.API{}
		plugin := Plugin{}
		channels := make([]*model.Channel, 0)
		api.On("GetPublicChannelsForTeam", mock.Anything, mock.Anything, mock.Anything).Return(channels, nil)
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything)
		plugin.SetAPI(api)
		ranks, err := plugin.getRankingSinceForTeam("", 0)
		assert.Nil(err)
		assert.Equal(0, len(ranks))
	})

	t.Run("GetPublicChannelsForTeam many channels, GetPostsSince error", func(t *testing.T) {
		api := &plugintest.API{}
		plugin := Plugin{}
		channels := make([]*model.Channel, 1)
		channels[0] = &model.Channel{Id: "channelId"}
		api.On("GetPublicChannelsForTeam", mock.Anything, mock.Anything, mock.Anything).Return(channels, nil)
		api.On("GetPostsSince", mock.Anything, mock.Anything).Return(nil, model.NewAppError("", "", nil, "", 404))
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything)
		plugin.SetAPI(api)
		_, err := plugin.getRankingSinceForTeam("", 0)
		assert.NotNil(err)
	})

	t.Run("GetPublicChannelsForTeam many channels, no error", func(t *testing.T) {
		api := &plugintest.API{}
		plugin := Plugin{}
		channelID := "channel1"
		channels := make([]*model.Channel, 1)
		channels[0] = &model.Channel{Id: channelID}
		postList, correctRanks := createMockPostList()
		api.On("GetPublicChannelsForTeam", mock.Anything, 0, mock.Anything).Return(channels, nil)
		api.On("GetPublicChannelsForTeam", mock.Anything, 1, mock.Anything).Return(make([]*model.Channel, 0), nil)
		api.On("GetPostsSince", channelID, mock.Anything).Return(postList, nil)
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything)
		plugin.SetAPI(api)
		ranks, err := plugin.getRankingSinceForTeam("", 0)
		assert.Nil(err)
		assert.Equal(correctRanks, ranks)
	})
}


*/
