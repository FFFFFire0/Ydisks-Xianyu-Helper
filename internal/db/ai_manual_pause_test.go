package db

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

// TestAIManualPausesLifecycle 覆盖人工接管暂停窗口的写入、读取、过期与清除。
func TestAIManualPausesLifecycle(t *testing.T) {
	ctx := context.Background()
	rawDB, dialect, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	store := NewStore(rawDB, dialect)
	defer rawDB.Close()
	if _, err := store.DB.ExecContext(ctx,
		`INSERT INTO users (id, username, email, password_hash) VALUES (1,'u','u@example.com','x')`); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := store.DB.ExecContext(ctx,
		`INSERT INTO cookies (id, value, user_id) VALUES ('cid','cookie',1)`); err != nil {
		t.Fatalf("insert cookie: %v", err)
	}

	// 未写入时不应处于暂停状态。
	pause, err := store.AIManualPauses.Active(ctx, "cid", "chat-1")
	if err != nil {
		t.Fatalf("read empty pause: %v", err)
	}
	if pause.Active {
		t.Fatal("expected inactive before upsert")
	}

	// 写入未来时间后应处于暂停状态，并保留只暂停 AI 的标记。
	until := time.Now().Add(10 * time.Minute).UnixMilli()
	if err := store.AIManualPauses.Upsert(ctx, "cid", "chat-1", until, true); err != nil {
		t.Fatalf("upsert pause: %v", err)
	}
	pause, err = store.AIManualPauses.Active(ctx, "cid", "chat-1")
	if err != nil {
		t.Fatalf("read active pause: %v", err)
	}
	if !pause.Active || !pause.AIOnly || pause.UntilAt != until {
		t.Fatalf("unexpected pause state: %+v", pause)
	}

	// 覆盖为已过期时间后应自动视为未暂停。
	if err := store.AIManualPauses.Upsert(ctx, "cid", "chat-1", time.Now().Add(-time.Minute).UnixMilli(), false); err != nil {
		t.Fatalf("upsert expired pause: %v", err)
	}
	pause, err = store.AIManualPauses.Active(ctx, "cid", "chat-1")
	if err != nil {
		t.Fatalf("read expired pause: %v", err)
	}
	if pause.Active {
		t.Fatal("expected inactive after expiry")
	}

	// 清除后记录应消失。
	if err := store.AIManualPauses.Upsert(ctx, "cid", "chat-1", until, false); err != nil {
		t.Fatalf("upsert before clear: %v", err)
	}
	if err := store.AIManualPauses.Clear(ctx, "cid", "chat-1"); err != nil {
		t.Fatalf("clear pause: %v", err)
	}
	pause, err = store.AIManualPauses.Active(ctx, "cid", "chat-1")
	if err != nil {
		t.Fatalf("read cleared pause: %v", err)
	}
	if pause.Active {
		t.Fatal("expected inactive after clear")
	}
}

// TestAIReplySettingsManualPauseRoundTrip 覆盖账号级人工接管配置的读写。
func TestAIReplySettingsManualPauseRoundTrip(t *testing.T) {
	ctx := context.Background()
	rawDB, dialect, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	store := NewStore(rawDB, dialect)
	defer rawDB.Close()
	if _, err := store.DB.ExecContext(ctx,
		`INSERT INTO users (id, username, email, password_hash) VALUES (1,'u','u@example.com','x')`); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := store.DB.ExecContext(ctx,
		`INSERT INTO cookies (id, value, user_id) VALUES ('cid','cookie',1)`); err != nil {
		t.Fatalf("insert cookie: %v", err)
	}
	if err := store.AIReply.UpsertSettings(ctx, "cid", AIReplySettings{
		AIEnabled: true, MaxDiscountPercent: 10, MaxDiscountAmount: 100, MaxBargainRounds: 3,
		AIMode: "full_service", ManualPauseMinutes: 25, ManualPauseAIOnly: false,
	}); err != nil {
		t.Fatalf("upsert settings: %v", err)
	}
	settings, err := store.AIReply.Get(ctx, "cid")
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if settings.ManualPauseMinutes != 25 || settings.ManualPauseAIOnly {
		t.Fatalf("unexpected manual pause settings: %+v", settings)
	}
}
