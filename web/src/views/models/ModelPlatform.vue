<script setup lang="ts">
import { ref, onMounted, computed, nextTick } from 'vue'
import { ElMessage, ElMessageBox, ElPopconfirm } from 'element-plus'
import { ArrowLeft, Plus, Delete, EditPen, Search, Refresh, ArrowDown, ArrowRight } from '@element-plus/icons-vue'
import request from '@/api/request'

interface PlatformPreset {
  name: string
  type: string
  api_key: string
  base_url: string
  enabled: boolean
  models: string[]
}

interface ModelInfo {
  id: string
  owned_by: string
}

const providerTypes = [
  { label: 'OpenAI', value: 'openai' },
  { label: 'Claude (Anthropic)', value: 'claude' },
  { label: 'Google Gemini', value: 'gemini' },
  { label: '阿里通义千问 (DashScope)', value: 'dashscope' },
  { label: 'Azure OpenAI', value: 'azure' },
  { label: '自定义', value: 'custom' },
]

const providers = ref<PlatformPreset[]>([])
const loading = ref(false)
const keyword = ref('')
let searchTimer: ReturnType<typeof setTimeout>

const statusFilter = ref('all') // all, enabled, disabled

const currentView = ref<'list' | 'edit'>('list')
const isEdit = ref(false)
const originalName = ref('')
const saving = ref(false)
const form = ref<PlatformPreset>({
  name: '', type: 'openai', api_key: '', base_url: '',
  enabled: true, models: [],
})

const testing = ref(false)
const fetchingModels = ref(false)
const fetchedModels = ref<ModelInfo[]>([])
const selectedModels = ref<Set<string>>(new Set())
const collapsedGroups = ref<Set<string>>(new Set())
const newManualModel = ref('')
const manualInput = ref<any>(null)
const viewModelsDialog = ref(false)
const viewModelsList = ref<string[]>([])

const filteredProviders = computed(() => {
  let list = providers.value
  if (keyword.value) {
    const kw = keyword.value.toLowerCase()
    list = list.filter(p => {
      const nameMatch = p.name.toLowerCase().includes(kw) || p.type.toLowerCase().includes(kw)
      const modelMatch = p.models?.some(m => m.toLowerCase().includes(kw))
      return nameMatch || modelMatch
    })
  }
  if (statusFilter.value === 'enabled') list = list.filter(p => p.enabled)
  else if (statusFilter.value === 'disabled') list = list.filter(p => !p.enabled)
  return list
})

interface ModelGroup { owner: string; models: ModelInfo[] }

const groupedModels = computed(() => {
  const groups: Record<string, ModelInfo[]> = {}
  for (const m of fetchedModels.value) {
    const owner = m.owned_by || '其他'
    if (!groups[owner]) groups[owner] = []
    groups[owner].push(m)
  }
  return Object.entries(groups).map(([owner, models]) => ({ owner, models }))
})

function toggleModel(id: string) {
  if (selectedModels.value.has(id)) selectedModels.value.delete(id)
  else selectedModels.value.add(id)
}

function toggleGroup(owner: string) {
  if (collapsedGroups.value.has(owner)) collapsedGroups.value.delete(owner)
  else collapsedGroups.value.add(owner)
}

function selectAllFetched() {
  for (const m of fetchedModels.value) selectedModels.value.add(m.id)
}

function selectAll(owner: string) {
  const g = groupedModels.value.find(g => g.owner === owner)
  if (!g) return
  const allSelected = g.models.every(m => selectedModels.value.has(m.id))
  for (const m of g.models) {
    if (allSelected) selectedModels.value.delete(m.id)
    else selectedModels.value.add(m.id)
  }
}

function addManualModel() {
  const m = newManualModel.value.trim()
  if (!m) return
  selectedModels.value.add(m)
  newManualModel.value = ''
  nextTick(() => manualInput.value?.focus())
}

function removeSelectedModel(m: string) {
  selectedModels.value.delete(m)
}

