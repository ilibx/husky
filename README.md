# Husky - 企业级智能工单系统

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## 项目概述

Husky 是一个基于 Go + GORM + Gin 实现的企业级智能工单处理系统。集成 RAG 知识库、ReAct + SOP 工作流引擎、LLM Agent 系统、HITL（人在回路）交互流程，支持多渠道接入（飞书），提供完整的工单管理与统计分析能力。内置嵌入式单页 Admin 管理 UI。

## 核心特性

### 智能工单系统
- 完整的工单生命周期管理（创建→分配→处理→关闭）
- 灵活的状态机配置和自动流转规则
- 智能工单分配（轮询、负载均衡、技能匹配）
- 工单评论、附件、协作、关注功能
- SLA 超时升级机制

### AI 智能能力
- **RAG 知识库**: 基于 pgvector 向量检索的增强生成，提供精准智能问答，无向量时 ILIKE 降级
- **SOP 管理**: 标准作业程序定义（react/human/condition/notification 步骤类型），自动创建 Workflow 实例
- **ReAct 推理引擎**: SOP 中 `react` 类型步骤使用 ReAct（Reasoning+Acting）循环执行，LLM 在 search_kb/reply_user/update_ticket/complete/fail 工具间自主决策，滑动窗口历史（最近 5 轮），最大 10 轮
- **Agent 系统**: rule（自动派单/改状态/改优先级）、llm（知识检索+自动回复）、hybrid 三种类型，按 ticket_created/ticket_updated 事件触发
- **ChatService**: 独立的 LLM 对话服务，与 EmbeddingService 解耦，避免类型断言的 leaky abstraction

### 多渠道接入
- 飞书 (Feishu) — 群聊自动创建、消息卡片、自动回复、人工升级
- Webhook 通道 — Lark / 钉钉 / 企业微信
- Restful API

### Admin 管理后台
- 嵌入式 SPA（Vue 3 + Element Plus，Go embed 编译进二进制）
- 仪表盘（工单概览统计图表）
- 工单管理（CRUD、筛选、详情、标签、评论）
- 资料管理（文档管理 / 向量库配置 / 资料检索）
- Agent 配置管理 / SOP 流程管理 / Skill 管理 / MCP 服务
- Workflow 工作流跟踪
- 通知管理（通知中心 / SLA 配置）
- 系统管理（用户/角色/部门/分类/标签/菜单管理）
- 渠道管理（渠道配置、渠道用户、渠道群组）
- 大模型管理（系统配置）

### 安全与权限
- JWT Token 认证 + Token 刷新
- RBAC 基于角色的权限控制
- 操作审计日志

## 技术架构

```
┌──────────────────────────────────────────────────────────────────┐
│                         Husky 系统架构                            │
├──────────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────────┐    │
│  │                     接入层                                │    │
│  │  飞书 │ Lark │ 钉钉 │ 企微 │ Webhook │ REST API          │    │
│  └──────────────────────────────────────────────────────────┘    │
│                              │                                    │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │                   API Gateway (Gin)                       │    │
│  │           认证(JWT) │ 限流 │ 路由 │ CORS                 │    │
│  └──────────────────────────────────────────────────────────┘    │
│                              │                                    │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │                    业务服务层                              │    │
│  │  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌──────┐   │    │
│  │  │ 工单    │ │ 用户    │ │ 知识库  │ │ SOP    │ │ Agent│   │    │
│  │  │ Service│ │ Service│ │ Service│ │ Service│ │Engine│   │    │
│  │  └────────┘ └────────┘ └────────┘ └────────┘ └──┬───┘   │    │
│  │  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐    │       │    │
│  │  │Workflow│ │ 通知    │ │ 统计    │ │ 渠道    │    │       │    │
│  │  │Service │ │ Service│ │ Service│ │ Service│    │       │    │
│  │  └───┬────┘ └────────┘ └────────┘ └────────┘    │       │    │
│  │      │   ReAct 循环                               │       │    │
│  │      │   ┌─────────────────────────┐              │       │    │
│  │      │   │ LLM → 思考 → 决策工具    │              │       │    │
│  │      │   │ search_kb / reply_user   │ ←───────────┘       │    │
│  │      │   │ update_ticket/complete   │                      │    │
│  │      │   └─────────────────────────┘                      │    │
│  └──────────────────────────────────────────────────────────┘    │
│                              │                                    │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │                  LLM 服务层                                │    │
│  │  ┌─────────────────┐  ┌─────────────────┐                 │    │
│  │  │ EmbeddingService │  │   ChatService    │                 │    │
│  │  │ Embed()/Batch()  │  │   Chat()         │                 │    │
│  │  └────────┬────────┘  └────────┬────────┘                 │    │
│  │           │                    │                            │    │
│  │  ┌────────▼────────────────────▼────────┐                  │    │
│  │  │    Provider (OpenAI / DashScope)      │                  │    │
│  │  │    EmbeddingProvider + ChatProvider    │                  │    │
│  │  └───────────────────────────────────────┘                  │    │
│  └──────────────────────────────────────────────────────────┘    │
│                              │                                    │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │                  Admin 管理 UI                            │    │
│  │    embedded: Go embed SPA (默认)                          │    │
│  │    external: 反向代理到前端 dev server                     │    │
│  └──────────────────────────────────────────────────────────┘    │
│                              │                                    │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │                  数据访问层 (Repository)                   │    │
│  │              GORM + 自定义查询 + pgvector                  │    │
│  └──────────────────────────────────────────────────────────┘    │
│                              │                                    │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │                  数据存储层                                │    │
│  │    PostgreSQL (+pgvector) │ Redis 缓存                    │    │
│  └──────────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────────┘
```

