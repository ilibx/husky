# Husky 智能工单系统

## 基础功能

### 用户管理

| 操作 | API | 权限 |
|------|-----|------|
| 注册 | `POST /api/v1/auth/register` | 公开 |
| 登录 | `POST /api/v1/auth/login` | 公开 |
| 个人信息 | `GET /api/v1/users/me` | 认证用户 |
| 用户列表 | `GET /api/v1/users` | 管理员 |
| 修改角色 | `PUT /api/v1/users/:id/role` | 管理员 |

### 工单管理

| 操作 | API | 说明 |
|------|-----|------|
| 创建工单 | `POST /api/v1/tickets` | 自动生单号 TK + 日期 + UUID |
| 工单列表 | `GET /api/v1/tickets` | 支持 status/priority/source 筛选 |
| 工单详情 | `GET /api/v1/tickets/:id` | 包含评论、附件、操作日志 |
| 更新工单 | `PUT /api/v1/tickets/:id` | 验证状态转换合法性 |
| 删除工单 | `DELETE /api/v1/tickets/:id` | 软删除 |
| 分配工单 | `POST /api/v1/tickets/:id/assign` | 自动通知被分配人 |
| 自动分配 | `POST /api/v1/tickets/:id/auto-assign` | 按负载均衡分配 |
| 认领工单 | `POST /api/v1/tickets/:id/claim` | 客服自行认领 |
| 更新状态 | `POST /api/v1/tickets/:id/status` | 校验合法流转 |
| 设置截止 | `POST /api/v1/tickets/:id/due` | SLA 到期设置 |
| 评价工单 | `POST /api/v1/tickets/:id/rate` | 1-5 分，仅已解决/关闭可评 |
| 评论 | `POST/GET /api/v1/tickets/:id/comments` | 支持内部/公开评论 |
| 附件 | `POST/GET/DELETE /api/v1/tickets/:id/attachments` | 文件上传/下载/删除 |
| 关注工单 | `POST/DELETE /api/v1/tickets/:id/watch` | 关注后接收评论通知 |
| 超期工单 | `GET /api/v1/tickets/overdue` | 列出已超 SLA 的工单 |

**状态转换规则：**

```
open ──→ in_progress ──→ pending ──→ resolved ──→ closed
  │          │              │            │
  └──────────┴──── closed ──┘            └── open
```

**优先级：** `low` → `medium` → `high` → `urgent`

### 知识库

| 操作 | API | 说明 |
|------|-----|------|
| 创建知识 | `POST /api/v1/knowledge` | 自动生成向量嵌入 |
| 知识列表 | `GET /api/v1/knowledge` | 分页查询 |
| 知识详情 | `GET /api/v1/knowledge/:id` | |
| 更新知识 | `PUT /api/v1/knowledge/:id` | 更新后重新生成向量 |
| 删除知识 | `DELETE /api/v1/knowledge/:id` | |
| 搜索知识 | `POST /api/v1/knowledge/search` | 向量相似度搜索，无向量时 ILIKE 降级 |
| 导入知识 | `POST /api/v1/knowledge/import` | 批量导入 |
| 导出知识 | `GET /api/v1/knowledge/export` | 批量导出 |

### 智能 Agent

| 操作 | API | 说明 |
|------|-----|------|
| 创建 Agent | `POST /api/v1/agents` | 类型：llm / rule / hybrid |
| Agent 列表 | `GET /api/v1/agents` | |
| 更新 Agent | `PUT /api/v1/agents/:id` | 管理 |
| Agent 触发 | 自动触发 | ticket_created / ticket_updated |

**Agent 执行规则：**

- **rule 类型**：匹配触发条件后执行预设动作（自动分配、改状态、改优先级）
- **llm 类型**：搜索知识库生成自动回复，以内部评论添加到工单
- **hybrid 类型**：先执行规则，再调用 LLM 生成回复

### 通知系统

| 操作 | API | 说明 |
|------|-----|------|
| 通知列表 | `GET /api/v1/notifications` | 当前用户的通知 |
| 未读数 | `GET /api/v1/notifications/unread-count` | |
| 标记已读 | `PUT /api/v1/notifications/:id/read` | |

**通知触发场景：** 工单分配、状态变更、新评论、SLA 到期

### SOP 流程

| 操作 | API |
|------|-----|
| 创建 SOP | `POST /api/v1/sop` |
| 列表 | `GET /api/v1/sop` |
| 详情 | `GET /api/v1/sop/:id` |
| 更新 | `PUT /api/v1/sop/:id` |
| 删除 | `DELETE /api/v1/sop/:id` |

