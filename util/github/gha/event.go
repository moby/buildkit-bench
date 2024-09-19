package gha

import (
	"os"

	"github.com/google/go-github/v65/github"
)

func ParseEventFile(name string, fp string) (interface{}, error) {
	dt, err := os.ReadFile(fp)
	if err != nil {
		return nil, err
	}
	return ParseEvent(name, dt)
}

func ParseEvent(name string, dt []byte) (interface{}, error) {
	event, err := github.ParseWebHook(name, dt)
	if err != nil {
		return nil, err
	}
	return event, nil
}
