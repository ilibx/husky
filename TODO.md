# Husky 项目实施现状

> 最后更新: 2026-06-07
> 
> 目标: 对标飞书服务台功能，逐步实现企业级智能工单系统。支持多渠道接入 -> 网关统一处理 -> Agent + SOP 智能执行 -> 人工协作的完整链路。

---

## 一、当前实现状态

### ✅ 已实现（核心架构）
| 模块 | 状态 | 说明 |
|------|------|------|
| 项目结构 | ✅ | 领域驱动：channel / gateway / ticket / agent / knowledge |
| 配置管理 | ✅ | 环境变量加载、Config 结构体 |
| 日志系统 | ✅ | Zap 日志封装 |
| 数据库连接 | ✅ | GORM + PostgreSQL 连接池 |
| AutoMigrate | ✅ | 自动迁移所有模型表 |
| JWT 认证中间件 | ✅ | Token 生成/验证/刷新 |
| RBAC 中间件 | ✅ | 角色字符串比对 |
| 健康检查 | ✅ | `/health` 端点 |
| 路由注册 | ✅ | Gin 路由框架、API v1 分组 |
| 工单 CRUD | ✅ | 完整 Create/Update/Delete/Get/List（含评论、附件、分配） |
| 工单状态机 | ✅ | 5 种状态 + 合法转换校验中间件 |
| 工单编号生成器 | ✅ | TK{YYYYMMDD}{UUID6} 格式 |
| 用户认证系统 | ✅ | 登录/注册/Token 刷新/密码 bcrypt |
| 统计分析 | ✅ | 概览 + 按渠道/优先级/状态分布 |
| 满意度评价 | ✅ | 1-5 分评分 + 评论文本 + 聚合统计 |
| 知识库 CRUD | ✅ | Handler + Service + VectorStore 完整实现 |
| AI Agent 系统 | ✅ | Rule/LLM/Hybrid 三种 Agent + 触发匹配 |
| SOP 流程引擎 | ✅ | CRUD + LLM 智能选择 + 步骤拆分 + ReAct 执行 |
| HITL 人机协作 | ✅ | Human step 超时/分配/完成接口 + ReAct 用户交互 |
| 工作流引擎 | ✅ | ReAct 循环 + Condition/Notification/Human 步骤 |
| 网关层 | ✅ | Gateway 统一入口 + 用户上下文富化 + 渠道分发 |
| 渠道接入 | ✅ | 飞书/Lark/钉钉/企微 Webhook 事件模型 |
| 飞书卡片 | ✅ | 创建群组 + 发送卡片的 TicketGroupService |
| 前端 WebUI | ✅ | Vue 3 + Element Plus + TypeScript，嵌入 Go 二进制；支持 external proxy 模式 |
| 飞书用户信息 | ✅ | FeishuEnricher 缓存 + 工单 metadata 注入 |
| 缓存层 | ✅ | Redis 解耦（已降级为 NoopCache，可选启用） |
| 仪表盘 KPI | ✅ | CSS Grid 布局，10 张统计卡片自适应排列，卡片可点击跳转对应管理页面 |
| 角色层级 | ✅ | Role.Level 字段，admin=100/agent=50/user=10 |
| 工单可见范围 | ✅ | 支持 scope=all/my/my_team 筛选，上级可见下级工单 |
| 分类负责人 | ✅ | Category.ManagerID，创建工单自动分配给负责人 |

### 🔶 部分实现
| 模块 | 状态 | 说明 |
|------|------|------|
| Agent/Workflow/SOP 管理 | 🔶 | CRUD 已实现，前端页面已对接 |
| 错误处理 | 🔶 | AppError + ErrorResponse + 错误码体系基本就绪 |

---

## 二、Redis 依赖解耦（已修复 ✅）

**问题**: 硬依赖 Redis，不可用时服务无法启动。