## 快速开始

### 前置要求

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 14+（推荐带 pgvector 扩展）
- Redis 6+

### 本地开发

```bash
# 1. 启动基础设施 (PostgreSQL + Redis)
docker-compose up -d postgres redis

# 2. 配置环境变量
cp .env.example .env
# 编辑 .env 文件配置必要的环境变量

# 3. 下载依赖
go mod download

# 4. 运行数据库迁移（SQL 脚本方式，适用于生产）
go run cmd/migrate/main.go

# 或直接启动 API（内置 GORM AutoMigrate，适用于开发）
go run cmd/api/main.go

# 5. 访问服务
# API:        http://localhost:8080
# Admin UI:   http://localhost:8080/admin
# 健康检查:   http://localhost:8080/health
```

### 默认管理员账号

首次启动时自动创建默认管理员（仅当数据库无 admin 用户时）：

| 字段 | 值 |
|------|-----|
| 邮箱 | `admin@husky.local` |
| 密码 | `admin123` |
| 角色 | `admin` |

> ⚠️ 生产环境请立即修改密码！

### Admin UI 部署模式

```bash
# 模式一：嵌入式（默认）— SPA 内嵌到 Go 二进制，无需额外部署
go run ./cmd/api --config config.yaml

# 构建嵌入式部署二进制
make build
```

### 使用 Makefile

```bash
make build          # 编译 API + migrate 二进制（含 embedded SPA）
make run            # 启动 API 服务
make migrate        # 运行数据库迁移
make test           # 运行测试
make docker-up      # 启动所有 Docker 服务
```

### 验证安装

```bash
curl http://localhost:8080/health
# {"status":"healthy"}
```

## 项目结构

```
husky/
├── cmd/                    # 应用程序入口
│   ├── api/               # API 服务（含 GORM AutoMigrate）
│   └── migrate/           # 独立数据库迁移工具（SQL 脚本）
├── internal/              # 私有应用代码
│   ├── admin/             # Admin 管理 UI（Go embed SPA）
│   ├── agent/             # Agent 引擎 + SOP 工作流
│   ├── channel/           # 渠道管理（用户/群组/配置）
│   ├── config/            # 配置管理（环境变量）
│   ├── database/          # 数据库连接 + GORM AutoMigrate
│   ├── gateway/           # 渠道网关（统一入口 + 用户富化）
│   ├── handler/           # HTTP 处理器
│   ├── intent/            # 用户意图识别
│   ├── knowledge/         # 知识库服务
│   ├── ldap/              # LDAP 同步服务
│   ├── middleware/        # 中间件（CORS、认证、RBAC）
│   ├── model/             # GORM 数据模型
│   ├── repository/        # 数据访问层
│   ├── router/            # 路由定义 + 服务注入
│   ├── ticket/            # 工单服务（含 SLA/Workflow 配置）
│   └── service/           # 业务逻辑层
├── pkg/                   # 公共库代码
│   ├── cache/             # Redis / Noop 缓存
│   ├── errors/            # 错误处理
│   │   └── herr/          # HTTP 错误响应
│   ├── feishu/            # 飞书 SDK（群聊、消息、卡片）
│   ├── llm/               # LLM 服务
│   │   ├── provider.go    # EmbeddingProvider + ChatProvider 接口
│   │   ├── embedding.go   # EmbeddingService
│   │   ├── chat.go        # ChatService（独立对话服务）
│   │   ├── openai.go      # OpenAI 实现（Embedding + Chat）
│   │   └── dashscope.go   # DashScope 实现（仅 Embedding）
│   ├── logger/            # 结构化日志封装
│   └── validator/         # 参数验证
├── scripts/               # 脚本工具
│   └── migrations/        # SQL 迁移脚本
├── deployments/           # 部署配置
│   └── Dockerfile         # API 服务 Docker 镜像
├── docs/                  # 文档
│   ├── basic.md           # 基础功能与 API 文档
│   ├── architecture.md    # 系统架构文档
│   └── channels/
│       └── feishu.md      # 飞书渠道集成文档
├── cmd/api/               # API 入口
├── cmd/migrate/           # 迁移工具入口
├── Makefile               # 构建/测试/运行命令
├── docker-compose.yml     # Docker 编排
└── README.md
```

