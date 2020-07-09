package rest

import (
	"fmt"
	"net/url"
)

type httpStatusError struct {
	StatusCode int
	Url        *url.URL
}

func (e *httpStatusError) Error() string {
	return fmt.Sprintf("Internal service call from executor failed calling %s status code %d", e.Url, e.StatusCode)
}

func invalidPayload(msg string) error {
	return fmt.Errorf("invalid payload: %s", msg)
}