async function fetchProviders() {
  loading.value = true
  try {
    const res: any = await request.get('/system-config', { params: { page: 1, page_size: 200 } })
    const list: any[] = res.data?.data || []
    providers.value = list
      .filter(c => c.category === 'llm' && c.key.startsWith('preset:'))
      .map(c => {
        try { return JSON.parse(c.value) } catch { return null }
      })
      .filter(Boolean) as PlatformPreset[]
  } catch { providers.value = [] }
  finally { loading.value = false }
}

async function fetchProviderModels() {
  if (!form.value.base_url || !form.value.api_key) {
    ElMessage.warning('请先填写接口地址和授权信息')
    return
  }
  fetchingModels.value = true
  try {
    const res: any = await request.post('/system-config/fetch-models', {
      type: form.value.type, api_key: form.value.api_key, base_url: form.value.base_url,
    })
    const list: ModelInfo[] = res.data?.models || []
    if (list.length === 0) { ElMessage.info('上游返回的模型列表为空'); return }
    fetchedModels.value = list
    // Do not auto-select models now
    ElMessage.success(`已获取 ${list.length} 个模型`)
  } catch (e: any) {
    ElMessage.error(e?.message || '获取模型列表失败')
  } finally { fetchingModels.value = false }
}

async function testConnection() {
  if (!form.value.base_url || !form.value.api_key) {
    ElMessage.warning('请填写接口地址和授权信息')
    return
  }
  testing.value = true
  try {
    await request.post('/system-config/test-connection', {
      type: form.value.type, api_key: form.value.api_key, base_url: form.value.base_url,
    })
    ElMessage.success('连接成功')
  } catch { /* handled */ }
  finally { testing.value = false }
}

async function handleSave() {
  if (!form.value.name) { ElMessage.warning('请输入平台名称'); return }
  form.value.models = [...selectedModels.value]
  saving.value = true
  try {
    if (isEdit.value && originalName.value && originalName.value !== form.value.name) {
      const res: any = await request.get('/system-config', { params: { page: 1, page_size: 200 } })
      const old = (res.data?.data || []).find((c: any) => c.key === `preset:${originalName.value}`)
      if (old) await request.delete(`/system-config/${old.id}`)
    }
    await request.post('/system-config/upsert', {
      category: 'llm', key: `preset:${form.value.name}`, value: JSON.stringify(form.value),
    })
    ElMessage.success('平台配置已保存')
    goBack()
  } catch { /* handled */ }
  finally { saving.value = false }
}

async function handleDelete(name: string) {
  try {
    const res: any = await request.get('/system-config', { params: { page: 1, page_size: 200 } })
    const cfg = (res.data?.data || []).find((c: any) => c.key === `preset:${name}`)
    if (cfg) await request.delete(`/system-config/${cfg.id}`)
    ElMessage.success('已删除')
    fetchProviders()
  } catch { /* handled */ }
}

function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(fetchProviders, 300)
}

function openAdd() {
  isEdit.value = false
  originalName.value = ''
  form.value = { name: '', type: 'openai', api_key: '', base_url: '', enabled: true, models: [] }
  fetchedModels.value = []
  selectedModels.value = new Set()
  currentView.value = 'edit'
}

function openEdit(p: PlatformPreset) {
  isEdit.value = true
  originalName.value = p.name
  form.value = { ...p }
  selectedModels.value = new Set(p.models)
  fetchedModels.value = []
  currentView.value = 'edit'
}

function goBack() {
  currentView.value = 'list'
  fetchProviders()
}

function openModelsDialog(models: string[]) {
  viewModelsList.value = models
  viewModelsDialog.value = true
}
function closeModelsDialog() { viewModelsDialog.value = false }

onMounted(fetchProviders)
</script>

