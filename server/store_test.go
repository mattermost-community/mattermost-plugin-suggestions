package main

import (
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
	var channels []*recommendedChannel
	helpers.On("KVSetJSON", "randomUser", channels).Return((*model.AppError)(nil))
	plugin.SetHelpers(helpers)
	err := plugin.saveUserRecommendations("randomUser", channels)
	assert.Nil(err)
}

func TestSaveUserRecommendationsWithError(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	helpers := &plugintest.Helpers{}
	var channels []*recommendedChannel
	helpers.On("KVSetJSON", "randomUser", channels).Return(model.NewAppError("", "", nil, "", 404))
	plugin.SetHelpers(helpers)
	err := plugin.saveUserRecommendations("randomUser", channels)
	assert.NotNil(err)
}

func TestRetrieveUserRecommendationsNoError(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	helpers := &plugintest.Helpers{}

	helpers.On("KVGetJSON", "randomUser", mock.Anything).Return(true, (*model.AppError)(nil))
	plugin.SetHelpers(helpers)

	_, err := plugin.retrieveUserRecommendations("randomUser")
	assert.Nil(err)
}

func TestRetrieveUserRecommendationsWithError(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	helpers := &plugintest.Helpers{}
	helpers.On("KVGetJSON", "randomUser", mock.Anything).Return(false, model.NewAppError("", "", nil, "", 404))
	plugin.SetHelpers(helpers)
	_, err := plugin.retrieveUserRecommendations("randomUser")
	assert.NotNil(err)
}

func TestSaveTimestampNoError(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	helpers := &plugintest.Helpers{}
	helpers.On("KVSetJSON", timestampKey, int64(0)).Return((*model.AppError)(nil))
	plugin.SetHelpers(helpers)
	err := plugin.saveTimestamp(0)
	assert.Nil(err)
}

func TestSaveTimestampWithError(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	helpers := &plugintest.Helpers{}
	helpers.On("KVSetJSON", timestampKey, int64(0)).Return(model.NewAppError("", "", nil, "", 404))
	plugin.SetHelpers(helpers)
	err := plugin.saveTimestamp(0)
	assert.NotNil(err)
}

func TestRetrieveTimestampNoError(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	helpers := &plugintest.Helpers{}
	helpers.On("KVGetJSON", timestampKey, mock.Anything).Return(true, (*model.AppError)(nil))
	plugin.SetHelpers(helpers)

	_, err := plugin.retrieveTimestamp()
	assert.Nil(err)
}

func TestRetrieveTimestampWithError(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	helpers := &plugintest.Helpers{}
	helpers.On("KVGetJSON", timestampKey, mock.Anything).Return(false, model.NewAppError("", "", nil, "", 404))
	plugin.SetHelpers(helpers)
	_, err := plugin.retrieveTimestamp()
	assert.NotNil(err)
}
