package payloadmsg

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"mime"
	"sort"
	"strconv"
	"strings"
)

func NormalizeRequest(protocol Protocol, contentType string, raw []byte) ([]byte, error) {
	if !isJSON(contentType) || len(bytes.TrimSpace(raw)) == 0 {
		return Marshal(nil)
	}
	var root any
	if err := json.Unmarshal(raw, &root); err != nil {
		return Marshal(nil)
	}
	m, _ := root.(map[string]any)
	var messages []Message
	switch protocol {
	case ProtocolAnthropic:
		messages = anthropicRequest(m)
	case ProtocolOpenAIChat:
		messages = chatRequest(m)
	case ProtocolOpenAIResponses:
		messages = responsesRequest(m)
	case ProtocolGemini:
		messages = geminiRequest(m)
	}
	return Marshal(messages)
}

func NormalizeResponse(protocol Protocol, contentType string, raw []byte) ([]byte, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return Marshal(nil)
	}
	if strings.Contains(strings.ToLower(contentType), "text/event-stream") {
		return Marshal(normalizeSSE(protocol, raw))
	}
	var root any
	if err := json.Unmarshal(raw, &root); err != nil {
		return Marshal(nil)
	}
	if e := errorMessage(root); e != "" {
		return Marshal([]Message{{Role: "error", Content: clean(e)}})
	}
	var messages []Message
	switch protocol {
	case ProtocolAnthropic:
		messages = anthropicResponse(asMap(root))
	case ProtocolOpenAIChat:
		messages = chatResponse(asMap(root))
	case ProtocolOpenAIResponses:
		messages = responseItems(asSlice(asMap(root)["output"]))
	case ProtocolGemini:
		messages = geminiResponse(root)
	}
	return Marshal(messages)
}

func isJSON(contentType string) bool {
	if contentType == "" {
		return true
	}
	mt, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		mt = strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0])
	}
	return mt == "application/json" || strings.HasSuffix(mt, "+json")
}

func clean(s string) string      { return strings.ReplaceAll(s, "\x00", "") }
func asMap(v any) map[string]any { m, _ := v.(map[string]any); return m }
func asSlice(v any) []any        { a, _ := v.([]any); return a }
func str(v any) string           { s, _ := v.(string); return clean(s) }

func jsonString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return clean(string(b))
}

func role(s string) string {
	s = strings.ToLower(s)
	if s == "developer" {
		return "system"
	}
	if s == "model" {
		return "assistant"
	}
	switch s {
	case "system", "user", "assistant", "tool", "error":
		return s
	}
	return s
}

func errorMessage(v any) string {
	m := asMap(v)
	if e := m["error"]; e != nil {
		if s := str(e); s != "" {
			return s
		}
		if s := str(asMap(e)["message"]); s != "" {
			return s
		}
	}
	return str(m["message"])
}

func contentValue(v any) any {
	switch x := v.(type) {
	case string:
		return clean(x)
	case []any:
		parts, textOnly, text := contentParts(x)
		if textOnly {
			return text
		}
		if len(parts) > 0 {
			return parts
		}
	}
	return nil
}

func contentParts(items []any) ([]any, bool, string) {
	parts := make([]any, 0, len(items))
	var text strings.Builder
	textOnly := true
	for _, item := range items {
		m := asMap(item)
		typ := str(m["type"])
		switch typ {
		case "text", "input_text", "output_text":
			t := str(m["text"])
			text.WriteString(t)
			parts = append(parts, map[string]any{"type": "text", "text": t})
		case "image", "image_url", "input_image":
			textOnly = false
			u := binaryURL(m["url"])
			if typ == "image_url" {
				u = binaryURL(asMap(m["image_url"])["url"])
				if u == "" {
					u = binaryURL(m["image_url"])
				}
			}
			parts = append(parts, map[string]any{"type": "image", "url": u})
		case "document", "file", "input_file", "input_audio", "audio", "inlineData", "inline_data", "fileData", "file_data":
			textOnly = false
			parts = append(parts, map[string]any{"type": typ, "data": BinaryOmitted})
		default:
			if t := str(m["text"]); t != "" {
				text.WriteString(t)
				parts = append(parts, map[string]any{"type": "text", "text": t})
			}
		}
	}
	return parts, textOnly, text.String()
}

