// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package oauther

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-plugin-api/experimental/bot/logger"
	"github.com/mattermost/mattermost-plugin-api/experimental/common"
	"golang.org/x/oauth2"
)

const (
	oAuth2StateTimeToLive = 5 * time.Minute
)

type OAuther interface {
	GetToken(userID string) (*oauth2.Token, error)
	GetURL() string
	Deauthorize(userID string) error
}

type oAuther struct {
	store           common.KVStore
	onConnect       func(userID string, token *oauth2.Token)
	storePrefix     string
	pluginURL       string
	oAuthURL        string
	connectedString string
	config          *oauth2.Config
	logger          logger.Logger
}

func NewOAuther(
	r *mux.Router,
	store common.KVStore,
	pluginURL,
	oAuthURL,
	storePrefix,
	connectedString string,
	onConnect func(userID string, token *oauth2.Token),
	oAuthConfig *oauth2.Config,
	loggerBot logger.Logger,
	test http.Handler,
) OAuther {
	o := &oAuther{
		store:           store,
		onConnect:       onConnect,
		storePrefix:     storePrefix,
		pluginURL:       pluginURL,
		oAuthURL:        oAuthURL,
		config:          oAuthConfig,
		connectedString: connectedString,
		logger:          loggerBot,
	}

	o.config.RedirectURL = pluginURL + oAuthURL + "/complete"

	oauth2Router := r.PathPrefix(oAuthURL).Subrouter()
	oauth2Router.HandleFunc("/connect", o.oauth2Connect).Methods("GET")
	oauth2Router.HandleFunc("/complete", o.oauth2Complete).Methods("GET")

	return o
}

func (o *oAuther) GetURL() string {
	return o.pluginURL + o.oAuthURL + "/connect"
}

func (o *oAuther) GetToken(userID string) (*oauth2.Token, error) {
	var token *oauth2.Token
	err := o.store.Get(o.getTokenKey(userID), token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (o *oAuther) getTokenKey(userID string) string {
	return o.storePrefix + "token_" + userID
}

func (o *oAuther) getStateKey(userID string) string {
	return o.storePrefix + "state_" + userID
}

func (o *oAuther) Deauthorize(userID string) error {
	err := o.store.Delete(o.getTokenKey(userID))
	if err != nil {
		return err
	}

	return nil
}
