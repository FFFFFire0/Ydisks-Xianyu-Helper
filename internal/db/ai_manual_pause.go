package db

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// AIManualPause 描述一个会话的人工接管暂停状态。
type AIManualPause struct {
	// Active 表示当前是否仍在暂停窗口内。
	Active bool
	// AIOnly 表示只暂停 AI 回复；为 false 时整条自动回复链都跳过。
	AIOnly bool
	// UntilAt 是暂停截止时间，Unix 毫秒。
	UntilAt int64
}

// AIManualPauses 操作 ai_manual_pauses 表。
// 暂停记录按账号和会话持久化，容器重启后仍然有效。
type AIManualPauses struct {
	// DB 是数据库连接。
	DB *sql.DB
	// Dialect 是当前数据库方言。
	Dialect Dialect
}

// Upsert 保存或刷新一个会话的人工接管窗口。
func (p *AIManualPauses) Upsert(ctx context.Context, cookieID, chatID string, untilAt int64, aiOnly bool) error {
	if p == nil || p.DB == nil || cookieID == "" || chatID == "" {
		return nil
	}
	_, err := p.DB.ExecContext(ctx,
		`INSERT INTO ai_manual_pauses (cookie_id, chat_id, until_at, ai_only, updated_at)
		 VALUES (?,?,?,?,CURRENT_TIMESTAMP)`+dialectUpsert(p.Dialect, []string{"cookie_id", "chat_id"}, map[string]string{
			"until_at":   "EXCLUDED.until_at",
			"ai_only":    "EXCLUDED.ai_only",
			"updated_at": "CURRENT_TIMESTAMP",
		}), cookieID, chatID, untilAt, boolToInt(aiOnly))
	return err
}

// Active 读取会话当前的暂停状态；没有记录或已过期时返回未激活。
func (p *AIManualPauses) Active(ctx context.Context, cookieID, chatID string) (AIManualPause, error) {
	if p == nil || p.DB == nil || cookieID == "" || chatID == "" {
		return AIManualPause{}, nil
	}
	// untilAt 是暂停截止时间；aiOnly 是数据库保存的布尔整数。
	var untilAt int64
	var aiOnly int
	err := p.DB.QueryRowContext(ctx,
		`SELECT until_at, ai_only FROM ai_manual_pauses WHERE cookie_id=? AND chat_id=?`,
		cookieID, chatID).Scan(&untilAt, &aiOnly)
	if errors.Is(err, sql.ErrNoRows) {
		return AIManualPause{}, nil
	}
	if err != nil {
		return AIManualPause{}, err
	}
	if untilAt <= time.Now().UnixMilli() {
		return AIManualPause{UntilAt: untilAt}, nil
	}
	return AIManualPause{Active: true, AIOnly: aiOnly != 0, UntilAt: untilAt}, nil
}

// Clear 立即结束一个会话的人工接管窗口。
func (p *AIManualPauses) Clear(ctx context.Context, cookieID, chatID string) error {
	if p == nil || p.DB == nil || cookieID == "" || chatID == "" {
		return nil
	}
	_, err := p.DB.ExecContext(ctx, `DELETE FROM ai_manual_pauses WHERE cookie_id=? AND chat_id=?`, cookieID, chatID)
	return err
}