### 部门管理

| 操作 | API | 权限 |
|------|-----|------|
| 部门树 | `GET /api/v1/departments` | 认证用户 |
| 创建/更新/删除 | `POST/PUT/DELETE` | 管理员 |

### 分类管理

| 操作 | API | 权限 |
|------|-----|------|
| 分类列表 | `GET /api/v1/categories` | 认证用户 |
| 创建/更新/删除 | `POST/PUT/DELETE` | 管理员 |

### 标签管理

| 操作 | API | 说明 |
|------|-----|------|
| 标签列表 | `GET /api/v1/tags` | 全部标签 |
| 创建/更新/删除 | `POST/PUT/DELETE /api/v1/tags/:id` | 管理员 |

### 渠道配置

| 操作 | API | 说明 |
|------|-----|------|
| 渠道列表 | `GET /api/v1/channels` | |
| 创建/更新/删除 | `POST/PUT/DELETE` | 管理员 |

### 渠道用户管理

| 操作 | API | 说明 |
|------|-----|------|
| 渠道用户列表 | `GET /api/v1/channel-users` | 按渠道筛选 |
| 创建/更新/删除 | `POST/PUT/DELETE /api/v1/channel-users/:id` | 管理员 |
| 标签管理 | `POST /api/v1/channel-users/:id/tags` | 设置用户标签 |

### 渠道群组管理

| 操作 | API | 说明 |
|------|-----|------|
| 渠道群组列表 | `GET /api/v1/channel-groups` | 按渠道筛选 |
| 创建/更新/删除 | `POST/PUT/DELETE /api/v1/channel-groups/:id` | 管理员 |
| 标签管理 | `POST /api/v1/channel-groups/:id/tags` | 设置群组标签 |

### Bot 配置

| 操作 | API | 说明 |
|------|-----|------|
| 获取 Bot 配置 | `GET /api/v1/channels/:channel/bot-config` | |
| 设置 Bot 配置 | `POST /api/v1/channels/:channel/bot-config` | 含分类关联 |

### 菜单管理

| 操作 | API | 说明 |
|------|-----|------|
| 菜单树 | `GET /api/v1/menus` | 按角色返回可见菜单 |
| 全部菜单 | `GET /api/v1/menus/all` | 管理员 |
| 创建/更新/删除 | `POST/PUT/DELETE /api/v1/menus/:id` | 管理员 |

### 飞书渠道

参见 [feishu.md](channels/feishu.md)

### 渠道消息记录

| 操作 | API | 说明 |
|------|-----|------|
| 消息记录 | `GET /api/v1/messages` | 管理员，按渠道筛选 |

### Webhook 配置

| 操作 | API | 说明 |
|------|-----|------|
| Webhook 配置列表 | `GET /api/v1/webhook-configs` | 管理员 |
| 创建/更新/删除 | `POST/PUT/DELETE` | 管理员 |

### Webhook 接口

| 渠道 | URL | 说明 |
|------|-----|------|
| 飞书 | `POST /webhooks/lark` | 消息→创建工单 / 群聊消息→自动回复 |
| 钉钉 | `POST /webhooks/dingtalk` | 消息→创建工单 |
| 企业微信 | `POST /webhooks/wecom` | 消息→创建工单 |

### 统计报表

| 操作 | API |
|------|-----|
| 概览 | `GET /api/v1/stats/overview` |
| 工单分布 | `GET /api/v1/stats/tickets` |
| 绩效统计 | `GET /api/v1/stats/performance` |

### 分配策略

| 操作 | API | 说明 |
|------|-----|------|
| 获取分配策略 | `GET /api/v1/assign-config` | |
| 设置分配策略 | `POST /api/v1/assign-config` | 按分类配置 |

### Skill 管理

| 操作 | API | 说明 |
|------|-----|------|
| Skill 列表 | `GET /api/v1/skills` | |
| 创建/更新/删除 | `POST/PUT/DELETE /api/v1/skills/:id` | 管理员 |

### MCP 服务管理

| 操作 | API | 说明 |
|------|-----|------|
| MCP 列表 | `GET /api/v1/mcps` | |
| 创建/更新/删除 | `POST/PUT/DELETE /api/v1/mcps/:id` | 管理员 |

### 系统配置

| 操作 | API | 说明 |
|------|-----|------|
| 配置列表 | `GET /api/v1/system-config` | 分页查询 |
| 更新配置 | `POST /api/v1/system-config/upsert` | 按 category+key |

## 数据表

