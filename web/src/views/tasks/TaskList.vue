<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

interface Task {
  id: number
  name: string
  type: string
  status: string
  suggestion: string
  decision: string
  feedback: string
  rejection_reason: string
  created_at: string
  assignee: any
  agent: any
  workflow: {
    id: number
    ticket: {
      id: number
      ticket_no: string
      title: string
      description: string
    }
  }
}

const list = ref<Task[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const detailVisible = ref(false)
const selectedTask = ref<Task | null>(null)
const feedbackText = ref('')
const resultText = ref('')
const rejectReason = ref('')
const actionType = ref<'approve' | 'reject' | 'revise'>('approve')
const actionTitle = ref('审核')

function openAction(task: Task, action: 'approve' | 'reject' | 'revise') {
  selectedTask.value = task
  actionType.value = action
  actionTitle.value = action === 'approve' ? '审核通过' : action === 'reject' ? '拒绝' : '打回修改'
  feedbackText.value = ''
  resultText.value = task.suggestion || ''
  rejectReason.value = ''
  detailVisible.value = true
}

async function fetchData() {
  loading.value = true
  try {
    const res: any = await request.get('/workflows/tasks', { params: { page: page.value, page_size: pageSize.value } })
    list.value = res.data?.data || []
    total.value = res.data?.total || 0
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

async function submitAction() {
  if (!selectedTask.value) return
  const task = selectedTask.value
  try {
    if (actionType.value === 'approve') {
      await request.post(`/workflows/steps/${task.id}/approve`, { feedback: feedbackText.value, result: resultText.value })
      ElMessage.success('已审核通过')
    } else if (actionType.value === 'reject') {
      if (!rejectReason.value.trim()) { ElMessage.warning('请输入拒绝原因'); return }
      await request.post(`/workflows/steps/${task.id}/reject`, { reason: rejectReason.value })
      ElMessage.success('已拒绝')
    } else {
      if (!feedbackText.value.trim()) { ElMessage.warning('请输入修改意见'); return }
      await request.post(`/workflows/steps/${task.id}/revise`, { feedback: feedbackText.value })
      ElMessage.success('已打回修改')
    }
    detailVisible.value = false
    fetchData()
  } catch { /* handled by interceptor */ }
}

onMounted(fetchData)
</script>

<template>
  <div>
    <div class="page-header">
      <h2>我的待办</h2>
      <el-button type="primary" @click="fetchData" :loading="loading">刷新</el-button>
    </div>
    <el-card>
      <el-table :data="list" v-loading="loading" stripe border style="width: 100%">
        <el-table-column label="工单号" width="160">
          <template #default="{ row }">{{ row.workflow?.ticket?.ticket_no || '-' }}</template>
        </el-table-column>
        <el-table-column label="工单标题" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">{{ row.workflow?.ticket?.title || '-' }}</template>
        </el-table-column>
        <el-table-column prop="name" label="步骤" width="180" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }"><el-tag :type="row.status === 'running' ? 'warning' : 'info'">{{ row.status }}</el-tag></template>
        </el-table-column>
        <el-table-column label="AI 建议" min-width="300" show-overflow-tooltip>
          <template #default="{ row }"><div class="suggestion-preview">{{ row.suggestion || '无建议' }}</div></template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="success" @click="openAction(row, 'approve')">通过</el-button>
            <el-button size="small" type="danger" @click="openAction(row, 'reject')">拒绝</el-button>
            <el-button size="small" @click="openAction(row, 'revise')">打回</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="fetchData" />
      </div>
    </el-card>

    <el-dialog v-model="detailVisible" :title="actionTitle" width="600px" v-if="selectedTask">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="工单">{{ selectedTask.workflow?.ticket?.ticket_no }} - {{ selectedTask.workflow?.ticket?.title }}</el-descriptions-item>
        <el-descriptions-item label="步骤">{{ selectedTask.name }}</el-descriptions-item>
        <el-descriptions-item label="AI 建议" v-if="selectedTask.suggestion"><pre class="suggestion-text">{{ selectedTask.suggestion }}</pre></el-descriptions-item>
      </el-descriptions>
      <el-divider />
      <template v-if="actionType === 'approve'">
        <el-input v-model="resultText" type="textarea" :rows="4" placeholder="处理结果（可选，默认为 AI 建议）" />
        <el-input v-model="feedbackText" type="textarea" :rows="2" placeholder="反馈意见（可选）" style="margin-top: 12px" />
      </template>
      <template v-if="actionType === 'reject'">
        <el-input v-model="rejectReason" type="textarea" :rows="3" placeholder="请输入拒绝原因" />
      </template>
      <template v-if="actionType === 'revise'">
        <el-input v-model="feedbackText" type="textarea" :rows="3" placeholder="请输入修改意见" />
      </template>
      <template #footer>
        <el-button @click="detailVisible = false">取消</el-button>
        <el-button type="primary" @click="submitAction">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }
.page-header h2 { margin: 0; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.suggestion-preview { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; color: #666; font-size: 13px; }
.suggestion-text { white-space: pre-wrap; background: #f5f7fa; padding: 12px; border-radius: 4px; margin: 0; font-size: 13px; line-height: 1.6; }
</style>
