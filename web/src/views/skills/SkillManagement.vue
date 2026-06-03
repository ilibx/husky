<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { ElMessage, ElMessageBox, ElPopconfirm } from 'element-plus'
import { ArrowLeft, Plus, Delete, Edit, EditPen, Search, Refresh } from '@element-plus/icons-vue'
import request from '@/api/request'
import MarkdownEditor from '@/components/MarkdownEditor.vue'

interface Skill {
  id: number
  name: string
  description: string
  category: string
  enabled: boolean
  created_at: string
}

interface SkillMCP {
  id: number
  name: string
  description: string
  endpoint: string
  enabled: boolean
  skill_id: number
}

const list = ref<Skill[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const keyword = ref('')

const currentView = ref<'list' | 'edit'>('list')
const isEdit = ref(false)
const activeTab = ref('params')
const saving = ref(false)

const form = ref({
  id: 0,
  name: '',
  description: '',
  category: '',
  enabled: true,
})

const categories = ref<{ id: number; name: string }[]>([])
const skillMcps = ref<SkillMCP[]>([])
const mcpDialogVisible = ref(false)
const editingMcpIndex = ref(-1)
const mcpForm = ref({
  name: '',
  description: '',
  endpoint: '',
  enabled: true,
})

const skillContent = ref('')
const tableHeight = ref(500)

async function updateTableHeight() {
  await nextTick()
  const table = document.querySelector('.beauty-table') as HTMLElement | null
  if (!table) return
  const top = table.getBoundingClientRect().top
  const bottomGap = 96
  tableHeight.value = Math.max(360, window.innerHeight - top - bottomGap)
}

function buildSkillContent() {
  const f = form.value
  let s = '---\n'
  s += `name: ${f.name || ''}\n`
  s += `enabled: ${f.enabled}\n`
  s += '---\n\n'
  s += f.description || ''
  return s
}

function parseSkillContent(text: string): { frontmatter: Record<string,string>, body: string } {
  const lines = text.split('\n')
  let inFront = false
  let frontLines: string[] = []
  let bodyLines: string[] = []
  let frontDone = false
  for (const line of lines) {
    if (!frontDone && line.trim() === '---') {
      if (!inFront) { inFront = true; continue }
      else { frontDone = true; continue }
    }
    if (inFront && !frontDone) frontLines.push(line)
    else bodyLines.push(line)
  }
  const frontmatter: Record<string,string> = {}
  for (const line of frontLines) {
    const idx = line.indexOf(':')
    if (idx < 0) continue
    frontmatter[line.slice(0, idx).trim()] = line.slice(idx + 1).trim()
  }
  const body = bodyLines.join('\n').replace(/^\n+/, '').trimEnd()
  return { frontmatter, body }
}

async function fetchCategories() {
  try {
    const res: any = await request.get('/categories', { params: { page_size: 200 } })
    const raw = res.data?.data?.list || res.data?.list || res.data || []
    categories.value = Array.isArray(raw) ? raw.map((c: any) => ({ id: c.id, name: c.name })) : []
  } catch {
    categories.value = []
  }
}

async function fetchData() {
  loading.value = true
  try {
    const params: any = { page: page.value, page_size: pageSize.value }
    if (keyword.value) params.keyword = keyword.value
    const res: any = await request.get('/skills', { params })
    list.value = res?.list || res?.data?.list || res?.data?.data?.list || []
    total.value = res?.total || res?.data?.total || 0
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

async function openAdd() {
  isEdit.value = false
  form.value = { id: 0, name: '', description: '', category: '', enabled: true }
  skillMcps.value = []
  activeTab.value = 'params'
  skillContent.value = buildSkillContent()
  currentView.value = 'edit'
}

async function openEdit(row: Skill) {
  isEdit.value = true
  form.value = { id: row.id, name: row.name, description: row.description, category: row.category, enabled: row.enabled }
  activeTab.value = 'params'
  skillContent.value = buildSkillContent()
  currentView.value = 'edit'
  await fetchSkillMcps(row.id)
}

async function fetchSkillMcps(skillId: number) {
  try {
    const res: any = await request.get(`/skills/${skillId}/mcps`)
    skillMcps.value = res?.list || res?.data?.list || []
  } catch {
    skillMcps.value = []
  }
}

function goBack() {
  currentView.value = 'list'
  fetchData()
}

async function handleSave() {
  saving.value = true
  try {
    if (form.value.id) {
      await request.put(`/skills/${form.value.id}`, form.value)
    } else {
      const res: any = await request.post('/skills', form.value)
      form.value.id = res.data?.id || res.data?.data?.id || 0
    }
    ElMessage.success('保存成功')
    goBack()
  } catch {
    // handled by interceptor
  } finally {
    saving.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确认删除该 Skill 吗？', '提示')
    await request.delete(`/skills/${id}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch {
    // cancelled or error
  }
}

async function toggleSkillEnabled(row: Skill, val: boolean) {
  try {
    await request.put(`/skills/${row.id}`, { ...row, enabled: !!val })
    ElMessage.success(val ? '已启用' : '已停用')
  } catch {
    row.enabled = !val
  }
}

async function toggleMcpEnabled(row: SkillMCP, val: boolean) {
  try {
    await request.put(`/skills/${form.value.id}/mcps/${row.id}`, { ...row, enabled: !!val })
    ElMessage.success(val ? '已启用' : '已停用')
  } catch {
    row.enabled = !val
  }
}

function openAddMcp() {
  editingMcpIndex.value = -1
  mcpForm.value = { name: '', description: '', endpoint: '', enabled: true }
  mcpDialogVisible.value = true
}

function openEditMcp(index: number) {
  const mcp = skillMcps.value[index]
  editingMcpIndex.value = index
  mcpForm.value = {
    name: mcp.name,
    description: mcp.description,
    endpoint: mcp.endpoint,
    enabled: mcp.enabled,
  }
  mcpDialogVisible.value = true
}

async function saveMcp() {
  if (!mcpForm.value.name || !mcpForm.value.endpoint) {
    ElMessage.warning('名称和端点为必填')
    return
  }
  const skillId = form.value.id
  if (!skillId) {
    ElMessage.warning('请先保存技能基本信息')
    return
  }
  try {
    if (editingMcpIndex.value >= 0) {
      const mcp = skillMcps.value[editingMcpIndex.value]
      await request.put(`/skills/${skillId}/mcps/${mcp.id}`, mcpForm.value)
      ElMessage.success('工具已更新')
    } else {
      await request.post(`/skills/${skillId}/mcps`, mcpForm.value)
      ElMessage.success('工具已添加')
    }
    mcpDialogVisible.value = false
    await fetchSkillMcps(skillId)
  } catch {
    // handled by interceptor
  }
}

async function deleteMcp(index: number) {
  const mcp = skillMcps.value[index]
  try {
    await ElMessageBox.confirm(`确认删除工具 "${mcp.name}" 吗？`, '提示')
    await request.delete(`/skills/${form.value.id}/mcps/${mcp.id}`)
    ElMessage.success('工具已删除')
    skillMcps.value.splice(index, 1)
  } catch {
    // cancelled or error
  }
}

function stripMarkdown(text: string) {
  if (!text) return ''
  return text
    .replace(/^#{1,6}\s+/gm, '')
    .replace(/\*\*(.+?)\*\*/g, '$1')
    .replace(/\*(.+?)\*/g, '$1')
    .replace(/`{1,3}[^`]*`{1,3}/g, '')
    .replace(/\[(.+?)\]\(.+?\)/g, '$1')
    .replace(/^[-*+]\s+/gm, '')
    .replace(/^\d+\.\s+/gm, '')
    .replace(/^>\s+/gm, '')
    .replace(/---+/g, '')
    .trim()
}

const categoryColorMap: Record<string, string> = {
  development: 'primary',
  system: 'success',
  network: 'warning',
  security: 'danger',
  database: 'info',
  ai: '',
}

function getCategoryTagType(category: string): 'primary' | 'success' | 'warning' | 'info' | 'danger' | undefined {
  return categoryColorMap[category.toLowerCase()] as any || undefined
}

let searchTimer: ReturnType<typeof setTimeout>
function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => { page.value = 1; fetchData() }, 300)
}

function formatDate(val: string) {
  if (!val) return '-'
  return val.slice(0, 10)
}

watch(skillContent, (val) => {
  const { frontmatter, body } = parseSkillContent(val)
  if (frontmatter.name !== undefined) form.value.name = frontmatter.name
  if (frontmatter.enabled !== undefined) form.value.enabled = frontmatter.enabled === 'true'
  form.value.description = body
})

onMounted(() => {
  fetchData()
  fetchCategories()
  updateTableHeight()
  window.addEventListener('resize', updateTableHeight)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateTableHeight)
})
</script>

<template>
  <div class="page">
    <template v-if="currentView === 'list'">
      <div class="list-view">
        <div class="page-header">
          <h2>Skill 管理</h2>
        </div>
        <el-card shadow="never" class="list-card">
          <div class="table-toolbar">
            <div class="toolbar-left">
              <el-input v-model="keyword" placeholder="搜索技能名称..." clearable @input="onSearch" class="search-input">
                <template #prefix><el-icon><Search /></el-icon></template>
              </el-input>
            </div>
            <div class="toolbar-right">
              <el-button :icon="Refresh" circle @click="fetchData" />
              <el-button type="primary" :icon="Plus" @click="openAdd">新增 Skill</el-button>
            </div>
          </div>
          <el-table :data="list" v-loading="loading" stripe style="width:100%" class="beauty-table" :height="tableHeight">
            <el-table-column prop="id" label="ID" width="64" align="center" />
            <el-table-column prop="name" label="名称" min-width="150">
              <template #default="{ row }">
                <span class="name-cell">{{ row.name }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip>
              <template #default="{ row }">
                <span class="desc-preview">{{ stripMarkdown(row.description) || '-' }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="category" label="分类" width="130" align="center">
              <template #default="{ row }">
                <el-tag v-if="row.category" :type="getCategoryTagType(row.category)" size="small" effect="plain">
                  {{ row.category }}
                </el-tag>
                <span v-else class="no-category">-</span>
              </template>
            </el-table-column>
            <el-table-column prop="enabled" label="状态" width="72" align="center">
              <template #default="{ row }">
                <el-switch v-model="row.enabled" size="small" active-color="#67c23a" inactive-color="#c0c4cc" @change="(val: any) => toggleSkillEnabled(row, val)" />
              </template>
            </el-table-column>
            <el-table-column label="创建时间" width="108" align="center">
              <template #default="{ row }">
                <span class="date-cell">{{ formatDate(row.created_at) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right" align="center">
              <template #default="{ row }">
                <div class="action-group">
                  <el-button size="small" text :icon="EditPen" @click="openEdit(row)" />
                  <el-popconfirm title="确认删除?" @confirm="handleDelete(row.id)">
                    <template #reference>
                      <el-button size="small" text type="danger" :icon="Delete" />
                    </template>
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
    </template>

    <template v-else>
      <div class="edit-page">
        <div class="edit-topbar">
          <el-button text @click="goBack">
            <el-icon><ArrowLeft /></el-icon> 返回
          </el-button>
          <div class="topbar-title">{{ isEdit ? '编辑 Skill' : '新增 Skill' }}</div>
          <div class="topbar-actions">
            <el-button @click="goBack">取消</el-button>
            <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
          </div>
        </div>
        <el-card shadow="never" class="edit-card">
          <el-tabs v-model="activeTab">
            <el-tab-pane label="基本参数" name="params">
              <el-form :model="form" label-width="100px" class="edit-form">
                <el-form-item label="名称">
                  <el-input v-model="form.name" placeholder="如 网络故障排查" />
                </el-form-item>
                <el-form-item label="分类">
                  <el-select v-model="form.category" placeholder="选择分类" clearable filterable style="width:100%">
                    <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.name" />
                  </el-select>
                </el-form-item>
                <el-form-item label="启用">
                  <el-switch v-model="form.enabled" />
                </el-form-item>
              </el-form>
            </el-tab-pane>
            <el-tab-pane label="内容编辑" name="content">
              <MarkdownEditor v-model="skillContent" placeholder="---
name: my-skill
enabled: true

# Skill title

Markdown description..." />
            </el-tab-pane>
            <el-tab-pane label="工具管理" name="tools">
              <div class="tools-tab" v-if="form.id">
                <div class="tools-header">
                  <span class="tools-title">工具列表</span>
                  <el-button size="small" type="primary" :icon="Plus" @click="openAddMcp">添加工具</el-button>
                </div>
                <el-card shadow="never" class="tools-table-card">
                  <el-table :data="skillMcps" stripe style="width:100%" size="small" class="tools-table">
                    <el-table-column prop="name" label="名称" min-width="130">
                      <template #default="{ row }">
                        <span class="tools-name-cell">{{ row.name }}</span>
                      </template>
                    </el-table-column>
                    <el-table-column prop="endpoint" label="端点" min-width="200" show-overflow-tooltip>
                      <template #default="{ row }">
                        <code class="tools-endpoint">{{ row.endpoint }}</code>
                      </template>
                    </el-table-column>
                    <el-table-column prop="description" label="描述" min-width="150" show-overflow-tooltip>
                      <template #default="{ row }">
                        <span class="tools-desc">{{ row.description || '-' }}</span>
                      </template>
                    </el-table-column>
                    <el-table-column prop="enabled" label="状态" width="64" align="center">
                      <template #default="{ row }">
                        <el-switch v-model="row.enabled" size="small" active-color="#67c23a" inactive-color="#c0c4cc" @change="(val: any) => toggleMcpEnabled(row, val)" />
                      </template>
                    </el-table-column>
                    <el-table-column label="操作" width="90" align="center">
                      <template #default="{ row, $index }">
                        <div class="tool-action-group">
                          <el-button size="small" text :icon="EditPen" @click="openEditMcp($index)" />
                          <el-popconfirm title="确认删除?" @confirm="deleteMcp($index)">
                            <template #reference>
                              <el-button size="small" text type="danger" :icon="Delete" />
                            </template>
                          </el-popconfirm>
                        </div>
                      </template>
                    </el-table-column>
                  </el-table>
                  <div v-if="skillMcps.length === 0" class="tools-table-empty">
                    <el-empty description="暂无工具" :image-size="80" />
                  </div>
                </el-card>
              </div>
              <div v-else class="tools-placeholder">
                请先保存技能基本信息后再添加工具
              </div>
            </el-tab-pane>
          </el-tabs>
        </el-card>
      </div>
    </template>

    <el-dialog v-model="mcpDialogVisible" :title="editingMcpIndex >= 0 ? '编辑工具' : '添加工具'" width="500px">
      <el-form :model="mcpForm" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="mcpForm.name" placeholder="如 web-search" />
        </el-form-item>
        <el-form-item label="端点" required>
          <el-input v-model="mcpForm.endpoint" placeholder="https://api.example.com/mcp" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="mcpForm.description" type="textarea" :rows="2" placeholder="工具描述（可选）" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="mcpForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="mcpDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveMcp">确定</el-button>
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
.no-category {
  color: #c9cdd4;
}
.date-cell {
  color: #86909c;
  font-size: 13px;
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

.edit-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 120px);
}
.edit-topbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 0 16px;
  flex-shrink: 0;
}
.topbar-title {
  flex: 1;
  font-size: 16px;
  font-weight: 600;
  color: #1d2129;
}
.topbar-actions {
  display: flex;
  gap: 8px;
}
.edit-card {
  flex: 1;
  display: flex;
  flex-direction: column;
  border-radius: 12px;
  border: 1px solid #e5e6eb;
  overflow: hidden;
}
.edit-card :deep(.el-card__body) {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 20px 24px;
}
.edit-card :deep(.el-tabs) {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.edit-card :deep(.el-tabs__content) {
  flex: 1;
  position: relative;
}
.edit-card :deep(.el-tab-pane) {
  height: 100%;
  overflow-y: auto;
}
.edit-form {
  max-width: 800px;
}
.edit-card :deep(.md-editor) {
  display: flex;
  flex-direction: column;
  min-height: 400px;
  height: 100%;
}
.edit-card :deep(.md-editor-body) {
  flex: 1;
}
.edit-card :deep(.md-editor-textarea),
.edit-card :deep(.md-editor-preview) {
  height: 100% !important;
  min-height: 0 !important;
}

.tools-tab {
  display: flex;
  flex-direction: column;
  gap: 12px;
  height: 100%;
}
.tools-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}
.tools-title {
  font-size: 14px;
  font-weight: 600;
  color: #1d2129;
}
.tools-table-card {
  flex: 1;
  border-radius: 10px;
  border: 1px solid #ebeef5;
  overflow: hidden;
}
.tools-table-card :deep(.el-card__body) {
  padding: 0;
}
.tools-table {
  --el-table-border-color: #f0f0f0;
}
.tools-table :deep(.el-table__header th) {
  background: #f7f8fa;
  color: #4e5969;
  font-weight: 500;
}
.tools-name-cell {
  color: #1d2129;
  font-weight: 500;
}
.tools-endpoint {
  font-size: 12px;
  color: #86909c;
  background: #f7f8fa;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'SFMono-Regular', Consolas, monospace;
}
.tools-desc {
  color: #86909c;
  font-size: 13px;
}
.tool-action-group {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  white-space: nowrap;
}
.tool-action-group .el-button {
  margin-left: 0;
}
.tools-table-empty {
  padding: 20px 0;
}
.tools-placeholder {
  text-align: center;
  padding: 60px 0;
  color: #86909c;
  font-size: 14px;
}
</style>
