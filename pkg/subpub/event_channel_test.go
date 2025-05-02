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
	var testChannel1 = "test-channel"
	var testChannel2 = "test-channel2"
	var h = func(msg interface{}) {}
	var e = newEventChannel()

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
			name:     "TestSubscribeNegativeCasesWrongSubject",
			channels: newEventChannel(),
			args: args{
				subject: "",
				cb:      h,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:     "TestSubscribeNegativeCasesWrongCallBackFunction",
			channels: newEventChannel(),
			args: args{
				subject: "test-subject",
				cb:      nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "TestSubscribeNegativeCasesWrongEventChannelState",
			channels: &eventChannel{
				flagDone: true,
			},
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
			e := &eventChannel{
				channels: tt.channels.channels,
				flagDone: tt.channels.flagDone,
			}
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
	t.Run("TestPublishPositiveCasesCommonWorkCheck",
		func(t *testing.T) {
			var (
				queue       = make([]string, 0, 10)
				testChannel = "test-channel"
				testMessage = "test-message"
				handler     = func(msg interface{}) {
					queue = append(queue, msg.(string))
				}
			)
			e := newEventChannel()

			sub, _ := e.Subscribe(testChannel, handler)

			err := e.Publish(testChannel, testMessage)
			time.Sleep(time.Second * 1)

			assert.NoError(t, err, "expected the correct Publish executing")
			assert.Equal(t, []string{testMessage}, queue)

			sub.Unsubscribe()
			queue = []string{}

			err = e.Publish(testChannel, testMessage)
			time.Sleep(time.Second * 1)

			assert.NoError(t, err, "expected the correct Publish executing")
			assert.Equal(t, []string{}, queue)

		})

	t.Run("TestPublishPositiveCasesQueueOrderCheck",
		func(t *testing.T) {
			var (
				testChannel = "test-channel"
				testMessage = "test-message"
				queue       = make([]string, 0, 20)
			)
			e := newEventChannel()

			e.Subscribe(testChannel, func(msg interface{}) {
				queue = append(queue, msg.(string))
			})

			for i := 0; i != 20; i++ {
				e.Publish(testChannel, fmt.Sprintf("%s-%d", testMessage, i))
			}

			time.Sleep(time.Second * 5)
			assert.Equal(t, 20, len(queue),
				"expected full completing the publishsin: actual it hasn't been completed")

			for i := 0; i != 20; i++ {
				assert.Equal(t, fmt.Sprintf("%s-%d", testMessage, i), queue[i],
					"expected corresponding queue value: actual order is wrong")
			}
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
			name:     "TestPublishNegativeCasesWrongSubject",
			channels: confEventChannel,
			args: args{
				subject: "",
				msg:     testMsg,
			},
			wantErr: true,
		},
		{
			name:     "TestPublishNegativeCasesWrongMessage",
			channels: confEventChannel,
			args: args{
				subject: testSubject,
				msg:     nil,
			},
			wantErr: true,
		},
		{
			name: "TestPublishNegativeCasesWrongEventChannelState",
			channels: &eventChannel{
				flagDone: true,
			},
			args: args{
				subject: testSubject,
				msg:     testMsg,
			},
			wantErr: true,
		},
		{
			name:     "TestPublishNegativeCasesUnexistingChannel",
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
			e := &eventChannel{
				channels: tt.channels.channels,
				flagDone: tt.channels.flagDone,
			}
			if err := e.Publish(tt.args.subject, tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("eventChannel.Publish() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClose(t *testing.T) {
	t.Run("TestCloseCorrectShutDown",
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
			assert.Equal(t, true, e.flagDone, "expected shutdown condition after event channel closing")
		})

	t.Run("TestCloseIncorrectShutDown",
		func(t *testing.T) {
			e := newEventChannel()
			ctx, cancel := context.WithCancel(context.Background())

			cancel()
			if err := e.Close(ctx); err == nil {
				t.Error("expected error after closing: actual no error was got")
			}

			assert.Equal(t, true, e.flagDone, "expected shutdown condition after event channel closing")
		})
}
