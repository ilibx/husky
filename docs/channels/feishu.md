# 飞书渠道集成

## 配置

### 环境变量

```bash
# 飞书应用凭证（必填）
export FEISHU_APP_ID=cli_xxxxxxxxxxxxxx
export FEISHU_APP_SECRET=xxxxxxxxxxxxxxxxxxxxxxxx

# LLM API Key（可选，启用知识库自动回复）
export LLM_PROVIDER=openai
export LLM_API_KEY=sk-xxx
```

### 飞书应用权限

在飞书开发者后台为应用添加以下权限：

| 权限 | 说明 |
|------|------|
| `im:chat:create` | 创建群聊 |
| `im:chat:member:add` | 添加群成员 |
| `im:message:send` | 发送消息 |
| `im:message:receive` | 接收消息事件 |
| `contact:user.base` | 获取用户信息 |

### Webhook 配置

在飞书应用的事件订阅中，添加以下事件：

- `im.message.receive_v1` — 接收用户消息

URL 配置为：`https://your-domain/webhooks/lark`

## 功能

### 工单自动建群

工单创建时会自动创建一个飞书群聊：

- 群名称格式：`工单 TK20260101XXXXXX: 工单标题`
- 群描述包含工单号、优先级、状态、创建时间
- 自动发送欢迎卡片，包含工单信息和群链接

### 用户加入群聊

用户可以通过以下方式加入工单群：

1. **直接访问群链接**：`https://applink.feishu.cn/client/chat/open?chat_id=xxx`
2. **API 接口加入**：`POST /api/v1/feishu/join-group?ticket_id=1&user_id=open_id`

### 自动回复

当用户在群内发送消息时，机器人会自动：

1. 检测消息是否来自绑定工单的群聊
2. 搜索知识库获取相关内容
3. 以卡片消息格式返回匹配的知识条目

### 人工升级

当用户消息包含以下关键词时，自动转接人工客服：

- `人工客服`、`转人工`、`客服`、`转接`
- `human`、`agent`、`help`

转接后自动通知工单负责人和提交人。

## API

### 加入工单群

```
POST /api/v1/feishu/join-group
Authorization: Bearer <token>
Query: ticket_id=<工单ID>&user_id=<飞书open_id>
```

## 架构流程

```
用户发送消息 ──→ 飞书 Webhook ──→ 检测 chat_id
                                        │
                      ┌─────────────────┤
                      ▼                 ▼
                 群聊消息          私聊消息
                      │                 │
                      ▼                 ▼
             查找 ticket_groups   创建新工单
             绑定关系                 │
                      │                 ▼
                      ▼            自动创建群
             搜索知识库                 │
             自动回复卡片               ▼
                      │            发送欢迎卡片
                      ▼
             检测升级关键词
                      │
                      ▼
             拉入客服 + 通知
```
