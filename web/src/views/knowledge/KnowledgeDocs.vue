<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, ElPopconfirm } from 'element-plus'
import { Search, Plus, Refresh, EditPen, Delete, View } from '@element-plus/icons-vue'
import { Marked } from 'marked'
import request from '@/api/request'

const marked = new Marked({ gfm: true, breaks: true })

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

interface Category {
  id: number
  name: string
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
const categories = ref<Category[]>([])

async function fetchCategories() {
  try {
    const res: any = await request.get('/categories')
    categories.value = res.data?.data || []
  } catch {
    categories.value = []
  }
}

async function fetchData() {
  loading.value = true
  try {
    const params: any = { page: page.value, page_size: pageSize.value }
    if (keyword.value) params.keyword = keyword.value
    const res: any = await request.get('/knowledge', { params })
    list.value = res.data?.data || res.data || []
    total.value = res.data?.total || 0
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

let searchTimer: ReturnType<typeof setTimeout>
function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => { page.value = 1; fetchData() }, 300)
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
  handleView(row)
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
  fetchCategories()
})
</script>

<template>
  <div class="page">
    <div class="list-view">
      <div class="page-header">
        <h2>文档库管理</h2>
      </div>

      <el-card shadow="never" class="list-card">
        <div class="table-toolbar">
          <div class="toolbar-left">
            <el-input v-model="keyword" placeholder="搜索关键词..." clearable @input="onSearch" class="search-input">
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
          </div>
          <div class="toolbar-right">
            <el-button :icon="Refresh" circle @click="fetchData" />
            <el-button @click="handleExport">导出 CSV</el-button>
            <el-button @click="triggerImport">导入 CSV</el-button>
            <el-button :loading="uploadLoading" @click="triggerDocUpload">上传文档</el-button>
            <el-button type="primary" :icon="Plus" @click="openAdd">新增条目</el-button>
            <input ref="fileInput" type="file" accept=".csv" style="display:none" @change="handleImport" />
            <input ref="docFileInput" type="file" accept=".md,.txt,.pdf,.doc,.docx,.epub" style="display:none" @change="handleDocUpload" />
          </div>
        </div>

        <el-table :data="list" v-loading="loading" stripe style="width:100%" class="beauty-table" height="calc(100vh - 100px)">
          <el-table-column prop="id" label="ID" width="64" align="center" />
          <el-table-column prop="title" label="标题" min-width="150" show-overflow-tooltip>
            <template #default="{ row }"><span class="name-cell">{{ row.title }}</span></template>
          </el-table-column>
          <el-table-column label="来源" width="86" align="center">
            <template #default="{ row }"><el-tag size="small" effect="plain">{{ sourceLabel(row.source_type) }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="content" label="内容" min-width="200" show-overflow-tooltip>
            <template #default="{ row }"><span class="desc-preview">{{ row.content?.length > 80 ? row.content.slice(0, 80) + '...' : row.content || '-' }}</span></template>
          </el-table-column>
          <el-table-column prop="category" label="分类" width="120" show-overflow-tooltip />
          <el-table-column prop="tags" label="标签" width="130" show-overflow-tooltip />
          <el-table-column prop="status" label="状态" width="72" align="center">
            <template #default="{ row }">
              <el-tag :type="row.status === 'active' ? 'success' : 'warning'" size="small" effect="plain">{{ row.status === 'active' ? '已发布' : '已归档' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="view_count" label="浏览" width="70" align="center" />
          <el-table-column prop="created_at" label="创建时间" width="170" />
          <el-table-column label="操作" width="140" fixed="right" align="center">
            <template #default="{ row }">
              <div class="action-group">
                <el-button size="small" text :icon="View" @click.stop="openPreview(row)" />
                <el-button size="small" text :icon="EditPen" @click.stop="openEdit(row)" />
                <el-popconfirm title="确认删除?" @confirm="handleDelete(row.id)">
                  <template #reference><el-button size="small" text type="danger" :icon="Delete" @click.stop /></template>
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
        <el-form-item label="分类">
          <el-select v-model="form.category" filterable clearable placeholder="选择系统分类" style="width: 100%">
            <el-option v-for="item in categories" :key="item.id" :label="item.name" :value="item.name" />
          </el-select>
        </el-form-item>
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
  flex-wrap: wrap;
}
.search-input {
  width: 320px;
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
.name-cell {
  color: #1d2129;
  font-weight: 500;
}
.desc-preview {
  color: #86909c;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
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

/* Preview modal styles */
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
</style>
