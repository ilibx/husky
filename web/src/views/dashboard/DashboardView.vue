<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import request from '@/api/request'

const router = useRouter()

function notifTypeTag(type: string) {
  const map: Record<string, string> = { ticket_assigned: 'primary', sla: 'warning', ticket_comment: 'success', system: 'info' }
  return map[type] || 'info'
}
function notifTypeLabel(type: string) {
  const map: Record<string, string> = { ticket_assigned: '工单分配', sla: 'SLA 提醒', ticket_comment: '工单评论', system: '系统通知' }
  return map[type] || type
}
function formatTime(t: string) {
  if (!t) return ''
  return t.slice(0, 16).replace('T', ' ')
}

interface Overview {
  total: number
  open: number
  in_progress: number
  resolved: number
  closed: number
  overdue: number
}

interface BySource {
  [source: string]: number
}
interface ByPriority {
  [priority: string]: number
}

interface ChannelConfig {
  id: number
  name: string
  type: string
  enabled: boolean
  status: number
}

interface Category {
  id: number
  name: string
  description?: string
  status: number
}

const overview = ref<Overview>({ total: 0, open: 0, in_progress: 0, resolved: 0, closed: 0, overdue: 0 })
const bySource = ref<BySource>({})
const byPriority = ref<ByPriority>({})
const satisfaction = ref({ total: 0, average: 0, distribution: {} as Record<number, number> })
const channels = ref<ChannelConfig[]>([])
const categories = ref<Category[]>([])
const agentCount = ref(0)
const sopCount = ref(0)
const skillCount = ref(0)
const mcpCount = ref(0)
const recentNotifs = ref<any[]>([])
const loading = ref(true)

const priorityColor: Record<string, string> = {
  urgent: '#f56c6c',
  high: '#e6a23c',
  medium: '#409eff',
  low: '#909399',
}

const sourceLabel: Record<string, string> = {
  web: 'Web 端',
  api: 'API 接口',
  email: '邮件',
  feishu: '飞书',
  lark: '飞书',
  dingtalk: '钉钉',
  wecom: '企微',
  im: '即时通讯',
  phone: '电话',
}

const sourceIcon: Record<string, string> = {
  web: 'Monitor',
  api: 'Connection',
  email: 'Message',
  lark: 'ChatDotSquare',
  feishu: 'ChatDotSquare',
  dingtalk: 'ChatDotSquare',
  wecom: 'ChatDotSquare',
  im: 'ChatLineSquare',
  phone: 'Phone',
}

const priorityLabel: Record<string, string> = {
  urgent: '紧急',
  high: '高',
  medium: '中',
  low: '低',
}

const channelTypeLabel: Record<string, string> = {
  lark: '飞书',
  dingtalk: '钉钉',
  wecom: '企微',
  email: '邮件',
}

const channelTypeColor: Record<string, string> = {
  lark: '#3370ff',
  dingtalk: '#0089ff',
  wecom: '#07c160',
  email: '#909399',
}

const activeChannelCount = computed(() => channels.value.filter(c => c.enabled).length)
const totalPriority = computed(() => Object.values(byPriority.value).reduce((a, b) => a + b, 0) || 1)
const totalSource = computed(() => Object.values(bySource.value).reduce((a, b) => a + b, 0) || 1)

const distributedSatisfaction = computed(() => {
  const dist = satisfaction.value.distribution
  const keys = Object.keys(dist).map(Number).sort((a, b) => b - a)
  const maxCount = Math.max(...Object.values(dist), 1)
  return keys.map(k => ({ score: k, count: dist[k], pct: (dist[k] / maxCount) * 100 }))
})

const sortedBySource = computed(() => {
  return Object.entries(bySource.value).sort((a, b) => b[1] - a[1])
})

const sortedByPriority = computed(() => {
  const order = ['urgent', 'high', 'medium', 'low']
  return Object.entries(byPriority.value).sort((a, b) => {
    return order.indexOf(a[0]) - order.indexOf(b[0])
  })
})

function goToTickets(filter: Record<string, string>) {
  router.push({ name: 'Tickets', query: filter })
}

function goToPage(name: string) {
  router.push({ name })
}

