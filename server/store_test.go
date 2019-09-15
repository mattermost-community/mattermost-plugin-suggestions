package main

import (
	"encoding/json"
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSaveUserRecommendationsNoError(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	helpers := &plugintest.Helpers{}
	helpers.On("KVSetJSON", mock.Anything, mock.Anything).Return((*model.AppError)(nil))
	plugin.SetHelpers(helpers)
	var channels []*recommendedChannel
	err := plugin.saveUserRecommendations("randomUser", channels)
	assert.Nil(err)
}

func TestSaveUserRecommendationsWithError(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	helpers := &plugintest.Helpers{}
	helpers.On("KVSetJSON", mock.Anything, mock.Anything).Return(model.NewAppError("", "", nil, "", 404))
	plugin.SetHelpers(helpers)
	var channels []*recommendedChannel
	err := plugin.saveUserRecommendations("randomUser", channels)
	assert.NotNil(err)
}

func TestRetreiveUserRecomendationsNoError(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	api := &plugintest.API{}
	helpers := &plugintest.Helpers{}

	channels := make([]*recommendedChannel, 1)

	channels[0] = &recommendedChannel{ChannelID: "chan", Score: 0.1}
	bytes, _ := json.Marshal(channels)

	api.On("KVGet", mock.Anything).Return(bytes, (*model.AppError)(nil))
	plugin.SetAPI(api)
	plugin.SetHelpers(helpers)

	c, err := plugin.retreiveUserRecomendations("randomUser")
	assert.Nil(err)
	assert.Equal(1, len(c))
	assert.Equal(channels[0], c[0])
}

func TestRetreiveUserRecomendationsWithError(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	api := &plugintest.API{}

	api.On("KVGet", mock.Anything).Return(nil, model.NewAppError("", "", nil, "", 404))
	api.On("LogError", mock.Anything, mock.Anything, mock.Anything)
	plugin.SetAPI(api)
	_, err := plugin.retreiveUserRecomendations("randomUser")
	assert.NotNil(err)
}

func TestSaveTimestampNoError(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	helpers := &plugintest.Helpers{}
	helpers.On("KVSetJSON", mock.Anything, mock.Anything).Return((*model.AppError)(nil))
	plugin.SetHelpers(helpers)
	err := plugin.saveTimestamp(0)
	assert.Nil(err)
}

func TestSaveTimestampWithError(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	helpers := &plugintest.Helpers{}
	helpers.On("KVSetJSON", mock.Anything, mock.Anything).Return(model.NewAppError("", "", nil, "", 404))
	plugin.SetHelpers(helpers)
	err := plugin.saveTimestamp(0)
	assert.NotNil(err)
}

func TestRetreiveTimestampNoError(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	api := &plugintest.API{}
	timestamp := int64(100)
	bytes, _ := json.Marshal(timestamp)
	api.On("KVGet", mock.Anything).Return(bytes, (*model.AppError)(nil))
	plugin.SetAPI(api)
	time, err := plugin.retreiveTimestamp()
	assert.Nil(err)
	assert.Equal(timestamp, time)
}

func TestRetreiveTimestampWithError(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	api := &plugintest.API{}

	api.On("KVGet", mock.Anything).Return(nil, model.NewAppError("", "", nil, "", 404))
	api.On("LogError", mock.Anything, mock.Anything, mock.Anything)
	plugin.SetAPI(api)
	_, err := plugin.retreiveTimestamp()
	assert.NotNil(err)
}

func TestUserChannelActivity(t *testing.T) {
	assert := assert.New(t)
	activity := make(map[string]map[string]int64)
	activity["user1"] = map[string]int64{"channel1": 100}
	bytes, _ := json.Marshal(activity)
	plugin := Plugin{}
	api := &plugintest.API{}
	helpers := &plugintest.Helpers{}

	api.On("KVGet", userChannelActivityKey).Return(bytes, nil)
	helpers.On("KVSetJSON", mock.Anything, mock.Anything).Return((*model.AppError)(nil))
	plugin.SetAPI(api)
	plugin.SetHelpers(helpers)

	err := plugin.saveUserChannelActivity(activity)
	assert.Nil(err)
	r, err := plugin.retreiveUserChannelActivity()
	assert.Nil(err)
	assert.Equal(activity, r)
}

func TestRetreive(t *testing.T) {
	t.Run("KVGet error", func(t *testing.T) {
		api := &plugintest.API{}
		plugin := Plugin{}
		api.On("KVGet", mock.Anything).Return(nil, &model.AppError{})
		api.On("LogError", mock.Anything, mock.Anything, mock.Anything)
		plugin.SetAPI(api)

		err := plugin.retreive("key", nil)
		assert.NotNil(t, err)
	})

	t.Run("Marshal error", func(t *testing.T) {
		api := &plugintest.API{}
		plugin := Plugin{}
		api.On("KVGet", mock.Anything).Return([]byte(`{`), nil)
		plugin.SetAPI(api)

		var value map[string]interface{}
		err := plugin.retreive("key", &value)
		assert.NotNil(t, err)
		assert.Nil(t, value)
	})

	t.Run("No error", func(t *testing.T) {
		api := &plugintest.API{}
		plugin := Plugin{}
		api.On("KVGet", mock.Anything).Return([]byte(`{"key": 100}`), nil)
		plugin.SetAPI(api)

		var value map[string]interface{}
		err := plugin.retreive("key", &value)
		assert.Nil(t, err)
		assert.Equal(t, map[string]interface{}{
			"key": float64(100),
		}, value)
	})
}
