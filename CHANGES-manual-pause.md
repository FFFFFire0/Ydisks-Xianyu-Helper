# 本次改动：人工接管后暂停自动回复（v1.0.15 基线）

> 打包日期：2026-11-02
> 基线：`Christ9038/Ydisks-Xianyu-Helper` main 分支（含 v1.0.15 全部功能）
> 构建验证：`go build ./...` 通过；`go test ./internal/engine ./internal/adapter ./internal/application/settings` 全通过；`internal/db` 仅剩两个 Windows 平台既有的 TempDir 清理失败（基线同样失败）；前端 `tsc --noEmit` 通过。

## 功能行为

| 场景 | 行为 |
|---|---|
| 你从闲鱼 App 或后台手动发消息 | 该会话写入暂停窗口（默认 10 分钟，按账号设置） |
| 暂停范围＝仅暂停 AI（默认开） | 关键词回复、默认回复照常；AI 不插话 |
| 暂停范围＝暂停全部 | 该会话整条自动回复链静默，商品首次回复也跳过 |
| 暂停分钟数＝0 | 关闭该能力 |
| 机器人自己发卡/关键词 | 回显被发送确认器消费，不会触发暂停 |
| 容器重启 | 暂停记录在数据库里，仍然有效 |
| 窗口过期 | 自动恢复，无需手动操作 |

配置位置：**账号管理 → AI 助手设置 → 人工接管**（开关 + 分钟数输入框）。

## 数据库

新增迁移 `00060_ai_manual_pause.sql`（sqlite / mysql / postgres 三方言）：

- `ai_reply_settings` 增加 `manual_pause_minutes`（默认 10）、`manual_pause_ai_only`（默认 1）
- 新表 `ai_manual_pauses(cookie_id, chat_id, until_at, ai_only, updated_at)`，主键 `(cookie_id, chat_id)`

## 改动文件清单

### 新增

| 文件 | 作用 |
|---|---|
| `internal/db/migrations/{sqlite,mysql,postgres}/00060_ai_manual_pause.sql` | 迁移 |
| `internal/db/ai_manual_pause.go` | 暂停窗口读写（Upsert / Active / Clear） |
| `internal/db/ai_manual_pause_test.go` | 迁移与读写单测 |
| `internal/engine/manual_pause.go` | 人工发送 → 读取账号配置 → 写入暂停窗口 |

### 修改（后端）

| 文件 | 改动 |
|---|---|
| `internal/db/ai_reply.go` | 设置结构体加两个字段；Get / ListForUser / UpsertSettings 读写新列（分钟数收敛 0–1440） |
| `internal/db/store.go` | 注册 `AIManualPauses` 子仓储 |
| `internal/engine/outgoing_echo_confirmation.go` | 新增 `observeMessageMatched` 等带返回值的观察方法，原函数签名不变 |
| `internal/engine/message_dispatcher.go` | 自身回显分支区分「自动化发送」与「人工发送」，人工时回调记录暂停 |
| `internal/engine/account.go` | 注入 `ObserveOutgoingMatched` 与 `ManualOutgoing` |
| `internal/engine/reply.go` | 回复链读取暂停状态：整段静默或仅跳过 AI；商品首次回复也受整段静默约束 |
| `internal/application/settings/service.go` | 应用层 DTO 加字段 + 0–1440 校验 |
| `internal/adapter/settings_repository.go` | db ↔ app 字段映射 |
| `internal/server/settings_handlers.go` | 接口请求/响应加字段 + 校验 |
| `internal/server/success_contract_support.go` | 响应 DTO 加字段 |
| `api/openapi.yaml` | 两个 schema 加字段 |
| `internal/db/migrate_test.go`、`internal/db/keyword_expressions_test.go` | 最新迁移版本断言 59 → 60 |

### 修改（前端）

| 文件 | 改动 |
|---|---|
| `frontend/app/features/accounts/models.ts` | 两处类型加可选字段 |
| `frontend/app/features/accounts/api.ts` | 提交载荷加字段（默认 10 / true） |
| `frontend/app/features/accounts/submoduleHooks.ts` | 默认值与加载合并 |
| `frontend/app/features/accounts/components/AccountAISettingsModal.tsx` | 「人工接管」区块：仅暂停 AI 开关 + 分钟数输入 |
| `frontend/shared/api-contract/generated/schema.ts` | 契约类型补字段（与 openapi.yaml 同步） |

### 构建

- 新增 `.github/workflows/fork-image.yml`：推到 main 自动构建 `ghcr.io/<你的用户名>/ydisks-xianyu-helper:sha-<commit>`。

## 已知差异 / 后续可做

1. **「在线客服中立即恢复」按钮未实现**：数据库层已有 `Clear`，但没有接口和聊天页按钮。暂停会自动过期，先不阻塞使用。需要的话下一步加 `DELETE /api/v1/chat/{chat_id}/manual-pause` + 聊天页按钮。
2. **按商品隔离未做**：当前按会话（账号 + 买家）暂停，符合「人在聊，不是商品的问题」的取舍。
3. **回显误判**：自动化发送的图片消息若正文匹配不上，会被当作人工发送，导致该会话暂停一个窗口。概率低，可把分钟数调小。