onMounted(async () => {
  try {
    const [ovRes, ticketRes, satRes, chRes, catRes, agRes, sopRes, skRes, mcpRes, notifRes] = await Promise.all([
      request.get('/stats/overview'),
      request.get('/stats/tickets'),
      request.get('/stats/performance'),
      request.get('/channels'),
      request.get('/categories'),
      request.get('/agents', { params: { page: 1, page_size: 1 } }),
      request.get('/sops', { params: { page: 1, page_size: 1 } }),
      request.get('/skills', { params: { page: 1, page_size: 1 } }),
      request.get('/mcps', { params: { page: 1, page_size: 1 } }),
      request.get('/notifications', { params: { page: 1, page_size: 5 } }),
    ])
    overview.value = ovRes.data || ovRes
    const td = ticketRes.data || ticketRes
    bySource.value = td.by_source || {}
    byPriority.value = td.by_priority || {}
    satisfaction.value = satRes.data || satRes
    channels.value = chRes.data?.data || chRes.data || []
    categories.value = catRes.data?.data || catRes.data || []
    agentCount.value = agRes.data?.total || agRes.total || 0
    sopCount.value = sopRes.data?.total || sopRes.total || 0
    skillCount.value = skRes.total || 0
    mcpCount.value = mcpRes.total || 0
    const nr = notifRes.data || notifRes
    recentNotifs.value = (nr.data || nr.list || []).slice(0, 5)
  } catch {
    // API not available
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="dashboard">
    <div class="page-header">
      <h2><el-icon><DataAnalysis /></el-icon> 仪表盘</h2>
    </div>

    <div v-loading="loading" class="dashboard-body">
      <!-- KPI Grid -->
      <div class="kpi-grid">
        <el-card shadow="never" class="kpi-card clickable" @click="goToTickets({})">
          <div class="kpi-body">
            <div class="kpi-icon" style="background:linear-gradient(135deg,#6366f1,#8b5cf6)">
              <el-icon :size="20"><Ticket /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ overview.total }}</div>
              <div class="kpi-label">工单总数</div>
            </div>
          </div>
        </el-card>
        <el-card shadow="never" class="kpi-card clickable" @click="goToTickets({ status: 'open' })">
          <div class="kpi-body">
            <div class="kpi-icon" style="background:linear-gradient(135deg,#f59e0b,#d97706)">
              <el-icon :size="20"><WarningFilled /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ overview.open }}</div>
              <div class="kpi-label">待处理</div>
            </div>
          </div>
        </el-card>
        <el-card shadow="never" class="kpi-card clickable" @click="goToTickets({ status: 'in_progress' })">
          <div class="kpi-body">
            <div class="kpi-icon" style="background:linear-gradient(135deg,#06b6d4,#0891b2)">
              <el-icon :size="20"><Loading /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ overview.in_progress }}</div>
              <div class="kpi-label">处理中</div>
            </div>
          </div>
        </el-card>
        <el-card shadow="never" class="kpi-card clickable" @click="goToTickets({ status: 'resolved' })">
          <div class="kpi-body">
            <div class="kpi-icon" style="background:linear-gradient(135deg,#22c55e,#16a34a)">
              <el-icon :size="20"><Select /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ overview.resolved }}</div>
              <div class="kpi-label">已解决</div>
            </div>
          </div>
        </el-card>
        <el-card shadow="never" class="kpi-card clickable" @click="goToTickets({ status: 'overdue' })">
          <div class="kpi-body">
            <div class="kpi-icon" style="background:linear-gradient(135deg,#ef4444,#dc2626)">
              <el-icon :size="20"><Clock /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value" style="color:#ef4444">{{ overview.overdue }}</div>
              <div class="kpi-label">已逾期</div>
            </div>
          </div>
        </el-card>
        <el-card shadow="never" class="kpi-card clickable" @click="goToPage('Channels')">
          <div class="kpi-body">
            <div class="kpi-icon" style="background:linear-gradient(135deg,#0ea5e9,#6366f1)">
              <el-icon :size="20"><Connection /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ activeChannelCount }}<span class="kpi-unit">/{{ channels.length }}</span></div>
              <div class="kpi-label">渠道接入</div>
            </div>
          </div>
        </el-card>
        <el-card shadow="never" class="kpi-card clickable" @click="goToPage('Agents')">
          <div class="kpi-body">
            <div class="kpi-icon" style="background:linear-gradient(135deg,#8b5cf6,#a855f7)">
              <el-icon :size="20"><Cpu /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ agentCount }}</div>
              <div class="kpi-label">Agent 数量</div>
            </div>
          </div>
        </el-card>
        <el-card shadow="never" class="kpi-card clickable" @click="goToPage('SOPs')">
          <div class="kpi-body">
            <div class="kpi-icon" style="background:linear-gradient(135deg,#f97316,#ea580c)">
              <el-icon :size="20"><Document /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ sopCount }}</div>
              <div class="kpi-label">SOP 数量</div>
            </div>
          </div>
        </el-card>
        <el-card shadow="never" class="kpi-card clickable" @click="goToPage('Skills')">
          <div class="kpi-body">
            <div class="kpi-icon" style="background:linear-gradient(135deg,#14b8a6,#0d9488)">
              <el-icon :size="20"><MagicStick /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ skillCount }}</div>
              <div class="kpi-label">Skill 数量</div>
            </div>
          </div>
        </el-card>
        <el-card shadow="never" class="kpi-card clickable" @click="goToPage('MCPs')">
          <div class="kpi-body">
            <div class="kpi-icon" style="background:linear-gradient(135deg,#6366f1,#4f46e5)">
              <el-icon :size="20"><Connection /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ mcpCount }}</div>
              <div class="kpi-label">MCP 数量</div>
            </div>
          </div>
        </el-card>
      </div>

      <!-- Distribution Row -->
      <el-row :gutter="16" class="chart-row">
        <el-col :xs="24" :md="12">
          <el-card shadow="never" class="section-card">
            <template #header>
              <span class="section-title">工单优先级分布</span>
              <el-button text size="small" @click="goToTickets({})">查看全部 <el-icon><ArrowRight /></el-icon></el-button>
            </template>
            <div class="dist-list">
              <div v-for="[key, count] in sortedByPriority" :key="key" class="dist-item clickable" @click="goToTickets({ priority: key })">
                <div class="dist-label">
                  <span class="dist-dot" :style="{ background: priorityColor[key] || '#999' }"></span>
                  <span>{{ priorityLabel[key] || key }}</span>
                </div>
                <div class="dist-bar-track">
                  <div class="dist-bar-fill" :style="{ width: (count / totalPriority * 100) + '%', background: priorityColor[key] || '#999' }"></div>
                </div>
                <div class="dist-count">{{ count }}</div>
              </div>
              <div v-if="Object.keys(byPriority).length === 0" class="empty-tip">暂无数据</div>
            </div>
          </el-card>
        </el-col>
        <el-col :xs="24" :md="12">
          <el-card shadow="never" class="section-card">
            <template #header>
              <span class="section-title">工单来源分布</span>
              <el-button text size="small" @click="goToTickets({})">查看全部 <el-icon><ArrowRight /></el-icon></el-button>
            </template>
            <div class="dist-list">
              <div v-for="[key, count] in sortedBySource" :key="key" class="dist-item clickable" @click="goToTickets({ source: key })">
                <div class="dist-label">
                  <span class="channel-source-icon" :style="{ background: channelTypeColor[key] || '#6366f1' }">
                    <el-icon :size="14"><component :is="sourceIcon[key] || 'ChatDotSquare'" /></el-icon>
                  </span>
                  <span>{{ sourceLabel[key] || key }}</span>
                </div>
                <div class="dist-bar-track">
                  <div class="dist-bar-fill" :style="{ width: (count / totalSource * 100) + '%', background: channelTypeColor[key] || '#6366f1' }"></div>
                </div>
                <div class="dist-count">{{ count }}</div>
              </div>
              <div v-if="Object.keys(bySource).length === 0" class="empty-tip">暂无数据</div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- Channel Status + Categories + Satisfaction + Notifications -->
      <el-row :gutter="16" class="chart-row">
        <el-col :xs="24" :md="12" :lg="6">
          <el-card shadow="never" class="section-card">
            <template #header>
              <span class="section-title">接入渠道概览</span>
              <el-button text size="small" @click="goToTickets({})">查看全部 <el-icon><ArrowRight /></el-icon></el-button>
            </template>
            <div v-if="channels.length > 0" class="channel-list">
              <div v-for="ch in channels" :key="ch.id" class="channel-item clickable" @click="goToTickets({ source: ch.type })">
                <div class="channel-left">
                  <span class="channel-icon" :style="{ background: channelTypeColor[ch.type] || '#999' }">
                    <el-icon :size="16"><Connection /></el-icon>
                  </span>
                  <div class="channel-info">
                    <div class="channel-name">{{ ch.name }}</div>
                    <div class="channel-type">{{ channelTypeLabel[ch.type] || ch.type }}</div>
                  </div>
                </div>
                <el-tag :type="ch.enabled ? 'success' : 'info'" size="small" effect="plain">{{ ch.enabled ? '已启用' : '已停用' }}</el-tag>
              </div>
            </div>
            <div v-else class="empty-tip">暂未接入渠道</div>
          </el-card>
        </el-col>
        <el-col :xs="24" :md="12" :lg="6">
          <el-card shadow="never" class="section-card">
            <template #header>
              <span class="section-title">分类概览</span>
            </template>
            <div v-if="categories.length > 0" class="category-list">
              <div v-for="cat in categories" :key="cat.id" class="category-item">
                <div class="category-left">
                  <span class="category-dot" :class="cat.status === 1 ? 'active' : 'inactive'"></span>
                  <span class="category-name">{{ cat.name }}</span>
                </div>
                <el-tag v-if="cat.description" size="small" effect="plain" style="border:none;background:transparent;color:#86909c;font-size:12px">{{ cat.description }}</el-tag>
              </div>
            </div>
            <div v-else class="empty-tip">暂未配置分类</div>
          </el-card>
        </el-col>
        <el-col :xs="24" :md="12" :lg="6">
          <el-card shadow="never" class="section-card">
            <template #header>
              <span class="section-title">通知概览</span>
              <el-button text size="small" @click="router.push({ name: 'Notifications' })">查看全部 <el-icon><ArrowRight /></el-icon></el-button>
            </template>
            <div v-if="recentNotifs.length > 0" class="notif-list">
              <div v-for="n in recentNotifs" :key="n.id" class="notif-item" :class="{ unread: !n.is_read }">
                <div class="notif-left">
                  <span class="notif-dot" :class="n.is_read ? 'read' : 'unread'"></span>
                  <div class="notif-info">
                    <div class="notif-title">{{ n.title }}</div>
                    <div class="notif-meta">
                      <el-tag :type="notifTypeTag(n.type)" size="small" effect="plain">{{ notifTypeLabel(n.type) }}</el-tag>
                      <span class="notif-time">{{ formatTime(n.created_at) }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div v-else class="empty-tip">暂无通知</div>
          </el-card>
        </el-col>
        <el-col :xs="24" :md="12" :lg="6">
          <el-card shadow="never" class="section-card">
            <template #header>
              <span class="section-title">满意度评分</span>
            </template>
            <div class="sat-header">
              <div class="sat-score">{{ satisfaction.average.toFixed(1) }}</div>
              <div class="sat-stars">
                <el-icon v-for="i in 5" :key="i" :color="i <= Math.round(satisfaction.average) ? '#f59e0b' : '#d9d9d9'" :size="16">
                  <StarFilled />
                </el-icon>
              </div>
              <div class="sat-total">共 {{ satisfaction.total }} 条评价</div>
            </div>
            <div class="sat-dist">
              <div v-for="item in distributedSatisfaction" :key="item.score" class="sat-dist-item">
                <div class="sat-dist-label">{{ item.score }} 分</div>
                <div class="sat-dist-track">
                  <div class="sat-dist-fill" :style="{ width: item.pct + '%' }"></div>
                </div>
                <div class="sat-dist-count">{{ item.count }}</div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<style scoped>
.dashboard {
  height: 100%;
  display: flex;
  flex-direction: column;
}
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  flex-shrink: 0;
}
.page-header h2 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 600;
  color: #1d2129;
}
.dashboard-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 20px;
  overflow: hidden;
  padding-bottom: 8px;
}

