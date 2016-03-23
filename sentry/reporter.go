package sentry

import (
	"fmt"

	"github.com/alderanalytics/snitch"
	"github.com/getsentry/raven-go"
)

// SentryReporter Reporter implements ErrorReporter via Sentry.
type SentryReporter struct {
	client *raven.Client
}

// NewSentryReporter creates a new sentry reporter from a raven client.
func NewSentryReporter(client *raven.Client) *SentryReporter {
	return &SentryReporter{client: client}
}

// Notify notifies errors via the SentryReporter.
func (sr SentryReporter) Notify(ectx *snitch.ErrorContext) {
	tags := make(map[string]string)
	for k, v := range ectx.Details {
		tags[k] = fmt.Sprint(v)
	}

	sr.client.CaptureMessage(ectx.Error, tags)
}
