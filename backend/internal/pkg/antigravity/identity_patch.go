package antigravity

// GetDefaultIdentityPatch returns the built-in identity patch prompt used for Antigravity upstreams.
//
// It is used by health checks / test requests that do not carry a model name context.
func GetDefaultIdentityPatch() string {
	// Use a generic model name here; the prompt is mainly used to satisfy upstream requirements.
	return defaultIdentityPatch("claude")
}

