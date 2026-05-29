<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

interface Ticket {
  id: number
  ticket_no: string
  title: string
  status: string
  priority: string
  source: string
  assignee_name: string
  created_at: string
}

const list = ref<Ticket[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const detailVisible = ref(false)
const detail = ref<any>(null)
const commentContent = ref('')

async function fetchData() {
  loading.value = true
  try {
    const res: any = await request.get('/tickets', { params: { page: page.value, page_size: pageSize.value } })
    list.value = res.data?.list || res.data || []
    total.value = res.data?.total || res.total || 0
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

async function viewDetail(row: Ticket) {
  try {
    const res: any = await request.get(`/tickets/${row.id}`)
    detail.value = res.data || res
    commentContent.value = ''
    detailVisible.value = true
  } catch {
    ElMessage.error('获取工单详情失败')
  }
}

async function addComment() {
  if (!commentContent.value.trim()) return
  try {
    await request.post(`/tickets/${detail.value.id}/comments`, { content: commentContent.value })
    commentContent.value = ''
    ElMessage.success('评论已添加')
    viewDetail(detail.value)
  } catch {
    // handled by interceptor
  }
}

const statusMap: Record<string, string> = { open: 'info', pending: 'warning', processing: 'primary', resolved: 'success', closed: 'info', rejected: 'danger' }
function statusTag(status: string) {
  return (statusMap[status] || 'info') as 'info' | 'primary' | 'success' | 'warning' | 'danger'
}

const priorityMap: Record<string, string> = { low: 'info', medium: 'warning', high: 'danger', urgent: 'danger' }
function priorityTag(p: string) {
  return (priorityMap[p] || 'info') as 'info' | 'primary' | 'success' | 'warning' | 'danger'
}

onMounted(fetchData)
</script>

<template>
  <div>
    <div class="page-header">
      <h2>工单管理</h2>
    </div>

    <el-card>
      <el-table :data="list" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="ticket_no" label="工单号" width="140" />
        <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="80">
          <template #default="{ row }">
            <el-tag :type="priorityTag(row.priority)" size="small">{{ row.priority }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="source" label="来源" width="100" />
        <el-table-column prop="assignee_name" label="处理人" width="120" />
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="viewDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="fetchData"
        />
      </div>
    </el-card>

    <el-dialog v-model="detailVisible" title="工单详情" width="700px" v-if="detail">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="工单号">{{ detail.ticket_no }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTag(detail.status)">{{ detail.status }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="标题" :span="2">{{ detail.title }}</el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">{{ detail.description }}</el-descriptions-item>
        <el-descriptions-item label="优先级">
          <el-tag :type="priorityTag(detail.priority)" size="small">{{ detail.priority }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="来源">{{ detail.source }}</el-descriptions-item>
        <el-descriptions-item label="处理人">{{ detail.assignee_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ detail.created_at }}</el-descriptions-item>
      </el-descriptions>

      <el-divider />
      <h4>评论</h4>
      <div v-for="c in (detail.comments || [])" :key="c.id" class="comment-item">
        <div class="comment-meta">
          <strong>{{ c.author_name || '系统' }}</strong>
          <span class="comment-time">{{ c.created_at }}</span>
          <el-tag v-if="c.is_internal" size="small" type="warning">内部</el-tag>
        </div>
        <div class="comment-body">{{ c.content }}</div>
      </div>

      <div class="comment-input">
        <el-input v-model="commentContent" type="textarea" :rows="3" placeholder="输入评论..." />
        <el-button type="primary" style="margin-top: 8px" @click="addComment">添加评论</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.page-header h2 {
  margin: 0;
}
.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
.comment-item {
  border-bottom: 1px solid #f0f0f0;
  padding: 12px 0;
}
.comment-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}
.comment-time {
  color: #999;
  font-size: 12px;
}
.comment-body {
  white-space: pre-wrap;
  color: #333;
}
.comment-input {
  margin-top: 16px;
}
</style>
