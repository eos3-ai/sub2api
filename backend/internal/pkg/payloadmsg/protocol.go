package payloadmsg

import "strings"

// DetectProtocol identifies the client-facing protocol from the request path.
// Protocol conversion upstream is intentionally irrelevant here.
func DetectProtocol(path string) Protocol {
	p := strings.ToLower(strings.TrimSuffix(strings.SplitN(path, "?", 2)[0], "/"))
	switch {
	case strings.HasSuffix(p, "/messages"), strings.HasSuffix(p, "/messages/count_tokens"):
		return ProtocolAnthropic
	case strings.HasSuffix(p, "/chat/completions"):
		return ProtocolOpenAIChat
	case strings.HasSuffix(p, "/responses") || strings.Contains(p, "/responses/"):
		return ProtocolOpenAIResponses
	case (strings.Contains(p, "/v1beta/models/") || strings.Contains(p, "/antigravity/v1beta/models/")) &&
		(strings.HasSuffix(p, ":generatecontent") || strings.HasSuffix(p, ":streamgeneratecontent")):
		return ProtocolGemini
	default:
		return ProtocolUnknown
	}
}