**修复内容**:
1. `pkg/cache/cache.go` — 新增 `CacheInterface` 接口和 `NoopCache` 空实现
2. `cmd/api/main.go` — Redis 连接失败时 Warn 日志 + 降级为 NoopCache
3. `internal/repository/repository.go` — 新增 `DatabaseConnection` 结构体，只包含 `*gorm.DB`
4. `internal/repository/ticket/ticket_repository.go` — 此文件包含硬 Redis 依赖，当前未被 router.go 引用。建议后续合并或移除。

---

## 三、飞书服务台功能对标与差距分析

### 核心功能 GAP

| # | 飞书服务台功能 | Husky 状态 | 差距 | 优先级 |
|---|--------------|-----------|------|--------|
| 1 | 工单自动生成（与机器人互动触发） | ✅ 已实现 | 用户发消息→AI 意图识别（create_ticket/ask_knowledge/greeting/unknown）→自动创建工单 | P0 |
| 2 | 客服代创建工单 | ✅ 已实现 | admin/agent 可在创建时指定 requester_id，普通用户只能为自己创建 | P0 |
| 3 | 工单状态流转（状态机 + 约束规则） | ✅ 已实现 | ValidTransitions 合法转换表 + IsValidTransition 校验 + ResolvedAt/ClosedAt 自动钩子 + 通知 watchers | P0 |
| 4 | 机器人知识库问答 | ✅ 已实现 | RAG 流程：用户提问→向量检索→LLM 生成回答→返回 AnswerResponse | P0 |
| 5 | 工单分配（技能分流/指定客服） | ✅ 已实现 | 四种策略：least_busy / round_robin / skill_based / random，支持按分类配置 | P1 |
| 6 | 工单标签系统 | ✅ 已实现 | Tag CRUD + 工单标签关联/替换 API | P1 |
| 7 | 知识库分类管理 + 层级 | ✅ 已实现 | Category 增加 type 字段区分知识库/工单分类，知识库分类树 API，List 支持 category 筛选 | P1 |
| 8 | 工单评分/满意度调查 | ✅ 已实现 | RateTicket + GetSatisfaction API | P1 |
| 9 | 客服工单面板（备注/标签/状态编辑） | ✅ 已实现 | 前端详情页：状态下拉编辑、标签管理弹窗、内部备注复选框 | P1 |
| 10 | SLA 超时升级 | ✅ 已实现 | SLAEscalator 后台 goroutine, 5min 轮询+通知 | P1 |
| 11 | 系统操作日志 | ✅ 已实现 | AuditLog 写入（CreateAuditLog）+ 查询 API | P1 |
| 12 | 历史消息记录持久化 | ✅ 已实现 | 飞书/钉钉/企微 Webhook 消息保存到 WebhookMessage 表 | P1 |
| 13 | 机器人欢迎语/签名 | ✅ 已实现 | BotConfig 模型 + API + Gateway 自动发送欢迎消息 | P2 |
| 14 | 推荐问题/猜你想问 | ✅ 已实现 | ViewCount + HotScore 热度统计 + Recommend API | P2 |
| 15 | 知识库批量导入/导出 | ✅ 已实现 | JSON + CSV 两种格式，CSV 支持 multipart 上传和文件下载 | P2 |
| 16 | 多语言知识库 | ✅ 已实现 | zh/en 标题/内容字段 + Language 属性 | P2 |
| 17 | 工单合并/关联 | ✅ 已实现 | TicketRelation 模型 + CRUD API，支持 parent/child/related/duplicate/blocks | P2 |
| 18 | 渠道用户/群组管理 | ✅ 已实现 | ChannelUser/ChannelGroup 模型 + 标签关联，按渠道区分 Bot 职责 | P2 |
| 19 | 菜单管理 | ✅ 已实现 | Menu 模型 + 按角色筛选 + admin 管理界面 | P1 |
| 18 | 渠道用户/群组管理 | ✅ 已实现 | ChannelUser/ChannelGroup 模型 + 标签关联，按渠道区分 Bot 职责 | P2 |

### 需求不合理/不明确项

