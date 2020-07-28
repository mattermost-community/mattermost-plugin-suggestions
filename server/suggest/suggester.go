package suggest

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

const numberOfRecommendedChannels = 5
const numberOfNeighbors = 10 // TODO exact number of nearest neighbors should be determined from real data

// Service is the suggestions/service interface.
type Service interface {
	PreCalculateRecommendations() error
	GetChannelRecommendations(userID, teamID string) ([]*model.Channel, error)
	StartPreCalcJob(api plugin.API) error
	StopPreCalcJob() error
}