| 表名 | 说明 |
|------|------|
| `users` | 用户 |
| `tickets` | 工单 |
| `comments` | 评论 |
| `attachments` | 附件 |
| `categories` | 分类 |
| `knowledge` | 知识库（含向量） |
| `agents` | Agent 配置 |
| `sops` | SOP 流程 |
| `departments` | 部门 |
| `roles` | 角色 |
| `notifications` | 通知 |
| `audit_logs` | 审计日志 |
| `satisfactions` | 满意度 |
| `channels` | 渠道配置 |
| `channel_users` | 渠道用户 |
| `channel_groups` | 渠道群组 |
| `webhook_message_records` | 消息记录 |
| `webhook_configs` | Webhook 配置 |
| `ticket_watchers` | 工单关注 |
| `ticket_groups` | 工单-飞书群绑定 |
| `tags` | 标签 |
| `ticket_tags` | 工单-标签关联 |
| `ticket_relations` | 工单关联 |
| `workflows` | 工作流实例 |
| `workflow_steps` | 工作流步骤 |
| `sla_configs` | SLA 配置 |
| `bot_configs` | Bot 配置 |
| `menus` | 菜单配置 |
| `skills` | 技能 |
| `mcps` | MCP 服务 |
| `system_configs` | 系统配置 |
| `assign_configs` | 分配策略 |

#### Workflow 工作流

| 操作 | API | 说明 |
|------|-----|------|
| 工作流列表 | `GET /api/v1/workflows` | 支持按 ticket_id/status 筛选 |
| 工作流详情 | `GET /api/v1/workflows/:id` | 包含所有步骤 |
| 完成人工步骤 | `POST /api/v1/workflows/steps/:stepId/complete` | 需传入 result |
| 待办步骤 | `GET /api/v1/workflows/tasks` | 当前用户的 pending 人工步骤 |

**SOP 步骤类型：**

| 类型 | 说明 |
|------|------|
| `react` | LLM ReAct 循环，自动执行 search_kb / reply_user / update_ticket / complete / fail |
| `human` | 人工介入，可配置超时（默认 24h），通过 API 完成 |
| `condition` | 条件判断，按 status/priority/assignee 字段自动跳过或执行 |
| `notification` | 发送通知给指定用户 |

### Admin 管理后台

Admin UI 是一个嵌入式单页应用，通过 Go embed 编译进二进制，无需额外构建。

| 功能 | 说明 |
|------|------|
| 访问地址 | `GET /admin` |
| 登录 | 使用 JWT 认证，登录后管理工单/知识库/Agent/SOP/用户等 |
| 仪表盘 | 工单概览、状态分布、渠道统计 |

### 数据库迁移

项目提供两种迁移方式：

**1. GORM AutoMigrate（开发推荐）**
```bash
go run cmd/api/main.go
```
API 启动时自动调用 `database.AutoMigrate()`，根据 model 结构体同步表结构。

**2. SQL 脚本迁移（生产推荐）**
```bash
go run cmd/migrate/main.go
```
读取 `scripts/migrations/` 目录下的 SQL 文件顺序执行。当前脚本：
- `001_initial_schema.sql` — 基础表结构
- `002_ticket_groups.sql` — 工单-飞书群绑定表

### LLM 服务架构

```
┌─────────────────────┐
│   EmbeddingService  │   向量嵌入（知识库检索）
│   Embed()           │
│   BatchEmbed()      │
└─────────┬───────────┘
          │   EmbeddingProvider (OpenAI / DashScope)
          │
┌─────────▼───────────┐
│    ChatService       │   LLM 对话（Agent/Workflow ReAct）
│    Chat()            │
└─────────────────────┘
```

- **EmbeddingService** — 封装 EmbeddingProvider，提供 Embed/BatchEmbed 接口，供知识库创建和向量搜索使用
- **ChatService** — 封装 ChatProvider，提供 Chat 接口，供 Agent LLM 和 Workflow ReAct 循环使用
- `NewChatService(provider)` 在 provider 不支持对话时（如 DashScope）返回 nil，避免运行时类型断言

## 启动

```bash
# 配置环境变量
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=husky
export JWT_SECRET=your-secret-key

# 可选：飞书集成
export FEISHU_APP_ID=xxx
export FEISHU_APP_SECRET=xxx

# 可选：LLM 配置
export LLM_PROVIDER=openai
export LLM_API_KEY=sk-xxx
export LLM_BASE_URL=https://api.openai.com/v1

# 启动（自动迁移）
go run cmd/api/main.go
```
