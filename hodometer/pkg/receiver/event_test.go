package receiver

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmarshallJson(t *testing.T) {
	type test struct {
		name             string
		input            string
		expected         *Event
		expectedErrorMsg *regexp.Regexp
	}

	tests := []test{
		{
			name:             "no event name",
			input:            `{"properties": {"time": 1234, "token": "1234", "distinct_id": "1234", "$insert_id": "1234"}}`,
			expected:         nil,
			expectedErrorMsg: regexp.MustCompile("cannot unmarshal missing event name"),
		},
		{
			name:             "empty event name",
			input:            `{"event": "", "properties": {"time": 1234, "token": "1234", "distinct_id": "1234", "$insert_id": "1234"}}`,
			expected:         nil,
			expectedErrorMsg: regexp.MustCompile("cannot unmarshal missing event name"),
		},
		{
			name:             "whitespace = empty event name",
			input:            `{"event": "  \t ", "properties": {"time": 1234, "token": "1234", "distinct_id": "1234", "$insert_id": "1234"}}`,
			expected:         nil,
			expectedErrorMsg: regexp.MustCompile("cannot unmarshal missing event name"),
		},
		{
			name:             "no properties",
			input:            `{"event": "collect metrics"}`,
			expected:         nil,
			expectedErrorMsg: regexp.MustCompile("cannot unmarshal empty properties"),
		},
		{
			name:     "empty properties",
			input:    `{"event": "collect metrics", "properties": {}}`,
			expected: nil,
			expectedErrorMsg: regexp.MustCompile(
				"required fields not provided: " +
					"((time|token|distinct_id|\\$insert_id), ){3}" +
					"(time|token|distinct_id|\\$insert_id)",
			),
		},
		{
			name: "missing required properties",
			input: `
				{
					"event": "collect metrics",
					"properties": {
						"$insert_id": "qwerty",
						"time": 1652392000437,
						"token": "asdf1234",
						"foo": "bar.bar",
						"Baz": "quux"
					}
				}
			`,
			expected:         nil,
			expectedErrorMsg: regexp.MustCompile("required fields not provided: distinct_id"),
		},
		{
			name: "required properties - milliseconds",
			input: `
				{
					"event": "collect metrics",
					"properties": {
						"$insert_id": "qwerty",
						"distinct_id": "cluster1",
						"time": 1652392000437,
						"token": "asdf1234"
					}
				}
			`,
			expected: &Event{
				Event: "collect metrics",
				Properties: Properties{
					Token:      "asdf1234",
					Time:       1652392000437,
					DistinctId: "cluster1",
					InsertId:   "qwerty",
				},
			},
			expectedErrorMsg: nil,
		},
		{
			name: "extra properties - seconds",
			input: `
				{
					"event": "collect metrics",
					"properties": {
						"$insert_id": "qwerty",
						"distinct_id": "cluster1",
						"time": 1652392000437,
						"token": "asdf1234",
						"foo": "bar.bar",
						"Baz": "quux"
					}
				}
			`,
			expected: &Event{
				Event: "collect metrics",
				Properties: Properties{
					Token:      "asdf1234",
					Time:       1652392000437,
					DistinctId: "cluster1",
					InsertId:   "qwerty",
					Extra: map[string]interface{}{
						"foo": "bar.bar",
						"Baz": "quux",
					},
				},
			},
			expectedErrorMsg: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := &Event{}
			err := json.Unmarshal(
				([]byte)(tt.input),
				actual,
			)

			if tt.expectedErrorMsg == nil {
				require.Nil(t, err)
				require.Equal(t, tt.expected, actual)
			} else {
				require.Regexp(
					t,
					tt.expectedErrorMsg,
					err.Error(),
				)
			}
		})
	}
}

func TestMarshalJSON(t *testing.T) {
	type test struct {
		name          string
		event         *Event
		expected      string
		expectedError bool
	}

	tests := []test{
		{
			name: "only required properties",
			event: &Event{
				Event: "collect metrics",
				Properties: Properties{
					Token:      "asdf1234",
					Time:       1652704273,
					DistinctId: "cluster1",
					InsertId:   "999",
				},
			},
			expected: `
			{
				"event": "collect metrics",
				"properties": {
					"token": "asdf1234",
					"time": 1652704273,
					"distinct_id": "cluster1",
					"$insert_id": "999"
				}
			}
			`,
			expectedError: false,
		},
		{
			name: "with extra properties",
			event: &Event{
				Event: "collect metrics",
				Properties: Properties{
					Token:      "asdf1234",
					Time:       1652704273,
					DistinctId: "cluster1",
					InsertId:   "999",
					Extra: KeysAndValues{
						"foo": "bar",
						"baz": "0987",
					},
				},
			},
			expected: `
			{
				"event": "collect metrics",
				"properties": {
					"token": "asdf1234",
					"time": 1652704273,
					"distinct_id": "cluster1",
					"$insert_id": "999",
					"foo": "bar",
					"baz": "0987"
				}
			}
			`,
			expectedError: false,
		},
		{
			name:  "empty event",
			event: &Event{},
			expected: `
			{
				"event": "",
				"properties": {
					"token": "",
					"time": 0,
					"distinct_id": "",
					"$insert_id": ""
				}
			}
			`,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := json.Marshal(tt.event)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
				require.JSONEq(t, tt.expected, string(actual))
			}
		})
	}
}
