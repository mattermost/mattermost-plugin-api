// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package oauther

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
	"golang.org/x/oauth2"
)

func (o *oAuther) oauth2Connect(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		o.logger.Debugf("oauth2Connect: reached by non authed user")
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	token, _ := o.GetToken(userID)
	if token != nil {
		o.logger.Debugf("oauth2Connect: reached by connected user")
		http.Error(w, "user already has a token", http.StatusBadRequest)
		return
	}

	state := fmt.Sprintf("%v_%v", model.NewId()[0:15], userID)
	appErr := o.api.KVSetWithExpiry(o.getStateKey(userID), []byte(state), oAuth2StateTimeToLive)
	if appErr != nil {
		o.logger.Errorf("oauth2Connect: failed to store state, err=%s", appErr.Error())
		http.Error(w, "failed to store token state", http.StatusInternalServerError)
		return
	}

	redirectURL := o.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
