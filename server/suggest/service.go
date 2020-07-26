package suggest

import (
	"sort"
	"time"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-api/cluster"
	"github.com/mattermost/mattermost-plugin-suggestions/server/bot"
	"github.com/mattermost/mattermost-plugin-suggestions/server/config"
	"github.com/mattermost/mattermost-plugin-suggestions/server/ml"
	"github.com/mattermost/mattermost-plugin-suggestions/server/store"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// ServiceImpl holds the information needed by the InsightsService's methods to complete their functions.
type ServiceImpl struct {
	pluginAPI     *pluginapi.Client
	configService config.Service
	store         *store.Store
	poster        bot.Poster
	logger        bot.Logger
	job           *cluster.Job
}

type ChannelScore struct {
	ChannelID string  // identifier
	Score     float64 // score
}

// NewService creates a new insights ServiceImpl.
func NewService(pluginAPI *pluginapi.Client, store *store.Store, poster bot.Poster, configService config.Service, logger bot.Logger) *ServiceImpl {
	suggester := &ServiceImpl{
		pluginAPI:     pluginAPI,
		store:         store,
		poster:        poster,
		configService: configService,
		logger:        logger,
	}
	return suggester
}

func (s *ServiceImpl) StartPreCalcJob(api plugin.API) error {
	callback := func() { s.PreCalculateRecommendations() }
	job, err := cluster.Schedule(api, "job", cluster.MakeWaitForInterval(1*time.Minute), callback)
	if err != nil {
		return err
	}
	s.job = job
	return nil
}

func (s *ServiceImpl) StopPreCalcJob() error {
	return s.job.Close()
}

func (s *ServiceImpl) PreCalculateRecommendations() error {
	s.logger.Infof("preCalculateRecommendations")
	teams, err := s.pluginAPI.Team.List()
	if err != nil {
		return errors.Wrap(err, "Can't get teams.")
	}
	for _, team := range teams {
		if err := s.preCalculateRecommendationsForTeam(team.Id); err != nil {
			return errors.Wrapf(err, "Can't calculate recommendations for team %s", team.DisplayName)
		}
	}
	return nil
}

func (s *ServiceImpl) preCalculateRecommendationsForTeam(teamID string) error {
	if teamID != "z6aynysqmfrdbjzt15hwyps1jr" {
		return nil
	}
	// get total activity of all users
	channelActivity, err := s.store.GetChannelActivity(teamID)
	if err != nil {
		return errors.Wrap(err, "Can't get user activity.")
	}

	channels, err := s.store.GetChannelsForTeam(teamID)
	if err != nil {
		return errors.Wrap(err, "Can't get public channels for a user.")
	}

	k := min(numberOfNeighbors, len(channels)/2+1)

	params := map[string]interface{}{"k": k}
	knn := ml.NewSimpleKNN(params)
	knn.Fit(channelActivity)

	count := 0
	for userID := range channelActivity {
		recommendedChannels := make([]*ChannelScore, 0)
		for _, channel := range channels {
			if _, ok := channelActivity[userID][channel]; !ok {
				score, err := knn.Predict(userID, channel)
				if err != nil {
					// unknown user or unknown channel
					continue
				}

				if score != 0 {
					recommendedChannels = append(recommendedChannels, &ChannelScore{
						ChannelID: channel,
						Score:     score,
					})
				}
			}
		}

		err = s.saveChannelRecommendations(userID, teamID, recommendedChannels)
		if err != nil {
			s.logger.Infof("Can't save recommendations for", "user", userID)
		}
		count++
	}
	return nil
}

func (s *ServiceImpl) GetChannelRecommendations(userID, teamID string) ([]*model.Channel, error) {
	recommendations, err := s.retrieveChannelRecommendations(userID, teamID)
	if err != nil {
		return nil, errors.Wrap(err, "Can't retrieve user recommendations from store")
	}
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})
	n := min(numberOfRecommendedChannels, len(recommendations))
	channels := make([]*model.Channel, 0, n)
	for i := 0; i < n; i++ {
		channel, err := s.pluginAPI.Channel.Get(recommendations[i].ChannelID)
		if err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	return channels, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
