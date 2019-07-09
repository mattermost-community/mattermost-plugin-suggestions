# Mattermost Suggestions Plugin
[![CircleCI](https://circleci.com/gh/iomodo/mattermost-plugin-suggestions.svg?style=svg)](https://circleci.com/gh/iomodo/mattermost-plugin-suggestions)
[![codecov](https://codecov.io/gh/iomodo/mattermost-plugin-suggestions/branch/master/graph/badge.svg)](https://codecov.io/gh/iomodo/mattermost-plugin-suggestions)

This plugin delivers a channel suggestions for the users using [collaborative filtering](https://en.wikipedia.org/wiki/Collaborative_filtering).

Collaborative filtering is based on user activities. Basically if user `U1` and `U2` happen to be active in channels `C1`, `C2` and `C3`, and user `U3` is active in `C1` and `C2` we can suggest to the user `U3` that he/she will probably be active in channel `C3` as well.

## Features
* Implementation uses simple KNN method. Later on model could be changed and could be as complicated as it needs to be.
* Number of posts is used as the user activity score per channel. This also could be changed for more complicated model.
* Suggestions are precalculated. A job is spawned in OnActivate() method which calculates suggestions daily and saves them in KVStore.
* One can change precalculation period in the configuration.

## Installation
> git clone https://github.com/iomodo/mattermost-plugin-suggestions.git

> cd mattermost-plugin-suggestions

> make

`suggestions-0.1.0.tar.gz` will be generated in the `mattermost-plugin-suggestions/dist` folder. This file should be uploaded in the mattermost admin console. See details [here](https://docs.mattermost.com/administration/plugins.html#plugin-uploads)

## Usage
Trigger of the suggestion is the slash command `/suggest channels`. Other triggers will be added later.

## Future Work
* Change user activity score and add more features.
* Implement couple of other machine learning models
* Collect user data, perform tests and validation, optimize parameters, improve RMSE

#### Nice to have methods in Mattermost Plugin API
* `func (p *Plugin) GetAllUsers(page, perPage int) ([]*model.User, *model.AppError)`
* `func (p *Plugin) GetAllChannels(page, perPage int) ([]*model.Channel, *model.AppError)`
* `func (p *Plugin) GetAllPublicChannelsForUser(userID string) ([]*model.Channel, *model.AppError)`
* `func (p *Plugin) GetPostsSince(channelID string, since int64, page, perPage int) (*model.PostList, *model.AppError)`

