<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

interface ModelItem {
  name: string
  provider: string
  enabled: boolean
  max_tokens: number
  temperature: number
}

const models = ref<ModelItem[]>([])
const providers = ref<{ name: string }[]>([])
const loading = ref(false)
const modelDialog = ref(false)
const modelForm = ref<ModelItem>({ name: '', provider: 'openai', enabled: true, max_tokens: 4096, temperature: 0.7 })
const savingModel = ref(false)

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
          return { name: c.key.replace('model:', ''), ...v }
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
    temperature: 0.7,
  }
  modelDialog.value = true
}

function openEditModel(item: ModelItem) {
  modelForm.value = { ...item }
  modelDialog.value = true
}

async function handleSaveModel() {
  if (!modelForm.value.name) { ElMessage.warning('请输入模型名称'); return }
  savingModel.value = true
  try {
    const value = JSON.stringify({
      provider: modelForm.value.provider,
      enabled: modelForm.value.enabled,
      max_tokens: modelForm.value.max_tokens,
      temperature: modelForm.value.temperature,
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
    await ElMessageBox.confirm(`确认删除模型 "${name}"？`, '提示')
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
      request.post('/system-config/upsert', { category: 'llm', key: 'llm_temperature', value: String(item.temperature) }),
    ])
    ElMessage.success(`已激活模型 "${item.name}"`)
  } catch { /* handled */ }
}

async function handleToggleModel(item: ModelItem) {
  try {
    const value = JSON.stringify({
      provider: item.provider,
      enabled: item.enabled,
      max_tokens: item.max_tokens,
      temperature: item.temperature,
    })
    await request.post('/system-config/upsert', {
      category: 'llm', key: `model:${item.name}`, value,
    })
  } catch { item.enabled = !item.enabled }
}

onMounted(() => {
  fetchModels()
  fetchProviders()
})
</script>

<template>
  <div>
    <div class="page-header">
      <h2>模型管理</h2>
    </div>

    <el-card>
      <div style="margin-bottom:16px">
        <el-button type="primary" @click="openAddModel">添加模型</el-button>
      </div>

      <el-table :data="models" v-loading="loading" stripe border style="width:100%">
        <el-table-column prop="name" label="模型名称" min-width="200" />
        <el-table-column prop="provider" label="厂商名称" width="160" />
        <el-table-column prop="max_tokens" label="Max Tokens" width="120" />
        <el-table-column prop="temperature" label="Temperature" width="120">
          <template #default="{ row }">
            {{ row.temperature ?? 0.7 }}
          </template>
        </el-table-column>
        <el-table-column label="启用" width="100">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="handleToggleModel(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="210" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="success" @click="handleActivateModel(row)">激活</el-button>
            <el-button size="small" @click="openEditModel(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDeleteModel(row.name)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="modelDialog" :title="modelForm.name && !modelForm.name.includes('请输入') ? '编辑模型' : '添加模型'" width="450px">
      <el-form :model="modelForm" label-width="100px">
        <el-form-item label="模型名称">
          <el-input v-model="modelForm.name" placeholder="gpt-4o-mini" />
        </el-form-item>
        <el-form-item label="厂商名称">
          <el-select v-model="modelForm.provider" style="width:100%">
            <el-option v-for="p in providers" :key="p.name" :label="p.name" :value="p.name" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="Max Tokens">
          <el-input-number v-model="modelForm.max_tokens" :min="256" :max="128000" :step="1024" />
        </el-form-item>
        <el-form-item label="Temperature">
          <el-slider v-model="modelForm.temperature" :min="0" :max="2" :step="0.1" />
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
</style>
