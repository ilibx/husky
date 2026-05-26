# Husky - 企业级智能工单系统

[![Go Version](https://img.shields.io/badge/go-1.19+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-in%20development-yellow.svg)](TODO.md)

## 📖 项目概述

Husky 是一个基于 Golang 实现的企业级智能工单处理系统，类似飞书服务台功能。系统集成了 RAG 知识库、工单处理 SOP 管理、HITL（人在回路）交互流程，支持多渠道接入，提供完整的工单管理与统计分析能力。

## ✨ 核心特性

### 🎯 工单管理系统
- 完整的工单生命周期管理（创建→分配→处理→关闭）
- 灵活的状态机配置和自动流转规则
- 智能工单分配（轮询、负载均衡、技能匹配）
- 工单评论、附件、协作功能
- SLA 超时升级机制

### 🤖 AI 智能能力
- **RAG 知识库**: 基于向量检索的增强生成，提供精准的智能问答
- **SOP 管理**: 标准作业程序定义、执行引擎、流程跟踪
- **HITL 交互**: 人机协作、AI 建议人工审核、反馈学习
- **Agent 系统**: 内置多个智能 Agent（自动分类、派单、回复、情绪识别等）

### 📱 多渠道接入
- ✅ 飞书 (Feishu)
- ✅ Lark (国际版)
- ✅ 钉钉 (DingTalk)
- ✅ 企业微信 (WeCom)
- ✅ Email (IMAP/SMTP)
- ✅ WebUI (响应式管理后台)

### 📊 统计与分析
- 工单总量、状态分布、来源渠道统计
- 处理时效分析、SLA 达标率
- 客服绩效统计
- 自定义报表与数据导出

### 🔐 安全与权限
- JWT Token 认证
- RBAC 基于角色的权限控制
- 操作审计日志
- 数据加密存储

## 🏗️ 技术架构

```
┌─────────────────────────────────────────────────────────────┐
│                      Husky 系统架构                          │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────┐    │
│  │                   接入层 (Channels)                  │    │
│  │  飞书 │ Lark │ 钉钉 │ 企微 │ Email │ WebUI │ API   │    │
│  └─────────────────────────────────────────────────────┘    │
│                              │                                │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                    API Gateway                       │    │
│  │           (认证/限流/路由/日志)                       │    │
│  └─────────────────────────────────────────────────────┘    │
│                              │                                │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                   业务服务层                         │    │
│  │  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐     │    │
│  │  │工单  │ │用户  │ │知识  │ │SOP   │ │Agent │     │    │
│  │  │服务  │ │服务  │ │服务  │ │服务  │ │服务  │     │    │
│  │  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘     │    │
│  │  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐              │    │
│  │  │RAG   │ │通知  │ │统计  │ │渠道  │              │    │
│  │  │服务  │ │服务  │ │服务  │ │服务  │              │    │
│  │  └──────┘ └──────┘ └──────┘ └──────┘              │    │
│  └─────────────────────────────────────────────────────┘    │
│                              │                                │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                   数据访问层                         │    │
│  │         Repository 模式 + ORM (GORM)                 │    │
│  └─────────────────────────────────────────────────────┘    │
│                              │                                │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                   数据存储层                         │    │
│  │    PostgreSQL (+pgvector) │ Redis │ 对象存储        │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

## 🚀 快速开始

### 前置要求

- Go 1.19+
- Docker & Docker Compose
- PostgreSQL 14+ (推荐带 pgvector 扩展)
- Redis 6+

### 本地开发环境

#### 1. 克隆项目

```bash
git clone <repository-url>
cd husky
```

#### 2. 启动基础设施

```bash
# 使用 Docker Compose 启动 PostgreSQL 和 Redis
docker-compose up -d postgres redis
```

#### 3. 配置环境变量

```bash
cp configs/.env.example .env
# 编辑 .env 文件配置必要的环境变量
```

#### 4. 安装依赖

```bash
go mod download
```

#### 5. 运行应用

```bash
# 运行 API 服务
go run cmd/api/main.go

# 或者编译后运行
go build -o bin/husky ./cmd/api
./bin/husky
```

#### 6. 完整服务启动

```bash
# 启动所有服务（包括 API 和 Worker）
docker-compose up -d
```

### 验证安装

访问健康检查端点：

```bash
curl http://localhost:8080/health
```

预期响应：

```json
{
  "status": "healthy"
}
```

## 📁 项目结构

```
husky/
├── cmd/                    # 应用程序入口
│   ├── api/               # API 服务
│   └── worker/            # 异步任务处理器
├── internal/              # 私有应用代码
│   ├── config/           # 配置管理
│   ├── handler/          # HTTP 处理器
│   ├── middleware/       # 中间件
│   ├── model/            # 数据模型
│   ├── repository/       # 数据访问层
│   ├── router/           # 路由定义
│   ├── service/          # 业务逻辑层
│   └── utils/            # 工具函数
├── pkg/                   # 公共库代码
│   ├── cache/            # 缓存封装
│   ├── errors/           # 错误处理
│   ├── logger/           # 日志封装
│   └── validator/        # 参数验证
├── web/                   # 前端代码
│   ├── src/              # 源代码
│   └── public/           # 静态资源
├── configs/               # 配置文件
├── deployments/           # 部署配置
│   ├── Dockerfile        # API 服务 Docker 镜像
│   └── docker-compose.yml
├── docs/                  # 文档
├── scripts/               # 脚本工具
├── tests/                 # 测试代码
├── go.mod                # Go 模块定义
├── go.sum                # 依赖校验
├── docker-compose.yml    # Docker Compose 配置
├── README.md             # 项目说明
└── TODO.md               # 项目实施计划
```

## 📋 开发计划

详细的开发计划和任务清单请查看 [TODO.md](TODO.md)

### 当前进度

| 阶段 | 状态 | 完成度 |
|------|------|--------|
| 阶段一：项目基础架构搭建 | 🟢 进行中 | 80% |
| 阶段二：核心数据模型与数据库设计 | ⚪ 待开始 | 0% |
| 阶段三：用户认证与权限系统 | ⚪ 待开始 | 0% |
| 阶段四：工单核心系统 | ⚪ 待开始 | 0% |
| 阶段五：RAG 知识库系统 | ⚪ 待开始 | 0% |
| 阶段六：SOP 管理系统 | ⚪ 待开始 | 0% |
| 阶段七：HITL 交互流程 | ⚪ 待开始 | 0% |
| 阶段八：多渠道接入系统 | ⚪ 待开始 | 0% |
| 阶段九：Web 管理后台 | ⚪ 待开始 | 0% |
| 阶段十：工单查询与统计系统 | ⚪ 待开始 | 0% |

**图例**: 🟢 进行中 | 🔵 已完成 | ⚪ 待开始 | 🔴 阻塞

### 里程碑

- **M1** (第 2-3 周): 基础架构、数据库、认证系统
- **M2** (第 5-7 周): 工单核心系统、RAG 知识库
- **M3** (第 9-12 周): SOP 管理、HITL、渠道接入（至少 2 个）
- **M4** (第 13-16 周): Web 管理后台、统计系统
- **M5** (第 17-20 周): Agent 系统、通知系统、API 文档
- **M6** (第 21-24 周): 安全加固、性能优化、完整测试
- **M7** (第 25-27 周): 生产部署、监控、文档
- **M8** (持续): MVP 发布、迭代优化

## 🔧 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `SERVER_PORT` | 服务器端口 | `8080` |
| `SERVER_MODE` | 运行模式 (debug/release/test) | `debug` |
| `LOG_LEVEL` | 日志级别 (debug/info/warn/error) | `info` |
| `LOG_FORMAT` | 日志格式 (json/console) | `json` |
| `DB_HOST` | 数据库主机 | `localhost` |
| `DB_PORT` | 数据库端口 | `5432` |
| `DB_USER` | 数据库用户 | `postgres` |
| `DB_PASSWORD` | 数据库密码 | `postgres` |
| `DB_NAME` | 数据库名称 | `husky` |
| `REDIS_HOST` | Redis 主机 | `localhost` |
| `REDIS_PORT` | Redis 端口 | `6379` |
| `JWT_SECRET` | JWT 密钥 | - |
| `LLM_PROVIDER` | LLM 服务商 | `openai` |
| `LLM_API_KEY` | LLM API 密钥 | - |

完整配置项请参考 [configs/.env.example](configs/.env.example)

## 📚 API 文档

API 文档将在开发完成后通过 Swagger/OpenAPI 自动生成。

预览地址：`http://localhost:8080/swagger/index.html`

### 主要 API 端点

- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/logout` - 用户登出
- `GET /api/v1/tickets` - 获取工单列表
- `POST /api/v1/tickets` - 创建工单
- `GET /api/v1/knowledge` - 获取知识库列表
- `POST /api/v1/knowledge` - 创建知识文档
- `GET /api/v1/stats/overview` - 获取统计概览

详细 API 文档请参考 [docs/API.md](docs/API.md) (待完善)

## 🧪 测试

```bash
# 运行单元测试
go test ./...

# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# 运行集成测试
go test -tags=integration ./tests/integration/...
```

## 📝 开发指南

### 代码规范

本项目遵循 Go 语言最佳实践和代码规范：

- 使用 `gofmt` 格式化代码
- 遵循 [Effective Go](https://golang.org/doc/effective_go.html)
- 使用 `golint` 进行代码检查
- 编写清晰的注释和文档

### 提交规范

```
feat: 新功能
fix: 修复 bug
docs: 文档更新
style: 代码格式调整
refactor: 重构代码
test: 测试相关
chore: 构建/工具链相关
```

### 分支策略

- `main` - 主分支，生产环境代码
- `develop` - 开发分支
- `feature/*` - 功能分支
- `bugfix/*` - 修复分支
- `release/*` - 发布分支

## 🤝 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建你的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交你的修改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启一个 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📞 联系方式

- 项目 Issues: [GitHub Issues](../../issues)
- 讨论区：[GitHub Discussions](../../discussions)

## 🙏 致谢

感谢以下开源项目：

- [Gin](https://github.com/gin-gonic/gin) - Web 框架
- [GORM](https://github.com/go-gorm/gorm) - ORM 库
- [Zap](https://github.com/uber-go/zap) - 日志库
- [pgvector](https://github.com/pgvector/pgvector) - 向量数据库扩展
- [Redis](https://redis.io/) - 缓存数据库

---

**Husky** - 让工单处理更智能、更高效！🚀
