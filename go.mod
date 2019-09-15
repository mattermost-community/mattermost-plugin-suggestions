module github.com/mattermost/mattermost-plugin-suggestions

go 1.12

require (
	github.com/blang/semver v3.6.1+incompatible // indirect
	//github.com/mattermost/mattermost-server v5.12.0+incompatible
	github.com/mattermost/mattermost-server v1.4.1-0.20190911153151-98489b9e67d9
	github.com/pkg/errors v0.8.1
	github.com/robfig/cron v1.2.0
	github.com/stretchr/testify v1.3.0
	willnorris.com/go/imageproxy v0.9.0
)

// Workaround for https://github.com/golang/go/issues/30831 and fallout.
//replace github.com/golang/lint => github.com/golang/lint v0.0.0-20190227174305-8f45f776aaf1

//replace willnorris.com/go/imageproxy@v0.8.1-0.20190326225038-d4246a08fdec => willnorris.com/go/imageproxy v0.9.0
