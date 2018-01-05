package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func Test_parseAction(t *testing.T) {
	type Test struct {
		Title   string
		Payload string
		Output  string
	}

	tests := []Test{
		Test{
			Title:   "pull request event",
			Payload: `{"action": "append", "number": 1}`,
			Output:  "append",
		},
		Test{
			Title:   "push created",
			Payload: `{"created": true, "deleted": false, "forced": false}`,
			Output:  "created",
		},
		Test{
			Title:   "push deleted",
			Payload: `{"created": false, "deleted": true, "forced": false}`,
			Output:  "deleted",
		},
		Test{
			Title:   "push forced",
			Payload: `{"created": false, "deleted": false, "forced": true}`,
			Output:  "forced",
		},
		Test{
			Title:   "push deleted and forced",
			Payload: `{"created": false, "deleted": true, "forced": true}`,
			Output:  "deleted",
		},
		Test{
			Title:   "other case: fork",
			Payload: `{"forkee": {"id": 123456, "name": "foo"}}`,
			Output:  "",
		},
	}

	for _, test := range tests {
		t.Run(test.Title, func(t *testing.T) {
			var payload interface{}
			if err := json.NewDecoder(strings.NewReader(test.Payload)).Decode(&payload); err != nil {
				t.Fatal("error parse json")
			}

			if action := parseAction(payload); action != test.Output {
				t.Errorf("error parse action. include:%s output:%s", test.Payload, test.Output)
			}
		})
	}
}
