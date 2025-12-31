package handler

import (
	"net/http"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type PromotionHandler struct {
	cfg              *config.Config
	promotionService *service.PromotionService
}

func NewPromotionHandler(cfg *config.Config, promotionService *service.PromotionService) *PromotionHandler {
	return &PromotionHandler{
		cfg:              cfg,
		promotionService: promotionService,
	}
}

// GetStatus returns current user's promotion status.
// GET /api/v1/user/promotion/status
func (h *PromotionHandler) GetStatus(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	if h.cfg == nil || !h.cfg.Promotion.Enabled || h.promotionService == nil {
		response.Success(c, dto.PromotionStatusResponse{Enabled: false})
		return
	}

	promotion, err := h.promotionService.GetUserPromotion(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if promotion == nil {
		promotion, err = h.promotionService.EnsureUserPromotion(c.Request.Context(), subject.UserID)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		if promotion == nil {
			response.Success(c, dto.PromotionStatusResponse{
				Enabled:       true,
				Status:        "none",
				Tiers:         dto.PromotionTiersFromConfig(h.cfg.Promotion.Tiers),
				DurationHours: h.cfg.Promotion.DurationHours,
				MinAmountUSD:  h.cfg.Promotion.MinAmount,
			})
			return
		}
	}

	now := time.Now()
	status := promotion.Status
	if status == service.PromotionStatusActive && now.After(promotion.ExpireAt) {
		status = service.PromotionStatusExpired
	}

	remaining := int64(time.Until(promotion.ExpireAt).Seconds())
	if remaining < 0 {
		remaining = 0
	}

	currentTier := dto.CurrentPromotionTier(h.cfg.Promotion.Tiers, promotion.ActivatedAt, promotion.ExpireAt, now)
	var currentTierEndsAt *time.Time
	var currentTierRemainingSeconds int64
	if currentTier != nil {
		endAt := promotion.ActivatedAt.Add(time.Duration(currentTier.Hours) * time.Hour)
		currentTierEndsAt = &endAt
		currentTierRemainingSeconds = int64(time.Until(endAt).Seconds())
		if currentTierRemainingSeconds < 0 {
			currentTierRemainingSeconds = 0
		}
	}

	response.Success(c, dto.PromotionStatusResponse{
		Enabled:                     true,
		Status:                      status,
		Promotion:                   dto.UserPromotionFromService(promotion),
		CurrentTier:                 currentTier,
		CurrentTierEndsAt:           currentTierEndsAt,
		CurrentTierRemainingSeconds: currentTierRemainingSeconds,
		RemainingSeconds:            remaining,
		Tiers:                       dto.PromotionTiersFromConfig(h.cfg.Promotion.Tiers),
		DurationHours:               h.cfg.Promotion.DurationHours,
		MinAmountUSD:                h.cfg.Promotion.MinAmount,
	})
}

func (h *PromotionHandler) Health(c *gin.Context) {
	c.Status(http.StatusOK)
}
