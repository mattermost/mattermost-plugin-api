// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package oauther

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (o *oAuther) oauth2Complete(w http.ResponseWriter, r *http.Request) {
	authedUserID := r.Header.Get("Mattermost-User-ID")
	if authedUserID == "" {
		o.logger.Debugf("oauth2Complete: reached by non authed user")
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		o.logger.Debugf("oauth2Complete: reached with no code")
		http.Error(w, "missing authorization code", http.StatusBadRequest)
		return
	}
	state := r.URL.Query().Get("state")

	storedState, appErr := o.api.KVGet(o.getStateKey(authedUserID))
	if appErr != nil {
		o.logger.Warnf("oauth2Complete: cannot get state, err=%s", appErr.Error())
		http.Error(w, "cannot get stored state", http.StatusInternalServerError)
		return
	}

	if string(storedState) != state {
		o.logger.Debugf("oauth2Complete: state mismatch")
		http.Error(w, "state does not mach", http.StatusUnauthorized)
		return
	}

	userID := strings.Split(state, "_")[1]
	if userID != authedUserID {
		o.logger.Debugf("oauth2Complete: authed user mismatch")
		http.Error(w, "authed user is not the same as state user", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()
	tok, err := o.config.Exchange(ctx, code)
	if err != nil {
		o.logger.Warnf("oauth2Complete: could not generate token, err=%s", err.Error())
		http.Error(w, "could not generate the token", http.StatusUnauthorized)
		return
	}

	rawToken, err := json.Marshal(tok)
	if err != nil {
		o.logger.Errorf("oauth2Complete: could not marshal token, err=%s", err.Error())
		http.Error(w, "cannot marshal the token", http.StatusInternalServerError)
		return
	}
	appErr = o.api.KVSet(o.getTokenKey(userID), rawToken)
	if appErr != nil {
		o.logger.Errorf("oauth2Complete: cannot store the token, err=%s", appErr.Error())
		http.Error(w, "cannot store token", http.StatusInternalServerError)
	}

	html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
			<head>
				<script>
					window.close();
				</script>
			</head>
			<body>
				<p>%s</p>
			</body>
		</html>
		`, o.connectedString)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))

	if o.onConnect != nil {
		o.onConnect(userID, tok)
	}
}
