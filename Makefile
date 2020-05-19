GO ?= go
GO_TEST_FLAGS ?= -race

all: test

test:
	$(GO) test $(GO_TEST_FLAGS) -v ./...

coverage:
	$(GO) test $(GO_TEST_FLAGS) -coverprofile=coverage.txt ./...
	$(GO) tool cover -html=coverage.txt

check-style:
	@if ! [ -x "$$(command -v golangci-lint)" ]; then \
		echo "golangci-lint is not installed. Please see https://github.com/golangci/golangci-lint#install for installation instructions."; \
		exit 1; \
	fi; \

	@echo Running golangci-lint
	golangci-lint run ./...

## Generates mock golang interfaces for testing
mock:
	go install github.com/golang/mock/mockgen
	mockgen -destination utils/panel/mocks/mock_panel.go -package mock_panel github.com/mattermost/mattermost-plugin-api/utils/panel Panel
	mockgen -destination utils/panel/mocks/mock_panelStore.go -package mock_panel github.com/mattermost/mattermost-plugin-api/utils/panel PanelStore
	mockgen -destination utils/panel/mocks/mock_setting.go -package mock_panel github.com/mattermost/mattermost-plugin-api/utils/panel/settings Setting
	mockgen -destination utils/flow/mocks/mock_flow.go -package mock_flow github.com/mattermost/mattermost-plugin-api/utils/flow Flow
	mockgen -destination utils/flow/mocks/mock_controller.go -package mock_flow github.com/mattermost/mattermost-plugin-api/utils/flow FlowController
	mockgen -destination utils/flow/mocks/mock_store.go -package mock_flow github.com/mattermost/mattermost-plugin-api/utils/flow FlowStore
	mockgen -destination utils/flow/mocks/mock_step.go -package mock_flow github.com/mattermost/mattermost-plugin-api/utils/flow/steps Step
	mockgen -destination utils/bot/mocks/mock_bot.go -package mock_bot github.com/mattermost/mattermost-plugin-api/utils/bot Bot
	mockgen -destination utils/bot/mocks/mock_admin.go -package mock_bot github.com/mattermost/mattermost-plugin-api/utils/bot Admin
	mockgen -destination utils/bot/mocks/mock_logger.go -package mock_bot github.com/mattermost/mattermost-plugin-api/utils/bot/logger Logger
	mockgen -destination utils/bot/mocks/mock_poster.go -package mock_bot github.com/mattermost/mattermost-plugin-api/utils/bot/poster Poster
	mockgen -destination utils/freetext_fetcher/mocks/mock_fetcher.go -package mock_freetext_fetcher github.com/mattermost/mattermost-plugin-api/utils/freetext_fetcher FreetextFetcher
	mockgen -destination utils/freetext_fetcher/mocks/mock_manager.go -package mock_freetext_fetcher github.com/mattermost/mattermost-plugin-api/utils/freetext_fetcher Manager
	mockgen -destination utils/freetext_fetcher/mocks/mock_store.go -package mock_freetext_fetcher github.com/mattermost/mattermost-plugin-api/utils/freetext_fetcher FreetextStore