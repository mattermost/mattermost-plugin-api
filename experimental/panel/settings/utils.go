package settings

import "github.com/mattermost/mattermost-server/server/v8/model"

func stringsToOptions(in []string) []*model.PostActionOptions {
	out := make([]*model.PostActionOptions, len(in))
	for i, o := range in {
		out[i] = &model.PostActionOptions{
			Text:  o,
			Value: o,
		}
	}
	return out
}
