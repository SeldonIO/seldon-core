package receiver

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	listenPort = 8080
)

func makeRequest(
	t *testing.T,
	method string,
	verbose bool,
	queryData string,
	body io.Reader,
) *http.Request {
	queryParams := url.Values{}
	if queryData != "" {
		encodedData := base64.URLEncoding.EncodeToString([]byte(queryData))
		queryParams.Set("data", encodedData)
	}
	if verbose {
		queryParams.Set("verbose", "1")
	}
	uri := trackPath + "?" + queryParams.Encode()

	r, err := http.NewRequest(http.MethodGet, uri, body)
	if err != nil {
		t.FailNow()
	}
	return r
}

func TestHandleTrack(t *testing.T) {
	type test struct {
		name           string
		request        *http.Request
		expectedCode   int
		expectedError  string
		expectedStatus string
	}

	tests := []test{
		{
			name:          "no event should return 400 error",
			request:       makeRequest(t, http.MethodGet, false, "", http.NoBody),
			expectedCode:  http.StatusBadRequest,
			expectedError: "no event provided",
		},
		{
			name:          "ill-formatted query params should return 400 error",
			request:       makeRequest(t, http.MethodGet, false, `{"event": "foo",`, http.NoBody),
			expectedCode:  http.StatusBadRequest,
			expectedError: "unable to interpret request as an event",
		},
		{
			name: "ill-formatted request body should return 400 error",
			request: makeRequest(
				t, http.MethodPost, false, "",
				bytes.NewBuffer([]byte(`{"event": "foo",`)),
			),
			expectedCode:  http.StatusBadRequest,
			expectedError: "unable to interpret request as an event",
		},
		{
			name:          "incomplete event in query params should return 400 error",
			request:       makeRequest(t, http.MethodGet, false, `{"event": "foo"}`, http.NoBody),
			expectedCode:  http.StatusBadRequest,
			expectedError: "unable to interpret request as an event",
		},
		{
			name: "incomplete event in request body should return 400 error",
			request: makeRequest(
				t, http.MethodPost, false, "",
				bytes.NewBuffer([]byte(`{"event": "foo", "properties": {}}`)),
			),
			expectedCode:  http.StatusBadRequest,
			expectedError: "unable to interpret request as an event",
		},
		{
			name: "event in query params should succeed",
			request: makeRequest(
				t, http.MethodGet, false,
				`
				{
					"event": "collect metrics",
					"properties": {
						"token": "asdf1234",
						"time": 1234,
						"distinct_id": "cluster1",
						"$insert_id": "4321"
					}
				}
				`,
				http.NoBody,
			),
			expectedCode:  http.StatusOK,
			expectedError: "",
		},
		{
			name: "event in request body should succeed",
			request: makeRequest(
				t, http.MethodPost, false, "",
				bytes.NewBuffer([]byte(`
					{
						"event": "collect metrics",
						"properties": {
							"token": "asdf1234",
							"time": 1234,
							"distinct_id": "cluster1",
							"$insert_id": "4321"
						}
					}
				`,
				)),
			),
			expectedCode:  http.StatusOK,
			expectedError: "",
		},
		{
			name: "verbose - event in query params should succeed",
			request: makeRequest(
				t, http.MethodGet, true,
				`
				{
					"event": "collect metrics",
					"properties": {
						"token": "asdf1234",
						"time": 1234,
						"distinct_id": "cluster1",
						"$insert_id": "4321"
					}
				}
				`,
				http.NoBody,
			),
			expectedCode:   http.StatusOK,
			expectedStatus: `{"status": 1, "error": ""}`,
		},
		{
			name: "verbose - event in request body should succeed",
			request: makeRequest(
				t, http.MethodPost, true, "",
				bytes.NewBuffer([]byte(`
					{
						"event": "collect metrics",
						"properties": {
							"token": "asdf1234",
							"time": 1234,
							"distinct_id": "cluster1",
							"$insert_id": "4321"
						}
					}
				`,
				)),
			),
			expectedCode:   http.StatusOK,
			expectedStatus: `{"status": 1, "error": ""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New()
			// Effectively disable logging
			logger.SetLevel(logrus.PanicLevel)

			recorder := NewNoopRecorder()
			responseWriter := httptest.NewRecorder()
			r := NewReceiver(logger, listenPort, recorder)

			handler := http.HandlerFunc(r.handleTrack)
			handler.ServeHTTP(responseWriter, tt.request)

			require.Equal(t, tt.expectedCode, responseWriter.Code)
			if tt.expectedError != "" {
				require.Equal(
					t,
					tt.expectedError,
					strings.TrimSpace(responseWriter.Body.String()),
				)
			}
			if tt.expectedStatus != "" {
				require.JSONEq(t, tt.expectedStatus, responseWriter.Body.String())
			}
		})
	}

}
