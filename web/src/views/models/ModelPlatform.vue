<script setup lang="ts">
import { ref, onMounted, nextTick, computed } from 'vue'
import { ElMessage, ElMessageBox, ElPopconfirm } from 'element-plus'
import { Plus, Delete, EditPen, Search, Refresh } from '@element-plus/icons-vue'
import request from '@/api/request'

interface PlatformPreset {
  name: string
  type: string
  api_key: string
  base_url: string
  enabled: boolean
  models: string[]
}

const providerTypes = [
  { label: 'OpenAI', value: 'openai' },
  { label: 'Azure OpenAI', value: 'azure' },
  { label: 'Anthropic', value: 'anthropic' },
  { label: 'Google', value: 'google' },
  { label: '自定义', value: 'custom' },
]

const providers = ref<PlatformPreset[]>([])
const loading = ref(false)
const providerDialog = ref(false)
const editingProvider = ref(false)
const keyword = ref('')
let searchTimer: ReturnType<typeof setTimeout>
const providerForm = ref<PlatformPreset>({
  name: '', type: 'openai', api_key: '', base_url: '',
  enabled: true,
  models: ['gpt-4o'],
})

const filteredProviders = computed(() => {
  if (!keyword.value) return providers.value
  const kw = keyword.value.toLowerCase()
  return providers.value.filter(p => p.name.toLowerCase().includes(kw) || p.type.toLowerCase().includes(kw))
})

const newModel = ref('')
const modelInput = ref<any>(null)

function addModel() {
  const m = newModel.value.trim()
  if (m && !providerForm.value.models.includes(m)) {
    providerForm.value.models.push(m)
  }
  newModel.value = ''
  nextTick(() => modelInput.value?.focus())
}

function removeModel(m: string) {
  providerForm.value.models = providerForm.value.models.filter(x => x !== m)
}

async function fetchProviders() {
  loading.value = true
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
  finally { loading.value = false }
}

async function handleSaveActive(p: PlatformPreset) {
  const map: Record<string, string> = {
    provider: 'llm_provider', api_key: 'llm_api_key', base_url: 'llm_base_url',
  }
  try {
    await Promise.all(
      Object.entries(map).map(([field, key]) =>
        request.post('/system-config/upsert', {
          category: 'llm', key, value: String((p as any)[field] || ''),
        })
      )
    )
    ElMessage.success(`已激活 "${p.name}"`)
  } catch { /* handled */ }
}

async function handleToggleProvider(p: PlatformPreset) {
  try {
    const value = JSON.stringify(p)
    await request.post('/system-config/upsert', {
      category: 'llm', key: `preset:${p.name}`, value,
    })
  } catch { p.enabled = !p.enabled }
}

function openAddProvider() {
  editingProvider.value = false
  providerForm.value = { name: '', type: 'openai', api_key: '', base_url: '',
    enabled: true, models: ['gpt-4o'] }
  providerDialog.value = true
}

function openEditProvider(p: PlatformPreset) {
  editingProvider.value = true
  providerForm.value = { ...p }
  providerDialog.value = true
}

async function handleSaveProvider() {
  if (!providerForm.value.name) { ElMessage.warning('请输入平台名称'); return }
  try {
    const value = JSON.stringify(providerForm.value)
    await request.post('/system-config/upsert', {
      category: 'llm', key: `preset:${providerForm.value.name}`, value,
    })
    ElMessage.success('平台配置已保存')
    providerDialog.value = false
    fetchProviders()
  } catch { /* handled */ }
}

async function handleDeleteProvider(name: string) {
  try {
    const res: any = await request.get('/system-config', { params: { page: 1, page_size: 200 } })
    const list: any[] = res.data?.data || []
    const cfg = list.find((c: any) => c.key === `preset:${name}`)
    if (cfg) await request.delete(`/system-config/${cfg.id}`)
    ElMessage.success('已删除')
    fetchProviders()
  } catch { /* cancelled or error */ }
}

function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(fetchProviders, 300)
}

