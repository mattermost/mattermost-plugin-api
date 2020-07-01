package common

import (
	"net/url"
	"strings"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
)

func GetPluginURL(client *pluginapi.Client, pluginID string) string {
	mattermostSiteURL := client.Configuration.GetConfig().ServiceSettings.SiteURL
	if mattermostSiteURL == nil {
		return ""
	}
	_, err := url.Parse(*mattermostSiteURL)
	if err != nil {
		return ""
	}

	pluginURLPath := "/plugins/" + pluginID
	return strings.TrimRight(*mattermostSiteURL, "/") + pluginURLPath
}
