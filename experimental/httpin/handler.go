package httpin

import (
	"github.com/gorilla/mux"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
)

type Handler struct {
	pluginAPI *pluginapi.Client
}

func NewHandler(pluginAPI *pluginapi.Client, router *mux.Router) *Handler {
	handler := &Handler{
		pluginAPI: pluginAPI,
	}

	return handler
}