/* === Row Layout === */
.chart-row {
  flex-shrink: 0;
  margin-left: 0 !important;
  margin-right: 0 !important;
  margin-bottom: -12px;
}
.chart-row .el-col {
  display: flex;
  margin-bottom: 8px;
}

/* KPI Grid */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 16px;
  flex-shrink: 0;
}

/* === KPI Cards === */
.kpi-card {
  border-radius: 10px;
  border: 1px solid #f0f0f0;
  margin-bottom: 0;
  width: 100%;
  min-width: 0;
  overflow: hidden;
  transition: border-color 0.15s, box-shadow 0.15s;
}
.kpi-card.clickable {
  cursor: pointer;
}
.kpi-card.clickable:hover {
  border-color: #c0c4cc;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
}
.kpi-body {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 4px 0;
}
.kpi-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}
.kpi-info {
  flex: 1;
  min-width: 0;
}
.kpi-value {
  font-size: 22px;
  font-weight: 700;
  color: #1d2129;
  line-height: 1.2;
  white-space: nowrap;
}
.kpi-unit {
  font-size: 13px;
  font-weight: 400;
  color: #86909c;
}
.kpi-label {
  margin-top: 2px;
  font-size: 12px;
  color: #86909c;
  white-space: nowrap;
  display: flex;
  align-items: center;
  gap: 4px;
}
.kpi-sub {
  color: #ef4444;
  font-weight: 600;
}