func binaryURL(v any) string {
	s := str(v)
	if strings.HasPrefix(strings.ToLower(s), "http://") || strings.HasPrefix(strings.ToLower(s), "https://") {
		return s
	}
	return BinaryOmitted
}

func parseToolCalls(v any) []ToolCall {
	var out []ToolCall
	for _, x := range asSlice(v) {
		m := asMap(x)
		f := asMap(m["function"])
		name := str(f["name"])
		args := f["arguments"]
		if name == "" {
			name = str(m["name"])
			args = m["input"]
		}
		argstr := str(args)
		if argstr == "" && args != nil {
			argstr = jsonString(args)
		}
		out = append(out, ToolCall{ID: str(m["id"]), Type: "function", Function: ToolFunction{Name: name, Arguments: argstr}})
	}
	return out
}

func anthropicRequest(root map[string]any) []Message {
	var out []Message
	if s := anthropicBlocks(root["system"], "system"); s != nil {
		out = append(out, *s)
	}
	for _, x := range asSlice(root["messages"]) {
		m := asMap(x)
		r := role(str(m["role"]))
		blocks, ok := m["content"].([]any)
		if !ok {
			if s := str(m["content"]); s != "" {
				out = append(out, Message{Role: r, Content: s})
			}
			continue
		}
		msg := Message{Role: r}
		var texts strings.Builder
		var multimodal []any
		hasMedia := false
		for _, b := range blocks {
			bm := asMap(b)
			switch str(bm["type"]) {
			case "text":
				t := str(bm["text"])
				texts.WriteString(t)
				multimodal = append(multimodal, map[string]any{"type": "text", "text": t})
			case "thinking", "redacted_thinking":
				msg.ReasoningContent += str(bm["thinking"])
			case "tool_use":
				args := jsonString(bm["input"])
				msg.ToolCalls = append(msg.ToolCalls, ToolCall{ID: str(bm["id"]), Type: "function", Function: ToolFunction{Name: str(bm["name"]), Arguments: args}})
			case "tool_result":
				out = appendMessage(out, msg, texts.String(), multimodal, hasMedia)
				msg = Message{Role: r}
				texts.Reset()
				multimodal = nil
				hasMedia = false
				out = append(out, Message{Role: "tool", ToolCallID: str(bm["tool_use_id"]), Content: contentValue(bm["content"])})
			case "image", "document":
				hasMedia = true
				multimodal = append(multimodal, map[string]any{"type": str(bm["type"]), "data": BinaryOmitted})
			}
		}
		out = appendMessage(out, msg, texts.String(), multimodal, hasMedia)
	}
	return out
}

func anthropicBlocks(v any, r string) *Message {
	if s := str(v); s != "" {
		return &Message{Role: r, Content: s}
	}
	a := asSlice(v)
	if len(a) == 0 {
		return nil
	}
	parts, only, text := contentParts(a)
	m := Message{Role: r}
	if only {
		m.Content = text
	} else {
		m.Content = parts
	}
	return &m
}

func appendMessage(out []Message, m Message, text string, parts []any, media bool) []Message {
	if media {
		m.Content = parts
	} else if text != "" {
		m.Content = text
	}
	if m.Content != nil || m.ReasoningContent != "" || len(m.ToolCalls) > 0 || m.Refusal != "" {
		return append(out, m)
	}
	return out
}

func chatRequest(root map[string]any) []Message {
	var out []Message
	for _, x := range asSlice(root["messages"]) {
		if m := chatMessage(asMap(x)); m != nil {
			out = append(out, *m)
		}
	}
	return out
}
func chatMessage(m map[string]any) *Message {
	r := role(str(m["role"]))
	if r == "" {
		return nil
	}
	x := Message{Role: r, Name: str(m["name"]), ToolCallID: str(m["tool_call_id"]), ReasoningContent: str(m["reasoning_content"]), Refusal: str(m["refusal"]), ToolCalls: parseToolCalls(m["tool_calls"])}
	x.Content = contentValue(m["content"])
	return &x
}

func responsesRequest(root map[string]any) []Message {
	var out []Message
	if s := str(root["instructions"]); s != "" {
		out = append(out, Message{Role: "system", Content: s})
	}
	if s := str(root["input"]); s != "" {
		out = append(out, Message{Role: "user", Content: s})
	} else {
		out = append(out, responseItems(asSlice(root["input"]))...)
	}
	return out
}

