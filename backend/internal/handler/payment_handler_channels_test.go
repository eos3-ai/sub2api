package handler

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestPaymentHandlerAvailableUserPaymentChannels(t *testing.T) {
	t.Run("wechat_only_from_env", func(t *testing.T) {
		h := &PaymentHandler{
			cfg: &config.Config{
				Payment: config.PaymentConfig{
					Enabled: true,
					Zpay: config.ZpayConfig{
						Enabled:        true,
						PaymentMethods: "wxpay",
					},
				},
			},
		}

		require.Equal(t, []string{"wechat"}, h.availableUserPaymentChannels())
		require.False(t, h.isCreateChannelEnabled("alipay"))
		require.True(t, h.isCreateChannelEnabled("wechat"))
		require.True(t, h.isCreateChannelEnabled("zpay"))
		require.Equal(t, "wechat", h.resolveCreateChannel("zpay"))
	})

	t.Run("alipay_and_wechat", func(t *testing.T) {
		h := &PaymentHandler{
			cfg: &config.Config{
				Payment: config.PaymentConfig{
					Enabled: true,
					Zpay: config.ZpayConfig{
						Enabled:        true,
						PaymentMethods: "alipay,wxpay",
					},
				},
			},
		}

		require.Equal(t, []string{"alipay", "wechat"}, h.availableUserPaymentChannels())
		require.True(t, h.isCreateChannelEnabled("alipay"))
		require.True(t, h.isCreateChannelEnabled("wechat"))
		require.Equal(t, "alipay", h.resolveCreateChannel("zpay"))
	})

	t.Run("payment_methods_empty_keep_legacy_behavior", func(t *testing.T) {
		h := &PaymentHandler{
			cfg: &config.Config{
				Payment: config.PaymentConfig{
					Enabled: true,
					Zpay: config.ZpayConfig{
						Enabled:        true,
						PaymentMethods: "",
					},
				},
			},
		}

		require.Equal(t, []string{"alipay", "wechat"}, h.availableUserPaymentChannels())
	})

	t.Run("stripe_channel_switch", func(t *testing.T) {
		h := &PaymentHandler{
			cfg: &config.Config{
				Payment: config.PaymentConfig{
					Enabled: true,
					Stripe: config.StripeConfig{
						Enabled: true,
					},
				},
			},
		}

		require.True(t, h.isCreateChannelEnabled("stripe"))
		h.cfg.Payment.Stripe.Enabled = false
		require.False(t, h.isCreateChannelEnabled("stripe"))
	})
}
