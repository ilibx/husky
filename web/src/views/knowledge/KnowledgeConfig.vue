<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox, ElPopconfirm } from 'element-plus'
import { Search, Plus, EditPen, Delete, Refresh } from '@element-plus/icons-vue'
import request from '@/api/request'

type StoreKind = 'vector' | 'graph'
type StoreProvider = 'postgresql' | 'neo4j' | 'milvus' | 'elasticsearch' | 'other'

interface KnowledgeStoreSource {
  id?: number
  name: string
  kind: StoreKind
  provider: StoreProvider
  host: string
  port: string
  user: string
  password: string
  database: string
  sslmode: string
  enabled: boolean
  is_default: boolean
  description: string
}

const sources = ref<KnowledgeStoreSource[]>([])
const loading = ref(false)
const saving = ref(false)
const keyword = ref('')
let searchTimer: ReturnType<typeof setTimeout>
function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => fetchSources(), 300)
}
const sourceDialog = ref(false)
const editingSource = ref(false)
const editingOriginalName = ref('')

const emptySource = (): KnowledgeStoreSource => ({
  name: '',
  kind: 'vector',
  provider: 'postgresql',
  host: 'localhost',
  port: '5432',
  user: 'postgres',
  password: '',
  database: 'husky',
  sslmode: 'disable',
  enabled: true,
  is_default: false,
  description: '',
})

const sourceForm = reactive<KnowledgeStoreSource>(emptySource())

const providerOptions = [
  { label: 'PGVector', value: 'postgresql', kind: 'vector' },
  { label: 'Neo4j', value: 'neo4j', kind: 'graph' },
  { label: 'Milvus', value: 'milvus', kind: 'vector' },
  { label: 'Elasticsearch', value: 'elasticsearch', kind: 'vector' },
  { label: '其他', value: 'other', kind: 'vector' },
]

const legacyVectorKeyMap: Record<string, string> = {
  provider: 'vector_provider', host: 'vector_host', port: 'vector_port',
  user: 'vector_user', password: 'vector_password', database: 'vector_database', sslmode: 'vector_sslmode',
}

function resetForm(source?: KnowledgeStoreSource) {
  Object.assign(sourceForm, emptySource(), source || {})
}

function kindLabel(kind: StoreKind) {
  return kind === 'graph' ? '知识图谱库' : '向量库'
}

function providerLabel(provider: string) {
  return providerOptions.find(item => item.value === provider)?.label || provider
}

function onProviderChange(provider: StoreProvider) {
  if (provider === 'neo4j') {
    sourceForm.kind = 'graph'
    if (!sourceForm.port || sourceForm.port === '5432') sourceForm.port = '7687'
    if (!sourceForm.user || sourceForm.user === 'postgres') sourceForm.user = 'neo4j'
    if (!sourceForm.database || sourceForm.database === 'husky') sourceForm.database = 'neo4j'
    return
  }
  sourceForm.kind = 'vector'
  if (provider === 'postgresql' && (!sourceForm.port || sourceForm.port === '7687')) sourceForm.port = '5432'
}

