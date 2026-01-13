package service

import "context"

// APIKeyAuthCacheInvalidator defines an interface for invalidating API key auth cache.
//
// It is intentionally small so that *APIKeyService can implement it without exposing internal details.
type APIKeyAuthCacheInvalidator interface {
	InvalidateAuthCacheByKey(ctx context.Context, key string)
	InvalidateAuthCacheByUserID(ctx context.Context, userID int64)
	InvalidateAuthCacheByGroupID(ctx context.Context, groupID int64)
}

