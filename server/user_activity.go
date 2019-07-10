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
	perPage := 100
	for page := 0; ; page++ {
		posts, err := p.API.GetPostsForChannel(channelID, page, perPage)
		if err != nil {
			return nil, err
		}
		if len(posts.Order) == 0 {
			break
		}
		pageActivity := getActivityFromPosts(posts, channelID)
		activityUnion(activity, pageActivity)
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
