<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
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
    const res: any = await request.get('/system-config', { params: { page: 1, page_size: 500 } })
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
  <div>
    <div class="page-header">
      <div>
        <h2>知识库管理</h2>
        <p>配置多个外置向量库或知识图谱库，例如 PGVector、Neo4j、Milvus 等。</p>
      </div>
      <el-button type="primary" @click="openAdd">新增连接</el-button>
    </div>

    <el-card class="store-card" shadow="never">
      <el-table :data="sources" v-loading="loading" class="store-table" row-key="name" style="width:100%">
        <el-table-column label="连接名称" min-width="180">
          <template #default="{ row }">
            <div class="store-name">
              <strong>{{ row.name }}</strong>
              <el-tag v-if="row.is_default" size="small" type="success">默认</el-tag>
            </div>
            <div class="store-desc">{{ row.description || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="120">
          <template #default="{ row }">
            <el-tag :type="row.kind === 'graph' ? 'warning' : 'primary'">{{ kindLabel(row.kind) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="引擎" width="130">
          <template #default="{ row }">{{ providerLabel(row.provider) }}</template>
        </el-table-column>
        <el-table-column label="地址" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">{{ row.host }}:{{ row.port }}/{{ row.database }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="removeSource(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && sources.length === 0" description="暂无外置知识库连接" />
    </el-card>

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
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}
.page-header h2 {
  margin: 0;
}
.page-header p {
  margin: 8px 0 0;
  color: #8c8c8c;
  font-size: 13px;
}
.store-card {
  border-radius: 12px;
}
.store-table :deep(.el-table__header th) {
  background: #f7f9fc;
  color: #606266;
  font-weight: 600;
}
.store-name {
  display: flex;
  align-items: center;
  gap: 8px;
}
.store-desc {
  margin-top: 4px;
  color: #a8abb2;
  font-size: 12px;
}
.form-help {
  margin-left: 12px;
  color: #909399;
  font-size: 12px;
}
</style>
