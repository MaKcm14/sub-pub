package subpub

import (
	"sync"
	"sync/atomic"
)

// channelSub defines the logic of the channel's definite subscription.
type channelSub struct {
	// handler defines the logic of message's handling after the publisher's publishing.
	handler MessageHandler

	// flagSub defines whether the current subscription is still active.
	flagSub atomic.Bool

	// mut defines the logic of handler synchronization.
	mut sync.Mutex
}

// Unsubscribe defines the logic of the subscription's refusing.
func (c *channelSub) Unsubscribe() {
	c.flagSub.Store(false)
}

// channelConfig defines the channel's configuration.
type channelConfig struct {
	handlers []*channelSub
}

func newChannelConfig() channelConfig {
	return channelConfig{
		handlers: make([]*channelSub, 0, 10),
	}
}

// addSub adds a new subscription to the channel.
func (c *channelConfig) addSub(h MessageHandler) *channelSub {
	sub := &channelSub{
		handler: h,
	}
	sub.flagSub.Store(true)
	c.handlers = append(c.handlers, sub)
	return sub
}

// updateSub updates the subscriptions on the current channel.
func (c *channelConfig) updateSub() {
	newHandler := make([]*channelSub, 0, len(c.handlers))

	for _, sub := range c.handlers {
		if sub.flagSub.Load() {
			newHandler = append(newHandler, sub)
		}
	}

	c.handlers = newHandler
}
