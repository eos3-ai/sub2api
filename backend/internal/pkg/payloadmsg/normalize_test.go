package payloadmsg

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func decodeEnvelope(t *testing.T, raw []byte) Envelope {
	t.Helper()
	var got Envelope
	require.NoError(t, json.Unmarshal(raw, &got))
	return got
}

func TestDetectProtocol(t *testing.T) {
	tests := map[string]Protocol{
		"/v1/messages":                 ProtocolAnthropic,
		"/antigravity/v1/messages":     ProtocolAnthropic,
		"/v1/chat/completions":         ProtocolOpenAIChat,
		"/backend-api/codex/responses": ProtocolOpenAIResponses,
		"/v1beta/models/gemini-2.5-pro:streamGenerateContent?alt=sse": ProtocolGemini,
		"/v1/embeddings": ProtocolUnknown,
	}
	for path, want := range tests {
		t.Run(path, func(t *testing.T) { require.Equal(t, want, DetectProtocol(path)) })
	}
}

func TestAnthropicRequest(t *testing.T) {
	raw := []byte(`{"system":[{"type":"text","text":"sys"}],"messages":[{"role":"user","content":[{"type":"text","text":"hello"},{"type":"image","source":{"type":"base64","data":"secret"}}]},{"role":"assistant","content":[{"type":"thinking","thinking":"why"},{"type":"tool_use","id":"call-1","name":"lookup","input":{"q":"x"}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"call-1","content":"result"}]}]}`)
	b, err := NormalizeRequest(ProtocolAnthropic, "application/json", raw)
	require.NoError(t, err)
	got := decodeEnvelope(t, b)
	require.Len(t, got.Messages, 4)
	require.Equal(t, "sys", got.Messages[0].Content)
	require.Equal(t, "why", got.Messages[2].ReasoningContent)
	require.Equal(t, "lookup", got.Messages[2].ToolCalls[0].Function.Name)
	require.Equal(t, "tool", got.Messages[3].Role)
	require.Equal(t, "call-1", got.Messages[3].ToolCallID)
	require.NotContains(t, string(b), "secret")
}

func TestOpenAIChatStream(t *testing.T) {
	raw := []byte("data: {\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"reasoning_content\":\"think\",\"content\":\"hel\"}}]}\n\n" +
		"data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"lo\",\"tool_calls\":[{\"index\":0,\"id\":\"c1\",\"function\":{\"name\":\"f\",\"arguments\":\"{\\\"x\\\":\"}}]}}]}\n\n" +
		"data: {\"choices\":[{\"index\":0,\"delta\":{\"tool_calls\":[{\"index\":0,\"function\":{\"arguments\":\"1}\"}}]}}]}\n\n" +
		"data: [DONE]\n\n")
	b, err := NormalizeResponse(ProtocolOpenAIChat, "text/event-stream", raw)
	require.NoError(t, err)
	got := decodeEnvelope(t, b)
	require.Len(t, got.Messages, 1)
	require.Equal(t, "hello", got.Messages[0].Content)
	require.Equal(t, "think", got.Messages[0].ReasoningContent)
	require.Equal(t, `{"x":1}`, got.Messages[0].ToolCalls[0].Function.Arguments)
}

func TestResponsesCompletedStreamWins(t *testing.T) {
	raw := []byte("event: response.output_text.delta\ndata: {\"type\":\"response.output_text.delta\",\"output_index\":0,\"delta\":\"partial\"}\n\n" +
		"event: response.completed\ndata: {\"type\":\"response.completed\",\"response\":{\"output\":[{\"type\":\"message\",\"role\":\"assistant\",\"content\":[{\"type\":\"output_text\",\"text\":\"complete\"}]}]}}\n\n")
	b, err := NormalizeResponse(ProtocolOpenAIResponses, "text/event-stream", raw)
	require.NoError(t, err)
	got := decodeEnvelope(t, b)
	require.Equal(t, "complete", got.Messages[0].Content)
}

func TestInvalidAndBinaryBodiesBecomeEmpty(t *testing.T) {
	for _, tc := range []struct {
		ct  string
		raw []byte
	}{{"application/json", []byte("{")}, {"multipart/form-data; boundary=x", []byte("anything")}} {
		b, err := NormalizeRequest(ProtocolOpenAIChat, tc.ct, tc.raw)
		require.NoError(t, err)
		require.JSONEq(t, `{"messages":[]}`, string(b))
	}
}
