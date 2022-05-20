package hodometer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"time"

	"github.com/dukex/mixpanel"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
)

const (
	apiKey = "e943e8a2ebd1352338ecac1f9fde8c7c"

	eventName = "collect metrics"
)

const (
	retryMaxAttempts    = 20
	retryMinWaitSeconds = 1 * time.Second
	retryMaxWaitSeconds = 30 * time.Second
)

type levelLogger struct {
	logger logrus.FieldLogger
}

var _ retryablehttp.LeveledLogger = (*levelLogger)(nil)

func (il *levelLogger) Error(msg string, fields ...interface{}) {
	il.logger.Error(msg, fields)
}

func (il *levelLogger) Warn(msg string, fields ...interface{}) {
	il.logger.Warn(msg, fields)
}

func (il *levelLogger) Info(msg string, fields ...interface{}) {
	il.logger.Info(msg, fields)
}

func (il *levelLogger) Debug(msg string, fields ...interface{}) {
	il.logger.Debug(msg, fields)
}

type properties = map[string]interface{}

type Publisher interface {
	Publish(ctx context.Context, metrics *UsageMetrics) error
}

type urlAndClient struct {
	url    string
	client mixpanel.Mixpanel
}

type JsonPublisher struct {
	clients []urlAndClient
	logger  logrus.FieldLogger
	apiUrl  string
}

var _ Publisher = (*JsonPublisher)(nil)

func NewJsonPublisher(
	logger logrus.FieldLogger,
	publishUrls []*url.URL,
) (*JsonPublisher, error) {
	logger = logger.WithField("source", "JsonPublisher")

	if len(publishUrls) == 0 {
		return nil, errors.New("no URLs provided to publish to")
	}

	// TODO - provide TLS & compression settings
	clients := []urlAndClient{}
	for _, u := range publishUrls {
		apiUrl := u.String()
		retryClient := makeRetryClient(logger)
		client := mixpanel.NewFromClient(retryClient, apiKey, apiUrl)

		clients = append(clients, urlAndClient{url: apiUrl, client: client})
	}
	return &JsonPublisher{
		clients: clients,
		logger:  logger,
	}, nil
}

func makeRetryClient(logger logrus.FieldLogger) *http.Client {
	client := retryablehttp.NewClient()
	client.RetryMax = retryMaxAttempts
	client.RetryWaitMin = retryMinWaitSeconds
	client.RetryWaitMax = retryMaxWaitSeconds
	client.Logger = &levelLogger{logger: logger}

	return client.StandardClient()
}

func (jp *JsonPublisher) Publish(ctx context.Context, metrics *UsageMetrics) error {
	logger := jp.logger.WithField("func", "Publish")
	event := jp.makeEvent(metrics)

	wg := sync.WaitGroup{}
	wg.Add(len(jp.clients))

	for _, c := range jp.clients {
		urlAndClient := c
		go func() {
			defer wg.Done()
			logger.Infof("publishing usage metrics to %s", urlAndClient.url)

			err := urlAndClient.client.Track(
				metrics.ClusterId,
				eventName,
				event,
			)
			if err != nil {
				logger.
					WithError(err).
					Errorf("failed to publish usage metrics to %s", urlAndClient.url)
			}
			logger.Infof("published usage metrics to %s", urlAndClient.url)
		}()
	}

	wg.Wait()
	return nil
}

func (jp *JsonPublisher) makeEvent(metrics *UsageMetrics) *mixpanel.Event {
	eventTime := time.Now().UTC()
	insertId := fmt.Sprintf("%d", eventTime.UnixNano())
	p := properties{
		"$insert_id": insertId,
	}
	flattenStructToProperties(p, metrics)

	return &mixpanel.Event{
		IP:         "0", // Disable IP collection
		Timestamp:  &eventTime,
		Properties: p,
	}
}

func metricsToProperties(metrics *UsageMetrics) map[string]interface{} {
	p := properties{}
	flattenStructToProperties(p, metrics)
	return p
}

// flattenStructToProperties puts the simple fields from `s` into `p` as key-value pairs.
//
// `s` can be either a struct or a pointer to a struct, and may contain other
// nested or embedded structs (or pointers to structs).
// `p` will be mutated during this 'flattening' procedure.
// The fields in `s` must be simple for the flattening to work properly,
// i.e. string, int, bool, but NOT list-like, map-like, or pointers.
func flattenStructToProperties(p properties, s interface{}) {
	// TODO - Go 1.18+ generics will allow UsageMetrics || *UsageMetrics

	st := reflect.TypeOf(s)
	sv := reflect.ValueOf(s)

	if st.Kind() != reflect.Struct {
		if !(st.Kind() == reflect.Ptr && st.Elem().Kind() == reflect.Struct) {
			return
		}
		st = st.Elem()
		sv = sv.Elem()
	}

	for i := 0; i < st.NumField(); i++ {
		ft := st.Field(i)
		fv := sv.Field(i)

		switch ft.Type.Kind() {
		case reflect.Struct, reflect.Ptr:
			flattenStructToProperties(p, fv.Interface())
		default:
			// Assume plain field - no support for pointers/other
			jsonTag, ok := ft.Tag.Lookup("json")
			if !ok || jsonTag == "-" {
				continue
			}

			p[jsonTag] = fv.Interface()
		}
	}
}
