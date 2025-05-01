package subpub

import (
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
	var (
		queue1       = make([]string, 0, 10)
		testChannel1 = "test-channel1"
		testMessage1 = "test-message1"
		handler1     = func(msg interface{}) {
			queue1 = append(queue1, msg.(string))
		}
	)
	e := newEventChannel()

	sub, _ := e.Subscribe(testChannel1, handler1)

	err := e.Publish(testChannel1, testMessage1)
	time.Sleep(time.Second * 1)

	assert.NoError(t, err, "expected the correct Publish executing")
	assert.Equal(t, []string{testMessage1}, queue1)

	sub.Unsubscribe()
	queue1 = []string{}

	err = e.Publish(testChannel1, testMessage1)
	time.Sleep(time.Second * 1)

	assert.NoError(t, err, "expected the correct Publish executing")
	assert.Equal(t, []string{}, queue1)
}

func TestPublishNegativeCases(t *testing.T) {
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
			channels: newEventChannel(),
			args: args{
				subject: "",
				msg:     "test-msg",
			},
			wantErr: true,
		},
		{
			name:     "TestPublishNegativeCasesWrongMessage",
			channels: newEventChannel(),
			args: args{
				subject: "test-subject",
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
				subject: "test-subject",
				msg:     "test-msg",
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
			// TODO: test here the correct shutdown through the uncanceled ctx.
		})

	t.Run("TestCloseIncorrectShutDown",
		func(t *testing.T) {
			// TODO: test here the incorrect shutdown through the canceled ctx.
		})
}
