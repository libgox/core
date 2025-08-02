package dynamic

import "time"

type options struct {
	pollInterval time.Duration
}

type Option func(o *options)

func WithPollInterval(pollInterval time.Duration) Option {
	return func(o *options) {
		o.pollInterval = pollInterval
	}
}
