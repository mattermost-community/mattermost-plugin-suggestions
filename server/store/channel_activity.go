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
func (db *Store) GetChannelActivity(teamID string) (userChannelActivity, error) {
	channels, err := db.GetChannelsForTeam(teamID)
	if err != nil {
		return nil, err
	}
	users, err := db.getUsersForTeam(teamID)
	if err != nil {
		return nil, err
	}
	activity, err := db.getUserActivityForUsersInChannels(users, channels)
	if err != nil {
		return nil, err
	}
	return activity, nil
}

// GetChannelsForTeam returns most active channels in the team
func (db *Store) GetChannelsForTeam(teamID string) ([]string, error) {
	since := model.GetMillis() - period

	query := db.sq.Select("C.Id").
		From("Posts AS P").
		LeftJoin("Channels AS C ON P.ChannelId = C.Id").
		Where(sq.Gt{"P.CreateAt": since}).
		Where(sq.Eq{"C.Type": model.CHANNEL_OPEN}).
		Where(sq.Eq{"C.TeamId": teamID}).
		Where(sq.Eq{"C.DeleteAt": 0}).
		GroupBy("C.Id").
		OrderBy("Count(P.Id) DESC").
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

// getUsersForTeam returns most active users in the team
func (db *Store) getUsersForTeam(teamID string) ([]string, error) {
	since := model.GetMillis() - period

	query := db.sq.Select("U.Id").
		From("Users AS U").
		Join("TeamMembers tm ON ( tm.UserId = U.Id AND tm.DeleteAt = 0 )").Where("tm.TeamId = ?", teamID).
		LeftJoin("Bots ON U.Id = Bots.UserId").Where("Bots.UserId IS NULL").
		LeftJoin("Posts AS P ON P.UserId = U.Id").
		Where(sq.Gt{"P.CreateAt": since}).
		Where(sq.Eq{"P.DeleteAt": 0}).
		GroupBy("U.Id").
		OrderBy("Count(P.Id) DESC").
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

	query := db.sq.Select("P.Id, P.ChannelId, P.UserId").
		From("Posts AS P").
		Where(sq.Gt{"P.CreateAt": since}).
		Where(sq.Eq{"P.ChannelId": channelIDs}).
		Where(sq.Eq{"P.UserId": userIDs})

	rows, err := query.Query()
	if err != nil {
		return nil, errors.Wrap(err, "Can't query posts in GetUserActivityForUsersInChannels")
	}
	defer rows.Close()

	activity := make(userChannelActivity)

	for rows.Next() {
		var postID, userID, channelID string
		if err := rows.Scan(&postID, &channelID, &userID); err != nil {
			return nil, errors.Wrap(err, "Can't scan rows in GetUserActivityForUsersInChannels")
		}
		if _, ok := activity[userID]; !ok {
			activity[userID] = make(map[string]int64)
		}
		activity[userID][channelID]++
	}
	return activity, nil
}