/* Notification List */
.notif-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
}
.notif-item {
  display: flex;
  align-items: flex-start;
  padding: 8px 0;
  border-bottom: 1px solid #f5f5f5;
}
.notif-item:last-child {
  border-bottom: none;
}
.notif-left {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  width: 100%;
  min-width: 0;
}
.notif-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  margin-top: 6px;
  flex-shrink: 0;
}
.notif-dot.unread {
  background: #409eff;
}
.notif-dot.read {
  background: #d9d9d9;
}
.notif-info {
  flex: 1;
  min-width: 0;
}
.notif-title {
  font-size: 13px;
  color: #1d2129;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.notif-item.unread .notif-title {
  font-weight: 600;
}
.notif-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 4px;
}
.notif-time {
  font-size: 11px;
  color: #a8abb2;
  flex-shrink: 0;
}

/* === Section Cards === */
.section-card {
  border-radius: 10px;
  border: 1px solid #f0f0f0;
  width: 100%;
  display: flex;
  flex-direction: column;
}
.section-card :deep(.el-card__header) {
  padding: 14px 20px;
  border-bottom: 1px solid #f5f5f5;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.section-card :deep(.el-card__body) {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.section-title {
  font-size: 14px;
  font-weight: 600;
  color: #1d2129;
}

/* === Distribution Bars === */
.dist-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  flex: 1;
  justify-content: center;
}
.dist-item {
  display: flex;
  align-items: center;
  gap: 10px;
}
.dist-item.clickable {
  cursor: pointer;
  border-radius: 6px;
  padding: 2px 4px;
  margin: 0 -4px;
  transition: background 0.15s;
}
.dist-item.clickable:hover {
  background: #f5f6f7;
}
.dist-label {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 90px;
  font-size: 13px;
  color: #4e5969;
  flex-shrink: 0;
}
.dist-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.channel-source-icon {
  width: 22px;
  height: 22px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}
