package main

import (
	"net/http"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-suggestions/server/bot"
	"github.com/mattermost/mattermost-plugin-suggestions/server/command"
	"github.com/mattermost/mattermost-plugin-suggestions/server/config"
	"github.com/mattermost/mattermost-plugin-suggestions/server/store"
	"github.com/mattermost/mattermost-plugin-suggestions/server/suggest"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	config    *config.ServiceImpl
	bot       *bot.Bot
	suggester suggest.Service
}

// OnActivate Called when this plugin is activated.
func (p *Plugin) OnActivate() error {
	pluginAPIClient := pluginapi.NewClient(p.API)
	p.config = config.NewConfigService(pluginAPIClient, manifest)
	pluginapi.ConfigureLogrus(logrus.New(), pluginAPIClient)

	botID, err := pluginAPIClient.Bot.EnsureBot(&model.Bot{
		Username:    "suggest",
		DisplayName: "Suggester Bot",
		Description: "A bot suggesting different insights in Mattermost.",
	},
		pluginapi.ProfileImagePath("assets/profile.jpeg"),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to ensure suggester bot")
	}

	err = p.config.UpdateConfiguration(func(c *config.Configuration) {
		c.BotUserID = botID
		c.AdminLogLevel = "debug"
	})
	if err != nil {
		return errors.Wrapf(err, "failed save bot to config")
	}

	p.bot = bot.New(pluginAPIClient, p.config.GetConfiguration().BotUserID, p.config)
	config := p.API.GetUnsanitizedConfig()
	st := store.NewStore(*config.SqlSettings.DriverName, pluginAPIClient, p.bot)
	p.suggester = suggest.NewService(pluginAPIClient, st, p.bot, p.config, p.bot)
	p.suggester.StartPreCalcJob(p.API)
	if err := command.RegisterCommands(p.API.RegisterCommand); err != nil {
		return errors.Wrapf(err, "failed register commands")
	}

	p.API.LogDebug("Suggestions plugin Activated")
	return nil
}

// ExecuteCommand executes a given command and returns a command response.
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	com := command.NewCommand(args, p.bot, pluginapi.NewClient(p.API), p.bot, p.suggester)
	if err := com.Handle(); err != nil {
		return nil, model.NewAppError("suggestions.ExecuteCommand", "Unable to execute command.", nil, err.Error(), http.StatusInternalServerError)
	}

	return &model.CommandResponse{}, nil
}

// OnDeactivate Called when this plugin is deactivated.
func (p *Plugin) OnDeactivate() error {
	return p.suggester.StopPreCalcJob()
}
