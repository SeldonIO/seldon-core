package receiver

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	trackPath      = "/track"
	listEventsPath = "/events"
)

type Receiver struct {
	mux      *http.ServeMux
	recorder Recorder
	logger   logrus.FieldLogger
	port     uint
}

func NewReceiver(l logrus.FieldLogger, port uint, recorder Recorder) *Receiver {
	r := &Receiver{
		mux:      http.NewServeMux(),
		recorder: recorder,
		logger:   l.WithField("source", "Receiver"),
		port:     port,
	}

	r.addRoutes()

	return r
}

func (r *Receiver) addRoutes() {
	r.mux.HandleFunc(trackPath, r.handleTrack)
	r.mux.HandleFunc(listEventsPath, r.handleListEvents)
}

func (r *Receiver) handleTrack(resp http.ResponseWriter, req *http.Request) {
	logger := r.logger.WithField("func", "handleTrack")
	logger.Debugf("received %s request", req.Method)

	event, err := r.eventFrom(req)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}

	r.recorder.Record(event)

	status, err := r.makeStatus(req)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = resp.Write(status)
}

func (r *Receiver) eventFrom(req *http.Request) (*Event, error) {
	logger := r.logger.WithField("func", "eventFrom")
	queryParams := req.URL.Query()

	var decoder *json.Decoder
	if queryParams.Has("data") {
		logger.Debug("decoding event from query parameters")
		decoder = json.NewDecoder(
			base64.NewDecoder(
				base64.URLEncoding,
				strings.NewReader(
					queryParams.Get("data"),
				),
			),
		)
	} else if body := req.Body; body != http.NoBody {
		logger.Debug("decoding event from request body")
		decoder = json.NewDecoder(req.Body)
	} else {
		return nil, errors.New("no event provided")
	}

	event := &Event{}
	err := decoder.Decode(event)
	if err != nil {
		errorMsg := "unable to interpret request as an event"
		logger.WithError(err).Error(errorMsg)
		return nil, errors.New(errorMsg)
	}

	return event, nil
}

func (r *Receiver) makeStatus(req *http.Request) ([]byte, error) {
	logger := r.logger.WithField("func", "makeStatus")

	verbose := req.URL.Query().Get("verbose")
	if len(verbose) == 0 || string(verbose[0]) != "1" {
		logger.Debug("verbose status not requested")
		return nil, nil
	}

	s := Status{
		Status: uint(1),
		Error:  "",
	}

	asJson, err := json.Marshal(s)
	if err != nil {
		errorMsg := "unable to create status response"
		logger.WithError(err).Error(errorMsg)
		return nil, errors.New(errorMsg)
	}

	return asJson, nil
}

func (r *Receiver) handleListEvents(resp http.ResponseWriter, req *http.Request) {
	logger := r.logger.WithField("func", "handleListEvents")
	logger.Debug("received request")

	details := r.recorder.Details()
	asJson, err := json.Marshal(details)
	if err != nil {
		logger.WithError(err).Error()
		http.Error(resp, "unable to create event details", http.StatusInternalServerError)
		return
	}

	_, _ = resp.Write(asJson)
}

// Listen blocks the caller and listens on the Receiver's port for
// inbound connections to serve.
func (r *Receiver) Listen() error {
	logger := r.logger.WithField("func", "Listen")
	logger.Infof("Starting to listen on port %d", r.port)

	err := http.ListenAndServe(
		fmt.Sprintf(":%d", r.port),
		r.mux,
	)
	if err != nil {
		logger.WithError(err).Infof("Unable to listen on port %d", r.port)
	}
	return err
}