async function fetchSources() {
  loading.value = true
  try {
    const params: any = { page: 1, page_size: 500 }
    if (keyword.value) params.keyword = keyword.value
    const res: any = await request.get('/system-config', { params })
    const list: any[] = res.data?.data || []
    sources.value = list
      .filter((c: any) => c.category === 'vector' && c.key.startsWith('source:'))
      .map((c: any) => {
        try {
          return { id: c.id, ...emptySource(), ...JSON.parse(c.value), name: JSON.parse(c.value).name || c.key.replace('source:', '') }
        } catch {
          return null
        }
      })
      .filter(Boolean)
  } catch {
    sources.value = []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  editingSource.value = false
  editingOriginalName.value = ''
  resetForm()
  sourceDialog.value = true
}

function openEdit(row: KnowledgeStoreSource) {
  editingSource.value = true
  editingOriginalName.value = row.name
  resetForm(row)
  sourceDialog.value = true
}

async function syncDefaultVector(source: KnowledgeStoreSource) {
  if (!source.is_default || source.kind !== 'vector') return
  await Promise.all(Object.entries(legacyVectorKeyMap).map(([field, key]) =>
    request.post('/system-config/upsert', {
      category: 'vector',
      key,
      value: String((source as any)[field] || ''),
    }),
  ))
}

async function saveSource() {
  if (!sourceForm.name.trim()) {
    ElMessage.error('请输入连接名称')
    return
  }
  saving.value = true
  try {
    const payload = { ...sourceForm, name: sourceForm.name.trim() }
    if (payload.is_default) {
      sources.value = sources.value.map(item => ({ ...item, is_default: item.name === editingOriginalName.value }))
      await Promise.all(sources.value
        .filter(item => item.name !== editingOriginalName.value && item.is_default)
        .map(item => request.post('/system-config/upsert', {
          category: 'vector', key: `source:${item.name}`, value: JSON.stringify({ ...item, is_default: false }),
        })))
    }
    if (editingSource.value && editingOriginalName.value && editingOriginalName.value !== payload.name) {
      const old = sources.value.find(item => item.name === editingOriginalName.value)
      if (old?.id) await request.delete(`/system-config/${old.id}`)
    }
    await request.post('/system-config/upsert', {
      category: 'vector',
      key: `source:${payload.name}`,
      value: JSON.stringify(payload),
      enabled: payload.enabled,
    })
    await syncDefaultVector(payload)
    ElMessage.success('知识库连接已保存')
    sourceDialog.value = false
    fetchSources()
  } catch {
    // handled by interceptor
  } finally {
    saving.value = false
  }
}

async function removeSource(row: KnowledgeStoreSource) {
  if (!row.id) return
  try {
    await ElMessageBox.confirm(`确认删除连接「${row.name}」？`, '提示')
    await request.delete(`/system-config/${row.id}`)
    ElMessage.success('删除成功')
    fetchSources()
  } catch {
    // cancelled or handled
  }
}

onMounted(fetchSources)
</script>

<template>
  <div class="page">
    <div class="list-view">
      <div class="page-header">
        <h2>知识库管理</h2>
      </div>

      <el-card shadow="never" class="list-card">
        <div class="table-toolbar">
          <div class="toolbar-left">
            <el-input v-model="keyword" placeholder="搜索连接名称..." clearable @input="onSearch" class="search-input">
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
          </div>
          <div class="toolbar-right">
            <el-button :icon="Refresh" circle @click="fetchSources" />
            <el-button type="primary" :icon="Plus" @click="openAdd">新增连接</el-button>
          </div>
        </div>
        <el-table :data="sources" v-loading="loading" class="beauty-table" row-key="name" style="width:100%" height="calc(100vh - 100px)">
          <el-table-column label="连接名称" min-width="180">
            <template #default="{ row }">
              <div class="store-name">
                <span class="name-cell">{{ row.name }}</span>
                <el-tag v-if="row.is_default" size="small" type="success">默认</el-tag>
              </div>
              <div class="store-desc">{{ row.description || '-' }}</div>
            </template>
          </el-table-column>
          <el-table-column label="类型" width="100" align="center">
            <template #default="{ row }">
              <el-tag :type="row.kind === 'graph' ? 'warning' : 'primary'" size="small" effect="plain">{{ kindLabel(row.kind) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="引擎" width="120">
            <template #default="{ row }">{{ providerLabel(row.provider) }}</template>
          </el-table-column>
          <el-table-column label="地址" min-width="220" show-overflow-tooltip>
            <template #default="{ row }">{{ row.host }}:{{ row.port }}/{{ row.database }}</template>
          </el-table-column>
          <el-table-column label="状态" width="72" align="center">
            <template #default="{ row }">
              <el-tag :type="row.enabled ? 'success' : 'info'" size="small" effect="plain">{{ row.enabled ? '启用' : '停用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120" fixed="right" align="center">
            <template #default="{ row }">
              <div class="action-group">
                <el-button size="small" text :icon="EditPen" @click="openEdit(row)" />
                <el-popconfirm title="确认删除?" @confirm="removeSource(row)">
                  <template #reference><el-button size="small" text type="danger" :icon="Delete" /></template>
                </el-popconfirm>
              </div>
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-if="!loading && sources.length === 0" description="暂无外置知识库连接" />
      </el-card>
    </div>

    <el-dialog v-model="sourceDialog" :title="editingSource ? '编辑知识库连接' : '新增知识库连接'" width="640px">
        <el-form :model="sourceForm" label-width="120px">
          <el-form-item label="连接名称" required>
            <el-input v-model="sourceForm.name" placeholder="如 主知识库、图谱库" />
          </el-form-item>
          <el-form-item label="库类型">
            <el-radio-group v-model="sourceForm.kind">
              <el-radio-button label="vector">向量库</el-radio-button>
              <el-radio-button label="graph">知识图谱库</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="引擎">
            <el-select v-model="sourceForm.provider" style="width:100%" @change="onProviderChange">
              <el-option v-for="item in providerOptions" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
          </el-form-item>
          <el-form-item label="主机"><el-input v-model="sourceForm.host" placeholder="localhost" /></el-form-item>
          <el-form-item label="端口"><el-input v-model="sourceForm.port" placeholder="5432 / 7687" /></el-form-item>
          <el-form-item label="用户名"><el-input v-model="sourceForm.user" /></el-form-item>
          <el-form-item label="密码"><el-input v-model="sourceForm.password" type="password" show-password /></el-form-item>
          <el-form-item label="数据库/库名"><el-input v-model="sourceForm.database" /></el-form-item>
          <el-form-item label="SSL 模式" v-if="sourceForm.provider === 'postgresql'">
            <el-select v-model="sourceForm.sslmode" style="width:100%">
              <el-option label="disable" value="disable" />
              <el-option label="require" value="require" />
              <el-option label="verify-ca" value="verify-ca" />
              <el-option label="verify-full" value="verify-full" />
            </el-select>
          </el-form-item>
          <el-form-item label="描述"><el-input v-model="sourceForm.description" type="textarea" :rows="3" /></el-form-item>
          <el-form-item label="状态">
            <el-switch v-model="sourceForm.enabled" active-text="启用" inactive-text="停用" />
          </el-form-item>
          <el-form-item label="默认检索库">
            <el-switch v-model="sourceForm.is_default" />
            <div class="form-help">默认向量库会同步到运行时检索配置；知识图谱库作为图谱连接配置保存。</div>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="sourceDialog = false">取消</el-button>
          <el-button type="primary" :loading="saving" @click="saveSource">保存</el-button>
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
.store-name {
  display: flex;
  align-items: center;
  gap: 8px;
}
.name-cell {
  color: #1d2129;
  font-weight: 500;
}
.store-desc {
  margin-top: 4px;
  color: #86909c;
  font-size: 12px;
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
.form-help {
  margin-left: 12px;
  color: #909399;
  font-size: 12px;
}
</style>
