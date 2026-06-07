<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox, ElPopconfirm } from 'element-plus'
import { Plus, Delete, EditPen, Search, Refresh } from '@element-plus/icons-vue'
import request from '@/api/request'

interface SLAConfig {
  id: number
  priority: string
  category_id: number | null
  category_name?: string
  response_minutes: number
  resolution_minutes: number
  warning_threshold: number
  warning_level: string
  enabled: boolean
  created_at: string
}

interface Category {
  id: number
  name: string
}

const slas = ref<SLAConfig[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const keyword = ref('')
let searchTimer: ReturnType<typeof setTimeout>

const dialogVisible = ref(false)
const formTitle = ref('新增SLA配置')
const form = ref<SLAConfig>({
  id: 0,
  priority: 'medium',
  category_id: null,
  response_minutes: 60,
  resolution_minutes: 480,
  warning_threshold: 0.8,
  warning_level: 'warning',
  enabled: true,
  created_at: '',
})

const categories = ref<Category[]>([])
const levelOptions = ref<{ name: string; key: string; color: string }[]>([])

async function fetchLevels() {
  try {
    const res: any = await request.get('/system-config/lookup', { params: { category: 'push', key: 'levels' } })
    if (res.data?.value) levelOptions.value = JSON.parse(res.data.value)
  } catch { levelOptions.value = [] }
}

async function fetchData() {
  loading.value = true
  try {
    const params: any = { page: page.value, page_size: pageSize.value }
    if (keyword.value) params.keyword = keyword.value
    const res: any = await request.get('/sla-configs', { params })
    slas.value = res.data?.data || res.data || []
    total.value = res.data?.total || res.total || 0
  } catch {
    slas.value = []
  } finally {
    loading.value = false
  }
}

async function fetchCategories() {
  try {
    const res: any = await request.get('/categories')
    categories.value = res.data?.data || res.data || []
  } catch {
    categories.value = []
  }
}

function openAdd() {
  formTitle.value = '新增SLA配置'
  form.value = {
    id: 0, priority: 'medium', category_id: null, response_minutes: 60,
    resolution_minutes: 480, warning_threshold: 0.8, warning_level: 'warning',
    enabled: true, created_at: '',
  }
  dialogVisible.value = true
}

function openEdit(row: SLAConfig) {
  formTitle.value = '编辑SLA配置'
  form.value = { ...row }
  dialogVisible.value = true
}

async function handleSave() {
  try {
    if (form.value.id) {
      await request.put(`/sla-configs/${form.value.id}`, form.value)
      ElMessage.success('更新成功')
    } else {
      await request.post('/sla-configs', form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch { /* handled */ }
}

async function handleDelete(id: number) {
  try {
    await request.delete(`/sla-configs/${id}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch { /* handled */ }
}

async function toggleSLA(row: SLAConfig) {
  try {
    await request.put(`/sla-configs/${row.id}`, { ...row, enabled: !row.enabled })
  } catch { row.enabled = !row.enabled }
}

function priorityType(priority: string): 'primary' | 'success' | 'warning' | 'info' | 'danger' {
  const map: Record<string, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = { low: 'info', medium: 'warning', high: 'danger', urgent: 'danger' }
  return map[priority] || 'info'
}

function priorityLabel(priority: string): string {
  const map: Record<string, string> = { low: '低', medium: '中', high: '高', urgent: '紧急' }
  return map[priority] || priority
}

function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => { page.value = 1; fetchData() }, 300)
}

onMounted(() => {
  fetchData()
  fetchCategories()
  fetchLevels()
})
</script>

<template>
  <div class="page">
    <div class="list-view">
      <div class="page-header">
        <h2>SLA 配置</h2>
      </div>

      <el-card shadow="never" class="list-card">
        <div class="table-toolbar">
          <div class="toolbar-left">
            <el-input v-model="keyword" placeholder="搜索 SLA 配置..." clearable @input="onSearch" class="search-input">
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
          </div>
          <div class="toolbar-right">
            <el-button :icon="Refresh" circle @click="fetchData" />
            <el-button type="primary" :icon="Plus" @click="openAdd">新增SLA配置</el-button>
          </div>
        </div>

        <el-table :data="slas" v-loading="loading" class="beauty-table" style="width:100%" height="calc(100vh - 100px)">
          <el-table-column prop="id" label="ID" width="60" align="center" />
          <el-table-column prop="priority" label="优先级" width="100" align="center">
            <template #default="{ row }">
              <el-tag :type="priorityType(row.priority)" size="small" effect="plain">{{ priorityLabel(row.priority) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="category_name" label="分类" min-width="120">
            <template #default="{ row }">
              <span>{{ row.category_name || '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="response_minutes" label="响应时间(分)" width="130" align="center" />
          <el-table-column prop="resolution_minutes" label="解决时间(分)" width="130" align="center" />
          <el-table-column prop="warning_threshold" label="预警阈值" width="100" align="center">
            <template #default="{ row }">
              <span>{{ (row.warning_threshold * 100).toFixed(0) }}%</span>
            </template>
          </el-table-column>
          <el-table-column prop="warning_level" label="预警等级" width="110" align="center">
            <template #default="{ row }">
              <el-tag v-if="row.warning_level" :color="levelOptions.find(l => l.key === row.warning_level)?.color || '#909399'" style="color:#fff" size="small" effect="dark">
                {{ levelOptions.find(l => l.key === row.warning_level)?.name || row.warning_level }}
              </el-tag>
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column label="启用" width="80" align="center">
            <template #default="{ row }">
              <el-switch v-model="row.enabled" size="small" active-color="#67c23a" inactive-color="#c0c4cc" @change="toggleSLA(row)" />
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="180" align="center" />
          <el-table-column label="操作" width="120" fixed="right" align="center">
            <template #default="{ row }">
              <div class="action-group">
                <el-button size="small" text :icon="EditPen" @click="openEdit(row)" />
                <el-popconfirm title="确认删除?" @confirm="handleDelete(row.id)">
                  <template #reference><el-button size="small" text type="danger" :icon="Delete" /></template>
                </el-popconfirm>
              </div>
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

    <el-dialog v-model="dialogVisible" :title="formTitle" width="600px">
      <el-form :model="form" label-width="140px">
        <el-form-item label="优先级">
          <el-select v-model="form.priority">
            <el-option label="低" value="low" />
            <el-option label="中" value="medium" />
            <el-option label="高" value="high" />
            <el-option label="紧急" value="urgent" />
          </el-select>
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="form.category_id" clearable placeholder="请选择分类">
            <el-option v-for="cat in categories" :key="cat.id" :label="cat.name" :value="cat.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="响应时间(分钟)">
          <el-input-number v-model="form.response_minutes" :min="1" />
        </el-form-item>
        <el-form-item label="解决时间(分钟)">
          <el-input-number v-model="form.resolution_minutes" :min="1" />
        </el-form-item>
        <el-form-item label="预警阈值">
          <div class="threshold-wrap">
            <el-slider v-model="form.warning_threshold" :min="0.1" :max="1" :step="0.05" style="width:200px" />
            <span class="threshold-value">{{ (form.warning_threshold * 100).toFixed(0) }}%</span>
          </div>
          <div class="form-help">当 SLA 时间消耗达到该百分比时触发预警推送</div>
        </el-form-item>
        <el-form-item label="预警等级">
          <el-select v-model="form.warning_level" style="width:200px">
            <el-option v-for="l in levelOptions" :key="l.key" :label="l.name" :value="l.key" />
          </el-select>
          <div class="form-help">推送管理中将按此等级匹配对应的推送方式。在"通知管理→等级管理"中配置</div>
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
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
.toolbar-left { flex: 1; }
.toolbar-right { display: flex; align-items: center; gap: 8px; }
.search-input { width: 320px; }
.beauty-table { --el-table-border-color: #f0f0f0; flex: 1; min-height: 350px; }
.beauty-table :deep(.el-table__header th) { background: #f7f8fa; color: #4e5969; font-weight: 500; }
.action-group { display: flex; align-items: center; justify-content: center; gap: 4px; white-space: nowrap; }
.action-group .el-button { margin-left: 0; }
.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
  flex-shrink: 0;
}
.threshold-wrap {
  display: flex;
  align-items: center;
  gap: 12px;
}
.threshold-value {
  font-weight: 600;
  min-width: 40px;
}
.form-help {
  font-size: 12px;
  color: var(--text-muted, #94a3b8);
  margin-top: 4px;
}
</style>
