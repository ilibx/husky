# Husky 改善方案

## 一、致命问题（需立即修复）

### 1.1 创建工单后执行路径重复
**问题**：`internal/ticket/handler.go:74-97`（goroutine）和 `internal/gateway/gateway.go:115,337`（onTicketCreated）都会执行 CreateGroupForTicket + OnTicketCreated。通过渠道创建的工单会触发两次。

**方案**：统一在 Service 层的事件回调中处理，消除 handler 中的 goroutine。或在 gateway 中创建后不再触发 handler goroutine。

### 1.2 SLA 升级器 context 泄漏
**问题**：`internal/ticket/sla.go:45` 使用 `context.Background()` 而不是传递给 Start 的 ctx，导致优雅关闭无法停止正在进行的 SLA 检查。

**方案**：将 ctx 存入结构体，在 tick 时使用存储的 ctx。

### 1.3 人工步骤超时 goroutine 泄漏
**问题**：`internal/agent/workflow.go:398-409` 中 goroutine 使用 `time.After` 且 `ctx` 是 `context.Background()`，步骤提前完成后 goroutine 直到 timeout 才释放。

**方案**：使用 `context.WithTimeout` 传入可取消的上下文，并在步骤完成时调用 cancel。

### 1.4 OnUserMessage 静默丢弃消息
**问题**：`internal/agent/workflow.go:593-595` 使用 `sync.Map` 做互斥锁，ReAct 循环运行时用户消息被丢弃。

**方案**：使用带缓冲 channel 的消息队列，循环结束后重放。

---

## 二、架构改进

### 2.1 Agent 引擎拆分
**问题**：`internal/agent/engine.go:46-49` 中 `OnTicketCreated` 同时处理 Agent 执行和 SOP→Workflow 创建。

**方案**：将 SOP 匹配抽离为独立的 `SOPMatcher` 服务，分别注入事件回调。

### 2.2 runRuleAgent 绕过 Service 层
**问题**：`internal/agent/engine.go:247-283` 直接调用 `ticketRepo.UpdateFields()`，跳过审计日志、状态验证和通知。

**方案**：将规则 Agent 的动作委托给对应的 Service 方法。

### 2.3 KnowledgeBase UUID 主键兼容性
**问题**：`internal/model/knowledge.go:13` 使用 UUID string 主键和 CreatedBy，其他模型使用 uint。

**方案**：统一使用 uint 主键，或至少增加 CreatedBy 与 User 表的关联映射。

### 2.4 全局可变状态
**问题**：`internal/config/config.go:67` 的 `var Conf *Config` 和 `internal/agent/engine.go:17` 的 `var AgentUserID uint`。

**方案**：通过依赖注入传递配置，去除包级全局变量。

---

## 三、性能优化

### 3.1 飞书 enricher 缓存淘汰
**问题**：`internal/gateway/enricher_feishu.go:71-83` 中 TTL 检查后不删除条目，map 无限增长。

**方案**：TTL 过期后删除条目，或使用 `sync.Map` + 定期清理。

### 3.2 executeAgents/selectSOP 使用 SQL 过滤
**问题**：`internal/agent/engine.go:56,97` 每次加载所有 Agent/SOP 后在 Go 代码过滤。

**方案**：在 Repository 层增加带 WHERE 条件的查询方法。

### 3.3 知识库导出无分页
**问题**：`internal/knowledge/service.go:173` 使用 `limit=99999`。

**方案**：使用流式导出或分页，或改为服务端流式写入。

### 3.4 LLM 调用无限速
**问题**：所有 LLM 调用直接并发，无令牌桶或信号量。

**方案**：在 ChatService 层实现令牌桶限速。

---

## 四、AI/LLM 智能改进

### 4.1 意图分类 JSON 解析健壮性
**问题**：`internal/intent/service.go:89-91` 简单的字符串 Trim，对 LLM 输出格式变化敏感。

**方案**：使用更鲁棒的正则提取 JSON，增加重试机制。

### 4.2 ReAct Agent 选择策略
**问题**：`internal/agent/workflow.go:346` 默认选择 ID 最小的 Agent。

**方案**：增加权重/负载感知的选择策略，或支持配置化策略。

### 4.3 降级搜索得分动态化
**问题**：`internal/repository/vector_store.go:137` 降级搜索固定返回 Score=0.5。

**方案**：根据 ILIKE 匹配程度计算动态得分（完全匹配 1.0，模糊匹配逐渐降低）。

### 4.4 HITL 反馈闭环
**问题**：`internal/agent/workflow.go:641-738` 的 Decision/Feedback/RejectionReason 仅存储不用于模型改进。

**方案**：将反馈累积为 few-shot 示例库，定期优化提示词。
