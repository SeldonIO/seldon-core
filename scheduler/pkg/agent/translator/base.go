package translator

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

type Translator interface {
	TranslateToOIP(req *http.Request, logger log.FieldLogger) (*http.Request, error)
	TranslateFromOIP(res *http.Response, logger log.FieldLogger) (*http.Response, error)
}
