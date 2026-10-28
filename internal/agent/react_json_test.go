package agent

import (
	"encoding/json"
	"testing"
)

func TestParseReActAction(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    ReActAction
		wantErr bool
	}{
		{
			name: "裸 JSON",
			raw:  `{"thought":"t","action":"complete","action_input":"ok"}`,
			want: ReActAction{Thought: "t", Action: "complete", ActionInput: "ok"},
		},
		{
			name: "json 围栏",
			raw:  "```json\n{\"thought\":\"t\",\"action\":\"complete\",\"action_input\":\"ok\"}\n```",
			want: ReActAction{Thought: "t", Action: "complete", ActionInput: "ok"},
		},
		{
			name: "无语言标记的围栏",
			raw:  "```\n{\"action\":\"fail\",\"action_input\":\"nope\"}\n```",
			want: ReActAction{Action: "fail", ActionInput: "nope"},
		},
		{
			name: "JSON 前有解释文字",
			raw:  "Sure! Here is my action:\n{\"action\":\"complete\",\"action_input\":\"done\"}",
			want: ReActAction{Action: "complete", ActionInput: "done"},
		},
		{
			name: "action_input 里含 JSON",
			raw:  `{"action":"update_ticket","action_input":"{\"status\":\"resolved\"}"}`,
			want: ReActAction{Action: "update_ticket", ActionInput: `{"status":"resolved"}`},
		},
		{
			name: "action 前后有空格",
			raw:  `{"action":"  complete  ","action_input":"ok"}`,
			want: ReActAction{Action: "complete", ActionInput: "ok"},
		},
		{name: "空响应", raw: "   ", wantErr: true},
		{name: "完全没有 JSON", raw: "I think we should just reply to the user.", wantErr: true},
		{name: "JSON 语法错误", raw: `{"action": "complete",}`, wantErr: true},
		{name: "缺 action 字段", raw: `{"thought":"t","action_input":"ok"}`, wantErr: true},
		{name: "JSON 没闭合", raw: `{"action":"complete"`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseReActAction(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("期望报错，实际解析出 %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("不应报错: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestParseReActActionUnmarshalBaseline 固定住「直接 Unmarshal 整段输出」必然失败的
// 事实：这是本 PR 的动机，写成测试以免以后有人误以为可以简化回去。
func TestParseReActActionUnmarshalBaseline(t *testing.T) {
	for _, raw := range []string{
		"```json\n{\"action\":\"complete\"}\n```",
		"Here is my action:\n{\"action\":\"complete\"}",
	} {
		var action ReActAction
		if err := json.Unmarshal([]byte(raw), &action); err == nil {
			t.Fatalf("预期直接 Unmarshal 会失败: %q", raw)
		}
		if _, err := parseReActAction(raw); err != nil {
			t.Fatalf("parseReActAction 应能处理 %q: %v", raw, err)
		}
	}
}