<template>
  <div class="page">
    <template v-if="currentView === 'list'">
      <div class="list-view">
        <div class="page-header">
          <h2>平台管理</h2>
        </div>
        <el-card shadow="never" class="list-card">
          <div class="table-toolbar">
            <div class="toolbar-left">
              <el-input v-model="keyword" placeholder="搜索平台名称/模型..." clearable class="search-input">
                <template #prefix><el-icon><Search /></el-icon></template>
              </el-input>
            </div>
            <div class="toolbar-right">
              <el-select v-model="statusFilter" placeholder="状态" style="width:120px;margin-right:8px">
                <el-option label="全部" value="all" />
                <el-option label="启用" value="enabled" />
                <el-option label="禁用" value="disabled" />
              </el-select>
              <el-button :icon="Refresh" circle @click="fetchProviders" />
              <el-button type="primary" :icon="Plus" @click="openAdd">添加平台</el-button>
            </div>
          </div>
          <el-table :data="filteredProviders" v-loading="loading" class="beauty-table" style="width:100%" height="calc(100vh - 120px)">
            <el-table-column prop="name" label="平台名称" width="160">
              <template #default="{ row }"><span class="name-cell">{{ row.name }}</span></template>
            </el-table-column>
            <el-table-column prop="type" label="接口类型" width="120" align="center">
              <template #default="{ row }"><el-tag size="small" effect="plain">{{ providerTypes.find(t => t.value === row.type)?.label || row.type }}</el-tag></template>
            </el-table-column>
            <el-table-column label="授权信息" min-width="200">
              <template #default="{ row }"><span v-if="row.api_key">{{ row.api_key.slice(0,8) }}********</span><span v-else style="color:#999">未设置</span></template>
            </el-table-column>
            <el-table-column prop="base_url" label="接口地址" min-width="200" show-overflow-tooltip />
            <el-table-column label="模型" min-width="220">
              <template #default="{ row }">
                <div class="model-line">
                    <template v-if="row.models && row.models.length">
                      <el-tag v-for="(m, idx) in row.models.slice(0, 2)" :key="m" size="small" effect="plain" style="margin:1px 2px">{{ m }}</el-tag>
                      <el-tag v-if="row.models.length > 2" class="more-count" size="small" effect="plain" style="margin:1px 2px; cursor:pointer; border-radius:50%; width:20px; height:20px; line-height:20px; text-align:center;" @click="openModelsDialog(row.models)">{{ row.models.length - 2 }}</el-tag>
                    </template>
                  <span v-else style="color:#999">未配置</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="启用" width="80" align="center">
              <template #default="{ row }"><el-switch v-model="row.enabled" size="small" active-color="#67c23a" inactive-color="#c0c4cc" @change="handleToggleProvider(row)" /></template>
            </el-table-column>
            <el-table-column label="操作" width="200" fixed="right" align="center">
              <template #default="{ row }">
                <div class="action-group" style="display:inline-flex;margin-left:4px">
                  <el-button size="small" text :icon="EditPen" @click="openEdit(row)" />
                  <el-popconfirm title="确认删除?" @confirm="handleDelete(row.name)"><template #reference><el-button size="small" text type="danger" :icon="Delete" /></template></el-popconfirm>
                </div>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </div>
    </template>
    <template v-else>
      <div class="edit-page">
        <div class="edit-topbar">
          <el-button text @click="goBack"><el-icon><ArrowLeft /></el-icon> 返回</el-button>
          <div class="topbar-title">{{ isEdit ? '编辑平台' : '添加平台' }}</div>
          <div class="topbar-actions">
            <el-button :loading="testing" @click="testConnection">测试</el-button>
            <el-button @click="goBack">取消</el-button>
            <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
          </div>
        </div>
        <div class="edit-body">
          <el-card shadow="never" class="edit-card">
            <div class="edit-split">
              <div class="edit-left">
                <div class="section-title">基础参数</div>
                <el-form :model="form" label-width="100px">
                  <el-form-item label="平台名称"><el-input v-model="form.name" placeholder="如 Azure-1, Anthropic-Pro" :disabled="isEdit" /></el-form-item>
                  <el-form-item label="接口类型"><el-select v-model="form.type" style="width:100%"><el-option v-for="t in providerTypes" :key="t.value" :label="t.label" :value="t.value" /></el-select></el-form-item>
                  <el-form-item label="接口地址"><el-input v-model="form.base_url" placeholder="https://api.openai.com/v1" /></el-form-item>
                  <el-form-item label="授权信息"><el-input v-model="form.api_key" type="password" show-password placeholder="sk-..." /></el-form-item>
                  <el-form-item label="启用"><el-switch v-model="form.enabled" /></el-form-item>
                </el-form>
                <div class="section-title" style="margin-top:24px">已选模型</div>
                <div class="model-tags" v-if="selectedModels.size > 0"><el-tag v-for="m in [...selectedModels]" :key="m" closable size="small" @close="removeSelectedModel(m)">{{ m }}</el-tag></div>
                <span v-else style="color:#999;font-size:13px">暂未选择模型</span>
              </div>
              <div class="edit-right">
                <div class="section-title">模型同步</div>
                <div class="right-body">
                  <div class="sync-card sync-manual">
                    <div class="sync-header"><span class="sync-label">手动添加</span></div>
                    <div class="manual-row"><el-input v-model="newManualModel" ref="manualInput" placeholder="输入模型名称" @keyup.enter="addManualModel" /><el-button @click="addManualModel">添加</el-button></div>
                  </div>
                  <div class="sync-card sync-upstream">
                    <div class="sync-header"><span class="sync-label">同步上游</span><div class="sync-actions"><el-button v-if="fetchedModels.length" size="small" @click="selectAllFetched">全选</el-button><el-button size="small" :loading="fetchingModels" @click="fetchProviderModels">{{ fetchedModels.length ? '重新拉取' : '拉取模型' }}</el-button></div></div>
                    <div class="upstream-body" v-if="fetchedModels.length > 0">
                      <div v-for="g in groupedModels" :key="g.owner" class="model-group">
                        <div class="group-header"><span class="collapse-btn" @click="toggleGroup(g.owner)"><span :class="['arrow', collapsedGroups.has(g.owner) ? 'collapsed' : '']">></span></span><el-checkbox :indeterminate="g.models.some(m => selectedModels.has(m.id)) && !g.models.every(m => selectedModels.has(m.id))" :model-value="g.models.every(m => selectedModels.has(m.id))" @click.stop @change="() => selectAll(g.owner)" /><span class="group-owner" @click="selectAll(g.owner)">{{ g.owner }}</span><span class="group-count">{{ g.models.length }}</span></div>
                        <div class="group-items" v-show="!collapsedGroups.has(g.owner)">
                          <div v-for="m in g.models" :key="m.id" class="model-item" @click="toggleModel(m.id)"><el-checkbox :model-value="selectedModels.has(m.id)" @click.stop @change="() => toggleModel(m.id)" /><span class="model-name">{{ m.id }}</span></div>
                        </div>
                      </div>
                    </div>
                    <div class="upstream-empty" v-else><span>点击拉取上游模型列表</span></div>
                  </div>
                </div>
              </div>
            </div>
          </el-card>
        </div>
      </div>
    </template>
    <el-dialog v-model="viewModelsDialog" title="模型列表" width="600px">
      <div v-for="m in viewModelsList" :key="m" style="margin-bottom:4px">{{ m }}</div>
      <template #footer><el-button @click="closeModelsDialog">关闭</el-button></template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page { height: 100%; display: flex; flex-direction: column; }
