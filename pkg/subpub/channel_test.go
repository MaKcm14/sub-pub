package subpub

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateSub(t *testing.T) {
	t.Run("TestUpdateSubPositiveCases",
		func(t *testing.T) {
			c := newChannelConfig()

			c.addSub(nil)
			c.addSub(nil)
			c.addSub(nil)
			c.handlers[0].flagSub = false

			c.updateSub()

			assert.Equal(t, 2, len(c.handlers), "try to update the channelSubs: wrong result was got")

			for _, sub := range c.handlers {
				assert.Equal(t, true, sub.flagSub, "expected active flagSub value: deactive flagSub was got")
			}
		})

	t.Run("TestUpdateSubMarginalPositiveCases",
		func(t *testing.T) {
			c := newChannelConfig()

			c.updateSub()

			assert.NotPanics(t, c.updateSub,
				"try to update the empty channelSubs: panic was generated on configured object")

			assert.Equal(t, len(c.handlers), 0, "try to update the empty channelSubs: new subs were appeared")
		})
}

func TestUnsubscribe(t *testing.T) {
	t.Run("TestUnsubscribePositiveCases",
		func(t *testing.T) {
			c := channelSub{
				flagSub: true,
			}

			c.Unsubscribe()

			assert.Equal(t, c.flagSub, false, "try to deactivate the subscription: wrong result was got")
		})
}
