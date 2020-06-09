package settings

import (
	"github.com/mattermost/mattermost-plugin-api/utils/freetext_fetcher"
)

type baseSetting struct {
	title       string
	description string
	id          string
	dependsOn   string
}

func (s *baseSetting) GetID() string {
	return s.id
}

func (s *baseSetting) GetTitle() string {
	return s.title
}

func (s *baseSetting) GetDescription() string {
	return s.description
}

func (s *baseSetting) GetDependency() string {
	return s.dependsOn
}

func (s *baseSetting) IsDisabled(foreignValue interface{}) bool {
	return false
}

func (s *baseSetting) GetFreetextFetcher() freetext_fetcher.FreetextFetcher {
	return nil
}