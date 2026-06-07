<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox, ElPopconfirm } from 'element-plus'
import { Search, Plus, EditPen, Delete, Refresh, View } from '@element-plus/icons-vue'
import request from '@/api/request'

const route = useRoute()

interface Ticket {
  id: number; ticket_no: string; title: string; status: string; priority: string
  source: string; assignee_name: string; created_at: string; category_id?: number
  description?: string; sla_status?: string; due_at?: string
}
interface Tag { id: number; name: string; color: string }
interface Category { id: number; name: string }

const list = ref<Ticket[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const keyword = ref('')
let searchTimer: ReturnType<typeof setTimeout>
function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => { page.value = 1; fetchData() }, 300)
}

const detailVisible = ref(false)
const detail = ref<any>(null)
const commentContent = ref('')
const commentInternal = ref(false)
const allTags = ref<Tag[]>([])
const tagSelectVisible = ref(false)
const selectedTagIDs = ref<number[]>([])

const categories = ref<Category[]>([])

const createVisible = ref(false)
const createForm = ref({ title: '', description: '', priority: 'medium', category_id: null as number | null })
const editVisible = ref(false)
const editId = ref(0)
const editForm = ref({ title: '', description: '', priority: 'medium', category_id: null as number | null })
async function fetchData() {
  loading.value = true
  try {
    const params: any = { page: page.value, page_size: pageSize.value }
    if (keyword.value) params.keyword = keyword.value
    if (route.query.priority) params.priority = route.query.priority
    if (route.query.source) params.source = route.query.source
    if (route.query.status) params.status = route.query.status
    const res: any = await request.get('/tickets', { params })
    list.value = res.data?.data || res.data || []
    total.value = res.data?.total || 0
  } catch { list.value = [] }
  finally { loading.value = false }
}

async function fetchMeta() {
  try {
    const res: any = await request.get('/categories')
    categories.value = res.data?.data || res.data || []
  } catch { /* ignore */ }
}

async function fetchAllTags() {
  try { const res: any = await request.get('/tags'); allTags.value = res.data || [] } catch { /* ignore */ }
}

async function viewDetail(row: Ticket) {
  try {
    const res: any = await request.get(`/tickets/${row.id}`)
    detail.value = res.data || res
    commentContent.value = ''
    commentInternal.value = false
    detailVisible.value = true
    await fetchTags()
  } catch { ElMessage.error('获取工单详情失败') }
}

async function fetchTags() {
  if (!detail.value) return
  try { const res: any = await request.get(`/tickets/${detail.value.id}/tags`); detail.value.tags = res.data || [] } catch { /* ignore */ }
}

async function addComment() {
  if (!commentContent.value.trim()) return
  try {
    await request.post(`/tickets/${detail.value.id}/comments`, { content: commentContent.value, is_internal: commentInternal.value })
    commentContent.value = ''; commentInternal.value = false
    ElMessage.success('评论已添加')
    viewDetail(detail.value)
  } catch { /* handled by interceptor */ }
}

async function updateStatus(ticketId: number, status: string) {
  try {
    await request.post(`/tickets/${ticketId}/status`, { status })
    ElMessage.success('状态已更新')
    fetchData()
    if (detail.value?.id === ticketId) detail.value.status = status
  } catch { /* handled by interceptor */ }
}

async function openTagSelector() {
  await fetchAllTags()
  if (!detail.value?.tags) detail.value.tags = []
  selectedTagIDs.value = detail.value.tags.map((t: Tag) => t.id)
  tagSelectVisible.value = true
}

async function saveTags() {
  if (!detail.value) return
  try {
    await request.put(`/tickets/${detail.value.id}/tags`, { tag_ids: selectedTagIDs.value })
    ElMessage.success('标签已更新')
    tagSelectVisible.value = false
    await viewDetail(detail.value)
  } catch { /* handled by interceptor */ }
}

async function removeTag(tagId: number) {
  if (!detail.value) return
  try {
    await request.delete(`/tickets/${detail.value.id}/tags/${tagId}`)
    ElMessage.success('标签已移除')
    await fetchTags()
  } catch { /* handled by interceptor */ }
}

function openCreate() {
  createForm.value = { title: '', description: '', priority: 'medium', category_id: null }
  createVisible.value = true
}

async function handleCreate() {
  if (!createForm.value.title) { ElMessage.warning('请输入标题'); return }
  try {
    const payload: any = {
      title: createForm.value.title,
      description: createForm.value.description,
      priority: createForm.value.priority,
    }
    if (createForm.value.category_id !== null) payload.category_id = createForm.value.category_id
    await request.post('/tickets', payload)
    ElMessage.success('创建成功')
    createVisible.value = false
    fetchData()
  } catch { /* handled by interceptor */ }
}

function openEdit(row: Ticket) {
  editId.value = row.id
  editForm.value = {
    title: row.title || '',
    description: row.description || '',
    priority: row.priority || 'medium',
    category_id: row.category_id ?? null,
  }
  editVisible.value = true
}