.dist-bar-track {
  flex: 1;
  height: 8px;
  background: #f2f3f5;
  border-radius: 4px;
  overflow: hidden;
}
.dist-bar-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.6s ease;
}
.dist-count {
  width: 32px;
  text-align: right;
  font-size: 13px;
  font-weight: 600;
  color: #1d2129;
  flex-shrink: 0;
}

/* === Channel List === */
.channel-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
}
.channel-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #f5f5f5;
}
.channel-item:last-child {
  border-bottom: none;
}
.channel-item.clickable {
  cursor: pointer;
  border-radius: 6px;
  padding: 8px 4px;
  margin: 0 -4px;
  transition: background 0.15s;
}
.channel-item.clickable:hover {
  background: #f5f6f7;
}
.channel-left {
  display: flex;
  align-items: center;
  gap: 10px;
}
.channel-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}
.channel-info {
  display: flex;
  flex-direction: column;
}
.channel-name {
  font-size: 13px;
  font-weight: 600;
  color: #1d2129;
}
.channel-type {
  font-size: 11px;
  color: #86909c;
  margin-top: 1px;
}

/* === Category List === */
.category-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
}
.category-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #f5f5f5;
}
.category-item:last-child {
  border-bottom: none;
}
.category-left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.category-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.category-dot.active {
  background: #22c55e;
}
.category-dot.inactive {
  background: #d9d9d9;
}
.category-name {
  font-size: 13px;
  font-weight: 500;
  color: #1d2129;
}

/* === Satisfaction === */
.sat-header {
  text-align: center;
  padding: 4px 0 12px;
  border-bottom: 1px solid #f5f5f5;
  margin-bottom: 12px;
}
.sat-score {
  font-size: 30px;
  font-weight: 700;
  color: #1d2129;
  line-height: 1.2;
}
.sat-stars {
  display: flex;
  justify-content: center;
  gap: 3px;
  margin: 4px 0;
}
.sat-total {
  font-size: 12px;
  color: #86909c;
}
.sat-dist {
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex: 1;
  justify-content: center;
}
.sat-dist-item {
  display: flex;
  align-items: center;
  gap: 8px;
}
.sat-dist-label {
  width: 36px;
  font-size: 12px;
  color: #4e5969;
  flex-shrink: 0;
}
.sat-dist-track {
  flex: 1;
  height: 6px;
  background: #f2f3f5;
  border-radius: 3px;
  overflow: hidden;
}
.sat-dist-fill {
  height: 100%;
  border-radius: 3px;
  background: linear-gradient(90deg, #f59e0b, #f97316);
  transition: width 0.6s ease;
}
.sat-dist-count {
  width: 24px;
  text-align: right;
  font-size: 12px;
  color: #86909c;
  flex-shrink: 0;
}
.empty-tip {
  text-align: center;
  color: #c0c4cc;
  font-size: 13px;
  padding: 24px 0;
}
</style>
