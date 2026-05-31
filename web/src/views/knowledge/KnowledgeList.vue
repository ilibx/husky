<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Marked } from 'marked'
import request from '@/api/request'

const marked = new Marked({ gfm: true, breaks: true })
const activeTab = ref('docs')

// ============ 数据源管理 ============
type VectorProvider = 'postgresql' | 'neo4j'

interface VectorSource {
  provider: VectorProvider
  host: string
  port: string
  user: string
  password: string
  database: string
  sslmode: string
  label: string
}

const sources = ref<VectorSource[]>([])
const sourceDialog = ref(false)
const editingSource = ref(false)
const sourceForm = reactive<VectorSource>({
  provider: 'postgresql', host: 'localhost', port: '5432',
  user: 'postgres', password: '', database: 'husky', sslmode: 'disable',
  label: '',
})

const vectorKeyMap: Record<string, string> = {
  provider: 'vector_provider', host: 'vector_host', port: 'vector_port',
  user: 'vector_user', password: 'vector_password', database: 'vector_database', sslmode: 'vector_sslmode',
}

async function fetchSources() {
  try {
    const res: any = await request.get('/system-config/vector')
    if (res.data) {
      sourceForm.provider = res.data.provider || 'postgresql'
      sourceForm.host = res.data.host || 'localhost'
      sourceForm.port = res.data.port || '5432'
      sourceForm.user = res.data.user || 'postgres'
      sourceForm.password = res.data.password || ''
      sourceForm.database = res.data.database || 'husky'
      sourceForm.sslmode = res.data.sslmode || 'disable'
    }
  } catch { /* default */ }
  // Also load named sources from system_configs
  try {
    const res: any = await request.get('/system-config', { params: { page: 1, page_size: 100 } })
    const list: any[] = res.data?.data || []
    sources.value = list
      .filter((c: any) => c.category === 'vector' && c.key.startsWith('source:'))
      .map((c: any) => {
        try { return { label: c.key.replace('source:', ''), ...JSON.parse(c.value) } }
        catch { return null }
      })
      .filter(Boolean)
  } catch { /* ignore */ }
}

async function handleSaveSource() {
  try {
    const promises = Object.entries(vectorKeyMap).map(([field, key]) =>
      request.post('/system-config/upsert', {
        category: 'vector', key, value: String((sourceForm as any)[field]),
      })
    )
    await Promise.all(promises)
    ElMessage.success('向量库配置已保存')
    sourceDialog.value = false
    fetchSources()
  } catch { /* handled */ }
}

// ============ 资料库查询 ============
const queryText = ref('')
const queryLoading = ref(false)
const queryResults = ref<any[]>([])
const queryCategory = ref('')
const queryLimit = ref(5)

async function handleQuery() {
  if (!queryText.value.trim()) return
  queryLoading.value = true
  try {
    const res: any = await request.post('/knowledge/search', {
      query: queryText.value,
      category: queryCategory.value || undefined,
      limit: queryLimit.value,
    })
    queryResults.value = res.data?.data || res.data?.results || []
  } catch {
    queryResults.value = []
  } finally {
    queryLoading.value = false
  }
}

const queryQR = ref('')
const queryQALoading = ref(false)
const queryQAResult = ref('')

async function handleAsk() {
  if (!queryQR.value.trim()) return
  queryQALoading.value = true
  queryQAResult.value = ''
  try {
    const res: any = await request.post('/knowledge/ask', { query: queryQR.value })
    queryQAResult.value = res.data?.answer || res.data?.content || JSON.stringify(res.data)
  } catch {
    queryQAResult.value = '查询失败'
  } finally {
    queryQALoading.value = false
  }
}

// ============ 文档管理 ============
interface Knowledge {
  id: number
  title: string
  content: string
  language: string
  category: string
  source_type: string
  source_url: string
  tags: string
  status: string
  view_count: number
  created_at: string
}

