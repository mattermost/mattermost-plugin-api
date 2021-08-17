package pluginapi

import (
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
)

const (
	internalKeyPrefix = "mmi_"
	botUserKey        = internalKeyPrefix + "botid"
)

// BotService exposes methods to manipulate bots.
type BotService struct {
	api plugin.API
}

// Get returns a bot by botUserID.
//
// Minimum server version: 5.10
func (b *BotService) Get(botUserID string, includeDeleted bool) (*model.Bot, error) {
	bot, appErr := b.api.GetBot(botUserID, includeDeleted)

	return bot, normalizeAppErr(appErr)
}

// BotListOption is an option to configure a bot List() request.
type BotListOption func(*model.BotGetOptions)

// BotOwner option configures bot list request to only retrieve the bots that matches with
// owner's id.
func BotOwner(id string) BotListOption {
	return func(o *model.BotGetOptions) {
		o.OwnerId = id
	}
}

// BotIncludeDeleted option configures bot list request to also retrieve the deleted bots.
func BotIncludeDeleted() BotListOption {
	return func(o *model.BotGetOptions) {
		o.IncludeDeleted = true
	}
}

// BotOnlyOrphans option configures bot list request to only retrieve orphan bots.
func BotOnlyOrphans() BotListOption {
	return func(o *model.BotGetOptions) {
		o.OnlyOrphaned = true
	}
}

// List returns a list of bots by page, count and options.
//
// Minimum server version: 5.10
func (b *BotService) List(page, perPage int, options ...BotListOption) ([]*model.Bot, error) {
	opts := &model.BotGetOptions{
		Page:    page,
		PerPage: perPage,
	}
	for _, o := range options {
		o(opts)
	}
	bots, appErr := b.api.GetBots(opts)

	return bots, normalizeAppErr(appErr)
}

// Create creates the bot and corresponding user.
//
// Minimum server version: 5.10
func (b *BotService) Create(bot *model.Bot) error {
	createdBot, appErr := b.api.CreateBot(bot)
	if appErr != nil {
		return normalizeAppErr(appErr)
	}

	*bot = *createdBot

	return nil
}

// Patch applies the given patch to the bot and corresponding user.
//
// Minimum server version: 5.10
func (b *BotService) Patch(botUserID string, botPatch *model.BotPatch) (*model.Bot, error) {
	bot, appErr := b.api.PatchBot(botUserID, botPatch)

	return bot, normalizeAppErr(appErr)
}

// UpdateActive marks a bot as active or inactive, along with its corresponding user.
//
// Minimum server version: 5.10
func (b *BotService) UpdateActive(botUserID string, isActive bool) (*model.Bot, error) {
	bot, appErr := b.api.UpdateBotActive(botUserID, isActive)

	return bot, normalizeAppErr(appErr)
}

// DeletePermanently permanently deletes a bot and its corresponding user.
//
// Minimum server version: 5.10
func (b *BotService) DeletePermanently(botUserID string) error {
	return normalizeAppErr(b.api.PermanentDeleteBot(botUserID))
}