## 核心设计

### ReAct + SOP 工作流

SOP 流程定义包含四种步骤类型，创建工单时自动实例化为 Workflow：

| 步骤类型 | 说明 |
|---------|------|
| `react` | LLM 驱动的 ReAct 循环：思考→决策→执行（search_kb/reply_user/update_ticket/complete/fail），最多 10 轮 |
| `human` | 人工介入步骤，等待用户处理（可配置超时，默认 24h） |
| `condition` | 条件判断，根据工单字段（status/priority/assignee）自动决定是否跳过 |
| `notification` | 发送通知给指定用户 |

Workflow 步骤使用 `sync.Map` 互斥锁防止并发 ReAct 循环冲突。

### Agent 引擎

Agent 在工单创建/更新时自动触发执行：

- **rule** 类型：执行预设动作（auto_assign/set_status/set_priority）
- **llm** 类型：搜索知识库 → LLM 生成回复 → 添加为内部评论
- **hybrid** 类型：先执行 rule，再执行 llm

Agent 匹配规则支持按触发事件（ticket_created/ticket_updated/any）筛选。

### ChatService 与 EmbeddingService

- **EmbeddingService**: 封装向量嵌入（Embed/BatchEmbed），供知识库向量检索使用
- **ChatService**: 封装 LLM 对话（Chat），供 Workflow ReAct 循环和 Agent LLM 使用
- `NewChatService(provider)` 在 provider 不支持 Chat 时返回 nil（如 DashScope），避免运行时类型断言失败

## 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `SERVER_PORT` | 服务器端口 | `8080` |
| `SERVER_MODE` | 运行模式 (debug/release) | `debug` |
| `LOG_LEVEL` | 日志级别 | `info` |
| `LOG_FORMAT` | 日志格式 (json/console) | `json` |
| `DB_HOST` | 数据库主机 | `localhost` |
| `DB_PORT` | 数据库端口 | `5432` |
| `DB_USER` | 数据库用户 | `postgres` |
| `DB_PASSWORD` | 数据库密码 | `postgres` |
| `DB_NAME` | 数据库名称 | `husky` |
| `DB_SSLMODE` | SSL 模式 | `disable` |
| `DB_MAX_IDLE_CONNS` | 最大空闲连接数 | `10` |
| `DB_MAX_OPEN_CONNS` | 最大打开连接数 | `100` |
| `DB_CONN_MAX_LIFETIME` | 连接最大生命周期 | `3600` |
| `REDIS_HOST` | Redis 主机 | `localhost` |
| `REDIS_PORT` | Redis 端口 | `6379` |
| `REDIS_PASSWORD` | Redis 密码 | `` |
| `REDIS_DB` | Redis 数据库编号 | `0` |
| `REDIS_POOL_SIZE` | Redis 连接池大小 | `100` |
| `JWT_SECRET` | JWT 签名密钥 | `husky-secret-key-...` |
| `JWT_EXPIRE_HOUR` | JWT 过期时间（小时） | `24` |
| `LLM_PROVIDER` | LLM 服务商 (openai/dashscope) | `openai` |
| `LLM_API_KEY` | LLM API 密钥 | `` |
| `LLM_BASE_URL` | LLM API 地址 | `https://api.openai.com/v1` |
| `FEISHU_APP_ID` | 飞书应用 App ID | `` |
| `FEISHU_APP_SECRET` | 飞书应用 App Secret | `` |
| `ADMIN_MODE` | 管理后台部署模式 (embedded/external) | `embedded` |
| `ADMIN_URL` | external 模式下前端 dev server URL | `http://localhost:5173` |

完整配置项请参考 [.env.example](.env.example)

## API 概要

