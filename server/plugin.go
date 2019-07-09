package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	// a job for pre-calculating channel recommendations for users.
	preCalcJob    *cron.Cron
	preCalcPeriod string
	botUserID     string
}

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, world!")
}

// See https://developers.mattermost.com/extend/plugins/server/reference/

type readFile func(filename string) ([]byte, error)

func (p *Plugin) setupBot(reader readFile) error {
	botID, err := p.Helpers.EnsureBot(&model.Bot{
		Username:    "suggestions",
		DisplayName: "Suggestions",
		Description: "Created by the Suggestions plugin.",
	})
	if err != nil {
		return errors.Wrap(err, "failed to ensure github bot")
	}
	p.botUserID = botID
	bundlePath, err := p.API.GetBundlePath()
	if err != nil {
		return errors.Wrap(err, "couldn't get bundle path")
	}

	profileImage, err := reader(filepath.Join(bundlePath, "assets", "profile.jpeg"))
	if err != nil {
		return errors.Wrap(err, "couldn't read profile image")
	}

	appErr := p.API.SetProfileImage(botID, profileImage)
	if appErr != nil {
		return errors.Wrap(appErr, "couldn't set profile image")
	}
	return nil
}

func (p *Plugin) startPrecalcJob() error {
	config := p.getConfiguration()
	p.preCalcPeriod = "@daily" // default once a day
	if config.PreCalculationPeriod != "" {
		p.preCalcPeriod = config.PreCalculationPeriod
	}
	c := cron.New()
	if err := c.AddFunc(p.preCalcPeriod, func() {
		p.preCalculateRecommendations()
	}); err != nil {
		return err
	}
	c.Start()
	p.preCalcJob = c
	return nil
}

// OnActivate will be run on plugin activation.
func (p *Plugin) OnActivate() error {
	p.API.RegisterCommand(getCommand())
	err := p.initStore()
	if err != nil {
		return err
	}
	err = p.setupBot(ioutil.ReadFile)
	if err != nil {
		return err
	}
	err = p.startPrecalcJob()
	if err != nil {
		return err
	}
	go p.preCalculateRecommendations() //Run pre-calculation at once

	return nil
}

// OnDeactivate will be run on plugin deactivation.
func (p *Plugin) OnDeactivate() error {
	if p.preCalcJob != nil {
		p.preCalcJob.Stop()
	}
	return nil
}
