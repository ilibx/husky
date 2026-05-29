<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

interface Ticket {
  id: number; ticket_no: string; title: string; status: string; priority: string
  source: string; assignee_name: string; created_at: string; category_id?: number
  description?: string; sla_status?: string; due_at?: string
}
interface Tag { id: number; name: string; color: string }
interface TicketField {
  id: number; name: string; field_key: string; field_type: string; options: string
  required: boolean; sort_order: number; placeholder: string; enabled: boolean
}
interface Category { id: number; name: string }

const list = ref<Ticket[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const detailVisible = ref(false)
const detail = ref<any>(null)
const commentContent = ref('')
const commentInternal = ref(false)
const detailFieldValues = ref<Record<string, string>>({})

const allTags = ref<Tag[]>([])
const tagSelectVisible = ref(false)
const selectedTagIDs = ref<number[]>([])

const ticketFields = ref<TicketField[]>([])
const categories = ref<Category[]>([])

const createVisible = ref(false)
const createForm = ref({ title: '', description: '', priority: 'medium', category_id: null as number | null })
const createFieldVals = ref<Record<string, string>>({})

const editVisible = ref(false)
const editId = ref(0)
const editForm = ref({ title: '', description: '', priority: 'medium', category_id: null as number | null })
const editFieldVals = ref<Record<string, string>>({})

const sortedFields = computed(() =>
  [...ticketFields.value].filter(f => f.enabled).sort((a, b) => a.sort_order - b.sort_order)
)

async function fetchData() {
  loading.value = true
  try {
    const res: any = await request.get('/tickets', { params: { page: page.value, page_size: pageSize.value } })
    list.value = res.data?.list || res.data || []
    total.value = res.data?.total || res.total || 0
  } catch { list.value = [] }
  finally { loading.value = false }
}

async function fetchMeta() {
  try {
    const [fRes, cRes] = await Promise.all([
      request.get('/tickets/fields/definitions'),
      request.get('/categories'),
    ])
    ticketFields.value = fRes.data || []
    categories.value = cRes.data || []
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
    await Promise.all([fetchTags(), fetchDetailFields(row.id)])
  } catch { ElMessage.error('获取工单详情失败') }
}

async function fetchDetailFields(ticketId: number) {
  try {
    const res: any = await request.get(`/tickets/${ticketId}/fields`)
    const vals: Record<string, string> = {}
    ;(res.data || []).forEach((fv: any) => { vals[String(fv.field_id)] = fv.value })
    detailFieldValues.value = vals
  } catch { detailFieldValues.value = {} as Record<string, string> }
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
  createFieldVals.value = {}
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
    const res: any = await request.post('/tickets', payload)
    const ticketId = res.data?.id || res.id
    await saveFieldValues(ticketId, createFieldVals.value)
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
  editFieldVals.value = {}
  request.get(`/tickets/${row.id}/fields`).then((res: any) => {
    const vals: Record<string, string> = {}
    ;(res.data || []).forEach((fv: any) => { vals[String(fv.field_id)] = fv.value })
    editFieldVals.value = vals
  }).catch(() => { /* ignore */ })
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
    await saveFieldValues(editId.value, editFieldVals.value)
    ElMessage.success('更新成功')
    editVisible.value = false
    fetchData()
  } catch { /* handled by interceptor */ }
}

