# Husky 系统架构

## 目录

1. [整体架构](#1-整体架构)
2. [LLM 服务层](#2-llm-服务层)
3. [Agent 引擎](#3-agent-引擎)
4. [ReAct + SOP 工作流](#4-react--sop-工作流)
5. [飞书渠道集成](#5-飞书渠道集成)
6. [Admin 管理 UI](#6-admin-管理-ui)
7. [数据流与关键路径](#7-数据流与关键路径)

---

## 1. 整体架构

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
│  │      └───└─────────────────────────┘                      │    │
│  └──────────────────────────────────────────────────────────┘    │
│                              │                                    │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │                  LLM 服务层                                │    │
│  │  ┌─────────────────┐  ┌─────────────────┐                 │    │
│  │  │ EmbeddingService │  │   ChatService    │                 │    │
│  │  └────────┬────────┘  └────────┬────────┘                 │    │
│  │           │                    │                            │    │
│  │  ┌────────▼────────────────────▼────────┐                  │    │
│  │  │    Provider (OpenAI / DashScope)      │                  │    │
│  └──────────────────────────────────────────────────────────┘    │
│                              │                                    │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │                  Admin 管理 UI (Go embed)                 │    │
│  └──────────────────────────────────────────────────────────┘    │
│                              │                                    │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │                  数据访问层 (Repository)                   │    │
│  └──────────────────────────────────────────────────────────┘    │
│                              │                                    │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │                  数据存储层                                │    │
│  │    PostgreSQL (+pgvector) │ Redis                        │    │
│  └──────────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────────┘
```

## 2. LLM 服务层

### 2.1 接口定义

`pkg/llm/provider.go` 定义了两个核心接口：

```go
// EmbeddingProvider 文本向量化
type EmbeddingProvider interface {
    Embed(ctx context.Context, text string) ([]float32, error)
    BatchEmbed(ctx context.Context, texts []string) ([][]float32, error)
    Name() string
}

// ChatProvider LLM 对话
type ChatProvider interface {
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
}
```

### 2.2 EmbeddingService

`pkg/llm/embedding.go`

职责：封装向量嵌入能力，供知识库使用。

- `Embed(text)` — 单段文本向量化
- `Dimension()` — 返回向量维度（OpenAI: 1536, DashScope: 1536）

使用场景：知识库创建时生成向量、搜索时将查询文本转为向量用于 pgvector 余弦相似度检索。

### 2.3 ChatService

`pkg/llm/chat.go`

职责：封装 LLM 对话能力，供 ReAct 循环和 Agent LLM 使用。

- `Chat(req)` — 发送对话请求，返回 LLM 回复

使用场景：
- WorkflowService.reActLoop — LLM 在 search_kb/reply_user/update_ticket/complete/fail 间决策
- AgentEngine.runLLMAgent — 搜索知识库后生成回复

### 2.4 为什么分离

原有设计通过 `EmbeddingService.Provider().(ChatProvider)` 类型断言获取对话能力，存在两个问题：

1. **泄漏抽象**：对话能力不应通过嵌入服务暴露
2. **静默失败**：DashScope 仅支持 Embedding 不支持 Chat，运行时类型断言失败但被 `!ok` 检查静默吞掉

重构后：
- `ChatService` 独立构造，provider 不支持 Chat 时返回 nil
- 调用方显式检查 `chatSvc == nil`，语义清晰
- 未来可以支持不同的 Embedding 和 Chat 提供商组合

### 2.5 Provider 实现

| Provider | Embedding | Chat | 文件 |
|----------|-----------|------|------|
| OpenAI | ✅ text-embedding-ada-002 | ✅ gpt-3.5-turbo/gpt-4 | `openai.go` |
| DashScope | ✅ text-embedding-v2 | ❌ | `dashscope.go` |

## 3. Agent 引擎

`internal/service/agent_engine.go`

### 3.1 触发机制

Agent 在工单生命周期事件中自动触发：

```
OnTicketCreated → executeAgents("ticket_created", ticket)
OnTicketUpdated → executeAgents("ticket_updated", ticket)
```

### 3.2 执行流程

```
executeAgents()
  ├─ 遍历所有启用的 Agent
  │   └─ matchesTrigger() → 检查 Agent 配置的 trigger_on 字段
  │       └─ runAgent()
  │           ├─ type=rule    → runRuleAgent()   预设动作
  │           ├─ type=llm     → runLLMAgent()    知识检索+LLM回复
  │           └─ type=hybrid  → runRuleAgent() + runLLMAgent()
  │
  └─ matchAndCreateWorkflows()
      └─ 遍历启用的 SOP
          └─ 匹配 trigger_type + TriggerConfig
              └─ CreateFromSOP() → 实例化 Workflow
```

### 3.3 Agent 类型

#### rule（规则型）
执行预设动作，无需 LLM 调用：
- `auto_assign` — 自动分配给负载最轻的客服
- `set_status` — 设置工单状态
- `set_priority` — 设置工单优先级

#### llm（LLM 型）
```
搜索知识库 → 构建 Prompt → LLM 生成回复 → 添加为内部评论
```

#### hybrid（混合型）
先执行 rule 动作，再执行 llm 生成回复。

### 3.4 SOP 匹配

Agent 引擎同时负责匹配 SOP 并创建 Workflow。SOP 的匹配条件：
1. `status == 1` (启用)
2. `trigger_type == event_type || trigger_type == "event"`
3. `TriggerConfig` 中的 source/priority 条件匹配

### 3.5 Agent 选择（Workflow ReAct）

当 Workflow 的 react 步骤未指定 AgentID 时，`selectAgent()` 按以下规则选择：
1. 如果步骤指定了 AgentID，使用该 Agent
2. 否则在所有启用的 llm/hybrid Agent 中选择 ID 最小的（确定性稳定排序）

## 4. ReAct + SOP 工作流

`internal/service/workflow_service.go`

### 4.1 SOP 步骤定义

SOP 以 JSON 数组定义步骤：

```json
[
  {"name": "分析问题", "type": "react", "goal": "收集信息并分析工单描述"},
  {"name": "客服确认", "type": "human", "assign_to": "requester", "config": {"timeout_minutes": 1440}},
  {"name": "是否已解决", "type": "condition", "config": {"field": "status", "operator": "eq", "value": "resolved"}},
  {"name": "发送通知", "type": "notification", "config": {"title": "工单已关闭", "content": "..."}}
]
```

### 4.2 步骤类型

| 类型 | 执行器 | 说明 |
|------|--------|------|
| `react` | `reActLoop()` | LLM 驱动的自主推理循环 |
| `human` | `executeHumanStep()` | 等待人工处理，超时自动标记失败 |
| `condition` | `executeConditionStep()` | 字段值判断，匹配则继续，否则跳过 |
| `notification` | `executeNotificationStep()` | 创建通知记录 |

### 4.3 ReAct 循环

```
reActLoop(step, ticket, wf)
  │
  ├─ 1. 检查 ChatService 是否可用
  ├─ 2. 选择 Agent（获取 model/temperature/max_tokens）
  ├─ 3. 搜索知识库构建上下文
  ├─ 4. 构建 System Prompt（包含可用的工具、目标、知识上下文）
  │
  └─ for round = 0..9:
      ├─ 构建消息（system prompt + 轮次历史滑动窗口最近5轮）
      ├─ 调用 LLM Chat
      ├─ 解析 JSON 格式的 ReActAction
      │   └─ 解析失败 → 记录错误并要求 LLM 重新输出 JSON
      │
      ├─ switch action:
      │   ├─ search_kb    → 向量搜索，追加观测结果到历史
      │   ├─ reply_user   → 添加公开评论
      │   ├─ update_ticket → 更新 status/priority
      │   ├─ complete     → 标记步骤完成，跳出
      │   └─ fail         → 标记步骤失败，跳出
      │
      └─ 达到 10 轮 → 自动 complete
```

**关键设计：**
- **滑动窗口**：只保留最近 5 轮历史，避免 systemPrompt 无限制增长
- **并发锁**：`sync.Map` 按 `stepID` 加锁，防止同一步骤的重复 ReAct 循环
- **JSON 重试**：LLM 输出非 JSON 时重新要求输出，不中断流程
- **确定性选择**：stable sort by Agent.ID 确保多次执行结果一致

### 4.4 工作流生命周期

```
SOP → CreateFromSOP()
  │
  ├─ 检查是否已有 Workflow（幂等）
  ├─ 解析 SOP Steps JSON
  ├─ 创建 Workflow（status=pending）
  ├─ 创建每个 WorkflowStep（status=pending）
  ├─ 设置 status=running
  │
  └─ executeNextStep():
      ├─ 获取 pending 步骤
      ├─ 设为 running
      ├─ executeStep():
      │   ├─ react        → reActLoop() (goroutine)
      │   ├─ human        → executeHumanStep() (goroutine + timeout)
      │   ├─ condition    → executeConditionStep() (同步)
      │   └─ notification → executeNotificationStep() (同步)
      │
      └─ completeStep() → executeNextStep() 递归
          └─ 无 pending → 设 workflow status=completed

OnUserMessage():
  └─ 找到 running 的 react 步骤
      └─ reActLoop() (带 sync.Map 互斥锁)
```

## 5. 飞书渠道集成

`pkg/feishu/client.go` + `internal/service/ticket_group_service.go`

### 5.1 工单自动建群

创建工单时：自动创建飞书群 → 添加相关人员 → 发送欢迎卡片

### 5.2 群聊消息处理

```
用户发消息 → 飞书 Webhook → /webhooks/lark
  └─ 解析消息类型:
      ├─ 群聊消息:
      │   ├─ 查找 ticket_groups 绑定关系
      │   ├─ 搜索知识库自动回复（卡片消息）
      │   └─ 检测人工升级关键词 → 通知负责人
      │
      └─ 私聊消息:
          ├─ 创建新工单
          ├─ 自动创建飞书群
          └─ 发送欢迎卡片
```

## 6. Admin 管理 UI

`internal/admin/`

### 6.1 实现方式

- 单页 HTML 文件（`static/index.html`），全部 CSS/JS 内联
- 通过 `//go:embed static/*` 编译进 Go 二进制
- 零构建工具链，无需 npm/webpack
- SPA fallback：未知路径返回 index.html

### 6.2 路由

```go
GET /admin        → index.html
GET /admin/*      → index.html (SPA fallback for non-file paths)
GET /admin/*.css  → static file
GET /admin/*.js   → static file
```

### 6.3 功能

- 登录页 → JWT 认证
- 仪表盘 → 工单概览统计
- 工单管理 → CRUD、筛选、详情、评论
- 知识库管理 → 创建/编辑/搜索/导入/导出
- Agent 管理 → 创建/编辑 llm/rule/hybrid Agent
- SOP 管理 → 定义步骤流程
- Workflow 跟踪 → 查看运行中的工作流
- 用户管理 → 列表、角色变更
- 部门/分类/渠道配置

## 7. 数据流与关键路径

### 7.1 创建工单 → Agent → SOP → Workflow → ReAct

```
POST /api/v1/tickets
  → TicketHandler.CreateTicket()
    → ticketService.Create()
    → ticketGroupSvc.CreateTicketGroup() (飞书建群)
    → agentEngine.OnTicketCreated()
      → executeAgents("ticket_created")
        → runAgent() (rule/llm/hybrid)
        → matchAndCreateWorkflows()
          → CreateFromSOP()
            → executeNextStep()
              → reActLoop()
```

### 7.2 知识库搜索

```
POST /api/v1/knowledge/search
  → KnowledgeHandler.SearchKnowledge()
    → knowledgeService.SearchKnowledge()
      → embedding.Embed(query) → 向量
      → vectorRepo.SearchSimilar(query, vector) → pgvector 余弦相似度
      → fallback: ILIKE 文本搜索（无向量时）
```

### 7.3 Agent LLM 回复

```
ticket created/updated
  → AgentEngine.runLLMAgent()
    → kbRepo.SearchSimilar() → 检索相关知识
    → chatSvc.Chat()         → LLM 生成回复
    → ticketRepo.AddComment() → 添加为内部评论
```
