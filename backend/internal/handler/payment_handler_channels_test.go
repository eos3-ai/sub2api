package handler

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

func TestLegacyPaymentChannelOptionsFollowProviderSwitches(t *testing.T) {
	tests := []struct {
		name            string
		payment         config.PaymentConfig
		wantChannels    []string
		enabledChannels []string
		disabledChannel string
	}{
		{
			name: "payment enabled but providers disabled exposes no user channels",
			payment: config.PaymentConfig{
				Enabled: true,
				Zpay:    config.ZpayConfig{Enabled: false},
				Stripe:  config.StripeConfig{Enabled: false},
			},
			wantChannels:    nil,
			disabledChannel: "zpay_alipay",
		},
		{
			name: "zpay only exposes zpay methods",
			payment: config.PaymentConfig{
				Enabled: true,
				Zpay: config.ZpayConfig{
					Enabled:        true,
					PaymentMethods: "wxpay",
				},
				Stripe: config.StripeConfig{Enabled: false},
			},
			wantChannels:    []string{"zpay_wechat"},
			enabledChannels: []string{"zpay_wechat", "wechat"},
			disabledChannel: "stripe_wechat",
		},
		{
			name: "stripe only exposes stripe methods",
			payment: config.PaymentConfig{
				Enabled: true,
				Zpay:    config.ZpayConfig{Enabled: false},
				Stripe: config.StripeConfig{
					Enabled:        true,
					PaymentMethods: "alipay,wechat_pay",
				},
			},
			wantChannels:    []string{"stripe_alipay", "stripe_wechat"},
			enabledChannels: []string{"stripe_alipay", "stripe_wechat", "stripe", "alipay", "wechat"},
			disabledChannel: "zpay_alipay",
		},
		{
			name: "both providers expose both provider method sets",
			payment: config.PaymentConfig{
				Enabled: true,
				Zpay: config.ZpayConfig{
					Enabled:        true,
					PaymentMethods: "alipay,wxpay",
				},
				Stripe: config.StripeConfig{
					Enabled:        true,
					PaymentMethods: "alipay,wechat_pay",
				},
			},
			wantChannels:    []string{"zpay_alipay", "zpay_wechat", "stripe_alipay", "stripe_wechat"},
			enabledChannels: []string{"zpay_alipay", "zpay_wechat", "stripe_alipay", "stripe_wechat"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &PaymentHandler{cfg: &config.Config{Payment: tt.payment}}
			gotOptions := h.availableUserPaymentChannelOptions()
			gotChannels := make([]string, 0, len(gotOptions))
			for _, option := range gotOptions {
				gotChannels = append(gotChannels, option.Channel)
			}
			if !stringSlicesEqual(gotChannels, tt.wantChannels) {
				t.Fatalf("channels = %#v, want %#v", gotChannels, tt.wantChannels)
			}

			for _, channel := range tt.enabledChannels {
				if !h.isCreateChannelEnabled(channel) {
					t.Fatalf("expected %q to be enabled", channel)
				}
			}
			if tt.disabledChannel != "" && h.isCreateChannelEnabled(tt.disabledChannel) {
				t.Fatalf("expected %q to be disabled", tt.disabledChannel)
			}
		})
	}
}

func TestLegacyBarePaymentMethodSelectionPrefersZpayThenFallsBack(t *testing.T) {
	stripeOnly := &PaymentHandler{cfg: &config.Config{Payment: config.PaymentConfig{
		Enabled: true,
		Zpay:    config.ZpayConfig{Enabled: false},
		Stripe: config.StripeConfig{
			Enabled:        true,
			PaymentMethods: "alipay,wechat_pay",
		},
	}}}
	got, ok := stripeOnly.resolveCreateChannel("wechat")
	if !ok {
		t.Fatal("expected bare wechat payload to be enabled for stripe-only config")
	}
	if got.Provider != "stripe" || got.Method != "wechat" {
		t.Fatalf("selection = %#v, want stripe/wechat", got)
	}

	bothProviders := &PaymentHandler{cfg: &config.Config{Payment: config.PaymentConfig{
		Enabled: true,
		Zpay: config.ZpayConfig{
			Enabled:        true,
			PaymentMethods: "wxpay",
		},
		Stripe: config.StripeConfig{
			Enabled:        true,
			PaymentMethods: "wechat_pay",
		},
	}}}
	got, ok = bothProviders.resolveCreateChannel("wechat")
	if !ok {
		t.Fatal("expected bare wechat payload to be enabled when both providers support it")
	}
	if got.Provider != "zpay" || got.Method != "wechat" {
		t.Fatalf("selection = %#v, want zpay/wechat", got)
	}
}

func TestLegacyPaymentChannelOptionsIgnoreUnsupportedConfiguredMethods(t *testing.T) {
	h := &PaymentHandler{cfg: &config.Config{Payment: config.PaymentConfig{
		Enabled: true,
		Stripe: config.StripeConfig{
			Enabled:        true,
			PaymentMethods: "card,link",
		},
	}}}

	if got := h.availableUserPaymentChannelOptions(); len(got) != 0 {
		t.Fatalf("channels = %#v, want empty", got)
	}
	if h.isCreateChannelEnabled("stripe") {
		t.Fatal("expected legacy stripe payload to be disabled when no user-visible stripe method is configured")
	}
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