1. **"支持 6 个渠道"范围过大** — 飞书和 Lark 本质是同一平台（国内/国际版），应合并为一个渠道。建议 MVP 仅支持飞书 + WebUI，其他渠道 P1 后处理。
2. **"Agent 系统"定义过广** — README 列出 7 种 Agent，但 Agent 模型仅定义了 LLM 参数（type/config/model/temperature），无触发条件、执行逻辑、回调机制设计。建议先实现 1-2 个核心 Agent（自动分类/智能回复）。
3. **RBAC 权限控制** — ✅ 已完善。`Role.Permissions` 支持 JSON 权限矩阵，`PermissionsLoader` 从 DB 加载，`PermissionMiddleware(resource, action)` 精确校验，admin 角色默认拥有全部权限，`LoadPermissionsMiddleware` 自动注入到请求上下文。
4. **HITL 交互协议未设计** — HITL 需要定义：AI 建议格式 → 人工审核接口 → 反馈回写 → 模型更新链路。✅ 已实现完整 HITL 协议：`executeHumanStep` 自动生成 AI 建议，`Suggestion`/`Decision`/`Feedback`/`ApprovedBy` 字段记录全链路，API：approve/reject/revise，前端「我的待办」页面。
5. **`SERVER_MODE` vs `APP_ENV` 命名冲突** — `.env.example` 使用 `APP_ENV`，但 `config.go` 使用 `SERVER_MODE` 字段。建议统一为 `APP_ENV`。
6. **工单状态机规则未明确** — 5 种状态之间的合法转换路径、转换条件、触发动作均未定义。

---

## 四、代码质量问题

| # | 问题 | 文件 | 修复状态 |
|---|------|------|---------|
| 1 | 两个 TicketRepository 并存 | `repository/repository.go` vs `repository/ticket/ticket_repository.go` | ✅ 已移除冗余副本 |
| 2 | 接口签名不一致 | `GetByID(id string)` vs `GetByID(id uint)` | ✅ 已统一 |
| 3 | GORM 模型 `uint` ID vs SQL 迁移 `UUID PK` | `model/models.go` vs `scripts/migrations/*.sql` | ✅ 已决策：使用 GORM AutoMigrate + `uint` 自增主键 |
| 4 | 两个知识库模型 | `model.Knowledge` (uint, models.go) vs `model.KnowledgeBase` (string UUID, knowledge.go) | ✅ 已统一：移除 `Knowledge`，`KnowledgeBase` 表重命名为 `knowledge` |
| 5 | `strings.Split` 重复实现 | `handler/knowledge.go` — 自实现 splitString | ✅ 已修复 |
| 6 | `io/ioutil` 已弃用 | `cmd/migrate/main.go` | ✅ 已修复 |
| 7 | 字符串拼接 bug | `handler/webhook.go:124` — `"工单已创建：" + ticket.ID` | ✅ 已修复 |

---

## 五、开发路线图

### Phase 0 — 基础可运行 ✅
- [x] Redis 解耦，不依赖 Redis 可启动
- [x] 领域驱动重构（channel / gateway / ticket / agent / knowledge）
- [x] 工单完整 CRUD（评论、附件、分配、认领、状态机）
- [x] 用户认证 + RBAC
- [x] AI Agent 执行引擎 + SOP 流程引擎 + 工作流
- [x] 前端 Vue 3 管理后台（嵌入 Go 二进制，`/admin` 访问）
- [x] 多渠道 Webhook 接入（飞书/钉钉/企微）
- [x] Gateway 网关 + 用户上下文富化
- [x] 知识库 CRUD + 文本搜索

### Phase 1 — 增强（已完成 ✅）
- [x] 向量搜索：接入真实 LLM Embedding，替换全 0 占位向量（维度 768→1536，score 修复，UpdateKnowledge 重嵌入）
- [x] SLA 超时自动升级逻辑 + 通知（5min 后台 goroutine，breach 状态 + 通知 assignee/requester）
- [x] 飞书/钉钉/企微群消息持久化：WebhookMessage 表保存每条 webhook 消息
- [x] Notifications 对接 Agent/Workflow：human step 通知 assignee，completeStep 通知 requester，failStep 通知 assignee
- [x] AuditLog 查询 API：`GET /tickets/:id/audit-logs`（资源/资源ID过滤 + 分页 + 用户预加载）

