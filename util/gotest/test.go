package gotest

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gotest.tools/gotestsum/testjson"
)

type Test struct {
	Name   string
	Ref    string
	Run    int
	Output string
}

func newTest(event testjson.TestEvent) (ParseEntry, error) {
	t, ok, err := parseTest(event.Test)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse test: %s", event.Test)
	} else if !ok {
		return nil, nil
	}
	return t, nil
}

func (t *Test) ID() string {
	return fmt.Sprintf("%s/ref=%s", t.Name, t.Ref)
}

func (t *Test) Update(event testjson.TestEvent) error {
	t.Output += event.Output
	return nil
}

func parseTest(t string) (*Test, bool, error) {
	if t == "" {
		return nil, false, nil
	}

	tt := &Test{}

	var attrs []string
	for _, part := range strings.Split(t, "/") {
		if !strings.Contains(part, "=") {
			if len(tt.Name) > 0 {
				tt.Name += "/"
			}
			tt.Name += part
		} else {
			attrs = append(attrs, part)
		}
	}
	if len(attrs) == 0 {
		return nil, false, nil
	}

	csvAttrs := strings.Join(attrs, ",")
	csvReader := csv.NewReader(strings.NewReader(csvAttrs))
	fields, err := csvReader.Read()
	if err != nil {
		return nil, false, errors.Wrapf(err, "failed to read test attributes: %s", csvAttrs)
	}

	for _, field := range fields {
		key, value, ok := strings.Cut(field, "=")
		if !ok {
			return nil, false, errors.Errorf("invalid value %s", field)
		}
		switch key {
		case "ref":
			tt.Ref = value
		case "run":
			r, err := strconv.Atoi(value)
			if err != nil {
				return nil, false, errors.Wrapf(err, "failed to parse test run value: %s", value)
			}
			tt.Run = r
		}
	}

	if tt.Ref == "" || tt.Run == 0 {
		return nil, false, nil
	}
	return tt, true, nil
}
