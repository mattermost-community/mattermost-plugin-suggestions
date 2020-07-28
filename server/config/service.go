package config

import (
	"reflect"
	"sync"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
)

// ServiceImpl holds access to the plugin's Configuration.
type ServiceImpl struct {
	api *pluginapi.Client

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *Configuration

	// configChangeListeners will be notified when the OnConfigurationChange event has been called.
	configChangeListeners map[string]func()

	// manifest is the plugin manifest
	manifest *model.Manifest
}

// NewConfigService Creates a new ServiceImpl struct.
func NewConfigService(api *pluginapi.Client, manifest *model.Manifest) *ServiceImpl {
	c := &ServiceImpl{
		manifest: manifest,
	}
	c.api = api
	c.configuration = new(Configuration)
	c.configChangeListeners = make(map[string]func())

	// api.LoadPluginConfiguration never returns an error, so ignore it.
	_ = api.Configuration.LoadPluginConfiguration(c.configuration)

	return c
}

// GetConfiguration retrieves the active configuration under lock, making it safe to use
// concurrently. The active configuration may change underneath the client of this method, but
// the struct returned by this API call is considered immutable.
func (c *ServiceImpl) GetConfiguration() *Configuration {
	c.configurationLock.RLock()
	defer c.configurationLock.RUnlock()

	if c.configuration == nil {
		return &Configuration{}
	}

	return c.configuration
}

// UpdateConfiguration updates the config. Any parts of the config that are persisted in the plugin's
// section in the server's config will be saved to the server.
func (c *ServiceImpl) UpdateConfiguration(f func(*Configuration)) error {
	c.configurationLock.Lock()

	if c.configuration == nil {
		c.configuration = &Configuration{}
	}

	oldStorableConfig := c.configuration.serialize()
	f(c.configuration)
	newStorableConfig := c.configuration.serialize()
	// Don't hold the lock longer than necessary, especially since we're calling the api and then listeners.
	c.configurationLock.Unlock()

	if !reflect.DeepEqual(oldStorableConfig, newStorableConfig) {
		if appErr := c.api.Configuration.SavePluginConfig(newStorableConfig); appErr != nil {
			return errors.New(appErr.Error())
		}
	}

	for _, f := range c.configChangeListeners {
		f()
	}

	return nil
}

// RegisterConfigChangeListener registers a function that will called when the config might have
// been changed. Returns an id which can be used to unregister the listener.
func (c *ServiceImpl) RegisterConfigChangeListener(listener func()) string {
	if c.configChangeListeners == nil {
		c.configChangeListeners = make(map[string]func())
	}

	id := model.NewId()
	c.configChangeListeners[id] = listener
	return id
}

// UnregisterConfigChangeListener unregisters the listener function identified by id.
func (c *ServiceImpl) UnregisterConfigChangeListener(id string) {
	delete(c.configChangeListeners, id)
}

// OnConfigurationChange is invoked when configuration changes may have been made.
// This method satisfies the interface expected by the server. Embed config.Config in the plugin.
func (c *ServiceImpl) OnConfigurationChange() error {
	// Have we been setup by OnActivate?
	if c.api == nil {
		return nil
	}

	var configuration = new(Configuration)

	// Load the public configuration fields from the Mattermost server configuration.
	if err := c.api.Configuration.LoadPluginConfiguration(configuration); err != nil {
		return errors.Wrapf(err, "failed to load plugin configuration")
	}

	configuration.BotUserID = c.configuration.BotUserID

	c.setConfiguration(configuration)

	for _, f := range c.configChangeListeners {
		f()
	}

	return nil
}

// GetManifest gets the plugin manifest.
func (c *ServiceImpl) GetManifest() *model.Manifest {
	return c.manifest
}

// setConfiguration replaces the active configuration under lock.
//
// Do not call setConfiguration while holding the configurationLock, as sync.Mutex is not
// reentrant. In particular, avoid using the plugin API entirely, as this may in turn trigger a
// hook back into the plugin. If that hook attempts to acquire this lock, a deadlock may occur.
//
// This method panics if setConfiguration is called with the existing configuration. This almost
// certainly means that the configuration was modified without being cloned and may result in
// an unsafe access.
func (c *ServiceImpl) setConfiguration(configuration *Configuration) {
	c.configurationLock.Lock()
	defer c.configurationLock.Unlock()

	if configuration != nil && c.configuration == configuration {
		// Ignore assignment if the configuration struct is empty. Go will optimize the
		// allocation for same to point at the same memory address, breaking the check
		// above.
		if reflect.ValueOf(*configuration).NumField() == 0 {
			return
		}

		panic("setConfiguration called with the existing configuration")
	}

	c.configuration = configuration
}
