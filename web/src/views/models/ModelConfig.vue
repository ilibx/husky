<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox, ElPopconfirm } from 'element-plus'
import { Plus, Delete, EditPen, Search, Refresh } from '@element-plus/icons-vue'
import request from '@/api/request'

interface ModelItem {
  name: string
  provider: string
  enabled: boolean
  max_tokens: number
}

interface RemoteModelItem {
  id: string
  name: string
  provider: string
  provider_name: string
  max_tokens: number
  context_tokens: number
  modalities: string
}

const models = ref<ModelItem[]>([])
const providers = ref<{ name: string }[]>([])
const loading = ref(false)
const keyword = ref('')
let searchTimer: ReturnType<typeof setTimeout>
const modelDialog = ref(false)
const importDialog = ref(false)
const importLoading = ref(false)
const remoteModels = ref<RemoteModelItem[]>([])
const selectedRemoteModels = ref<RemoteModelItem[]>([])
const modelForm = ref<ModelItem>({ name: '', provider: 'openai', enabled: true, max_tokens: 4096 })
const savingModel = ref(false)

const filteredModels = computed(() => {
  if (!keyword.value) return models.value
  const kw = keyword.value.toLowerCase()
  return models.value.filter(m => m.name.toLowerCase().includes(kw))
})

async function fetchModels() {
  loading.value = true
  try {
    const res: any = await request.get('/system-config', { params: { page: 1, page_size: 200 } })
    const list: any[] = res.data?.data || []
    models.value = list
      .filter((c: any) => c.category === 'llm' && c.key.startsWith('model:'))
      .map((c: any) => {
        try {
          const v = JSON.parse(c.value)
          return { name: c.key.replace('model:', ''), max_tokens: 4096, ...v }
        } catch { return null }
      })
      .filter(Boolean)
  } catch { models.value = [] }
  finally { loading.value = false }
}

async function fetchProviders() {
  try {
    const res: any = await request.get('/system-config', { params: { page: 1, page_size: 200 } })
    const list: any[] = res.data?.data || []
    providers.value = list
      .filter((c: any) => c.category === 'llm' && c.key.startsWith('preset:'))
      .map((c: any) => {
        try { return JSON.parse(c.value) } catch { return null }
      })
      .filter(Boolean)
  } catch { providers.value = [] }
}

function openAddModel() {
  modelForm.value = {
    name: '',
    provider: providers.value[0]?.name || 'openai',
    enabled: true,
    max_tokens: 4096,
  }
  modelDialog.value = true
}

function openEditModel(item: ModelItem) {
  modelForm.value = { ...item }
  modelDialog.value = true
}

function parseRemoteModels(data: any): RemoteModelItem[] {
  const result: RemoteModelItem[] = []
  for (const provider of Object.values<any>(data || {})) {
    const models = provider?.models || {}
    for (const model of Object.values<any>(models)) {
      const input = model?.modalities?.input || []
      const output = model?.modalities?.output || []
      result.push({
        id: model.id || model.name,
        name: model.name || model.id,
        provider: provider.id || provider.name || '',
        provider_name: provider.name || provider.id || '',
        max_tokens: Number(model?.limit?.output || model?.limit?.context || 4096),
        context_tokens: Number(model?.limit?.context || 0),
        modalities: [...input, ...output].join(', '),
      })
    }
  }
  return result.sort((a, b) => `${a.provider}:${a.id}`.localeCompare(`${b.provider}:${b.id}`))
}

async function openImportModels() {
  importDialog.value = true
  if (remoteModels.value.length > 0) return
  importLoading.value = true
  try {
    const res: any = await request.get('/system-config/model-catalog')
    remoteModels.value = parseRemoteModels(res.data || res)
  } catch {
    ElMessage.error('拉取 models.dev 模型列表失败')
  } finally {
    importLoading.value = false
  }
}

function handleRemoteSelection(rows: RemoteModelItem[]) {
  selectedRemoteModels.value = rows
}

async function importSelectedModels() {
  if (selectedRemoteModels.value.length === 0) {
    ElMessage.warning('请选择要导入的模型')
    return
  }
  savingModel.value = true
  try {
    await Promise.all(selectedRemoteModels.value.map(item => request.post('/system-config/upsert', {
      category: 'llm',
      key: `model:${item.id}`,
      value: JSON.stringify({
        provider: item.provider,
        enabled: true,
        max_tokens: item.max_tokens,
      }),
    })))
    ElMessage.success(`已导入 ${selectedRemoteModels.value.length} 个模型`)
    importDialog.value = false
    selectedRemoteModels.value = []
    fetchModels()
  } catch { /* handled */ }
  finally { savingModel.value = false }
}

async function handleSaveModel() {
  if (!modelForm.value.name) { ElMessage.warning('请输入模型名称'); return }
  savingModel.value = true
  try {
    const value = JSON.stringify({
      provider: modelForm.value.provider,
      enabled: modelForm.value.enabled,
      max_tokens: modelForm.value.max_tokens,
    })
    await request.post('/system-config/upsert', {
      category: 'llm', key: `model:${modelForm.value.name}`, value,
    })
    ElMessage.success('模型已保存')
    modelDialog.value = false
    fetchModels()
  } catch { /* handled */ } finally { savingModel.value = false }
}