func responseItems(items []any) []Message {
	var out []Message
	for _, x := range items {
		m := asMap(x)
		typ := str(m["type"])
		switch typ {
		case "message":
			msg := Message{Role: role(str(m["role"])), Content: contentValue(m["content"]), Refusal: str(m["refusal"])}
			if msg.Role == "" {
				msg.Role = "assistant"
			}
			out = appendOrMergeAssistant(out, msg)
		case "function_call", "custom_tool_call":
			name := str(m["name"])
			args := str(m["arguments"])
			if args == "" && m["input"] != nil {
				args = jsonString(m["input"])
			}
			msg := Message{Role: "assistant", ToolCalls: []ToolCall{{ID: first(str(m["call_id"]), str(m["id"])), Type: "function", Function: ToolFunction{Name: name, Arguments: args}}}}
			out = appendOrMergeAssistant(out, msg)
		case "function_call_output", "custom_tool_call_output":
			out = append(out, Message{Role: "tool", ToolCallID: first(str(m["call_id"]), str(m["id"])), Content: contentValue(m["output"])})
		case "reasoning":
			var b strings.Builder
			for _, s := range asSlice(m["summary"]) {
				b.WriteString(str(asMap(s)["text"]))
			}
			if b.Len() == 0 {
				b.WriteString(str(m["text"]))
			}
			out = appendOrMergeAssistant(out, Message{Role: "assistant", ReasoningContent: b.String()})
		}
	}
	return out
}

func appendOrMergeAssistant(out []Message, m Message) []Message {
	if len(out) > 0 && out[len(out)-1].Role == "assistant" && m.Role == "assistant" {
		d := &out[len(out)-1]
		d.ReasoningContent += m.ReasoningContent
		d.Refusal += m.Refusal
		d.ToolCalls = append(d.ToolCalls, m.ToolCalls...)
		if s := str(m.Content); s != "" {
			if old := str(d.Content); old != "" {
				d.Content = old + s
			} else {
				d.Content = s
			}
		}
		return out
	}
	return appendMessage(out, m, str(m.Content), nil, false)
}

func geminiRequest(root map[string]any) []Message {
	var out []Message
	if m := geminiContent(asMap(root["systemInstruction"]), "system", 0); m != nil {
		out = append(out, *m)
	}
	for i, x := range asSlice(root["contents"]) {
		if m := geminiContent(asMap(x), role(str(asMap(x)["role"])), i); m != nil {
			out = append(out, *m)
		}
	}
	return out
}
func geminiResponse(root any) []Message {
	var out []Message
	roots := []any{root}
	if a, ok := root.([]any); ok {
		roots = a
	}
	for _, r := range roots {
		for i, c := range asSlice(asMap(r)["candidates"]) {
			cm := asMap(c)
			if m := geminiContent(asMap(cm["content"]), "assistant", i); m != nil {
				out = append(out, *m)
			}
		}
	}
	return out
}
func geminiContent(c map[string]any, r string, index int) *Message {
	if len(c) == 0 {
		return nil
	}
	m := Message{Role: r}
	var text strings.Builder
	for i, p := range asSlice(c["parts"]) {
		pm := asMap(p)
		if t := str(pm["text"]); t != "" {
			if b, ok := pm["thought"].(bool); ok && b {
				m.ReasoningContent += t
			} else {
				text.WriteString(t)
			}
		}
		if fc := asMap(pm["functionCall"]); len(fc) > 0 {
			m.ToolCalls = append(m.ToolCalls, ToolCall{ID: first(str(fc["id"]), fmt.Sprintf("%s-%d-%d", str(fc["name"]), index, i)), Type: "function", Function: ToolFunction{Name: str(fc["name"]), Arguments: jsonString(fc["args"])}})
		}
		if fr := asMap(pm["functionResponse"]); len(fr) > 0 {
			return &Message{Role: "tool", Name: str(fr["name"]), Content: jsonString(fr["response"])}
		}
		if pm["inlineData"] != nil || pm["inline_data"] != nil || pm["fileData"] != nil || pm["file_data"] != nil {
			m.Content = []any{map[string]any{"type": "file", "data": BinaryOmitted}}
		}
	}
	if text.Len() > 0 {
		m.Content = text.String()
	}
	if m.Content == nil && m.ReasoningContent == "" && len(m.ToolCalls) == 0 {
		return nil
	}
	return &m
}

