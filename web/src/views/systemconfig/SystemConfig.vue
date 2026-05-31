<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

// ---- 活跃 LLM 配置 (后端实际使用的) ----
interface ActiveLLM {
  provider: string
  api_key: string
  base_url: string
  model: string
  max_tokens: number
  temperature: number
}

// ---- 大模型厂商预设 ----
interface ProviderPreset {
  name: string
  type: string
  api_key: string
  base_url: string
  model: string
  max_tokens: number
  temperature: number
}

// ---- 模型条目 ----
interface ModelItem {
  name: string
  provider: string
  enabled: boolean
  max_tokens: number
}



// ============ 数据加载 ============

async function fetchActiveLLM() {
  loadingActive.value = true
  try {
    const res: any = await request.get('/system-config/llm')
    if (res.data) Object.assign(activeLLM, res.data)
  } catch { /* defaults */ } finally { loadingActive.value = false }
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
  } catch {
    providers.value = []
  }
}

async function fetchModels() {
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
  } catch {
    models.value = []
  }
}

async function fetchVectorConfig() {
  loadingVector.value = true
  try {
    const res: any = await request.get('/system-config/vector')
    if (res.data) Object.assign(vectorConfig, res.data)
  } catch { /* defaults */ } finally { loadingVector.value = false }
}

// ============ 活跃 LLM 配置保存 ============

const llmKeyMap: Record<keyof ActiveLLM, string> = {
  provider: 'llm_provider', api_key: 'llm_api_key', base_url: 'llm_base_url',
  model: 'llm_model', max_tokens: 'llm_max_tokens', temperature: 'llm_temperature',
}

async function handleSaveActive() {
  savingActive.value = true
  try {
    const promises = Object.entries(llmKeyMap).map(([field, key]) =>
      request.post('/system-config/upsert', {
        category: 'llm', key, value: String((activeLLM as any)[field]),
      })
    )
    await Promise.all(promises)
    ElMessage.success('当前 LLM 配置已保存')
  } catch { /* handled */ } finally { savingActive.value = false }
}

// ============ 厂商预设 CRUD ============

function openAddProvider() {
  editingProvider.value = false
  Object.assign(providerForm, { name: '', type: 'openai', api_key: '', base_url: '',
    model: 'gpt-4o', max_tokens: 4096, temperature: 0.7 })
  providerDialog.value = true
}

function openEditProvider(p: ProviderPreset) {
  editingProvider.value = true
  Object.assign(providerForm, p)
  providerDialog.value = true
}

async function handleSaveProvider() {
  if (!providerForm.name) { ElMessage.warning('请输入厂商名称'); return }
  try {
    const value = JSON.stringify(providerForm)
    await request.post('/system-config/upsert', {
      category: 'llm', key: `preset:${providerForm.name}`, value,
    })
    ElMessage.success('厂商配置已保存')
    providerDialog.value = false
    fetchProviders()
  } catch { /* handled */ }
}

async function handleDeleteProvider(name: string) {
  try {
    await ElMessageBox.confirm(`确认删除厂商 "${name}"？`, '提示')
    // Find config id from list
    const res: any = await request.get('/system-config', { params: { page: 1, page_size: 200 } })
    const list: any[] = res.data?.data || []
    const cfg = list.find((c: any) => c.key === `preset:${name}`)
    if (cfg) await request.delete(`/system-config/${cfg.id}`)
    ElMessage.success('已删除')
    fetchProviders()
  } catch { /* cancelled or error */ }
}

async function handleActivateProvider(p: ProviderPreset) {
  Object.assign(activeLLM, {
    provider: p.type, api_key: p.api_key, base_url: p.base_url,
    model: p.model, max_tokens: p.max_tokens, temperature: p.temperature,
  })
  await handleSaveActive()
  ElMessage.success(`已激活 "${p.name}" 配置`)
}

// ============ 模型 CRUD ============

function openAddModel() {
  modelForm.name = ''
  modelForm.provider = providers.value[0]?.name || 'openai'
  modelForm.enabled = true
  modelForm.max_tokens = 4096
  modelDialog.value = true
}

