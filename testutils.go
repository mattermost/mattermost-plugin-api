package pluginapi

import (
	"net/http"

	"github.com/mattermost/mattermost-server/server/v8/model"
)

func newAppError() *model.AppError {
	return model.NewAppError("here", "id", nil, "an error occurred", http.StatusInternalServerError)
}

type testMutex struct {
}

func (m testMutex) Lock()   {}
func (m testMutex) Unlock() {}
