<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox, ElPopconfirm } from 'element-plus'
import { Plus, Delete, EditPen, Search, Refresh } from '@element-plus/icons-vue'
import request from '@/api/request'

interface ToolItem {
  id: number
  name: string
  type: 'builtin' | 'mcp'
  key: string
  description: string
  endpoint: string
  enabled: boolean
  created_at: string
}

const list = ref<ToolItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const keyword = ref('')

const dialogVisible = ref(false)
const isBuiltin = ref(false)
const formTitle = ref('新增工具')
const form = ref({
  id: 0,
  name: '',
  type: 'mcp' as 'builtin' | 'mcp',
  key: '',
  description: '',
  endpoint: '',
  enabled: true,
})

const typeFilter = ref('')

async function fetchData() {
  loading.value = true
  try {
    const params: any = { page: page.value, page_size: pageSize.value }
    if (keyword.value) params.keyword = keyword.value
    if (typeFilter.value) params.type = typeFilter.value
    const res: any = await request.get('/mcps', { params })
    list.value = res?.list || res?.data?.list || res?.data?.data || res?.data || []
    total.value = res?.total || res?.data?.total || 0
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

function openAdd() {
  isBuiltin.value = false
  formTitle.value = '新增 MCP 工具'
  form.value = { id: 0, name: '', type: 'mcp', key: '', description: '', endpoint: '', enabled: true }
  dialogVisible.value = true
}

function openEdit(row: ToolItem) {
  isBuiltin.value = row.type === 'builtin'
  formTitle.value = isBuiltin.value ? `内置工具：${row.name}` : '编辑 MCP 工具'
  form.value = { id: row.id, name: row.name, type: row.type, key: row.key, description: row.description, endpoint: row.endpoint, enabled: row.enabled }
  dialogVisible.value = true
}

async function handleSave() {
  try {
    if (form.value.id) {
      await request.put(`/mcps/${form.value.id}`, form.value)
      ElMessage.success('更新成功')
    } else {
      await request.post('/mcps', form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch { /* handled */ }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确认删除该工具？', '提示')
    await request.delete(`/mcps/${id}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch { /* cancelled */ }
}

function formatDate(val: string) {
  if (!val) return '-'
  return val.slice(0, 10)
}

async function toggleEnabled(row: ToolItem, val: boolean) {
  try {
    await request.put(`/mcps/${row.id}`, { ...row, enabled: !!val })
    ElMessage.success(val ? '已启用' : '已停用')
  } catch {
    row.enabled = !val
  }
}

function typeLabel(t: string) {
  return t === 'builtin' ? '内置' : 'MCP'
}

onMounted(fetchData)
</script>

<template>
  <div class="page">
    <div class="list-view">
      <div class="page-header">
        <h2>工具管理</h2>
      </div>

      <el-card shadow="never" class="list-card">
        <div class="table-toolbar">
          <div class="toolbar-left">
            <el-input v-model="keyword" placeholder="搜索工具名称..." clearable @input="onSearch" class="search-input">
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
            <el-select v-model="typeFilter" placeholder="类型" clearable style="width:100px" @change="onSearch">
              <el-option label="内置" value="builtin" />
              <el-option label="MCP" value="mcp" />
            </el-select>
          </div>
          <div class="toolbar-right">
            <el-button :icon="Refresh" circle @click="fetchData" />
            <el-button type="primary" :icon="Plus" @click="openAdd">新增 MCP 工具</el-button>
          </div>
        </div>

        <el-table :data="list" v-loading="loading" stripe style="width:100%" class="beauty-table" height="calc(100vh - 100px)">
          <el-table-column prop="id" label="ID" width="56" align="center" />
          <el-table-column label="类型" width="64" align="center">
            <template #default="{ row }">
              <el-tag :type="row.type === 'builtin' ? 'success' : 'primary'" size="small" effect="plain">{{ typeLabel(row.type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="name" label="名称" min-width="140">
            <template #default="{ row }"><span class="name-cell">{{ row.name }}</span></template>
          </el-table-column>
          <el-table-column prop="key" label="标识" width="120">
            <template #default="{ row }"><code class="key-cell">{{ row.key }}</code></template>
          </el-table-column>
          <el-table-column prop="endpoint" label="服务地址" min-width="220" show-overflow-tooltip>
            <template #default="{ row }"><code class="endpoint-cell">{{ row.endpoint || '-' }}</code></template>
          </el-table-column>
          <el-table-column prop="description" label="描述" min-width="180" show-overflow-tooltip>
            <template #default="{ row }"><span class="desc-preview">{{ row.description || '-' }}</span></template>
          </el-table-column>
          <el-table-column prop="enabled" label="状态" width="72" align="center">
            <template #default="{ row }">
              <el-switch v-model="row.enabled" size="small" active-color="#67c23a" inactive-color="#c0c4cc" @change="(val: any) => toggleEnabled(row, val)" />
            </template>
          </el-table-column>
          <el-table-column label="创建时间" width="100" align="center">
            <template #default="{ row }"><span class="date-cell">{{ formatDate(row.created_at) }}</span></template>
          </el-table-column>
          <el-table-column label="操作" width="100" fixed="right" align="center">
            <template #default="{ row }">
              <div class="action-group">
                <el-button size="small" text :icon="EditPen" @click="openEdit(row)" />
                <el-popconfirm v-if="row.type !== 'builtin'" title="确认删除?" @confirm="handleDelete(row.id)">
                  <template #reference><el-button size="small" text type="danger" :icon="Delete" /></template>
                </el-popconfirm>
              </div>
            </template>
          </el-table-column>
        </el-table>

        <div class="pagination-wrap" v-if="total > 0">
          <el-pagination
            v-model:current-page="page"
            v-model:page-size="pageSize"
            :total="total"
            layout="total, prev, pager, next"
            @current-change="fetchData"
          />
        </div>
      </el-card>
    </div>

    <el-dialog v-model="dialogVisible" :title="formTitle" width="550px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="form.name" placeholder="工具名称" />
        </el-form-item>
        <el-form-item label="标识" v-if="isBuiltin">
          <el-input v-model="form.key" disabled />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="工具功能描述" />
        </el-form-item>
        <el-form-item label="服务地址">
          <el-input v-model="form.endpoint" placeholder="http://example.com/api" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
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
.table-toolbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; gap: 12px; }
.toolbar-left { flex: 1; display: flex; align-items: center; gap: 8px; }
.toolbar-right { display: flex; align-items: center; gap: 8px; }
.search-input { width: 280px; }
.beauty-table { --el-table-border-color: #f0f0f0; flex: 1; min-height: 350px; }
.beauty-table :deep(.el-table__header th) { background: #f7f8fa; color: #4e5969; font-weight: 500; }
.name-cell { color: #1d2129; font-weight: 500; }
.key-cell { font-size: 12px; color: #2468f2; background: #e8f1ff; padding: 2px 6px; border-radius: 4px; font-family: 'SFMono-Regular', Consolas, monospace; }
.desc-preview { color: #86909c; font-size: 13px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; display: block; }
.endpoint-cell { font-size: 12px; color: #86909c; background: #f7f8fa; padding: 2px 6px; border-radius: 4px; font-family: 'SFMono-Regular', Consolas, monospace; }
.date-cell { color: #86909c; font-size: 13px; }
.action-group { display: flex; align-items: center; justify-content: center; gap: 4px; white-space: nowrap; }
.action-group .el-button { margin-left: 0; }
</style>
