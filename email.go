package pluginapi

import "github.com/mattermost/mattermost-server/v5/plugin"

// MailService exposes methods to read and write the groups of a Mattermost server.
type MailService struct {
	api plugin.API
}

// Send sends an email to a specific address.
//
// Minimum server version: 5.7
func (m *MailService) Send(to, subject, htmlBody string) error {
	appErr := m.api.SendMail(to, subject, htmlBody)

	return normalizeAppErr(appErr)
}
