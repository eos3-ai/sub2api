package service

import (
	"context"
	"time"
)

// PaymentMaintenanceService runs background tasks for payment module.
type PaymentMaintenanceService struct {
	paymentService *PaymentService
	interval       time.Duration
	stopCh         chan struct{}
}

func NewPaymentMaintenanceService(paymentService *PaymentService, interval time.Duration) *PaymentMaintenanceService {
	if interval <= 0 {
		interval = time.Minute
	}
	return &PaymentMaintenanceService{
		paymentService: paymentService,
		interval:       interval,
		stopCh:         make(chan struct{}),
	}
}

func (s *PaymentMaintenanceService) Start() {
	if s == nil || s.paymentService == nil {
		return
	}
	ticker := time.NewTicker(s.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				_, _ = s.paymentService.CancelExpiredOrders(ctx)
				cancel()
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *PaymentMaintenanceService) Stop() {
	if s == nil {
		return
	}
	select {
	case <-s.stopCh:
		return
	default:
		close(s.stopCh)
	}
}

