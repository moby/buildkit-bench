package gha

import (
	"os"
)

func IsRunning() bool {
	return os.Getenv("GITHUB_ACTIONS") == "true"
}

func IsPullRequestEvent() bool {
	return getEventName() == "pull_request"
}

func getEventName() string {
	return os.Getenv("GITHUB_EVENT_NAME")
}
