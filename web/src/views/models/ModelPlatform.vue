<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
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
const providerForm = ref<PlatformPreset>({
  name: '', type: 'openai', api_key: '', base_url: '',
  enabled: true,
  models: ['gpt-4o'],
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
    enabled: true,
    models: ['gpt-4o'] }
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
    await ElMessageBox.confirm(`确认删除平台 "${name}"？`, '提示')
    const res: any = await request.get('/system-config', { params: { page: 1, page_size: 200 } })
    const list: any[] = res.data?.data || []
    const cfg = list.find((c: any) => c.key === `preset:${name}`)
    if (cfg) await request.delete(`/system-config/${cfg.id}`)
    ElMessage.success('已删除')
    fetchProviders()
  } catch { /* cancelled or error */ }
}

onMounted(() => {
  fetchProviders()
})
</script>

<template>
  <div>
    <div class="page-header">
      <h2><el-icon><Cpu /></el-icon> 平台管理</h2>
    </div>

    <el-card>
      <div style="margin-bottom:16px">
        <el-button type="primary" @click="openAddProvider">添加平台</el-button>
      </div>

      <el-table :data="providers" v-loading="loading" stripe border style="width:100%">
        <el-table-column prop="name" label="平台名称" width="160" />
        <el-table-column prop="type" label="平台类型" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="api_key" label="授权信息" min-width="200">
          <template #default="{ row }">
            <span v-if="row.api_key">{{ row.api_key.slice(0, 8) }}********</span>
            <span v-else style="color:#999">未设置</span>
          </template>
        </el-table-column>
        <el-table-column prop="base_url" label="接口地址" min-width="200" show-overflow-tooltip />
        <el-table-column label="模型" min-width="200">
          <template #default="{ row }">
            <el-tag v-for="m in (row.models || [])" :key="m" size="small" style="margin:1px 2px">{{ m }}</el-tag>
            <span v-if="!row.models?.length" class="text-muted">未配置</span>
          </template>
        </el-table-column>
        <el-table-column label="启用" width="80">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="handleToggleProvider(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="success" @click="handleSaveActive(row)">激活</el-button>
            <el-button size="small" @click="openEditProvider(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDeleteProvider(row.name)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

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
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.page-header h2 {
  margin: 0;
}
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