const list = ref<Knowledge[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const keyword = ref('')
const dialogVisible = ref(false)
const previewVisible = ref(false)
const previewItem = ref<Knowledge | null>(null)
const formTitle = ref('新增知识')
const form = ref({ id: 0, title: '', content: '', language: 'zh', category: '', tags: '', status: 'active', source_type: 'manual', source_url: '' })
const uploadLoading = ref(false)

async function fetchData() {
  loading.value = true
  try {
    const params: any = { page: page.value, page_size: pageSize.value }
    if (keyword.value) params.keyword = keyword.value
    const res: any = await request.get('/knowledge', { params })
    list.value = res.data?.list || res.data || []
    total.value = res.data?.total || res.total || 0
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

function onSearch() {
  page.value = 1
  fetchData()
}

const sourceLabel = (t: string) => {
  const map: Record<string, string> = { manual: '手动录入', word: 'Word', pdf: 'PDF', epub: 'EPUB', markdown: 'Markdown', text: '文本' }
  return map[t] || t
}

function openAdd() {
  formTitle.value = '新增知识'
  form.value = { id: 0, title: '', content: '', language: 'zh', category: '', tags: '', status: 'active', source_type: 'manual', source_url: '' }
  dialogVisible.value = true
}

function openEdit(row: Knowledge) {
  formTitle.value = '编辑知识'
  form.value = { id: row.id, title: row.title, content: row.content, language: row.language, category: row.category, tags: row.tags, status: row.status, source_type: row.source_type, source_url: row.source_url }
  dialogVisible.value = true
}

function openPreview(row: Knowledge) {
  previewItem.value = row
  previewVisible.value = true
}

async function handleSave() {
  try {
    if (form.value.id) {
      await request.put(`/knowledge/${form.value.id}`, form.value)
      ElMessage.success('更新成功')
    } else {
      await request.post('/knowledge', form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch { /* handled */ }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确认删除该知识吗？', '提示')
    await request.delete(`/knowledge/${id}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch { /* cancelled */ }
}

async function handleView(row: Knowledge) {
  try {
    await request.post(`/knowledge/${row.id}/view`)
    row.view_count++
  } catch { /* ignore */ }
}

async function handleExport() {
  try {
    const res: any = await request.get('/knowledge/export', { params: { format: 'csv' }, responseType: 'blob' })
    const url = URL.createObjectURL(new Blob([res]))
    const link = document.createElement('a')
    link.href = url
    link.download = 'knowledge.csv'
    link.click()
    URL.revokeObjectURL(url)
  } catch { /* handled */ }
}

const fileInput = ref<HTMLInputElement | null>(null)
const docFileInput = ref<HTMLInputElement | null>(null)

function triggerImport() { fileInput.value?.click() }
function triggerDocUpload() { docFileInput.value?.click() }

async function handleImport(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  const formData = new FormData()
  formData.append('file', file)
  try {
    await request.post('/knowledge/import', formData, { headers: { 'Content-Type': 'multipart/form-data' } })
    ElMessage.success('导入成功')
    fetchData()
  } catch { /* handled */ } finally { input.value = '' }
}

async function handleDocUpload(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  uploadLoading.value = true
  const formData = new FormData()
  formData.append('file', file)
  try {
    const res: any = await request.post('/knowledge/upload-doc', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    ElMessage.success(`已导入 ${res.data?.count || 1} 条知识`)
    fetchData()
  } catch {
    ElMessage.error('文档上传失败')
  } finally {
    uploadLoading.value = false
    input.value = ''
  }
}

const previewType = computed(() => {
  if (!previewItem.value) return 'none'
  const t = previewItem.value.source_type
  if (t === 'pdf') return 'pdf'
  if (t === 'markdown') return 'markdown'
  if (t === 'manual' && previewItem.value.content && !previewItem.value.content.startsWith('[文档]')) return 'markdown'
  return 'text'
})

const previewURL = computed(() => {
  if (!previewItem.value || !previewItem.value.source_url) return ''
  const filename = previewItem.value.source_url.split('/').pop()
  return `/api/knowledge/file/${filename}`
})

const renderedContent = computed(() => {
  if (!previewItem.value) return ''
  if (previewType.value === 'markdown') {
    try { return marked.parse(previewItem.value.content) as string }
    catch { return previewItem.value.content }
  }
  return previewItem.value.content.replace(/\n/g, '<br>')
})

onMounted(() => {
  fetchData()
  fetchSources()
})
</script>

<template>
  <div>
    <div class="page-header">
      <h2>资料库管理</h2>
    </div>

    <el-card>
      <el-tabs v-model="activeTab">
        <!-- ===== 数据源管理 ===== -->
        <el-tab-pane label="数据源管理" name="sources">
          <div style="margin-bottom:16px">
            <el-button type="primary" @click="sourceDialog = true; editingSource = false">配置向量库</el-button>
          </div>

          <div class="config-section">
            <h3>当前向量库</h3>
            <el-form :model="sourceForm" label-width="120px">
              <el-form-item label="类型">
                <el-select v-model="sourceForm.provider" style="width:300px">
                  <el-option label="PGVector" value="postgresql" />
                  <el-option label="Neo4j" value="neo4j" />
                </el-select>
              </el-form-item>
              <el-form-item label="主机"><el-input v-model="sourceForm.host" style="width:300px" placeholder="localhost" /></el-form-item>
              <el-form-item label="端口"><el-input v-model="sourceForm.port" style="width:200px" placeholder="5432" /></el-form-item>
              <el-form-item label="用户名"><el-input v-model="sourceForm.user" style="width:300px" placeholder="postgres" /></el-form-item>
              <el-form-item label="密码"><el-input v-model="sourceForm.password" type="password" show-password style="width:300px" placeholder="数据库密码" /></el-form-item>
              <el-form-item label="数据库名"><el-input v-model="sourceForm.database" style="width:300px" placeholder="husky" /></el-form-item>
              <el-form-item label="SSL 模式">
                <el-select v-model="sourceForm.sslmode" style="width:200px">
                  <el-option label="disable" value="disable" />
                  <el-option label="require" value="require" />
                  <el-option label="verify-ca" value="verify-ca" />
                  <el-option label="verify-full" value="verify-full" />
                </el-select>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="handleSaveSource">保存配置</el-button>
              </el-form-item>
            </el-form>
          </div>
        </el-tab-pane>

        <!-- ===== 资料库查询 ===== -->
        <el-tab-pane label="资料库查询" name="query">
          <div style="display:flex;flex-direction:column;gap:24px">
            <!-- 语义搜索 -->
            <div>
              <h3 style="margin:0 0 12px 0">语义搜索</h3>
              <div style="display:flex;gap:12px;margin-bottom:12px">
                <el-input v-model="queryText" placeholder="输入搜索关键词..." style="flex:1" clearable @keyup.enter="handleQuery" />
                <el-input-number v-model="queryLimit" :min="1" :max="20" style="width:100px" />
                <el-button type="primary" :loading="queryLoading" @click="handleQuery">搜索</el-button>
              </div>

              <div v-if="queryResults.length" style="display:flex;flex-direction:column;gap:12px">
                <el-card v-for="(item, idx) in queryResults" :key="idx" shadow="hover">
                  <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:8px">
                    <strong>{{ item.title || item.name || '未知' }}</strong>
                    <el-tag v-if="item.score !== undefined" size="small">{{ (item.score * 100).toFixed(1) }}%</el-tag>
                  </div>
                  <p style="color:#666;font-size:13px;margin:0;line-height:1.6;white-space:pre-wrap">{{ item.content?.slice(0, 300) }}</p>
                </el-card>
              </div>
              <el-empty v-else-if="queryText && !queryLoading" description="无结果" />
            </div>

            <el-divider />

            <!-- 问答查询 -->
            <div>
              <h3 style="margin:0 0 12px 0">知识问答</h3>
              <div style="display:flex;gap:12px;margin-bottom:12px">
                <el-input v-model="queryQR" placeholder="输入问题..." style="flex:1" clearable @keyup.enter="handleAsk" />
                <el-button type="success" :loading="queryQALoading" @click="handleAsk">提问</el-button>
              </div>
              <div v-if="queryQAResult" class="qa-result">
                <div class="qa-result-content">{{ queryQAResult }}</div>
              </div>
            </div>
          </div>
        </el-tab-pane>

        <!-- ===== 文档管理 ===== -->
        <el-tab-pane label="文档管理" name="docs">
          <div style="margin-bottom:16px;display:flex;justify-content:space-between">
            <div style="display:flex;gap:12px">
              <el-input v-model="keyword" placeholder="搜索关键词" style="width:300px" clearable @keyup.enter="onSearch" />
              <el-button type="primary" @click="onSearch">搜索</el-button>
            </div>
            <div>
              <el-button @click="handleExport">导出 CSV</el-button>
              <el-button @click="triggerImport">导入 CSV</el-button>
              <el-button :loading="uploadLoading" @click="triggerDocUpload">上传文档</el-button>
              <el-button type="primary" @click="openAdd">新增条目</el-button>
              <input ref="fileInput" type="file" accept=".csv" style="display:none" @change="handleImport" />
              <input ref="docFileInput" type="file" accept=".md,.txt,.pdf,.doc,.docx,.epub" style="display:none" @change="handleDocUpload" />
            </div>
          </div>

          <el-table :data="list" v-loading="loading" stripe border style="width:100%" @row-click="handleView">
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="title" label="标题" min-width="150" show-overflow-tooltip />
            <el-table-column label="来源" width="100">
              <template #default="{ row }"><el-tag size="small">{{ sourceLabel(row.source_type) }}</el-tag></template>
            </el-table-column>
            <el-table-column prop="content" label="内容" min-width="200" show-overflow-tooltip>
              <template #default="{ row }">{{ row.content?.length > 80 ? row.content.slice(0, 80) + '...' : row.content }}</template>
            </el-table-column>
            <el-table-column prop="category" label="分类" width="120" show-overflow-tooltip />
            <el-table-column prop="tags" label="标签" width="150" show-overflow-tooltip />
            <el-table-column prop="status" label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.status === 'active' ? 'success' : 'warning'">{{ row.status === 'active' ? '已发布' : '已归档' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="view_count" label="浏览" width="70" />
            <el-table-column prop="created_at" label="创建时间" width="180" />
            <el-table-column label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click.stop="openPreview(row)">预览</el-button>
                <el-button size="small" @click.stop="openEdit(row)">编辑</el-button>
                <el-button size="small" type="danger" @click.stop="handleDelete(row.id)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination-wrap">
            <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="fetchData" />
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 文档编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="formTitle" width="700px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="标题"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="内容"><el-input v-model="form.content" type="textarea" :rows="8" /></el-form-item>
        <el-form-item label="语言">
          <el-select v-model="form.language">
            <el-option label="中文" value="zh" />
            <el-option label="英文" value="en" />
          </el-select>
        </el-form-item>
        <el-form-item label="分类"><el-input v-model="form.category" placeholder="分类名称" /></el-form-item>
        <el-form-item label="标签"><el-input v-model="form.tags" placeholder="多个标签用逗号分隔" /></el-form-item>
        <el-form-item label="状态" v-if="form.id">
          <el-select v-model="form.status">
            <el-option label="已发布" value="active" />
            <el-option label="已归档" value="archived" />
          </el-select>
        </el-form-item>
        <el-form-item label="来源" v-if="form.id"><el-tag>{{ sourceLabel(form.source_type) }}</el-tag></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- 预览对话框 -->
    <el-dialog v-model="previewVisible" title="内容预览" width="800px" top="5vh">
      <div v-if="previewItem" class="preview-content">
        <h2>{{ previewItem.title }}</h2>
        <div class="preview-meta">
          <span>分类：{{ previewItem.category || '-' }}</span>
          <span>来源：{{ sourceLabel(previewItem.source_type) }}</span>
          <span>浏览：{{ previewItem.view_count }}</span>
          <span v-if="previewItem.source_url"><a :href="previewURL" target="_blank" download>下载原文件</a></span>
        </div>
        <el-divider />
        <div v-if="previewType === 'pdf'" class="preview-pdf">
          <iframe :src="previewURL" width="100%" height="600" style="border:none"></iframe>
        </div>
        <div v-else-if="previewType === 'markdown'" class="preview-body markdown-body" v-html="renderedContent"></div>
        <div v-else class="preview-body" v-html="renderedContent"></div>
      </div>
      <template #footer>
        <el-button @click="previewVisible = false">关闭</el-button>
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
.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
.config-section {
  max-width: 640px;
}
.config-section h3 {
  margin: 0 0 16px 0;
  font-weight: 500;
  color: #333;
}
.preview-content h2 {
  margin: 0 0 12px 0;
}
.preview-meta {
  display: flex;
  gap: 20px;
  color: #999;
  font-size: 13px;
}
.preview-body {
  line-height: 1.8;
  white-space: pre-wrap;
}
.markdown-body {
  white-space: normal;
}
.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3) {
  margin: 16px 0 8px;
}
.markdown-body :deep(p) {
  margin: 8px 0;
}
.markdown-body :deep(code) {
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 90%;
}
.markdown-body :deep(pre) {
  background: #f5f5f5;
  padding: 12px;
  border-radius: 4px;
  overflow-x: auto;
}
.markdown-body :deep(pre code) {
  background: none;
  padding: 0;
}
.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  padding-left: 20px;
}
.markdown-body :deep(blockquote) {
  border-left: 4px solid #ddd;
  padding-left: 12px;
  color: #666;
  margin: 8px 0;
}
.markdown-body :deep(table) {
  border-collapse: collapse;
  width: 100%;
}
.markdown-body :deep(th),
.markdown-body :deep(td) {
  border: 1px solid #ddd;
  padding: 6px 10px;
  text-align: left;
}
.preview-pdf {
  margin: 0 -20px;
}
.qa-result {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 16px;
  line-height: 1.8;
}
.qa-result-content {
  white-space: pre-wrap;
}
</style>
