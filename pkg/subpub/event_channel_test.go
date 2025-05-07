package subpub

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSubscribePositiveCases(t *testing.T) {
	var (
		testChannel1 = "test-channel"
		testChannel2 = "test-channel2"
		h            = func(msg interface{}) {}
		e            = newEventChannel()
	)
	_, err := e.Subscribe(testChannel1, h)

	assert.NoError(t, err, "expected nil error after the right case of Subscribe was called")
	assert.Equal(t, 1, len(e.channels[testChannel1].handlers))

	_, err = e.Subscribe(testChannel1, h)

	assert.NoError(t, err, "expected nil error after the right case of Subscribe was called")
	assert.Equal(t, 2, len(e.channels[testChannel1].handlers))

	_, err = e.Subscribe(testChannel2, h)

	assert.NoError(t, err, "expected nil error after the right case of Subscribe was called")
	assert.Equal(t, 1, len(e.channels[testChannel2].handlers))
}

func TestSubscribeNegativeCases(t *testing.T) {
	var h = func(msg interface{}) {}
	var closedEvenChannel = &eventChannel{}
	closedEvenChannel.flagDone.Store(true)

	type args struct {
		subject string
		cb      MessageHandler
	}
	tests := []struct {
		name     string
		channels *eventChannel
		args     args
		want     Subscription
		wantErr  bool
	}{
		{
			name:     "TestSubscribeNegativeCases_WrongSubject",
			channels: newEventChannel(),
			args: args{
				subject: "",
				cb:      h,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:     "TestSubscribeNegativeCases_WrongCallBackFunction",
			channels: newEventChannel(),
			args: args{
				subject: "test-subject",
				cb:      nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:     "TestSubscribeNegativeCases_WrongEventChannelState",
			channels: closedEvenChannel,
			args: args{
				subject: "test-subject",
				cb:      h,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.channels
			got, err := e.Subscribe(tt.args.subject, tt.args.cb)
			if (err != nil) != tt.wantErr {
				t.Errorf("eventChannel.Subscribe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("eventChannel.Subscribe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPublishPositiveCases(t *testing.T) {
	t.Run("TestPublishPositiveCases_CommonWorkCheck_SingleSubscribersBySingleChannel",
		func(t *testing.T) {
			var (
				testChannel = "test-channel"
				testMessage = "test-message"
			)
			var (
				queue1   = make([]string, 0, 10)
				queue2   = make([]string, 0, 10)
				queue3   = make([]string, 0, 10)
				handler1 = func(msg interface{}) {
					queue1 = append(queue1, msg.(string))
				}
				handler2 = func(msg interface{}) {
					queue2 = append(queue2, msg.(string))
				}
				handler3 = func(msg interface{}) {
					queue3 = append(queue3, msg.(string))
				}
			)
			e := newEventChannel()

			sub1, _ := e.Subscribe(fmt.Sprintf("%s-%d", testChannel, 1), handler1)
			sub2, _ := e.Subscribe(fmt.Sprintf("%s-%d", testChannel, 2), handler2)
			e.Subscribe(fmt.Sprintf("%s-%d", testChannel, 3), handler3)

			for i := 0; i != 3; i++ {
				err := e.Publish(fmt.Sprintf("%s-%d", testChannel, i+1), testMessage)
				assert.NoError(t, err, "expected correct Publish work on the configured channel")
			}
			time.Sleep(time.Second * 5)

			assert.Equal(t, []string{testMessage}, queue1, "expected correct queue1: actual wrong queue was got")
			assert.Equal(t, []string{testMessage}, queue2, "expected correct queue2: actual wrong queue was got")
			assert.Equal(t, []string{testMessage}, queue3, "expected correct queue3: actual wrong queue was got")

			sub1.Unsubscribe()
			sub2.Unsubscribe()

			queue1, queue2 = []string{}, []string{}

			for i := 0; i != 3; i++ {
				err := e.Publish(fmt.Sprintf("%s-%d", testChannel, i+1), testMessage)
				assert.NoError(t, err, "expected the correct Publish executing after the cancelation the sub")
			}
			time.Sleep(time.Second * 5)

			assert.Equal(t, []string{}, queue1, "expected empty queue1: actual wrong queue was got")
			assert.Equal(t, []string{}, queue2, "expected empty queue2: actual wrong queue was got")
			assert.Equal(t, []string{testMessage, testMessage}, queue3, "expected correct queue3: actual wrong queue was got")
		})

	t.Run("TestPublishPositiveCases_CommonWorkCheck_MultipleSubscribersBySingleChannel",
		func(t *testing.T) {
			var (
				testChannel = "test-channel"
				testMessage = "test-message"
			)
			var (
				queue1   = make([]string, 0, 10)
				queue2   = make([]string, 0, 10)
				queue3   = make([]string, 0, 10)
				handler1 = func(msg interface{}) {
					queue1 = append(queue1, msg.(string))
				}
				handler2 = func(msg interface{}) {
					queue2 = append(queue2, msg.(string))
				}
				handler3 = func(msg interface{}) {
					queue3 = append(queue3, msg.(string))
				}
			)
			e := newEventChannel()

			sub1, _ := e.Subscribe(testChannel, handler1)
			sub2, _ := e.Subscribe(testChannel, handler2)
			e.Subscribe(testChannel, handler3)

			for i := 0; i != 2; i++ {
				err := e.Publish(testChannel, testMessage)
				assert.NoError(t, err, "expected correct Publish work on the configured channel")
			}
			time.Sleep(time.Second * 5)

			assert.Equal(t, []string{testMessage, testMessage}, queue1, "expected correct queue1: actual wrong queue was got")
			assert.Equal(t, []string{testMessage, testMessage}, queue2, "expected correct queue2: actual wrong queue was got")
			assert.Equal(t, []string{testMessage, testMessage}, queue3, "expected correct queue3: actual wrong queue was got")

			sub1.Unsubscribe()
			sub2.Unsubscribe()

			queue1, queue2 = []string{}, []string{}

			err := e.Publish(testChannel, testMessage)
			assert.NoError(t, err, "expected the correct Publish executing after the cancelation the sub")

			time.Sleep(time.Second * 5)

			assert.Equal(t, []string{}, queue1, "expected empty queue1: actual wrong queue was got")
			assert.Equal(t, []string{}, queue2, "expected empty queue2: actual wrong queue was got")
			assert.Equal(t, []string{testMessage, testMessage, testMessage}, queue3, "expected correct queue3: actual wrong queue was got")
		})

	t.Run("TestPublishPositiveCases_QueueOrderCheck",
		func(t *testing.T) {
			var (
				testChannel = "test-channel"
				testMessage = "test-message"
				queue       = make([]string, 0, 10)
			)
			e := newEventChannel()

			e.Subscribe(testChannel, func(msg interface{}) {
				time.Sleep(time.Second * 2)
				queue = append(queue, msg.(string))
			})

			for i := 0; i != 5; i++ {
				e.Publish(testChannel, fmt.Sprintf("%s-%d", testMessage, i))
			}

			time.Sleep(time.Second * 15)
			assert.Equal(t, 5, len(queue),
				"expected full completing the publishsing: actual it hasn't been completed")

			for i := 0; i != 5; i++ {
				assert.Equal(t, fmt.Sprintf("%s-%d", testMessage, i), queue[i],
					"expected corresponding queue value: actual order is wrong")
			}
		})
}

func TestPublishCornerCases(t *testing.T) {
	t.Run("TestPublishCornerCases_CheckDeletingUnusedChannels",
		func(t *testing.T) {
			var (
				testChannel = "test-channel"
				testMessage = "test-message"
			)
			e := newEventChannel()

			subs := make([]*channelSub, 0, 5)
			for i := 0; i != 5; i++ {
				s := &channelSub{}
				s.flagSub.Store(false)
				subs = append(subs, s)
			}
			e.channels[testChannel] = channelConfig{
				handlers: subs,
			}

			e.Publish(testChannel, testMessage)

			_, ok := e.channels[testChannel]

			assert.Equal(t, ok, false, "expected unexisting of the empty channel")
		})
}

func TestPublishNegativeCases(t *testing.T) {
	var testSubject = "test-subject"
	var testMsg = "test-message"
	var confEventChannel = &eventChannel{
		channels: map[string]channelConfig{
			testSubject: newChannelConfig(),
		},
	}
	var closedEvenChannel = &eventChannel{}
	closedEvenChannel.flagDone.Store(true)

	type args struct {
		subject string
		msg     interface{}
	}
	tests := []struct {
		name     string
		channels *eventChannel
		args     args
		wantErr  bool
	}{
		{
			name:     "TestPublishNegativeCases_WrongSubject",
			channels: confEventChannel,
			args: args{
				subject: "",
				msg:     testMsg,
			},
			wantErr: true,
		},
		{
			name:     "TestPublishNegativeCases_WrongMessage",
			channels: confEventChannel,
			args: args{
				subject: testSubject,
				msg:     nil,
			},
			wantErr: true,
		},
		{
			name:     "TestPublishNegativeCases_WrongEventChannelState",
			channels: closedEvenChannel,
			args: args{
				subject: testSubject,
				msg:     testMsg,
			},
			wantErr: true,
		},
		{
			name:     "TestPublishNegativeCases_UnexistingChannel",
			channels: newEventChannel(),
			args: args{
				subject: testSubject,
				msg:     testMsg,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.channels
			if err := e.Publish(tt.args.subject, tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("eventChannel.Publish() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClose(t *testing.T) {
	t.Run("TestClose_CorrectShutDown",
		func(t *testing.T) {
			var (
				testChannel = "test-channel"
				testMessage = "test-message"
			)
			e := newEventChannel()
			ctx := context.Background()

			e.Subscribe(testChannel, func(msg interface{}) {
				time.Sleep(time.Millisecond * 200)
			})

			for i := 0; i != 30; i++ {
				e.Publish(testChannel, testMessage)
			}

			assert.NoError(t, e.Close(ctx), "expected nil error after closing: actual some err was got")
			assert.Equal(t, true, e.flagDone.Load(), "expected shutdown condition after event channel closing")
		})

	t.Run("TestClose_IncorrectShutDown",
		func(t *testing.T) {
			e := newEventChannel()
			ctx, cancel := context.WithCancel(context.Background())

			cancel()
			if err := e.Close(ctx); err == nil {
				t.Error("expected error after closing: actual no error was got")
			}

			assert.Equal(t, true, e.flagDone.Load(), "expected shutdown condition after event channel closing")
		})
}
