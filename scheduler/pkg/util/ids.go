package util

import "github.com/rs/xid"

const RequestIdHeader = "x-request-id"
const RequestIdHeaderCanonical = "X-Request-Id"

func CreateRequestId() string {
	return xid.New().String()
}
