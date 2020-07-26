package command

import (
	"strings"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-suggestions/server/bot"
	"github.com/mattermost/mattermost-plugin-suggestions/server/suggest"
	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	trigger       = "suggest"
	channelAction = "channels"
	resetAction   = "reset"
	computeAction = "compute"

	displayName = "Suggester Bot"
	desc        = "Mattermost Suggestions Plugin"
	resetText   = "Recommendations were cleared."
	computeText = "Recommendations were computed."
)

const commandHelp = "###### Mattermost Suggestions Plugin - Slash Command Help\n" +
	"* |/suggest channels| - Suggests relevant channels for the user." +
	""

// Command represents slash command of the plugin
type Command struct {
	args      *model.CommandArgs
	log       bot.Logger
	pluginAPI *pluginapi.Client
	poster    bot.Poster
	suggester suggest.Service
}

// NewCommand creates new command
func NewCommand(args *model.CommandArgs, logger bot.Logger, api *pluginapi.Client, poster bot.Poster, suggester suggest.Service) *Command {
	return &Command{
		args:      args,
		log:       logger,
		pluginAPI: api,
		poster:    poster,
		suggester: suggester,
	}
}

// Register is a function that allows the runner to register commands with the mattermost server.
type Register func(*model.Command) error

// RegisterCommands should be called by the plugin to register all necessary commands
func RegisterCommands(registerFunc Register) error {
	return registerFunc(getCommand())
}

func getCommand() *model.Command {
	return &model.Command{
		Trigger:          trigger,
		DisplayName:      displayName,
		Description:      desc,
		AutoComplete:     true,
		AutoCompleteDesc: "Available commands: channels, help",
		AutoCompleteHint: "[command]",
		AutocompleteData: createSuggestionsAutocompleteData(),
	}
}

// Handle .
func (c *Command) Handle() error {
	split := strings.Fields(c.args.Command)
	command := split[0]
	cmd := ""
	if len(split) > 1 {
		cmd = split[1]
	}

	if command != "/suggest" {
		return nil
	}

	switch cmd {
	case "":
		c.postCommandResponse(commandHelp)
	case "help":
		c.postCommandResponse(commandHelp)
	case channelAction:
		c.suggestChannelResponse()
	case computeAction:
		c.compute()
	}

	return nil
}

func (c *Command) postCommandResponse(text string) {
	c.poster.Ephemeral(c.args.UserId, c.args.ChannelId, "%s", text)
}

func createSuggestionsAutocompleteData() *model.AutocompleteData {
	suggestions := model.NewAutocompleteData("suggest", "[command]", "Available commands: channels, help")
	suggestions.AddCommand(createChannelsAutocompleteData())
	return suggestions
}

func (c *Command) compute() {
	if err := c.suggester.PreCalculateRecommendations(); err != nil {
		c.postCommandResponse("error while calculating recommendations" + err.Error())
		return
	}
	c.postCommandResponse(computeText)
}
