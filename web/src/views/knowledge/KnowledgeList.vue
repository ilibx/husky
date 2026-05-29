<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

interface Knowledge {
  id: number
  title: string
  content: string
  language: string
  category: string
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
const formTitle = ref('新增知识')
const form = ref({ id: 0, title: '', content: '', language: 'zh', category: '', tags: '', status: 'active' })

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

function openAdd() {
  formTitle.value = '新增知识'
  form.value = { id: 0, title: '', content: '', language: 'zh', category: '', tags: '', status: 'active' }
  dialogVisible.value = true
}

function openEdit(row: Knowledge) {
  formTitle.value = '编辑知识'
  form.value = { id: row.id, title: row.title, content: row.content, language: row.language, category: row.category, tags: row.tags, status: row.status }
  dialogVisible.value = true
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
  } catch {
    // handled by interceptor
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确认删除该知识吗？', '提示')
    await request.delete(`/knowledge/${id}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch {
    // cancelled or error
  }
}

async function handleView(row: Knowledge) {
  try {
    await request.post(`/knowledge/${row.id}/view`)
    row.view_count++
  } catch {
    // ignore
  }
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
  } catch {
    // handled by interceptor
  }
}

const fileInput = ref<HTMLInputElement | null>(null)

function triggerImport() {
  fileInput.value?.click()
}

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
  } catch {
    // handled by interceptor
  } finally {
    input.value = ''
  }
}

onMounted(fetchData)
</script>

<template>
  <div>
    <div class="page-header">
      <h2>知识管理</h2>
      <div>
        <el-button @click="handleExport">导出CSV</el-button>
        <el-button @click="triggerImport">导入CSV</el-button>
        <el-button type="primary" @click="openAdd">新增知识</el-button>
        <input ref="fileInput" type="file" accept=".csv" style="display:none" @change="handleImport" />
      </div>
    </div>

    <el-card>
      <div style="margin-bottom:16px;display:flex;gap:12px">
        <el-input v-model="keyword" placeholder="搜索关键词" style="width:300px" clearable @keyup.enter="onSearch" />
        <el-button type="primary" @click="onSearch">搜索</el-button>
      </div>

      <el-table :data="list" v-loading="loading" stripe border style="width:100%" @row-click="handleView">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="title" label="标题" min-width="150" show-overflow-tooltip />
        <el-table-column prop="content" label="内容" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.content?.length > 80 ? row.content.slice(0, 80) + '...' : row.content }}
          </template>
        </el-table-column>
        <el-table-column prop="language" label="语言" width="80">
          <template #default="{ row }">
            <el-tag :type="row.language === 'zh' ? 'success' : 'primary'">{{ row.language === 'zh' ? '中文' : '英文' }}</el-tag>
          </template>
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
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="fetchData"
        />
      </div>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="formTitle" width="600px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="标题">
          <el-input v-model="form.title" />
        </el-form-item>
        <el-form-item label="内容">
          <el-input v-model="form.content" type="textarea" :rows="4" />
        </el-form-item>
        <el-form-item label="语言">
          <el-select v-model="form.language">
            <el-option label="中文" value="zh" />
            <el-option label="英文" value="en" />
          </el-select>
        </el-form-item>
        <el-form-item label="分类">
          <el-input v-model="form.category" />
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="form.tags" placeholder="多个标签用逗号分隔" />
        </el-form-item>
        <el-form-item label="状态" v-if="form.id">
          <el-select v-model="form.status">
            <el-option label="已发布" value="active" />
            <el-option label="已归档" value="archived" />
          </el-select>
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
</style>
