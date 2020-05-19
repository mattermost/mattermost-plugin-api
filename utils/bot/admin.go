// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package bot

import (
	"errors"
	"strings"

	"github.com/mattermost/mattermost-plugin-api/utils/bot/poster"
)

type Admin interface {
	IsUserAdmin(mattermostUserID string) bool
	DMAdmins(format string, args ...interface{}) error
}

type admin struct {
	poster.Poster
	AdminUserIDs string
}

func NewAdmin(adminUserIDs string, p poster.Poster) Admin {
	return &admin{
		Poster:       p,
		AdminUserIDs: adminUserIDs,
	}
}

func (a *admin) IsUserAdmin(mattermostUserID string) bool {
	list := strings.Split(a.AdminUserIDs, ",")
	for _, u := range list {
		if mattermostUserID == strings.TrimSpace(u) {
			return true
		}
	}
	return false
}

// DM posts a simple Direct Message to the specified user
func (a *admin) DMAdmins(format string, args ...interface{}) error {
	if a.Poster == nil {
		return errors.New("current implementation cannot DM admins, Poster not set")
	}
	for _, id := range strings.Split(a.AdminUserIDs, ",") {
		_, err := a.DM(id, format, args)
		if err != nil {
			return err
		}
	}
	return nil
}
