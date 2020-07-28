# Mattermost Suggestions Plugin

This plugin delivers a channel suggestions for the users using [collaborative filtering](https://en.wikipedia.org/wiki/Collaborative_filtering).

Collaborative filtering is based on user activities. Basically if user `U1` and `U2` happen to be active in channels `C1`, `C2` and `C3`, and user `U3` is active in `C1` and `C2` we can suggest to the user `U3` that they will probably be active in channel `C3` as well.

## Features
* Implementation uses simple [KNN method](http://saedsayad.com/k_nearest_neighbors_reg.htm). Later on model could be changed and could be as complicated as it needs to be.
* Number of posts is used as the user activity score per channel. This also could be changed for a more complicated model.
* Suggestions are precalculated. A job is spawned in OnActivate() method which calculates suggestions daily and saves them in KVStore.
* One can change precalculation period in the configuration.

## Installation
> git clone https://github.com/mattermost/mattermost-plugin-suggestions.git

> cd mattermost-plugin-suggestions

> make

`suggestions-0.1.0.tar.gz` will be generated in the `mattermost-plugin-suggestions/dist` folder. This file should be uploaded in the Mattermost System Console. See details [here](https://docs.mattermost.com/administration/plugins.html#plugin-uploads)

## Usage
Trigger the suggestion via the slash command `/suggest channels`. Other triggers may be added later.

## Future Work
* Change user activity score and add more features.
* Implement other machine learning models
* Collect user data, perform tests and validation, optimize parameters, improve Root Mean Square Error