func anthropicResponse(root map[string]any) []Message {
	m := Message{Role: "assistant"}
	var text strings.Builder
	for _, x := range asSlice(root["content"]) {
		b := asMap(x)
		switch str(b["type"]) {
		case "text":
			text.WriteString(str(b["text"]))
		case "thinking", "redacted_thinking":
			m.ReasoningContent += str(b["thinking"])
		case "tool_use":
			m.ToolCalls = append(m.ToolCalls, ToolCall{ID: str(b["id"]), Type: "function", Function: ToolFunction{Name: str(b["name"]), Arguments: jsonString(b["input"])}})
		}
	}
	return appendMessage(nil, m, text.String(), nil, false)
}
func chatResponse(root map[string]any) []Message {
	var out []Message
	for _, x := range asSlice(root["choices"]) {
		if m := chatMessage(asMap(asMap(x)["message"])); m != nil {
			out = append(out, *m)
		}
	}
	return out
}

type sseEvent struct {
	Event string
	Data  []byte
}

func parseSSE(raw []byte) []sseEvent {
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 64*1024), 512*1024*1024)
	var out []sseEvent
	cur := sseEvent{}
	flush := func() {
		d := bytes.TrimSpace(cur.Data)
		if len(d) > 0 && !bytes.Equal(d, []byte("[DONE]")) {
			cur.Data = append([]byte(nil), d...)
			out = append(out, cur)
		}
		cur = sseEvent{}
	}
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			flush()
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		if strings.HasPrefix(line, "event:") {
			cur.Event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}
		if strings.HasPrefix(line, "data:") {
			if len(cur.Data) > 0 {
				cur.Data = append(cur.Data, '\n')
			}
			cur.Data = append(cur.Data, strings.TrimSpace(strings.TrimPrefix(line, "data:"))...)
		}
	}
	flush()
	return out
}
func normalizeSSE(p Protocol, raw []byte) []Message {
	events := parseSSE(raw)
	switch p {
	case ProtocolAnthropic:
		return anthropicSSE(events)
	case ProtocolOpenAIChat:
		return chatSSE(events)
	case ProtocolOpenAIResponses:
		return responsesSSE(events)
	case ProtocolGemini:
		var chunks []any
		for _, e := range events {
			var v any
			if json.Unmarshal(e.Data, &v) == nil {
				chunks = append(chunks, v)
			}
		}
		return geminiResponse(chunks)
	}
	return nil
}

