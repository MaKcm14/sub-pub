package subpub

import (
	"context"
	"fmt"
)

// eventChannel is the main channel for sub-pub logic implementation.
type eventChannel struct {
	channels map[string]channelConfig
	flagDone bool
}

func newEventChannel() *eventChannel {
	return &eventChannel{
		channels: make(map[string]channelConfig),
		flagDone: false,
	}
}

// Subscribe defines the logic of the subscription on the subject.
func (e *eventChannel) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	const op = "subpub.Subscribe"

	if e.flagDone {
		return nil, fmt.Errorf("error of the %s: try to subscribe after the work done", op)
	} else if cb == nil {
		return nil, fmt.Errorf("error of the %s: try to subscribe with the nil handler", op)
	} else if subject == "" {
		return nil, fmt.Errorf("error of the %s: try to subscribe on the empty subject", subject)
	}

	if conf, ok := e.channels[subject]; ok {
		sub := conf.addSub(cb)
		e.channels[subject] = conf
		return sub, nil
	}
	conf := newChannelConfig()
	sub := conf.addSub(cb)

	e.channels[subject] = conf

	return sub, nil
}

// Publish defines the logic of the publishing the event.
func (e *eventChannel) Publish(subject string, msg interface{}) error {
	const op = "subpub.Publish"

	if e.flagDone {
		return fmt.Errorf("error of the %s: try to subscribe after the work done", op)
	} else if subject == "" {
		return fmt.Errorf("error of the %s: try to publish into the empty subject", op)
	} else if msg == nil {
		return fmt.Errorf("error of the %s: try to publish the nil msg", op)
	} else if _, ok := e.channels[subject]; !ok {
		return fmt.Errorf("error of the %s: try to publish into the unexisting channel", op)
	}

	conf := e.channels[subject]
	conf.updateSub()

	e.channels[subject] = conf

	for _, sub := range e.channels[subject].handlers {
		go sub.handler(msg)
	}

	return nil
}

// Close shutdowns the eventChannel.
func (e *eventChannel) Close(ctx context.Context) error {
	<-ctx.Done()
	e.flagDone = true
	return nil
}