async function handleDeleteModel(name: string) {
  try {
    const res: any = await request.get('/system-config', { params: { page: 1, page_size: 200 } })
    const list: any[] = res.data?.data || []
    const cfg = list.find((c: any) => c.key === `model:${name}`)
    if (cfg) await request.delete(`/system-config/${cfg.id}`)
    ElMessage.success('已删除')
    fetchModels()
  } catch { /* cancelled or error */ }
}

async function handleActivateModel(item: ModelItem) {
  try {
    await Promise.all([
      request.post('/system-config/upsert', { category: 'llm', key: 'llm_model', value: item.name }),
      request.post('/system-config/upsert', { category: 'llm', key: 'llm_max_tokens', value: String(item.max_tokens) }),
    ])
    ElMessage.success(`已激活模型 "${item.name}"`)
  } catch { /* handled */ }
}

async function handleToggleModel(item: ModelItem) {
  try {
    const value = JSON.stringify({
      provider: item.provider, enabled: item.enabled, max_tokens: item.max_tokens,
    })
    await request.post('/system-config/upsert', {
      category: 'llm', key: `model:${item.name}`, value,
    })
  } catch { item.enabled = !item.enabled }
}

function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(fetchModels, 300)
}

onMounted(() => {
  fetchModels()
  fetchProviders()
})
</script>

<template>
  <div class="page">
    <div class="list-view">
      <div class="page-header">
        <h2>模型管理</h2>
      </div>

      <el-card shadow="never" class="list-card">
        <div class="table-toolbar">
          <div class="toolbar-left">
            <el-input v-model="keyword" placeholder="搜索模型名称..." clearable @input="onSearch" class="search-input">
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
          </div>
          <div class="toolbar-right">
            <el-button :icon="Refresh" circle @click="fetchModels" />
            <el-button type="primary" :icon="Plus" @click="openAddModel">添加模型</el-button>
            <el-button @click="openImportModels">从 models.dev 导入</el-button>
          </div>
        </div>

        <el-table :data="keyword ? filteredModels : models" v-loading="loading" class="beauty-table" style="width:100%" height="calc(100vh - 100px)">
          <el-table-column prop="name" label="模型名称" min-width="200">
            <template #default="{ row }"><span class="name-cell">{{ row.name }}</span></template>
          </el-table-column>
          <el-table-column prop="provider" label="厂商名称" width="180" />
          <el-table-column prop="max_tokens" label="Max Tokens" width="120" align="center" />
          <el-table-column label="启用" width="80" align="center">
            <template #default="{ row }">
              <el-switch v-model="row.enabled" size="small" active-color="#67c23a" inactive-color="#c0c4cc" @change="handleToggleModel(row)" />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right" align="center">
            <template #default="{ row }">
              <el-button size="small" type="success" @click="handleActivateModel(row)">激活</el-button>
              <div class="action-group" style="display:inline-flex;vertical-align:middle;margin-left:4px">
                <el-button size="small" text :icon="EditPen" @click="openEditModel(row)" />
                <el-popconfirm title="确认删除?" @confirm="handleDeleteModel(row.name)">
                  <template #reference><el-button size="small" text type="danger" :icon="Delete" /></template>
                </el-popconfirm>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <el-dialog v-model="modelDialog" :title="modelForm.name ? '编辑模型' : '添加模型'" width="450px">
      <el-form :model="modelForm" label-width="100px">
        <el-form-item label="模型名称">
          <el-input v-model="modelForm.name" placeholder="gpt-4o-mini" />
        </el-form-item>
        <el-form-item label="厂商名称">
          <el-select v-model="modelForm.provider" style="width:100%" filterable allow-create default-first-option>
            <el-option v-for="p in providers" :key="p.name" :label="p.name" :value="p.name" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="Max Tokens">
          <el-input-number v-model="modelForm.max_tokens" :min="256" :max="1024000" :step="1024" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="modelForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="modelDialog = false">取消</el-button>
        <el-button type="primary" :loading="savingModel" @click="handleSaveModel">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="importDialog" title="从 models.dev 导入模型" width="900px">
      <el-table :data="remoteModels" v-loading="importLoading" height="520" border @selection-change="handleRemoteSelection">
        <el-table-column type="selection" width="48" />
        <el-table-column prop="id" label="模型 ID" min-width="220" show-overflow-tooltip />
        <el-table-column prop="name" label="名称" min-width="180" show-overflow-tooltip />
        <el-table-column prop="provider_name" label="厂商" width="140" show-overflow-tooltip />
        <el-table-column prop="max_tokens" label="输出上限" width="110" />
        <el-table-column prop="context_tokens" label="上下文" width="110" />
        <el-table-column prop="modalities" label="模态" width="140" show-overflow-tooltip />
      </el-table>
      <template #footer>
        <el-button @click="importDialog = false">取消</el-button>
        <el-button type="primary" :loading="savingModel" @click="importSelectedModels">导入选中</el-button>
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
.name-cell { color: #1d2129; font-weight: 500; }
.action-group { display: flex; align-items: center; justify-content: center; gap: 4px; white-space: nowrap; }
.action-group .el-button { margin-left: 0; }
</style>
