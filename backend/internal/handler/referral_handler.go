package handler

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type ReferralHandler struct {
	cfg            *config.Config
	referralService *service.ReferralService
}

func NewReferralHandler(cfg *config.Config, referralService *service.ReferralService) *ReferralHandler {
	return &ReferralHandler{
		cfg:            cfg,
		referralService: referralService,
	}
}

// GetInfo returns current user's referral program info (code, link, stats, rules).
// GET /api/v1/user/referral/info
func (h *ReferralHandler) GetInfo(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	if h.cfg == nil || !h.cfg.Referral.Enabled || h.referralService == nil {
		response.Success(c, dto.ReferralInfoResponse{Enabled: false})
		return
	}

	code, err := h.referralService.GetOrCreateUserCode(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	stats, err := h.referralService.GetReferrerStats(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	inviteLink := buildInviteLink(h.inviteLinkBaseURL(), code.Code)

	response.Success(c, dto.ReferralInfoResponse{
		Enabled:                true,
		Code:                   code.Code,
		InviteLink:             inviteLink,
		RewardUSD:              h.cfg.Referral.RewardUSD,
		QualifiedRechargeCNY:   h.cfg.Referral.QualifiedRechargeCNY,
		QualifiedRechargeUSD:   h.cfg.Referral.QualifiedRechargeUSD,
		QualifiedRechargeTypes: h.cfg.Referral.QualifiedRechargeTypes,
		MaxInviteesPerUser:     h.cfg.Referral.MaxInviteesPerUser,
		Stats:                  dto.ReferralStatsFromService(stats),
	})
}

// ListInvitees lists current user's invitees.
// GET /api/v1/user/referral/invitees?page=1&page_size=20
func (h *ReferralHandler) ListInvitees(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	if h.cfg == nil || !h.cfg.Referral.Enabled || h.referralService == nil {
		response.Error(c, 400, "referral is disabled")
		return
	}

	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}

	invites, result, err := h.referralService.ListInvites(c.Request.Context(), subject.UserID, params)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.ReferralInvite, 0, len(invites))
	for i := range invites {
		out = append(out, *dto.ReferralInviteFromService(&invites[i]))
	}
	response.Paginated(c, out, result.Total, page, pageSize)
}

func (h *ReferralHandler) inviteLinkBaseURL() string {
	if h == nil || h.cfg == nil {
		return "/register"
	}
	base := strings.TrimSpace(h.cfg.Referral.LinkBaseURL)
	if base == "" {
		base = "/register"
	}

	// Prefer referral.base_url for invite link output when link_base_url is relative.
	if strings.HasPrefix(base, "/") {
		host := strings.TrimRight(strings.TrimSpace(h.cfg.Referral.BaseURL), "/")
		if host != "" {
			return host + base
		}
	}

	// Backwards compatibility: fall back to payment.base_url when referral.base_url is not set.
	if strings.HasPrefix(base, "/") {
		host := strings.TrimRight(strings.TrimSpace(h.cfg.Payment.BaseURL), "/")
		if host != "" {
			return host + base
		}
	}
	return base
}

func buildInviteLink(base string, code string) string {
	base = strings.TrimSpace(base)
	code = strings.TrimSpace(code)
	if base == "" || code == "" {
		return ""
	}
	u, err := url.Parse(base)
	if err != nil {
		// Basic fallback, avoid breaking the endpoint.
		sep := "?"
		if strings.Contains(base, "?") {
			sep = "&"
		}
		return fmt.Sprintf("%s%sinviter=%s", base, sep, url.QueryEscape(code))
	}
	q := u.Query()
	q.Set("inviter", code)
	u.RawQuery = q.Encode()
	return u.String()
}
