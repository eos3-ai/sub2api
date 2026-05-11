package service

import (
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

func TestResolveStripeSelectedPaymentMethod(t *testing.T) {
	tests := []struct {
		name           string
		channel        string
		enabledMethods []string
		want           string
		wantErr        bool
	}{
		{
			name:           "selects alipay from channel",
			channel:        "alipay",
			enabledMethods: []string{"alipay", "wechat_pay"},
			want:           "alipay",
		},
		{
			name:           "selects wechat pay from legacy wechat channel",
			channel:        "wechat",
			enabledMethods: []string{"alipay", "wechat_pay"},
			want:           "wechat_pay",
		},
		{
			name:           "selects wechat pay from provider method channel",
			channel:        "stripe_wechat",
			enabledMethods: []string{"wechat_pay"},
			want:           "wechat_pay",
		},
		{
			name:           "falls back to first user visible method when channel is omitted",
			channel:        "",
			enabledMethods: []string{"card", "alipay", "wechat_pay"},
			want:           "alipay",
		},
		{
			name:           "rejects unsupported configured methods",
			channel:        "",
			enabledMethods: []string{"card", "link"},
			wantErr:        true,
		},
		{
			name:           "rejects selected method when disabled",
			channel:        "wechat",
			enabledMethods: []string{"alipay"},
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveStripeSelectedPaymentMethod(tt.channel, tt.enabledMethods)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveStripeSelectedPaymentMethod returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("method = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStripeResolveReturnURL(t *testing.T) {
	s := &StripeService{
		cfg: &config.StripeConfig{
			SuccessURL: "/payment/return/stripe?order={ORDER_ID}&status=success",
		},
		paymentBaseURL: "https://pay.example.com/",
	}

	got, err := s.resolveReturnURL(&PaymentOrder{OrderNo: "order 123"})
	if err != nil {
		t.Fatalf("resolveReturnURL returned error: %v", err)
	}
	want := "https://pay.example.com/payment/return/stripe?order=order+123&status=success"
	if got != want {
		t.Fatalf("return url = %q, want %q", got, want)
	}
}

func TestStripeResolveReturnURLRequiresBaseForRelativeURL(t *testing.T) {
	s := &StripeService{
		cfg: &config.StripeConfig{
			SuccessURL: "/payment/return/stripe?order={ORDER_ID}",
		},
	}

	_, err := s.resolveReturnURL(&PaymentOrder{OrderNo: "order_123"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "payment.base_url") {
		t.Fatalf("error = %q, want payment.base_url hint", err.Error())
	}
}