async function handleSaveModel() {
  if (!modelForm.name) { ElMessage.warning('请输入模型名称'); return }
  savingModel.value = true
  try {
    const value = JSON.stringify({ provider: modelForm.provider, enabled: modelForm.enabled, max_tokens: modelForm.max_tokens })
    await request.post('/system-config/upsert', {
      category: 'llm', key: `model:${modelForm.name}`, value,
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

async function handleToggleModel(item: ModelItem) {
  try {
    const value = JSON.stringify({ provider: item.provider, enabled: item.enabled, max_tokens: item.max_tokens })
    await request.post('/system-config/upsert', {
      category: 'llm', key: `model:${item.name}`, value,
    })
  } catch { item.enabled = !item.enabled }
}

// ============ 向量库 ============

const vectorKeyMap: Record<keyof VectorDBConfig, string> = {
  provider: 'vector_provider', host: 'vector_host', port: 'vector_port',
  user: 'vector_user', password: 'vector_password', database: 'vector_database', sslmode: 'vector_sslmode',
}

async function handleSaveVector() {
  savingVector.value = true
  try {
    const promises = Object.entries(vectorKeyMap).map(([field, key]) =>
      request.post('/system-config/upsert', {
        category: 'vector', key, value: String((vectorConfig as any)[field]),
      })
    )
    await Promise.all(promises)
    ElMessage.success('向量库配置已保存')
  } catch { /* handled */ } finally { savingVector.value = false }
}

// ============ 初始化 ============
onMounted(() => {
  fetchActiveLLM()
  fetchProviders()
  fetchModels()
  fetchVectorConfig()
})
</script>

<template>
  <div>
    <div class="page-header">
      <h2>大模型管理</h2>
    </div>

    <el-card>
      <el-tabs v-model="activeTab">
        <!-- ===== 大模型厂商 ===== -->
        <el-tab-pane label="大模型厂商" name="providers">
          <div style="margin-bottom:16px">
            <el-button type="primary" @click="openAddProvider">添加厂商</el-button>
          </div>

          <el-table :data="providers" stripe border style="width:100%">
            <el-table-column prop="name" label="厂商名称" width="160" />
            <el-table-column prop="type" label="类型" width="120">
              <template #default="{ row }">
                <el-tag size="small">{{ row.type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="api_key" label="API Key" min-width="200">
              <template #default="{ row }">
                <span v-if="row.api_key">{{ row.api_key.slice(0, 8) }}********</span>
                <span v-else style="color:#999">未设置</span>
              </template>
            </el-table-column>
            <el-table-column prop="base_url" label="Base URL" min-width="200" show-overflow-tooltip />
            <el-table-column prop="model" label="默认模型" width="150" />
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="{ row }">
                <el-button size="small" type="success" @click="handleActivateProvider(row)">激活</el-button>
                <el-button size="small" @click="openEditProvider(row)">编辑</el-button>
                <el-button size="small" type="danger" @click="handleDeleteProvider(row.name)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-divider />
          <div class="config-section" v-loading="loadingActive">
            <h3>当前活跃配置</h3>
            <el-form :model="activeLLM" label-width="120px">
              <el-form-item label="类型">
                <el-select v-model="activeLLM.provider" style="width:300px">
                  <el-option v-for="t in providerTypes" :key="t.value" :label="t.label" :value="t.value" />
                </el-select>
              </el-form-item>
              <el-form-item label="API Key">
                <el-input v-model="activeLLM.api_key" type="password" show-password style="width:400px" placeholder="sk-..." />
              </el-form-item>
              <el-form-item label="Base URL">
                <el-input v-model="activeLLM.base_url" style="width:400px" placeholder="https://api.openai.com/v1" />
              </el-form-item>
              <el-form-item label="模型">
                <el-input v-model="activeLLM.model" style="width:300px" placeholder="gpt-4o" />
              </el-form-item>
              <el-form-item label="Max Tokens">
                <el-input-number v-model="activeLLM.max_tokens" :min="256" :max="128000" :step="1024" />
              </el-form-item>
              <el-form-item label="Temperature">
                <el-slider v-model="activeLLM.temperature" :min="0" :max="2" :step="0.1" style="width:300px" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :loading="savingActive" @click="handleSaveActive">保存活跃配置</el-button>
              </el-form-item>
            </el-form>
          </div>
        </el-tab-pane>

        <!-- ===== 可用模型 ===== -->
        <el-tab-pane label="可用模型" name="models">
          <div style="margin-bottom:16px">
            <el-button type="primary" @click="openAddModel">添加模型</el-button>
          </div>

          <el-table :data="models" stripe border style="width:100%">
            <el-table-column prop="name" label="模型名称" min-width="200" />
            <el-table-column prop="provider" label="所属厂商" width="160" />
            <el-table-column prop="max_tokens" label="Max Tokens" width="120" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-switch v-model="row.enabled" @change="handleToggleModel(row)" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120">
              <template #default="{ row }">
                <el-button size="small" type="danger" @click="handleDeleteModel(row.name)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>


      </el-tabs>
    </el-card>

    <!-- 厂商预设对话框 -->
    <el-dialog v-model="providerDialog" :title="editingProvider ? '编辑厂商' : '添加厂商'" width="550px">
      <el-form :model="providerForm" label-width="100px">
        <el-form-item label="厂商名称">
          <el-input v-model="providerForm.name" placeholder="如 Azure-1, Anthropic-Pro" :disabled="editingProvider" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="providerForm.type" style="width:100%">
            <el-option v-for="t in providerTypes" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="API Key">
          <el-input v-model="providerForm.api_key" type="password" show-password placeholder="sk-..." />
        </el-form-item>
        <el-form-item label="Base URL">
          <el-input v-model="providerForm.base_url" placeholder="https://api.openai.com/v1" />
        </el-form-item>
        <el-form-item label="默认模型">
          <el-input v-model="providerForm.model" placeholder="gpt-4o" />
        </el-form-item>
        <el-form-item label="Max Tokens">
          <el-input-number v-model="providerForm.max_tokens" :min="256" :max="128000" :step="1024" />
        </el-form-item>
        <el-form-item label="Temperature">
          <el-slider v-model="providerForm.temperature" :min="0" :max="2" :step="0.1" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="providerDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveProvider">保存</el-button>
      </template>
    </el-dialog>

    <!-- 模型对话框 -->
    <el-dialog v-model="modelDialog" title="添加模型" width="450px">
      <el-form :model="modelForm" label-width="100px">
        <el-form-item label="模型名称">
          <el-input v-model="modelForm.name" placeholder="gpt-4o-mini" />
        </el-form-item>
        <el-form-item label="所属厂商">
          <el-select v-model="modelForm.provider" style="width:100%">
            <el-option v-for="p in providers" :key="p.name" :label="p.name" :value="p.name" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="Max Tokens">
          <el-input-number v-model="modelForm.max_tokens" :min="256" :max="128000" :step="1024" />
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
.config-section {
  max-width: 640px;
  padding-top: 8px;
}
.config-section h3 {
  margin: 0 0 16px 0;
  font-weight: 500;
  color: #333;
}
</style>
