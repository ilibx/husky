package agent

import (
	"encoding/json"
	"fmt"
	"strings"
)

// parseReActAction 把模型输出解析成 ReActAction。
//
// 模型并不总是只返回一段裸 JSON：常见的是套一层 ```json 代码块，或者在 JSON
// 前面补一句 "Here is my action:"。直接对整段输出 json.Unmarshal 会解析失败，
// 循环于是重试、再失败，直到轮数耗尽，然后这一步被当成「已完成」（实际什么
// 都没做）。这里做三件事：剥代码块围栏、截取第一个完整的 JSON 对象、
// 校验 action 字段非空。
func parseReActAction(raw string) (ReActAction, error) {
	var action ReActAction
	text := strings.TrimSpace(raw)
	if text == "" {
		return action, fmt.Errorf("empty response")
	}

	text = stripCodeFence(text)
	payload := extractFirstJSONObject(text)
	if payload == "" {
		return action, fmt.Errorf("no JSON object found")
	}

	if err := json.Unmarshal([]byte(payload), &action); err != nil {
		return action, fmt.Errorf("invalid JSON: %w", err)
	}
	action.Action = strings.TrimSpace(action.Action)
	if action.Action == "" {
		return action, fmt.Errorf("missing action field")
	}
	return action, nil
}

// stripCodeFence 去掉 ```json ... ``` / ``` ... ``` 包裹。
func stripCodeFence(s string) string {
	if !strings.HasPrefix(s, "```") {
		return s
	}
	idx := strings.IndexByte(s, '\n')
	if idx < 0 {
		return s
	}
	s = s[idx+1:]
	if end := strings.LastIndex(s, "```"); end >= 0 {
		s = s[:end]
	}
	return strings.TrimSpace(s)
}

// maxReActRawEchoLength 是回灌解析失败原文的长度上限，
// 避免一次超长输出把整窗上下文挤掉。
const maxReActRawEchoLength = 500

// truncateForEcho 截断要回灌给模型的原文。
func truncateForEcho(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxReActRawEchoLength {
		return s
	}
	return s[:maxReActRawEchoLength] + "...(truncated)"
}

// extractFirstJSONObject 从任意文本里截出第一个花括号平衡的 JSON 对象。
// 会跳过字符串字面量里的花括号与转义引号，因此 action_input 里含 JSON 也能正确截断。
func extractFirstJSONObject(s string) string {
	start := strings.IndexByte(s, '{')
	if start < 0 {
		return ""
	}
	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if inString {
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return ""
}
