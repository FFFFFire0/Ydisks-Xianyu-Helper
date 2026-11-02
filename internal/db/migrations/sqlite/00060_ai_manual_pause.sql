-- +goose Up
-- 人工接管：账号级暂停时长与按会话持久化的暂停窗口。
ALTER TABLE ai_reply_settings ADD COLUMN manual_pause_minutes INTEGER NOT NULL DEFAULT 10;
ALTER TABLE ai_reply_settings ADD COLUMN manual_pause_ai_only INTEGER NOT NULL DEFAULT 1;

CREATE TABLE IF NOT EXISTS ai_manual_pauses (
    cookie_id TEXT NOT NULL,
    chat_id TEXT NOT NULL,
    until_at INTEGER NOT NULL DEFAULT 0,
    ai_only INTEGER NOT NULL DEFAULT 1,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (cookie_id, chat_id),
    FOREIGN KEY (cookie_id) REFERENCES cookies(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_ai_manual_pauses_until ON ai_manual_pauses(until_at);

-- +goose Down
DROP INDEX IF EXISTS idx_ai_manual_pauses_until;
DROP TABLE IF EXISTS ai_manual_pauses;
ALTER TABLE ai_reply_settings DROP COLUMN manual_pause_ai_only;
ALTER TABLE ai_reply_settings DROP COLUMN manual_pause_minutes;