onMounted(fetchProviders)
</script>

<template>
  <div class="page">
    <div class="list-view">
      <div class="page-header">
        <h2>平台管理</h2>
      </div>

      <el-card shadow="never" class="list-card">
        <div class="table-toolbar">
          <div class="toolbar-left">
            <el-input v-model="keyword" placeholder="搜索平台名称..." clearable @input="onSearch" class="search-input">
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
          </div>
          <div class="toolbar-right">
            <el-button :icon="Refresh" circle @click="fetchProviders" />
            <el-button type="primary" :icon="Plus" @click="openAddProvider">添加平台</el-button>
          </div>
        </div>

        <el-table :data="keyword ? filteredProviders : providers" v-loading="loading" class="beauty-table" style="width:100%" height="calc(100vh - 100px)">
          <el-table-column prop="name" label="平台名称" width="160">
            <template #default="{ row }"><span class="name-cell">{{ row.name }}</span></template>
          </el-table-column>
          <el-table-column prop="type" label="平台类型" width="120" align="center">
            <template #default="{ row }">
              <el-tag size="small" effect="plain">{{ providerTypes.find(t => t.value === row.type)?.label || row.type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="授权信息" min-width="200">
            <template #default="{ row }">
              <span v-if="row.api_key">{{ row.api_key.slice(0, 8) }}********</span>
              <span v-else style="color:#999">未设置</span>
            </template>
          </el-table-column>
          <el-table-column prop="base_url" label="接口地址" min-width="200" show-overflow-tooltip />
          <el-table-column label="模型" min-width="220">
            <template #default="{ row }">
              <el-tag v-for="m in (row.models || [])" :key="m" size="small" effect="plain" style="margin:1px 2px">{{ m }}</el-tag>
              <span v-if="!row.models?.length" style="color:#999">未配置</span>
            </template>
          </el-table-column>
          <el-table-column label="启用" width="80" align="center">
            <template #default="{ row }">
              <el-switch v-model="row.enabled" size="small" active-color="#67c23a" inactive-color="#c0c4cc" @change="handleToggleProvider(row)" />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right" align="center">
            <template #default="{ row }">
              <el-button size="small" type="success" @click="handleSaveActive(row)">激活</el-button>
              <div class="action-group" style="display:inline-flex;vertical-align:middle;margin-left:4px">
                <el-button size="small" text :icon="EditPen" @click="openEditProvider(row)" />
                <el-popconfirm title="确认删除?" @confirm="handleDeleteProvider(row.name)">
                  <template #reference><el-button size="small" text type="danger" :icon="Delete" /></template>
                </el-popconfirm>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <el-dialog v-model="providerDialog" :title="editingProvider ? '编辑平台' : '添加平台'" width="550px">
      <el-form :model="providerForm" label-width="100px">
        <el-form-item label="平台名称">
          <el-input v-model="providerForm.name" placeholder="如 Azure-1, Anthropic-Pro" :disabled="editingProvider" />
        </el-form-item>
        <el-form-item label="平台类型">
          <el-select v-model="providerForm.type" style="width:100%">
            <el-option v-for="t in providerTypes" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="授权信息">
          <el-input v-model="providerForm.api_key" type="password" show-password placeholder="sk-..." />
        </el-form-item>
        <el-form-item label="接口地址">
          <el-input v-model="providerForm.base_url" placeholder="https://api.openai.com/v1" />
        </el-form-item>
        <el-form-item label="支持模型">
          <div class="model-tags">
            <el-tag v-for="m in providerForm.models" :key="m" closable @close="removeModel(m)">
              {{ m }}
            </el-tag>
            <el-input
              ref="modelInput"
              v-model="newModel"
              class="model-input"
              placeholder="输入模型名回车添加"
              @keyup.enter="addModel"
              @blur="addModel"
            />
          </div>
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="providerForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="providerDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveProvider">保存</el-button>
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
.model-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}
.model-input {
  width: 180px;
  flex-shrink: 0;
}
</style>