func anthropicSSE(events []sseEvent) []Message {
	type block struct{ typ, id, name, text, reason, args string }
	blocks := map[int]*block{}
	var errs []Message
	for _, e := range events {
		var m map[string]any
		if json.Unmarshal(e.Data, &m) != nil {
			continue
		}
		typ := first(e.Event, str(m["type"]))
		if typ == "error" {
			if x := errorMessage(m); x != "" {
				errs = append(errs, Message{Role: "error", Content: x})
			}
			continue
		}
		idx := intNum(m["index"])
		if typ == "content_block_start" {
			b := asMap(m["content_block"])
			blocks[idx] = &block{typ: str(b["type"]), id: str(b["id"]), name: str(b["name"]), text: str(b["text"])}
			continue
		}
		d := asMap(m["delta"])
		b := blocks[idx]
		if b == nil {
			b = &block{}
			blocks[idx] = b
		}
		switch str(d["type"]) {
		case "text_delta":
			b.text += str(d["text"])
		case "thinking_delta":
			b.reason += str(d["thinking"])
		case "input_json_delta":
			b.args += str(d["partial_json"])
		}
	}
	keys := make([]int, 0, len(blocks))
	for k := range blocks {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	msg := Message{Role: "assistant"}
	var text strings.Builder
	for _, k := range keys {
		b := blocks[k]
		text.WriteString(b.text)
		msg.ReasoningContent += b.reason
		if b.typ == "tool_use" {
			msg.ToolCalls = append(msg.ToolCalls, ToolCall{ID: b.id, Type: "function", Function: ToolFunction{Name: b.name, Arguments: b.args}})
		}
	}
	out := appendMessage(nil, msg, text.String(), nil, false)
	return append(out, errs...)
}
func chatSSE(events []sseEvent) []Message {
	type call struct{ id, name, args string }
	type choice struct {
		role, text, reason, refusal string
		calls                       map[int]*call
	}
	cs := map[int]*choice{}
	var errs []Message
	for _, e := range events {
		var root map[string]any
		if json.Unmarshal(e.Data, &root) != nil {
			continue
		}
		if x := errorMessage(root); x != "" {
			errs = append(errs, Message{Role: "error", Content: x})
			continue
		}
		for _, x := range asSlice(root["choices"]) {
			m := asMap(x)
			i := intNum(m["index"])
			c := cs[i]
			if c == nil {
				c = &choice{calls: map[int]*call{}}
				cs[i] = c
			}
			d := asMap(m["delta"])
			c.role = first(c.role, str(d["role"]))
			c.text += str(d["content"])
			c.reason += first(str(d["reasoning_content"]), str(d["reasoning"]))
			c.refusal += str(d["refusal"])
			for j, t := range asSlice(d["tool_calls"]) {
				tm := asMap(t)
				ti := intNum(tm["index"])
				if _, ok := tm["index"]; !ok {
					ti = j
				}
				tc := c.calls[ti]
				if tc == nil {
					tc = &call{}
					c.calls[ti] = tc
				}
				f := asMap(tm["function"])
				tc.id = first(tc.id, str(tm["id"]))
				tc.name = first(tc.name, str(f["name"]))
				tc.args += str(f["arguments"])
			}
		}
	}
	keys := make([]int, 0, len(cs))
	for k := range cs {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	var out []Message
	for _, k := range keys {
		c := cs[k]
		m := Message{Role: first(c.role, "assistant"), Content: c.text, ReasoningContent: c.reason, Refusal: c.refusal}
		tk := make([]int, 0, len(c.calls))
		for i := range c.calls {
			tk = append(tk, i)
		}
		sort.Ints(tk)
		for _, i := range tk {
			x := c.calls[i]
			m.ToolCalls = append(m.ToolCalls, ToolCall{ID: x.id, Type: "function", Function: ToolFunction{Name: x.name, Arguments: x.args}})
		}
		out = appendMessage(out, m, c.text, nil, false)
	}
	return append(out, errs...)
}
func responsesSSE(events []sseEvent) []Message {
	var completed []Message
	items := map[int]map[string]any{}
	var errs []Message
	for _, e := range events {
		var m map[string]any
		if json.Unmarshal(e.Data, &m) != nil {
			continue
		}
		typ := first(e.Event, str(m["type"]))
		if typ == "response.completed" {
			completed = responseItems(asSlice(asMap(m["response"])["output"]))
			continue
		}
		if typ == "error" || typ == "response.failed" || typ == "response.incomplete" {
			if x := errorMessage(firstAny(m["error"], asMap(m["response"])["error"], m)); x != "" {
				errs = append(errs, Message{Role: "error", Content: x})
			}
			continue
		}
		i := intNum(m["output_index"])
		if typ == "response.output_item.added" {
			items[i] = asMap(m["item"])
			continue
		}
		it := items[i]
		if it == nil {
			it = map[string]any{"type": "message", "role": "assistant"}
			items[i] = it
		}
		switch typ {
		case "response.output_text.delta":
			it["_text"] = str(it["_text"]) + str(m["delta"])
		case "response.reasoning_summary_text.delta", "response.reasoning_text.delta":
			it["type"] = "reasoning"
			it["_reason"] = str(it["_reason"]) + str(m["delta"])
		case "response.refusal.delta":
			it["_refusal"] = str(it["_refusal"]) + str(m["delta"])
		case "response.function_call_arguments.delta":
			it["type"] = "function_call"
			it["arguments"] = str(it["arguments"]) + str(m["delta"])
			it["call_id"] = first(str(it["call_id"]), str(m["item_id"]))
		}
	}
	if completed != nil {
		return append(completed, errs...)
	}
	keys := make([]int, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	var normalized []any
	for _, k := range keys {
		it := items[k]
		switch str(it["type"]) {
		case "message":
			it["content"] = []any{map[string]any{"type": "output_text", "text": str(it["_text"])}}
			it["refusal"] = str(it["_refusal"])
		case "reasoning":
			it["text"] = str(it["_reason"])
		}
		normalized = append(normalized, it)
	}
	return append(responseItems(normalized), errs...)
}
func first(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
func firstAny(values ...any) any {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return nil
}
func intNum(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, _ := strconv.Atoi(n.String())
		return i
	}
	return 0
}
