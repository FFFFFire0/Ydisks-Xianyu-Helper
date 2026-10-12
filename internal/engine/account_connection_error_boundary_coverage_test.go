package engine

import (
	"context"
	"errors"
	"testing"

	"xianyu-go/internal/xianyu/ws"
)

// TestHandleWSConnectFailureClassifiesAllConnectionErrors 验证握手失败按服务端错误类型选择用户可见原因：
//   - 平台明确拒绝凭证（限流、token 失效、鉴权失败）→ 状态置为 AuthExpired 并返回原始错误；
//   - 其它传输层瞬时错误（DNS、拨号、网络等）→ 状态置为 Reconnecting 并返回 nil，让主循环继续。
func TestHandleWSConnectFailureClassifiesAllConnectionErrors(t *testing.T) {
	// cases 保存每种握手错误及其对应的运行状态原因。
	// 标注 transient==true 的样本期望 handleWSConnectFailure 返回 nil 并把状态置为 Reconnecting；
	// 否则期望返回原始错误并把状态置为 AuthExpired。
	cases := []struct {
		name      string
		err       error
		reason    string
		transient bool
	}{
		{name: "connect-limit", err: &ws.RegError{Kind: ws.RegErrorConnectLimit, Reason: "session remove"}, reason: "消息会话已被服务端移除"},
		{name: "invalid-token", err: &ws.RegError{Kind: ws.RegErrorInvalidToken, Reason: "invalid token"}, reason: "消息凭证被拒绝，请重新登录"},
		{name: "authentication", err: &ws.RegError{Kind: ws.RegErrorAuthentication, Reason: "not auth"}, reason: "消息凭证被拒绝，请重新登录"},
		{name: "generic", err: errors.New("dial failed"), reason: "消息服务连接失败，正在重试", transient: true},
		{name: "transient-timeout", err: errors.New("i/o timeout"), reason: "消息服务连接失败，正在重试", transient: true},
	}
	// testCase 表示当前子测试使用的一种握手错误分类样本。
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			// account 是仅用于验证错误分类的本地账号。
			account := New(Config{CookieID: testCase.name, CookieStr: "unb=1"})
			// returnedErr 是握手失败处理后保留的错误。
			returnedErr := account.handleWSConnectFailure(context.Background(), testCase.err)
			if testCase.transient {
				// 瞬时错误：必须返回 nil，不得让调用方把主循环当失败处理。
				if returnedErr != nil {
					t.Fatalf("瞬时错误应返回 nil 让主循环继续：got=%v", returnedErr)
				}
			} else {
				// 凭证问题：必须把原始错误透传出去，触发上层重登流程。
				if !errors.Is(returnedErr, testCase.err) {
					t.Fatalf("应保留原始握手错误：got=%v want=%v", returnedErr, testCase.err)
				}
			}
			// runtimeStatus 保存错误分类后供用户界面读取的账号状态。
			runtimeStatus := account.RuntimeStatus()
			if testCase.transient {
				if runtimeStatus.State != RuntimeReconnecting || runtimeStatus.Message != testCase.reason {
					t.Fatalf("瞬时错误应置为 Reconnecting：status=%+v wantReason=%q", runtimeStatus, testCase.reason)
				}
			} else {
				if runtimeStatus.State != RuntimeAuthExpired || runtimeStatus.Message != testCase.reason {
					t.Fatalf("错误分类不正确：status=%+v wantReason=%q", runtimeStatus, testCase.reason)
				}
			}
		})
	}
}
