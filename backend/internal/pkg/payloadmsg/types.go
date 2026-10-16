// Package payloadmsg normalizes gateway payloads into a protocol-independent
// messages representation suitable for payload logging.
package payloadmsg

import "encoding/json"

const BinaryOmitted = "<binary data omitted>"

type Protocol string

const (
	ProtocolUnknown         Protocol = ""
	ProtocolAnthropic       Protocol = "anthropic"
	ProtocolOpenAIChat      Protocol = "openai_chat"
	ProtocolOpenAIResponses Protocol = "openai_responses"
	ProtocolGemini          Protocol = "gemini"
)

type ToolFunction struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type ToolCall struct {
	ID       string       `json:"id,omitempty"`
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type Message struct {
	Role             string     `json:"role"`
	Content          any        `json:"content,omitempty"`
	Name             string     `json:"name,omitempty"`
	ToolCallID       string     `json:"tool_call_id,omitempty"`
	ToolCalls        []ToolCall `json:"tool_calls,omitempty"`
	ReasoningContent string     `json:"reasoning_content,omitempty"`
	Refusal          string     `json:"refusal,omitempty"`
}

type Envelope struct {
	Messages []Message `json:"messages"`
}

func Marshal(messages []Message) ([]byte, error) {
	if messages == nil {
		messages = []Message{}
	}
	return json.Marshal(Envelope{Messages: messages})
}