async function saveFieldValues(ticketId: number, vals: Record<string, string>) {
  const entries = Object.entries(vals).filter(([_, v]) => v !== '' && v !== undefined && v !== null)
  if (entries.length === 0) return
  await request.put(`/tickets/${ticketId}/fields`, entries.map(([fieldId, value]) => ({
    ticket_id: ticketId, field_id: Number(fieldId), value: String(value),
  })))
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确认删除该工单吗？', '提示')
    await request.delete(`/tickets/${id}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch { /* cancelled or error */ }
}

function fieldComponent(field: TicketField) {
  const map: Record<string, string> = { text: 'el-input', textarea: 'el-input', select: 'el-select', multi_select: 'el-select', number: 'el-input-number', date: 'el-date-picker', boolean: 'el-switch' }
  return map[field.field_type] || 'el-input'
}

function fieldOptions(field: TicketField): string[] {
  try { return JSON.parse(field.options || '[]') } catch { return [] }
}

const statusOptions = ['open', 'in_progress', 'pending', 'resolved', 'closed']
const statusMap: Record<string, string> = { open: 'info', in_progress: 'primary', pending: 'warning', resolved: 'success', closed: '' }
function statusTag(s: string) { return (statusMap[s] || 'info') as 'info' | 'primary' | 'success' | 'warning' | 'danger' | '' }

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
  <div>
    <div class="page-header">
      <h2>工单管理</h2>
      <el-button type="primary" @click="openCreate">新建工单</el-button>
    </div>

    <el-card>
      <el-table :data="list" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="ticket_no" label="工单号" width="140" />
        <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="140">
          <template #default="{ row }">
            <el-select :model-value="row.status" size="small" style="width: 110px" @change="(v: string) => updateStatus(row.id, v)">
              <el-option v-for="s in statusOptions" :key="s" :label="s" :value="s" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="80">
          <template #default="{ row }">
            <el-tag :type="priorityTag(row.priority)" size="small">{{ row.priority }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="source" label="来源" width="100" />
        <el-table-column prop="sla_status" label="SLA" width="100">
          <template #default="{ row }">
            <el-tag :type="slaType(row.sla_status)" size="small">{{ slaLabel(row.sla_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="assignee_name" label="处理人" width="120" />
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="viewDetail(row)">详情</el-button>
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="page" v-model:page-size="pageSize"
          :total="total" layout="total, prev, pager, next"
          @current-change="fetchData"
        />
      </div>
    </el-card>

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

        <el-divider v-if="sortedFields.length > 0" />
        <template v-for="f in sortedFields" :key="f.id">
          <el-form-item :label="f.name" v-if="f.field_type === 'text'">
            <el-input v-model="createFieldVals[String(f.id)]" :placeholder="f.placeholder || f.name" />
          </el-form-item>
          <el-form-item :label="f.name" v-else-if="f.field_type === 'textarea'">
            <el-input v-model="createFieldVals[String(f.id)]" type="textarea" :rows="3" :placeholder="f.placeholder || f.name" />
          </el-form-item>
          <el-form-item :label="f.name" v-else-if="f.field_type === 'select'">
            <el-select v-model="createFieldVals[String(f.id)]" :placeholder="f.placeholder || '请选择'" clearable>
              <el-option v-for="o in fieldOptions(f)" :key="o" :label="o" :value="o" />
            </el-select>
          </el-form-item>
          <el-form-item :label="f.name" v-else-if="f.field_type === 'multi_select'">
            <el-select v-model="createFieldVals[String(f.id)]" multiple collapse-tags :placeholder="f.placeholder || '请选择'">
              <el-option v-for="o in fieldOptions(f)" :key="o" :label="o" :value="o" />
            </el-select>
          </el-form-item>
          <el-form-item :label="f.name" v-else-if="f.field_type === 'number'">
            <el-input v-model="createFieldVals[String(f.id)]" type="number" :placeholder="f.placeholder || f.name" />
          </el-form-item>
          <el-form-item :label="f.name" v-else-if="f.field_type === 'date'">
            <el-date-picker v-model="createFieldVals[String(f.id)]" type="date" style="width:100%" />
          </el-form-item>
          <el-form-item :label="f.name" v-else-if="f.field_type === 'boolean'">
            <el-switch v-model="createFieldVals[String(f.id)]" :active-value="'true'" :inactive-value="'false'" />
          </el-form-item>
        </template>
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

        <el-divider v-if="sortedFields.length > 0" />
        <template v-for="f in sortedFields" :key="f.id">
          <el-form-item :label="f.name" v-if="f.field_type === 'text'">
            <el-input v-model="editFieldVals[String(f.id)]" :placeholder="f.placeholder || f.name" />
          </el-form-item>
          <el-form-item :label="f.name" v-else-if="f.field_type === 'textarea'">
            <el-input v-model="editFieldVals[String(f.id)]" type="textarea" :rows="3" :placeholder="f.placeholder || f.name" />
          </el-form-item>
          <el-form-item :label="f.name" v-else-if="f.field_type === 'select'">
            <el-select v-model="editFieldVals[String(f.id)]" :placeholder="f.placeholder || '请选择'" clearable>
              <el-option v-for="o in fieldOptions(f)" :key="o" :label="o" :value="o" />
            </el-select>
          </el-form-item>
          <el-form-item :label="f.name" v-else-if="f.field_type === 'multi_select'">
            <el-select v-model="editFieldVals[String(f.id)]" multiple collapse-tags :placeholder="f.placeholder || '请选择'">
              <el-option v-for="o in fieldOptions(f)" :key="o" :label="o" :value="o" />
            </el-select>
          </el-form-item>
          <el-form-item :label="f.name" v-else-if="f.field_type === 'number'">
            <el-input v-model="editFieldVals[String(f.id)]" type="number" :placeholder="f.placeholder || f.name" />
          </el-form-item>
          <el-form-item :label="f.name" v-else-if="f.field_type === 'date'">
            <el-date-picker v-model="editFieldVals[String(f.id)]" type="date" style="width:100%" />
          </el-form-item>
          <el-form-item :label="f.name" v-else-if="f.field_type === 'boolean'">
            <el-switch v-model="editFieldVals[String(f.id)]" :active-value="'true'" :inactive-value="'false'" />
          </el-form-item>
        </template>
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
            <el-option v-for="s in statusOptions" :key="s" :label="s" :value="s" />
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

      <el-divider v-if="sortedFields.length > 0" />
      <template v-if="sortedFields.length > 0">
        <h4>自定义字段</h4>
        <el-descriptions :column="2" border style="margin-top:8px">
          <el-descriptions-item v-for="f in sortedFields" :key="f.id" :label="f.name">
            <template v-if="f.field_type === 'boolean'">
              <el-tag :type="detailFieldValues[String(f.id)] === 'true' ? 'success' : 'info'">
                {{ detailFieldValues[String(f.id)] === 'true' ? '是' : '否' }}
              </el-tag>
            </template>
            <template v-else>
              {{ detailFieldValues[String(f.id)] || '-' }}
            </template>
          </el-descriptions-item>
        </el-descriptions>
      </template>

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
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }
.page-header h2 { margin: 0; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
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
