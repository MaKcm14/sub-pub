package subpub

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

// eventChannel is the main channel for sub-pub logic implementation.
type eventChannel struct {
	// channels defines the subscriptions on the channels and its corresponding handlers.
	channels map[string]channelConfig

	// wg defines the object for correct closing.
	wg sync.WaitGroup

	// flagDone defines the condition of the eventChannel.
	flagDone atomic.Bool

	// mut helps syncronize the access to the channels.
	mut sync.Mutex
}

func newEventChannel() *eventChannel {
	return &eventChannel{
		channels: make(map[string]channelConfig),
	}
}

// Subscribe defines the logic of the subscription on the subject.
func (e *eventChannel) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	const op = "subpub.Subscribe"

	if e.flagDone.Load() {
		return nil, fmt.Errorf("error of the %s: %w: try to subscribe after the work done", op, ErrSystemCondition)
	} else if cb == nil {
		return nil, fmt.Errorf("error of the %s: %w: try to subscribe with the nil handler", op, ErrInputData)
	} else if subject == "" {
		return nil, fmt.Errorf("error of the %s: %w: try to subscribe on the empty subject", subject, ErrInputData)
	}

	e.mut.Lock()
	defer e.mut.Unlock()

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

	if e.flagDone.Load() {
		return fmt.Errorf("error of the %s: %w: try to subscribe after the work done", op, ErrSystemCondition)
	} else if subject == "" {
		return fmt.Errorf("error of the %s: %w: try to publish into the empty subject", op, ErrInputData)
	} else if msg == nil {
		return fmt.Errorf("error of the %s: %w: try to publish the nil msg", op, ErrInputData)
	} else if _, ok := e.channels[subject]; !ok {
		return fmt.Errorf("error of the %s: %w: try to publish into the unexisting channel", op, ErrInputData)
	}

	e.mut.Lock()
	defer e.mut.Unlock()

	conf := e.channels[subject]
	conf.updateSub()

	e.channels[subject] = conf

	checkGoStart := atomic.Int64{}
	for _, sub := range e.channels[subject].handlers {
		e.wg.Add(1)
		go func() {
			defer e.wg.Done()

			checkGoStart.Add(1)

			sub.mut.Lock()
			sub.handler(msg)
			sub.mut.Unlock()
		}()
	}

	for checkGoStart.Load() != int64(len(e.channels[subject].handlers)) {
		continue
	}

	return nil
}

// Close shutdowns the eventChannel.
func (e *eventChannel) Close(ctx context.Context) error {
	const op = "subpub.Close"

	e.flagDone.Store(true)

	select {
	case <-ctx.Done():
		return fmt.Errorf("error of the %s: fast shutdown: %s", op, ctx.Err())

	default:
		e.wg.Wait()
	}

	return nil
}