async function handleEdit() {
  if (!editForm.value.title) { ElMessage.warning('请输入标题'); return }
  try {
    const payload: any = {
      title: editForm.value.title,
      description: editForm.value.description,
      priority: editForm.value.priority,
    }
    if (editForm.value.category_id !== null) payload.category_id = editForm.value.category_id
    await request.put(`/tickets/${editId.value}`, payload)
    ElMessage.success('更新成功')
    editVisible.value = false
    fetchData()
  } catch { /* handled by interceptor */ }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确认删除该工单吗？', '提示')
    await request.delete(`/tickets/${id}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch { /* cancelled or error */ }
}

const statusOptions = [
  { label: '待处理', value: 'open' },
  { label: '处理中', value: 'in_progress' },
  { label: '待确认', value: 'pending' },
  { label: '已解决', value: 'resolved' },
  { label: '已关闭', value: 'closed' },
]
const statusTypeMap: Record<string, string> = { open: 'info', in_progress: 'primary', pending: 'warning', resolved: 'success', closed: '' }
function statusTag(s: string) { return (statusTypeMap[s] || 'info') as 'info' | 'primary' | 'success' | 'warning' | 'danger' | '' }

const priorityMap: Record<string, string> = { low: 'info', medium: 'warning', high: 'danger', urgent: 'danger' }
function priorityTag(p: string) { return (priorityMap[p] || 'info') as 'info' | 'primary' | 'success' | 'warning' | 'danger' }

const slaTypeMap: Record<string, 'success' | 'warning' | 'danger'> = { normal: 'success', warning: 'warning', breached: 'danger' }
const slaLabelMap: Record<string, string> = { normal: '正常', warning: '预警', breached: '超期' }
function slaType(s: string | undefined) { return (s && slaTypeMap[s]) || 'info' }
function slaLabel(s: string | undefined) { return (s && slaLabelMap[s]) || '-' }

function tagColor(c: string) { return c || '#409eff' }

onMounted(() => { fetchData(); fetchMeta() })
</script>

<template>
  <div class="page">
    <div class="list-view">
      <div class="page-header">
        <h2>全部工单</h2>
      </div>

      <el-card shadow="never" class="list-card">
        <div class="table-toolbar">
          <div class="toolbar-left">
            <el-input v-model="keyword" placeholder="搜索工单标题..." clearable @input="onSearch" class="search-input">
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
          </div>
          <div class="toolbar-right">
            <el-button :icon="Refresh" circle @click="fetchData" />
            <el-button type="primary" :icon="Plus" @click="openCreate">新建工单</el-button>
          </div>
        </div>
        <el-table :data="list" v-loading="loading" stripe style="width: 100%" class="beauty-table" height="calc(100vh - 100px)">
          <el-table-column prop="id" label="ID" width="64" align="center" />
          <el-table-column prop="ticket_no" label="工单号" width="140" />
          <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip>
            <template #default="{ row }"><span class="name-cell">{{ row.title }}</span></template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="140">
            <template #default="{ row }">
              <el-select :model-value="row.status" size="small" style="width: 110px" @change="(v: string) => updateStatus(row.id, v)">
                <el-option v-for="s in statusOptions" :key="s.value" :label="s.label" :value="s.value" />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column prop="priority" label="优先级" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="priorityTag(row.priority)" size="small" effect="plain">{{ row.priority }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="source" label="来源" width="100" align="center" />
          <el-table-column prop="sla_status" label="SLA" width="90" align="center">
            <template #default="{ row }">
              <el-tag :type="slaType(row.sla_status)" size="small" effect="plain">{{ slaLabel(row.sla_status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="assignee_name" label="处理人" width="120" align="center" />
          <el-table-column prop="created_at" label="创建时间" width="170" />
          <el-table-column label="操作" width="140" fixed="right" align="center">
            <template #default="{ row }">
              <div class="action-group">
                <el-button size="small" text :icon="View" @click="viewDetail(row)" />
                <el-button size="small" text :icon="EditPen" @click="openEdit(row)" />
                <el-popconfirm title="确认删除?" @confirm="handleDelete(row.id)">
                  <template #reference><el-button size="small" text type="danger" :icon="Delete" /></template>
                </el-popconfirm>
              </div>
            </template>
          </el-table-column>
        </el-table>

        <div class="pagination-wrap" v-if="total > 0">
          <el-pagination
            v-model:current-page="page" v-model:page-size="pageSize"
            :total="total" layout="total, prev, pager, next"
            @current-change="fetchData"
          />
        </div>
      </el-card>
    </div>

    <!-- Create Dialog -->
    <el-dialog v-model="createVisible" title="新建工单" width="700px">
      <el-form :model="createForm" label-width="100px">
        <el-form-item label="标题" required>
          <el-input v-model="createForm.title" placeholder="工单标题" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="createForm.description" type="textarea" :rows="4" placeholder="问题描述" />
        </el-form-item>
        <el-form-item label="优先级">
          <el-select v-model="createForm.priority">
            <el-option label="低" value="low" />
            <el-option label="中" value="medium" />
            <el-option label="高" value="high" />
            <el-option label="紧急" value="urgent" />
          </el-select>
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="createForm.category_id" clearable placeholder="选择分类">
            <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>

      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- Edit Dialog -->
    <el-dialog v-model="editVisible" title="编辑工单" width="700px">
      <el-form :model="editForm" label-width="100px">
        <el-form-item label="标题" required>
          <el-input v-model="editForm.title" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="editForm.description" type="textarea" :rows="4" />
        </el-form-item>
        <el-form-item label="优先级">
          <el-select v-model="editForm.priority">
            <el-option label="低" value="low" />
            <el-option label="中" value="medium" />
            <el-option label="高" value="high" />
            <el-option label="紧急" value="urgent" />
          </el-select>
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="editForm.category_id" clearable placeholder="选择分类">
            <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>

      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" @click="handleEdit">保存</el-button>
      </template>
    </el-dialog>

    <!-- Detail Dialog -->
    <el-dialog v-model="detailVisible" title="工单详情" width="750px" v-if="detail">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="工单号">{{ detail.ticket_no }}</el-descriptions-item>
        <el-descriptions-item label="状态">
            <el-select :model-value="detail.status" size="small" @change="(v: string) => updateStatus(detail.id, v)">
              <el-option v-for="s in statusOptions" :key="s.value" :label="s.label" :value="s.value" />
            </el-select>
        </el-descriptions-item>
        <el-descriptions-item label="标题" :span="2">{{ detail.title }}</el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">{{ detail.description }}</el-descriptions-item>
        <el-descriptions-item label="优先级">
          <el-tag :type="priorityTag(detail.priority)" size="small">{{ detail.priority }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="SLA 状态">
          <el-tag :type="slaType(detail.sla_status)" size="small">{{ slaLabel(detail.sla_status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="截止时间">{{ detail.due_at || '-' }}</el-descriptions-item>
        <el-descriptions-item label="来源">{{ detail.source }}</el-descriptions-item>
        <el-descriptions-item label="处理人">{{ detail.assignee_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ detail.created_at }}</el-descriptions-item>
      </el-descriptions>

      <el-divider />
      <div class="section-header">
        <h4>标签</h4>
        <el-button size="small" @click="openTagSelector">管理标签</el-button>
      </div>
      <div class="tags-wrap">
        <el-tag v-for="t in (detail.tags || [])" :key="t.id"
          :color="tagColor(t.color)" style="color:#fff; margin-right:6px; margin-bottom:4px;" closable @close="removeTag(t.id)">
          {{ t.name }}
        </el-tag>
        <span v-if="!detail.tags?.length" class="empty-hint">暂无标签</span>
      </div>

      <el-divider />
      <h4>评论 &amp; 备注</h4>
      <div v-for="c in (detail.comments || [])" :key="c.id" class="comment-item">
        <div class="comment-meta">
          <strong>{{ c.author_name || '系统' }}</strong>
          <span class="comment-time">{{ c.created_at }}</span>
          <el-tag v-if="c.is_internal" size="small" type="warning">内部备注</el-tag>
        </div>
        <div class="comment-body">{{ c.content }}</div>
      </div>

      <div class="comment-input">
        <el-input v-model="commentContent" type="textarea" :rows="3" placeholder="输入评论或备注..." />
        <div class="comment-actions">
          <el-checkbox v-model="commentInternal">内部备注（仅客服可见）</el-checkbox>
          <el-button type="primary" @click="addComment">发送</el-button>
        </div>
      </div>
    </el-dialog>

    <el-dialog v-model="tagSelectVisible" title="选择标签" width="400px">
      <el-checkbox-group v-model="selectedTagIDs">
        <el-checkbox v-for="t in allTags" :key="t.id" :value="t.id" :label="t.id">{{ t.name }}</el-checkbox>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="tagSelectVisible = false">取消</el-button>
        <el-button type="primary" @click="saveTags">保存</el-button>
      </template>
    </el-dialog>
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
  font-size: 18px;
  font-weight: 600;
  color: #1d2129;
}
.table-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  gap: 12px;
  flex-shrink: 0;
}
.toolbar-left {
  flex: 1;
}
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}
.search-input {
  width: 320px;
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
.beauty-table {
  --el-table-border-color: #f0f0f0;
  flex: 1;
  min-height: 350px;
}
.beauty-table :deep(.el-table__header th) {
  background: #f7f8fa;
  color: #4e5969;
  font-weight: 500;
}
.name-cell {
  color: #1d2129;
  font-weight: 500;
}
.action-group {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  white-space: nowrap;
}
.action-group .el-button {
  margin-left: 0;
}
.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
  flex-shrink: 0;
}
.comment-item { border-bottom: 1px solid #f0f0f0; padding: 12px 0; }
.comment-meta { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
.comment-time { color: #999; font-size: 12px; }
.comment-body { white-space: pre-wrap; color: #333; }
.comment-input { margin-top: 16px; }
.comment-actions { display: flex; align-items: center; justify-content: space-between; margin-top: 8px; }
.section-header { display: flex; align-items: center; justify-content: space-between; }
.section-header h4 { margin: 0; }
.tags-wrap { margin-top: 8px; }
.empty-hint { color: #999; font-size: 13px; }
</style>
