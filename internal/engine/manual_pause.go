package engine

import (
	"context"
	"time"
)

// recordManualTakeover 读取账号的人工接管配置，并为当前会话写入暂停窗口。
// 该方法只处理人工发送产生的自身回显；自动化发送的回显已由发送确认器消费。
func (a *Account) recordManualTakeover(ctx context.Context, message OutgoingChatMessage) {
	if a == nil || a.store == nil || a.store.AIReply == nil || a.store.AIManualPauses == nil {
		return
	}
	if message.ChatID == "" {
		return
	}
	// settings 保存账号级 AI 设置，其中包含人工接管暂停时长与暂停范围。
	settings, err := a.store.AIReply.Get(ctx, a.CookieID)
	if err != nil {
		a.logger.WarnContext(ctx, "读取人工接管配置失败", "err", err)
		return
	}
	if settings == nil || settings.ManualPauseMinutes <= 0 {
		return
	}
	// until 是本次暂停的截止时间，按账号配置的分钟数向后推算。
	until := time.Now().Add(time.Duration(settings.ManualPauseMinutes) * time.Minute).UnixMilli()
	if err := a.store.AIManualPauses.Upsert(ctx, a.CookieID, message.ChatID, until, settings.ManualPauseAIOnly); err != nil {
		a.logger.WarnContext(ctx, "记录人工接管暂停窗口失败", "err", err, "chat_id", message.ChatID)
		return
	}
	a.logger.InfoContext(ctx, "人工接管暂停窗口已记录", "chat_id", message.ChatID,
		"minutes", settings.ManualPauseMinutes, "ai_only", settings.ManualPauseAIOnly)
}