### 认证
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/login` | 登录 |
| POST | `/api/v1/auth/register` | 注册 |
| POST | `/api/v1/auth/refresh` | 刷新 Token |

### 工单
| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | `/api/v1/tickets` | 列表/创建 |
| GET/PUT/DELETE | `/api/v1/tickets/:id` | 详情/更新/删除 |
| POST | `/api/v1/tickets/:id/assign` | 分配 |
| POST | `/api/v1/tickets/:id/auto-assign` | 自动分配 |
| POST | `/api/v1/tickets/:id/claim` | 认领 |
| POST | `/api/v1/tickets/:id/status` | 更新状态 |
| POST | `/api/v1/tickets/:id/due` | 设置截止时间 |
| POST | `/api/v1/tickets/:id/rate` | 评价 |
| POST/GET | `/api/v1/tickets/:id/comments` | 评论 |
| POST/GET/DELETE | `/api/v1/tickets/:id/attachments` | 附件 |
| POST/DELETE | `/api/v1/tickets/:id/watch` | 关注 |
| GET/PUT | `/api/v1/tickets/:id/tags` | 工单标签管理 |
| POST | `/api/v1/tickets/:id/relations` | 工单关联 |

### 知识库
| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | `/api/v1/knowledge` | 列表/创建 |
| GET/PUT/DELETE | `/api/v1/knowledge/:id` | 详情/更新/删除 |
| POST | `/api/v1/knowledge/search` | 向量搜索 |
| POST/GET | `/api/v1/knowledge/import` | 批量导入 |
| GET | `/api/v1/knowledge/export` | 批量导出 |
| GET | `/api/v1/knowledge/recommend` | 推荐知识 |
| GET | `/api/v1/knowledge/categories` | 知识库分类 |

### Agent
| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | `/api/v1/agents` | 列表/创建 |
| GET/PUT/DELETE | `/api/v1/agents/:id` | 详情/更新/删除 |

### SOP
| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | `/api/v1/sop` | 列表/创建 |
| GET/PUT/DELETE | `/api/v1/sop/:id` | 详情/更新/删除 |

### Workflow
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/workflows` | 列表 |
| GET | `/api/v1/workflows/:id` | 详情 |
| POST | `/api/v1/workflows/steps/:stepId/complete` | 完成人工步骤 |
| GET | `/api/v1/workflows/tasks` | 当前用户的待办步骤 |

### 标签
| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | `/api/v1/tags` | 列表/创建 |
| GET/PUT/DELETE | `/api/v1/tags/:id` | 详情/更新/删除 |

### 渠道
| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | `/api/v1/channels` | 列表/创建 |
| GET/PUT/DELETE | `/api/v1/channels/:id` | 详情/更新/删除 |
| GET/POST/PUT/DELETE | `/api/v1/channel-users` | 渠道用户管理 |
| GET/POST/PUT/DELETE | `/api/v1/channel-groups` | 渠道群组管理 |
| GET/POST | `/api/v1/channels/:channel/bot-config` | Bot 配置 |

### 菜单管理
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/menus` | 当前用户菜单树 |
| GET/POST | `/api/v1/menus/all` | 全部菜单（管理员） |
| GET/PUT/DELETE | `/api/v1/menus/:id` | 详情/更新/删除 |

### 其他
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/users/me` | 当前用户信息 |
| GET/PUT/DELETE | `/api/v1/users/:id` | 用户管理（管理员） |
| GET | `/api/v1/departments` | 部门树 |
| GET | `/api/v1/categories` | 分类列表 |
| GET/PUT | `/api/v1/notifications` | 通知 |
| GET | `/api/v1/notifications/unread-count` | 未读通知数 |
| GET/POST | `/api/v1/sla-configs` | SLA 配置管理 |
| GET/POST | `/api/v1/webhook-configs` | Webhook 配置管理 |
| GET/POST/PUT/DELETE | `/api/v1/roles` | 角色管理 |
| GET/POST | `/api/v1/assign-config` | 分配策略配置 |
| GET/POST/PUT/DELETE | `/api/v1/skills` | Skill 管理 |
| GET/POST/PUT/DELETE | `/api/v1/mcps` | MCP 服务管理 |
| GET | `/api/v1/stats/*` | 统计报表 |
| GET/POST | `/api/v1/system-config` | 系统配置 |
| POST | `/webhooks/lark` | 飞书 Webhook |
| POST | `/webhooks/dingtalk` | 钉钉 Webhook |
| POST | `/webhooks/wecom` | 企微 Webhook |
| GET | `/admin` | Admin 管理 UI |

详细 API 文档请参考 [docs/basic.md](docs/basic.md)

## 测试

```bash
go test ./...                    # 运行所有测试
go test -coverprofile=coverage.out ./...   # 覆盖率
```

## 开发指南

### 代码规范
- 使用 `gofmt -s -w .` 格式化代码
- 遵循 [Effective Go](https://golang.org/doc/effective_go.html)

### 提交规范
```
feat:     新功能
fix:      修复 bug
docs:     文档更新
refactor: 重构
test:     测试
chore:    构建/工具链
```

## 致谢

- [Gin](https://github.com/gin-gonic/gin) - Web 框架
- [GORM](https://github.com/go-gorm/gorm) - ORM 库
- [pgvector](https://github.com/pgvector/pgvector) - 向量数据库扩展
- [Redis](https://redis.io/) - 缓存
