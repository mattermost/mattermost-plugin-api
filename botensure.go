package pluginapi

import (
	"os"
	"path/filepath"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-api/cluster"
)

type mutex interface {
	Lock()
	Unlock()
}

const (
	botEnsureMutexKey = internalKeyPrefix + "bot_ensure"
)

type ensureBotOptions struct {
	ProfileImagePath string
}

type EnsureBotOption func(*ensureBotOptions)

func ProfileImagePath(path string) EnsureBotOption {
	return func(args *ensureBotOptions) {
		args.ProfileImagePath = path
	}
}

// EnsureBot either returns an existing bot user matching the given bot, or creates a bot user from the given bot.
// A profile image or icon image may be optionally passed in to be set for the existing or newly created bot.
// Returns the id of the resulting bot.
// EnsureBot can safely be called multiple instances of a plugin concurrently.
//
// Minimum server version: 5.10
func (b *BotService) EnsureBot(bot *model.Bot, options ...EnsureBotOption) (string, error) {
	m, err := cluster.NewMutex(b.api, botEnsureMutexKey)
	if err != nil {
		return "", errors.Wrap(err, "failed to create mutex")
	}

	return b.ensureBot(m, bot, options...)
}

func (b *BotService) ensureBot(m mutex, bot *model.Bot, options ...EnsureBotOption) (string, error) {
	err := ensureServerVersion(b.api, "5.10.0")
	if err != nil {
		return "", errors.Wrap(err, "failed to ensure bot")
	}

	// Default options
	o := &ensureBotOptions{
		ProfileImagePath: "",
	}

	for _, setter := range options {
		setter(o)
	}

	botID, err := b.ensureBotUser(m, bot)
	if err != nil {
		return "", err
	}

	if o.ProfileImagePath != "" {
		imageBytes, err := b.readFile(o.ProfileImagePath)
		if err != nil {
			return "", errors.Wrap(err, "failed to read profile image")
		}
		appErr := b.api.SetProfileImage(botID, imageBytes)
		if appErr != nil {
			return "", errors.Wrap(appErr, "failed to set profile image")
		}
	}

	return botID, nil
}

func (b *BotService) ensureBotUser(m mutex, bot *model.Bot) (retBotID string, retErr error) {
	// Must provide a bot with a username
	if bot == nil {
		return "", errors.New("passed a nil bot")
	}

	if bot.Username == "" {
		return "", errors.New("passed a bot with no username")
	}

	// Lock to prevent two plugins from racing to create the bot account
	m.Lock()
	defer m.Unlock()

	botIDBytes, kvGetErr := b.api.KVGet(botUserKey)
	if kvGetErr != nil {
		return "", errors.Wrap(kvGetErr, "failed to get bot")
	}

	// If the bot has already been created, use it
	if botIDBytes != nil {
		botID := string(botIDBytes)

		// ensure existing bot is synced with what is being created
		botPatch := &model.BotPatch{
			Username:    &bot.Username,
			DisplayName: &bot.DisplayName,
			Description: &bot.Description,
		}

		if _, err := b.api.PatchBot(botID, botPatch); err != nil {
			return "", errors.Wrap(err, "failed to patch bot")
		}

		return botID, nil
	}

	// Check for an existing bot user with that username. If one exists, then use that.
	if user, appErr := b.api.GetUserByUsername(bot.Username); appErr == nil && user != nil {
		if user.IsBot {
			if appErr := b.api.KVSet(botUserKey, []byte(user.Id)); appErr != nil {
				b.api.LogWarn("Failed to set claimed bot user id.", "userid", user.Id, "err", appErr)
			}
		} else {
			b.api.LogError("Plugin attempted to use an account that already exists. Convert user to a bot "+
				"account in the CLI by running 'mattermost user convert <username> --bot'. If the user is an "+
				"existing user account you want to preserve, change its username and restart the Mattermost server, "+
				"after which the plugin will create a bot account with that name. For more information about bot "+
				"accounts, see https://mattermost.com/pl/default-bot-accounts", "username",
				bot.Username,
				"user_id",
				user.Id,
			)
		}
		return user.Id, nil
	}

	// Create a new bot user for the plugin
	createdBot, appErr := b.api.CreateBot(bot)
	if appErr != nil {
		return "", errors.Wrap(appErr, "failed to create bot")
	}

	if appErr := b.api.KVSet(botUserKey, []byte(createdBot.UserId)); appErr != nil {
		b.api.LogWarn("Failed to set created bot user id.", "userid", createdBot.UserId, "err", appErr)
	}

	return createdBot.UserId, nil
}

func (b *BotService) readFile(path string) ([]byte, error) {
	bundlePath, err := b.api.GetBundlePath()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get bundle path")
	}

	imageBytes, err := os.ReadFile(filepath.Join(bundlePath, path))
	if err != nil {
		return nil, errors.Wrap(err, "failed to read image")
	}

	return imageBytes, nil
}
