package service

import "github.com/Wei-Shaw/sub2api/internal/payment"

// ConfigureCheckoutCompatibility wires the dependencies required by the
// v0.1.119 checkout flow without changing the legacy PaymentService constructor.
func (s *PaymentService) ConfigureCheckoutCompatibility(
	configService *PaymentConfigService,
	userRepo UserRepository,
	loadBalancer payment.LoadBalancer,
) {
	if s == nil {
		return
	}

	s.configService = configService
	s.userRepo = userRepo

	if configService != nil {
		s.resumeService = psNewPaymentResumeService(configService)
	}

	if loadBalancer != nil && configService != nil {
		s.loadBalancer = newVisibleMethodLoadBalancer(loadBalancer, configService)
		return
	}

	s.loadBalancer = loadBalancer
}

func (s *PaymentService) PaymentConfigService() *PaymentConfigService {
	if s == nil {
		return nil
	}
	return s.configService
}
