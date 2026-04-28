package service

import (
	"context"
	"strings"
)

func (s *UserService) GetProfileIdentitySummaries(ctx context.Context, userID int64, user *User) (UserIdentitySummarySet, error) {
	if user == nil {
		loaded, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			return UserIdentitySummarySet{}, err
		}
		user = loaded
	}

	records, err := s.userRepo.ListUserAuthIdentities(ctx, userID)
	if err != nil {
		return UserIdentitySummarySet{}, err
	}

	summaries := UserIdentitySummarySet{
		Email:   s.buildEmailIdentitySummary(user, records),
		LinuxDo: s.buildProviderIdentitySummary("linuxdo", user, records),
		OIDC:    s.buildProviderIdentitySummary("oidc", user, records),
		WeChat:  s.buildProviderIdentitySummary("wechat", user, records),
	}

	if !s.identityProviderEnabled(ctx, SettingKeyLinuxDoConnectEnabled) && !summaries.LinuxDo.Bound {
		summaries.LinuxDo.CanBind = false
		summaries.LinuxDo.BindStartPath = ""
	}
	if !s.identityProviderEnabled(ctx, SettingKeyOIDCConnectEnabled) && !summaries.OIDC.Bound {
		summaries.OIDC.CanBind = false
		summaries.OIDC.BindStartPath = ""
	}
	if !s.identityProviderEnabled(ctx, SettingKeyWeChatConnectEnabled) && !summaries.WeChat.Bound {
		summaries.WeChat.CanBind = false
		summaries.WeChat.BindStartPath = ""
	}

	return summaries, nil
}

func (s *UserService) PrepareIdentityBindingStart(ctx context.Context, req StartUserIdentityBindingRequest) (*StartUserIdentityBindingResult, error) {
	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	switch provider {
	case "linuxdo":
		if !s.identityProviderEnabled(ctx, SettingKeyLinuxDoConnectEnabled) {
			return nil, ErrIdentityProviderInvalid
		}
	case "oidc":
		if !s.identityProviderEnabled(ctx, SettingKeyOIDCConnectEnabled) {
			return nil, ErrIdentityProviderInvalid
		}
	case "wechat":
		if !s.identityProviderEnabled(ctx, SettingKeyWeChatConnectEnabled) {
			return nil, ErrIdentityProviderInvalid
		}
	default:
		return nil, ErrIdentityProviderInvalid
	}

	authorizeURL, err := buildUserIdentityBindAuthorizeURL(provider, req.RedirectTo)
	if err != nil {
		return nil, err
	}
	return &StartUserIdentityBindingResult{
		Provider:           provider,
		AuthorizeURL:       authorizeURL,
		Method:             "GET",
		UseBrowserRedirect: true,
	}, nil
}

func (s *UserService) UnbindUserAuthProvider(ctx context.Context, userID int64, provider string) (*User, error) {
	user, _, err := s.UnbindUserAuthProviderWithResult(ctx, userID, provider)
	return user, err
}

func (s *UserService) UnbindUserAuthProviderWithResult(ctx context.Context, userID int64, provider string) (*User, bool, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	switch provider {
	case "linuxdo", "oidc", "wechat":
	default:
		return nil, false, ErrIdentityProviderInvalid
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, false, err
	}
	summaries, err := s.GetProfileIdentitySummaries(ctx, userID, user)
	if err != nil {
		return nil, false, err
	}

	canUnbind := false
	switch provider {
	case "linuxdo":
		canUnbind = summaries.LinuxDo.CanUnbind
	case "oidc":
		canUnbind = summaries.OIDC.CanUnbind
	case "wechat":
		canUnbind = summaries.WeChat.CanUnbind
	}
	if !canUnbind {
		return nil, false, ErrIdentityUnbindLastMethod
	}

	if err := s.userRepo.UnbindUserAuthProvider(ctx, userID, provider); err != nil {
		return nil, false, err
	}
	if s.authCacheInvalidator != nil {
		s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, userID)
	}

	updatedUser, err := s.GetProfile(ctx, userID)
	if err != nil {
		return nil, false, err
	}
	return updatedUser, true, nil
}

func (s *UserService) identityProviderEnabled(ctx context.Context, key string) bool {
	if s == nil || s.settingRepo == nil {
		return true
	}
	value, err := s.settingRepo.GetValue(ctx, key)
	if err != nil {
		return true
	}
	return strings.TrimSpace(strings.ToLower(value)) != "false"
}
