package pluginapi

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// SlashCommandService exposes methods to manipulate slash commands.
type SlashCommandService struct {
	api plugin.API
}

// Register registers a custom slash command. When the command is triggered, your plugin
// can fulfill it via the ExecuteCommand hook.
//
// Minimum server version: 5.2
func (c *SlashCommandService) Register(command *model.Command) error {
	return c.api.RegisterCommand(command)
}

// Unregister unregisters a command previously registered via Register.
//
// Minimum server version: 5.2
func (c *SlashCommandService) Unregister(teamID, trigger string) error {
	return c.api.UnregisterCommand(teamID, trigger)
}

// ExecuteSlashCommand executes a slash command.
//
// Minimum server version: 5.26
func (c *SlashCommandService) ExecuteSlashCommand(command *model.CommandArgs) (*model.CommandResponse, error) {
	return c.api.ExecuteSlashCommand(command)
}

// CreateCommand creates a server-owned slash command that is not handled by the plugin
// itself, and which will persist past the life of the plugin. The command will have its
// CreatorId set to "" and its PluginId set to the id of the plugin that created it.
//
// Minimum server version: 5.28
func (c *SlashCommandService) CreateCommand(command *model.Command) (*model.Command, error) {
	return c.api.CreateCommand(command)
}

// ListCommands returns the list of all slash commands for teamID. E.g., custom commands
// (those created through the integrations menu, the REST api, or the plugin api CreateCommand),
// plugin commands (those created with plugin api RegisterCommand), and builtin commands
// (those added internally through RegisterCommandProvider).
//
// Minimum server version: 5.28
func (c *SlashCommandService) ListCommands(teamID string) ([]*model.Command, error) {
	return c.api.ListCommands(teamID)
}

// ListCustomCommands returns the list of slash commands for teamID that where created
// through the integrations menu, the REST api, or the plugin api CreateCommand.
//
// Minimum server version: 5.28
func (c *SlashCommandService) ListCustomCommands(teamID string) ([]*model.Command, error) {
	return c.api.ListCustomCommands(teamID)
}

// ListPluginCommands returns the list of slash commands for teamID that were created
// with the plugin api RegisterCommand.
//
// Minimum server version: 5.28
func (c *SlashCommandService) ListPluginCommands(teamID string) ([]*model.Command, error) {
	return c.api.ListPluginCommands(teamID)
}

// ListBuiltInCommands returns the list of slash commands that are builtin commands
// (those added internally through RegisterCommandProvider).
//
// Minimum server version: 5.28
func (c *SlashCommandService) ListBuiltInCommands() ([]*model.Command, error) {
	return c.api.ListBuiltInCommands()
}

// GetCommand returns the command definition based on a command id string.
//
// Minimum server version: 5.28
func (c *SlashCommandService) GetCommand(commandID string) (*model.Command, error) {
	return c.api.GetCommand(commandID)
}

// UpdateCommand updates a single command (commandID) with the information provided in the
// updatedCmd model.Command struct. The following fields in the command cannot be updated:
// Id, Token, CreateAt, DeleteAt, and PluginId. If updatedCmd.TeamId is blank, it
// will be set to commandID's TeamId.
//
// Minimum server version: 5.28
func (c *SlashCommandService) UpdateCommand(commandID string, updatedCmd *model.Command) (*model.Command, error) {
	return c.api.UpdateCommand(commandID, updatedCmd)
}

// DeleteCommand deletes a slash command (commandID).
//
// Minimum server version: 5.28
func (c *SlashCommandService) DeleteCommand(commandID string) error {
	return c.api.DeleteCommand(commandID)
}