### Phase 2 — 企业级（已完成 ✅）
- [x] 客服代创建工单（admin/agent 指定 requester_id）
- [x] 工单标签系统（Tag CRUD + 工单标签关联/替换）
- [x] 工单分配策略（least_busy / round_robin / skill_based / random，按分类配置）
- [x] 工单合并/关联（parent/child/related/duplicate/blocks 类型）
- [x] 知识库批量导入/导出（JSON + CSV 格式，embedding 自动生成）
- [x] RBAC 细粒度权限（permission-resource-action 矩阵 + PermissionMiddleware）
- [x] 多语言知识库（zh/en 标题/内容字段，language 筛选）
- [x] 推荐问题/猜你想问（ViewCount + HotScore 热度统计，Recommend API）
- [x] 机器人欢迎语/签名（BotConfig 模型 + API + Gateway 自动发送）
- [x] 客服工单面板（前端：状态编辑下拉、标签管理弹窗、内部备注复选框）

### Phase 2 — 剩余
- [ ] 多语言知识库

---

## 六、技术决策记录

### ADR-1: 主键策略
- **决定**: 使用 GORM AutoMigrate + `uint` 自增主键
- **例外**: `KnowledgeBase` 使用 `string` UUID（pgvector 兼容性）
- **理由**: GORM 对 `uint` 自增主键支持最完善；AutoMigrate 减少 migration 维护成本；SQL 迁移脚本（`scripts/migrations/`）已弃用
- **影响**: 所有新模型应嵌入 `model.Base`（`uint` ID）；向量相关的模型可例外使用 UUID

### ADR-2: Admin 部署模式
- **决定**: 默认嵌入式（静态资源打包进 Go 二进制），支持 external 代理模式
- **配置**: `ADMIN_MODE=embedded`（默认，Go embed 内嵌 SPA）/ `ADMIN_MODE=external` + `ADMIN_URL=http://localhost:5173`（反向代理到前端 dev server）
- **环境变量**: Vue 前端使用 `.env.embed`（`VITE_BASE_PATH=/admin/`），external 模式下由前端 dev server 自行管理代理
- **影响**: 开发时无需每次都 `make web-build`；生产部署用默认 embedded 模式。`ADMIN_URL` 仅在 external 模式生效

### ADR-3: 知识库模型
- **决定**: 使用 `model.KnowledgeBase` 作为唯一知识库模型，表名 `knowledge`
- **理由**: `KnowledgeBase` 包含 Vector 字段（pgvector 支持），已被 service/handler/vector store 全程使用；`Knowledge` 是未使用的死代码
- **影响**: 标准知识 CRUD 通过 `VectorStoreRepository` 操作；后期可添加 `Summary`/`ViewCount` 等字段到 `KnowledgeBase`

## 七、技术债务记录

| ID | 描述 | 影响 | 优先级 |
|----|------|------|--------|
| TD-1 | 向量搜索全 0 占位向量 | 搜索结果不可用 | ✅ 已修复（维度 768→1536，score 修复，UpdateKnowledge 重嵌入） |
| TD-2 | AuditLog 无写入逻辑 | 操作不可追溯 | ✅ 已修复（CreateAuditLog + ListAuditLogs API） |
| TD-3 | Feishu 升级通知使用 DB ID 而非 OpenID | 飞书通知不可用 | P1 |
| TD-4 | runRuleAgent 使用指针直接修改 ticket | 多 Agent 并发脏写（已修复为 UpdateFields） | ✅ |
| TD-5 | Dockerfile.worker 缺失 | docker-compose 无法启动 worker | P2 |
| TD-6 | `SERVER_MODE` vs `APP_ENV` 命名不统一 | 配置混淆 | P2 |
