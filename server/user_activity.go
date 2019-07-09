package main

import (
	"time"

	"github.com/mattermost/mattermost-server/model"
)

type userChannelActivity = map[string]map[string]int64 //map[userID]map[channelID]activity

// getActivity returns user activity in channels for every user from the beginning of times.
// For now user activity is the number of posts posted in a channel.
func (p *Plugin) getActivity() (userChannelActivity, error) {
	previousTimestamp, err := p.retreiveTimestamp()
	if err != nil {
		return nil, err
	}
	timestampNow := time.Now().Unix()
	activitySince, appErr := p.getActivitySince(previousTimestamp) // TODO what about the posts that where added between those lines?
	if appErr != nil {
		return nil, appErr
	}
	activityUntil, err := p.retreiveUserChannelActivity()
	if err != nil {
		return nil, err
	}

	activityUnion(activitySince, activityUntil)
	if err = p.saveTimestamp(timestampNow); err != nil {
		return nil, err
	}
	return activitySince, nil
}

// getActivitySince returns activities in channels for every user.
// if since equals to -1 method returns activity score based on
// all posts of the user, otherwise it score is based on posts since certain time.
func (p *Plugin) getActivitySince(since int64) (userChannelActivity, *model.AppError) {
	activity := make(userChannelActivity)
	channels, err := p.GetAllChannels()
	if err != nil {
		return nil, err
	}
	for channelID := range channels {
		var activityForChannel userChannelActivity
		if since == -1 {
			activityForChannel, err = p.getActivityForChannel(channelID)
		} else {
			activityForChannel, err = p.getActivityForChannelSince(channelID, since)
		}
		if err != nil {
			return nil, err
		}
		activityUnion(activity, activityForChannel)
	}
	return activity, nil
}

func (p *Plugin) getActivityForChannel(channelID string) (userChannelActivity, *model.AppError) {
	activity := make(userChannelActivity)
	page := 0
	perPage := 100
	for {
		posts, err := p.API.GetPostsForChannel(channelID, page, perPage)
		if err != nil {
			return nil, err
		}
		if len(posts.Order) == 0 {
			break
		}
		pageActivity := getActivityFromPosts(posts, channelID)
		activityUnion(activity, pageActivity)
		page++
	}
	return activity, nil
}

func (p *Plugin) getActivityForChannelSince(channelID string, since int64) (userChannelActivity, *model.AppError) {
	posts, err := p.API.GetPostsSince(channelID, since)
	if err != nil {
		return nil, err
	}
	return getActivityFromPosts(posts, channelID), nil
}

func getActivityFromPosts(posts *model.PostList, channelID string) userChannelActivity {
	activity := make(userChannelActivity)
	for _, post := range posts.ToSlice() {
		if _, ok := activity[post.UserId]; !ok {
			activity[post.UserId] = make(map[string]int64)
		}
		activity[post.UserId][channelID]++
	}
	return activity
}

func activityUnion(r1, r2 userChannelActivity) {
	for userID2, m2 := range r2 {
		if _, ok := r1[userID2]; !ok {
			r1[userID2] = make(map[string]int64)
		}
		for channelID2, activity2 := range m2 {
			r1[userID2][channelID2] += activity2
		}
	}
}

// getActivitySince returns activities in channels for every user.
// if since equals to -1 method returns activity score based on
// all posts of the user, otherwise it score is based on posts since certain time.
// func (p *Plugin) getActivitySince2(since int64) (userChannelActivity, *model.AppError) {
// 	activities := make(userChannelActivity)
// 	teams, err := p.API.GetTeams()
// 	if err != nil {
// 		p.API.LogError("can't get Teams", "err", err.Error())
// 		return nil, err
// 	}
// 	for _, team := range teams {
// 		rankingsForTeam, err := p.getRankingSinceForTeam(team.Id, since)
// 		if err != nil {
// 			return nil, err
// 		}
// 		activityUnion(activities, rankingsForTeam)
// 	}
// 	return activities, nil
// }

// func (p *Plugin) getRankingSinceForTeam(teamID string, since int64) (userChannelActivity, *model.AppError) {
// 	activities := make(userChannelActivity)
// 	page := 0
// 	perPage := 100
// 	for {
// 		channels, err := p.API.GetPublicChannelsForTeam(teamID, page, perPage)
// 		if err != nil {
// 			p.API.LogError("can't get public channels for a team", "err", err.Error())
// 			return nil, err
// 		}
// 		if len(channels) == 0 {
// 			break
// 		}
// 		for _, channel := range channels {
// 			mlog.Info(fmt.Sprintf("Channel;=======%v==========", channel.Id))
// 			rankingsForChannel, err := p.getActivityForChannelSince(channel.Id, since)
// 			mlog.Info(fmt.Sprintf("rankingsForChannel; = %v", rankingsForChannel))

// 			if err != nil {
// 				return nil, err
// 			}
// 			activityUnion(activities, rankingsForChannel)
// 		}
// 		page++
// 	}
// 	return activities, nil
// }
