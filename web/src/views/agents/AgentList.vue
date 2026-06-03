<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox, ElPopconfirm } from 'element-plus'
import { ArrowLeft, Plus, Delete, EditPen, Search, Refresh } from '@element-plus/icons-vue'
import request from '@/api/request'
import MarkdownEditor from '@/components/MarkdownEditor.vue'

interface Agent {
  id: number
  name: string
  description: string
  type: string
  config: string
  model: string
  temperature: number
  max_tokens: number
  enabled: boolean
  created_at: string
}

const agents = ref<Agent[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const keyword = ref('')

const currentView = ref<'list' | 'edit'>('list')
const isEdit = ref(false)
const activeTab = ref('params')
const saving = ref(false)
const form = ref({
  id: 0,
  name: '',
  description: '',
  type: 'llm',
  config: '',
  model: '',
  temperature: 0.7,
  max_tokens: 2048,
  enabled: true,
})

async function fetchData() {
  loading.value = true
  try {
    const params: any = { page: page.value, page_size: pageSize.value }
    if (keyword.value) params.keyword = keyword.value
    const res: any = await request.get('/agents', { params })
    agents.value = res?.list || res?.data?.list || res?.data?.data || res?.data || []
    total.value = res?.total || res?.data?.total || 0
  } catch {
    agents.value = []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  isEdit.value = false
  form.value = { id: 0, name: '', description: '', type: 'llm', config: '', model: '', temperature: 0.7, max_tokens: 2048, enabled: true }
  activeTab.value = 'params'
  currentView.value = 'edit'
}

function openEdit(row: Agent) {
  isEdit.value = true
  form.value = { id: row.id, name: row.name, description: row.description, type: row.type, config: row.config, model: row.model, temperature: row.temperature, max_tokens: row.max_tokens, enabled: row.enabled }
  activeTab.value = 'params'
  currentView.value = 'edit'
}

function goBack() {
  currentView.value = 'list'
  fetchData()
}

async function handleSave() {
  saving.value = true
  try {
    if (form.value.id) {
      await request.put(`/agents/${form.value.id}`, form.value)
      ElMessage.success('更新成功')
    } else {
      await request.post('/agents', form.value)
      ElMessage.success('创建成功')
    }
    goBack()
  } catch {
    // handled by interceptor
  } finally {
    saving.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确认删除该 Agent 吗？', '提示')
    await request.delete(`/agents/${id}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch {
    // cancelled or error
  }
}

let searchTimer: ReturnType<typeof setTimeout>
function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => { page.value = 1; fetchData() }, 300)
}

function stripMarkdown(text: string) {
  if (!text) return ''
  return text.replace(/^#{1,6}\s+/gm, '').replace(/\*\*(.+?)\*\*/g, '$1').replace(/\*(.+?)\*/g, '$1').replace(/`{1,3}[^`]*`{1,3}/g, '').replace(/\[(.+?)\]\(.+?\)/g, '$1').replace(/^[-*+]\s+/gm, '').replace(/^\d+\.\s+/gm, '').replace(/^>\s+/gm, '').trim()
}

function formatDate(val: string) {
  if (!val) return '-'
  return val.slice(0, 10)
}

async function toggleAgentEnabled(row: Agent, val: boolean) {
  try {
    await request.put(`/agents/${row.id}`, { ...row, enabled: !!val })
    ElMessage.success(val ? '已启用' : '已停用')
  } catch {
    row.enabled = !val
  }
}

onMounted(fetchData)
</script>

<template>
  <div class="page">
    <template v-if="currentView === 'list'">
      <div class="list-view">
        <div class="page-header">
          <h2>Agent 管理</h2>
        </div>
        <el-card shadow="never" class="list-card">
          <div class="table-toolbar">
            <div class="toolbar-left">
              <el-input v-model="keyword" placeholder="搜索 Agent 名称..." clearable @input="onSearch" class="search-input">
                <template #prefix><el-icon><Search /></el-icon></template>
              </el-input>
            </div>
            <div class="toolbar-right">
              <el-button :icon="Refresh" circle @click="fetchData" />
              <el-button type="primary" :icon="Plus" @click="openAdd">新增 Agent</el-button>
            </div>
          </div>
          <el-table :data="agents" v-loading="loading" stripe style="width:100%" class="beauty-table" height="calc(100vh - 100px)">
            <el-table-column prop="id" label="ID" width="64" align="center" />
            <el-table-column prop="name" label="名称" min-width="140">
              <template #default="{ row }"><span class="name-cell">{{ row.name }}</span></template>
            </el-table-column>
            <el-table-column prop="description" label="描述" min-width="180" show-overflow-tooltip>
              <template #default="{ row }"><span class="desc-preview">{{ stripMarkdown(row.description) || '-' }}</span></template>
            </el-table-column>
            <el-table-column prop="type" label="类型" width="86" align="center">
              <template #default="{ row }"><el-tag size="small" effect="plain">{{ row.type }}</el-tag></template>
            </el-table-column>
            <el-table-column prop="model" label="模型" min-width="110" show-overflow-tooltip />
            <el-table-column prop="temperature" label="温度" width="72" align="center" />
            <el-table-column prop="max_tokens" label="Token" width="86" align="center" />
            <el-table-column prop="enabled" label="状态" width="72" align="center">
              <template #default="{ row }">
                <el-switch v-model="row.enabled" size="small" active-color="#67c23a" inactive-color="#c0c4cc" @change="(val: any) => toggleAgentEnabled(row, val)" />
              </template>
            </el-table-column>
            <el-table-column label="创建时间" width="108" align="center">
              <template #default="{ row }"><span class="date-cell">{{ formatDate(row.created_at) }}</span></template>
            </el-table-column>
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
          <div class="pagination-wrap" v-if="total > 0">
            <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="fetchData" />
          </div>
        </el-card>
      </div>
    </template>

    <template v-else>
      <div class="edit-page">
        <div class="edit-topbar">
          <el-button text @click="goBack">
            <el-icon><ArrowLeft /></el-icon> 返回
          </el-button>
          <div class="topbar-title">{{ isEdit ? '编辑 Agent' : '新增 Agent' }}</div>
          <div class="topbar-actions">
            <el-button @click="goBack">取消</el-button>
            <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
          </div>
        </div>
        <el-card shadow="never" class="edit-card">
          <el-tabs v-model="activeTab">
            <el-tab-pane label="基本参数" name="params">
              <el-form :model="form" label-width="100px" class="edit-form">
                <el-row :gutter="24">
                  <el-col :span="12">
                    <el-form-item label="名称">
                      <el-input v-model="form.name" placeholder="Agent 名称" />
                    </el-form-item>
                  </el-col>
                  <el-col :span="12">
                    <el-form-item label="类型">
                      <el-select v-model="form.type" style="width:100%">
                        <el-option label="LLM" value="llm" />
                        <el-option label="规则" value="rule" />
                        <el-option label="混合" value="hybrid" />
                      </el-select>
                    </el-form-item>
                  </el-col>
                </el-row>
                <el-row :gutter="24">
                  <el-col :span="12">
                    <el-form-item label="模型">
                      <el-input v-model="form.model" placeholder="如 gpt-4o" />
                    </el-form-item>
                  </el-col>
                  <el-col :span="12">
                    <el-form-item label="温度">
                      <el-input-number v-model="form.temperature" :min="0" :max="2" :step="0.1" style="width:100%" />
                    </el-form-item>
                  </el-col>
                </el-row>
                <el-form-item label="启用">
                  <el-switch v-model="form.enabled" />
                </el-form-item>
                <el-form-item label="配置">
                  <el-input v-model="form.config" type="textarea" :rows="4" placeholder="Agent 配置 JSON（可选）" />
                </el-form-item>
              </el-form>
            </el-tab-pane>
            <el-tab-pane label="描述编辑" name="content">
              <MarkdownEditor v-model="form.description" />
            </el-tab-pane>
          </el-tabs>
        </el-card>
      </div>
    </template>
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
.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
  flex-shrink: 0;
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
.name-cell { color: #1d2129; font-weight: 500; }
.desc-preview { color: #86909c; font-size: 13px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; display: block; }
.date-cell { color: #86909c; font-size: 13px; }
.action-group { display: flex; align-items: center; justify-content: center; gap: 4px; white-space: nowrap; }
.action-group .el-button { margin-left: 0; }

/* Edit page */
.edit-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 120px);
}
.edit-topbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 0 16px;
  flex-shrink: 0;
}
.topbar-title {
  flex: 1;
  font-size: 16px;
  font-weight: 600;
  color: #1d2129;
}
.topbar-actions {
  display: flex;
  gap: 8px;
}
.edit-card {
  flex: 1;
  display: flex;
  flex-direction: column;
  border-radius: 12px;
  border: 1px solid #e5e6eb;
  overflow: hidden;
}
.edit-card :deep(.el-card__body) {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 20px 24px;
}
.edit-card :deep(.el-tabs) {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.edit-card :deep(.el-tabs__content) {
  flex: 1;
  position: relative;
}
.edit-card :deep(.el-tab-pane) {
  height: 100%;
  overflow-y: auto;
}
.edit-form {
  max-width: 800px;
}
.edit-card :deep(.md-editor) {
  display: flex;
  flex-direction: column;
  min-height: 400px;
  height: 100%;
}
.edit-card :deep(.md-editor-body) {
  flex: 1;
}
.edit-card :deep(.md-editor-textarea),
.edit-card :deep(.md-editor-preview) {
  height: 100% !important;
  min-height: 0 !important;
}
</style>
