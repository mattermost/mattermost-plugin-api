package command

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"
)

type BuildInfo struct {
	Manifest   model.Manifest
	CommitSHA  string
	BuildTime  time.Time
	ModulePath string
	GoVersion  string
}

func (b *BuildInfo) String() string {
	path := b.ModulePath

	matches := versionRegexp.FindAllString(path, -1)
	if len(matches) > 0 {
		path = strings.TrimSuffix(path, matches[len(matches)-1])
	}

	CommitSHAShort := b.CommitSHA[0:7]

	commit := fmt.Sprintf("[%s](https://%s/commit/%s)", CommitSHAShort, path, b.CommitSHA)

	return fmt.Sprintf("%s version: %s, %s, built %s with %s\n",
		b.Manifest.Name,
		b.Manifest.Version,
		commit,
		b.BuildTime.Format(time.RFC1123),
		b.GoVersion)
}

var versionRegexp = regexp.MustCompile(`/v\d$`)

func BuildInfoAutocomplete(cmd string) *model.AutocompleteData {
	return model.NewAutocompleteData(cmd, "", "Display build info")
}

func GetBuildInfo(manifest model.Manifest) (*BuildInfo, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil, errors.New("failed to read build info")
	}

	var (
		revision  string
		buildTime time.Time
	)
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value

		case "vcs.time":
			var err error
			buildTime, err = time.Parse(time.RFC3339, s.Value)

			if err != nil {
				return nil, err
			}
		}
	}

	return &BuildInfo{
		Manifest:   manifest,
		CommitSHA:  revision,
		BuildTime:  buildTime,
		ModulePath: info.Main.Path,
		GoVersion:  info.GoVersion,
	}, nil
}