.list-view { flex: 1; display: flex; flex-direction: column; min-height: 0; }
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; flex-shrink: 0; }
.page-header h2 { margin: 0; font-size: 18px; font-weight: 600; color: #1d2129; }
.list-card { flex: 1; display: flex; flex-direction: column; border-radius: 12px; border: 1px solid #e5e6eb; min-height: 0; }
.list-card :deep(.el-card__body) { flex: 1; display: flex; flex-direction: column; min-height: 0; padding: 16px 20px; }
.table-toolbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; gap: 12px; }
.toolbar-left { flex: 1; }
.toolbar-right { display: flex; align-items: center; gap: 8px; }
.search-input { width: 320px; }
.beauty-table { --el-table-border-color: #f0f0f0; flex: 1; min-height: 350px; }
.beauty-table :deep(.el-table__header th) { background: #f7f8fa; color: #4e5969; font-weight: 500; }
.name-cell { color: #1d2129; font-weight: 500; }
.model-line { display: flex; align-items: center; }
.model-text { max-width: 150px; }
.edit-page { display: flex; flex-direction: column; height: calc(100vh - 120px); }
.edit-topbar { display: flex; align-items: center; gap: 12px; padding: 0 0 16px; flex-shrink: 0; }
.topbar-title { flex: 1; font-size: 16px; font-weight: 600; color: #1d2129; }
.topbar-actions { display: flex; align-items: center; gap: 8px; }
.edit-body { flex: 1; display: flex; min-height: 0; }
.edit-card { flex: 1; border-radius: 12px; border: 1px solid #e5e6eb; overflow: hidden; }
.edit-card :deep(.el-card__body) { padding: 20px 24px; height: 100%; }
.edit-split { display: flex; gap: 24px; height: 100%; }
.edit-left { flex: 1; overflow-y: auto; min-width: 0; }
.edit-right { flex: 1; display: flex; flex-direction: column; min-width: 0; min-height: 0; border-left: 1px solid #e5e6eb; padding-left: 24px; }
.section-title { font-size: 15px; font-weight: 600; color: #1d2129; margin-bottom: 16px; display: flex; align-items: center; }
.right-body { flex: 1; display: flex; flex-direction: column; gap: 16px; min-height: 0; }
.sync-card { background: #f7f8fa; border-radius: 8px; padding: 14px 16px; }
.sync-manual { flex-shrink: 0; }
.sync-upstream { flex: 1; display: flex; flex-direction: column; min-height: 0; }
.sync-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 4px; }
.sync-actions { display: flex; align-items: center; gap: 6px; }
.manual-row { display: flex; gap: 8px; align-items: center; }
.manual-row .el-input { flex: 1; }
.model-tags { display: flex; flex-wrap: wrap; gap: 4px; align-items: center; margin-top: 8px; }
.upstream-body { flex: 1; overflow-y: auto; margin-top: 10px; display: flex; flex-direction: column; gap: 8px; min-height: 0; }
.upstream-empty { flex: 1; display: flex; align-items: center; justify-content: center; color: #c9cdd4; font-size: 13px; min-height: 80px; }
.model-group { border-bottom: 1px solid #edeff2; }
.model-group:last-child { border-bottom: none; }
.collapse-btn { width: 16px; text-align: center; cursor: pointer; flex-shrink: 0; user-select: none; opacity: 0.5; }
.collapse-btn:hover { opacity: 0.8; }
.arrow { display: inline-block; transition: transform 0.2s; font-size: 12px; font-family: monospace; }
.arrow.collapsed { transform: rotate(0deg); }
.arrow:not(.collapsed) { transform: rotate(90deg); }
.group-header { display: flex; align-items: center; gap: 8px; padding: 8px 0; }
.group-header:hover .group-owner { color: #165dff; }
.group-owner { font-size: 13px; font-weight: 600; color: #1d2129; flex: 1; cursor: pointer; }
.group-count { font-size: 11px; color: #86909c; background: #f0f0f0; padding: 0 6px; border-radius: 8px; line-height: 18px; }
.group-items { padding: 0 0 6px 24px; }
.model-item { display: flex; align-items: center; gap: 8px; padding: 5px 8px; border-radius: 4px; cursor: pointer; font-size: 13px; color: #4e5969; }
.model-item:hover { background: #f0f2f5; }
.model-name { font-family: 'SF Mono', 'Cascadia Code', 'Fira Code', monospace; font-size: 12px; color: #1d2129; }
</style>