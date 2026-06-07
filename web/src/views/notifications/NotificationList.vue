<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import request from '@/api/request'
import { useUserStore } from '@/stores/user'

interface Notification {
  id: number
  type: string
  title: string
  content: string
  reference_id?: number
  reference_type?: string
  is_read: boolean
  status: string
  created_at: string
}

const router = useRouter()
const userStore = useUserStore()
const isAdmin = computed(() => userStore.userInfo?.role === 'admin')

const notifications = ref<Notification[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const typeFilter = ref('')
const showAll = ref(true)
const unreadCount = ref(0)

const typeOptions = [
  { value: '', label: '全部' },
  { value: 'ticket_assigned', label: '工单分配' },
  { value: 'sla', label: 'SLA 提醒' },
  { value: 'ticket_comment', label: '工单评论' },
  { value: 'system', label: '系统通知' },
]

const typeTagMap: Record<string, string> = {
  ticket_assigned: 'primary',
  sla: 'warning',
  ticket_comment: 'success',
  system: 'info',
}

async function fetchData() {
  loading.value = true
  try {
    const params: Record<string, any> = { page: page.value, page_size: pageSize.value }
    if (typeFilter.value) params.type = typeFilter.value
    if (isAdmin.value && !showAll.value) params.scope = 'mine'
    const res: any = await request.get('/notifications', { params })
    notifications.value = res.data?.data || res.data || []
    total.value = res.data?.total || 0
  } catch {
    notifications.value = []
  } finally {
    loading.value = false
  }
}

async function fetchUnreadCount() {
  try {
    const res: any = await request.get('/notifications/unread-count')
    unreadCount.value = res.count ?? res.data?.count ?? 0
  } catch {
    unreadCount.value = 0
  }
}

async function markAsRead(row: Notification) {
  try {
    await request.put(`/notifications/${row.id}/read`, {})
    row.is_read = true
    row.status = 'read'
    if (unreadCount.value > 0) unreadCount.value--
    ElMessage.success('已标记为已读')
  } catch {
    // handled by interceptor
  }
}

async function markAllRead() {
  try {
    await ElMessageBox.confirm('确认将所有通知标记为已读？', '提示')
    await request.put('/notifications/all/read', {})
    fetchData()
    unreadCount.value = 0
    ElMessage.success('全部标记为已读')
  } catch {
    // cancelled or error
  }
}

function handleRowClick(row: Notification) {
  if (row.reference_id && row.reference_type === 'ticket') {
    router.push(`/tickets/${row.reference_id}`)
  } else if (row.reference_id) {
    ElMessage.info(`${row.reference_type} #${row.reference_id}`)
  }
}

function onTypeChange() {
  page.value = 1
  fetchData()
}

onMounted(() => {
  fetchData()
  fetchUnreadCount()
})
</script>

<template>
  <div class="page">
    <div class="list-view">
      <div class="page-header">
        <h2>
          通知列表
          <el-badge :value="unreadCount" :hidden="unreadCount === 0" class="unread-badge">
            <span />
          </el-badge>
        </h2>
      </div>

      <el-card shadow="never" class="list-card">
        <div class="table-toolbar">
          <div class="toolbar-left">
            <el-select v-model="typeFilter" placeholder="通知类型" clearable style="width: 160px" @change="onTypeChange">
              <el-option v-for="opt in typeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
            </el-select>
            <el-button v-if="isAdmin" :type="showAll ? 'primary' : 'default'" size="small" @click="showAll = !showAll; onTypeChange()" style="margin-left: 8px;">
              {{ showAll ? '所有通知' : '我的通知' }}
            </el-button>
          </div>
          <div class="toolbar-right">
            <el-button :disabled="unreadCount === 0" type="primary" @click="markAllRead">全部标记已读</el-button>
          </div>
        </div>

        <el-table :data="notifications" v-loading="loading" class="beauty-table" style="width:100%" height="calc(100vh - 280px)" @row-click="handleRowClick" highlight-current-row>
          <el-table-column prop="id" label="ID" width="60" align="center" />
          <el-table-column label="类型" width="140" align="center">
            <template #default="{ row }">
              <el-tag :type="(typeTagMap[row.type] || 'info') as any" size="small" effect="plain">
                {{ typeOptions.find(o => o.value === row.type)?.label || row.type }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="title" label="标题" min-width="140" show-overflow-tooltip />
          <el-table-column label="内容" min-width="200">
            <template #default="{ row }">
              <span class="content-cell">{{ row.content }}</span>
            </template>
          </el-table-column>
          <el-table-column label="关联" width="140" align="center">
            <template #default="{ row }">
              <span v-if="row.reference_id">{{ row.reference_type }} #{{ row.reference_id }}</span>
              <span v-else class="no-ref">-</span>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="90" align="center">
            <template #default="{ row }">
              <el-tag :type="row.is_read ? 'info' : 'primary'" size="small" effect="plain">
                {{ row.is_read ? '已读' : '未读' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="180" align="center" />
          <el-table-column label="操作" width="110" fixed="right" align="center">
            <template #default="{ row }">
              <el-button size="small" type="primary" :disabled="row.is_read" @click.stop="markAsRead(row)">
                {{ row.is_read ? '已读' : '标为已读' }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <div v-if="total > 0" class="pagination-wrap">
          <el-pagination
            v-model:current-page="page"
            v-model:page-size="pageSize"
            :total="total"
            layout="total, prev, pager, next"
            @current-change="fetchData"
          />
        </div>
      </el-card>
    </div>
  </div>
</template>

<style scoped>
.page {
  height: 100%;
  display: flex;
  flex-direction: column;
}
.list-view {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
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
.list-card {
  flex: 1;
  display: flex;
  flex-direction: column;
  border-radius: 12px;
  border: 1px solid #e5e6eb;
  min-height: 0;
}
.list-card :deep(.el-card__body) {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: 16px 20px;
}
.table-toolbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; gap: 12px; }
.toolbar-left { display: flex; align-items: center; gap: 8px; }
.toolbar-right { display: flex; align-items: center; gap: 8px; }
.beauty-table { --el-table-border-color: #f0f0f0; flex: 1; min-height: 350px; }
.beauty-table :deep(.el-table__header th) { background: #f7f8fa; color: #4e5969; font-weight: 500; }
.content-cell {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
}
.no-ref { color: #999; }
.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
  flex-shrink: 0;
}
.unread-badge :deep(.el-badge__content) {
  position: static;
  transform: none;
  margin-left: 4px;
}
</style>
