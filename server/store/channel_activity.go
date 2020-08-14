package store

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

const (
	channelLimit = 1000
	userLimit    = 1000
)

type userChannelActivity = map[string]map[string]int64 //map[userID]map[channelID]activity

// GetChannelActivity returns user activity in most active channels for most active users.
// For now user activity is the number of posts posted in a channel.
func (db *Store) GetChannelActivity(teamID string) (userChannelActivity, []string, error) {
	channels, err := db.GetChannelsForTeam(teamID)
	if err != nil {
		return nil, nil, err
	}
	users, err := db.getUsersForTeam(teamID)
	if err != nil {
		return nil, nil, err
	}
	activity, err := db.getUserActivityForUsersInChannels(users, channels)
	if err != nil {
		return nil, nil, err
	}
	return activity, channels, nil
}

// GetChannelsForTeam returns random 1000 channels in the team
func (db *Store) GetChannelsForTeam(teamID string) ([]string, error) {
	query := db.sq.Select("PC.Id").
		From("PublicChannels AS PC").
		Where(sq.Eq{"PC.TeamId": teamID}).
		Where(sq.Eq{"PC.DeleteAt": 0}).
		OrderBy("RAND()").
		Limit(channelLimit)

	rows, err := query.Query()
	if err != nil {
		return nil, errors.Wrap(err, "Can't query channels")
	}
	defer rows.Close()
	channels := []string{}
	for rows.Next() {
		var channelID string
		if err := rows.Scan(&channelID); err != nil {
			return nil, err
		}
		channels = append(channels, channelID)
	}
	return channels, nil
}

// getUsersForTeam returns random 1000 users in the team
func (db *Store) getUsersForTeam(teamID string) ([]string, error) {
	query := db.sq.Select("U.Id").
		From("Users AS U").
		Join("TeamMembers tm ON ( tm.UserId = U.Id AND tm.DeleteAt = 0 )").Where("tm.TeamId = ?", teamID).
		LeftJoin("Bots ON U.Id = Bots.UserId").Where("Bots.UserId IS NULL").
		OrderBy("RAND()").
		Limit(userLimit)

	rows, err := query.Query()
	if err != nil {
		return nil, errors.Wrap(err, "Can't query users")
	}
	defer rows.Close()
	users := []string{}
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		users = append(users, userID)
	}
	return users, nil
}

func (db *Store) getUserActivityForUsersInChannels(userIDs, channelIDs []string) (userChannelActivity, error) {
	since := model.GetMillis() - period

	query := db.sq.Select("P.ChannelId, P.UserId, COUNT(*)").
		From("Posts AS P").
		Where(sq.Gt{"P.CreateAt": since}).
		GroupBy("P.UserId, P.ChannelId")

	rows, err := query.Query()
	if err != nil {
		return nil, errors.Wrap(err, "Can't query posts in GetUserActivityForUsersInChannels")
	}
	defer rows.Close()

	activity := make(userChannelActivity)
	usersMap := sliceToMap(userIDs)
	channelsMap := sliceToMap(channelIDs)

	for rows.Next() {
		var userID, channelID string
		var count int64
		if err := rows.Scan(&channelID, &userID, &count); err != nil {
			return nil, errors.Wrap(err, "Can't scan rows in GetUserActivityForUsersInChannels")
		}
		if _, ok := usersMap[userID]; !ok {
			continue
		}
		if _, ok := channelsMap[channelID]; !ok {
			continue
		}

		if _, ok := activity[userID]; !ok {
			activity[userID] = make(map[string]int64)
		}
		activity[userID][channelID] = count
	}
	return activity, nil
}

func sliceToMap(s []string) map[string]bool {
	m := make(map[string]bool, len(s))
	for _, item := range s {
		m[item] = true
	}
	return m
}
