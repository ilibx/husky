<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

interface SLAConfig {
  id: number
  priority: string
  category_id: number | null
  category_name?: string
  response_minutes: number
  resolution_minutes: number
  warning_threshold: number
  enabled: boolean
  created_at: string
}

interface Category {
  id: number
  name: string
}

const slas = ref<SLAConfig[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const dialogVisible = ref(false)
const formTitle = ref('新增SLA配置')
const form = ref<SLAConfig>({
  id: 0,
  priority: 'medium',
  category_id: null,
  response_minutes: 60,
  resolution_minutes: 480,
  warning_threshold: 0.8,
  enabled: true,
  created_at: '',
})

const categories = ref<Category[]>([])

async function fetchData() {
  loading.value = true
  try {
    const res: any = await request.get('/sla-configs', { params: { page: page.value, page_size: pageSize.value } })
    slas.value = res.data?.list || res.data || []
    total.value = res.data?.total || res.total || 0
  } catch {
    slas.value = []
  } finally {
    loading.value = false
  }
}

async function fetchCategories() {
  try {
    const res: any = await request.get('/categories')
    categories.value = res.data?.list || res.data || []
  } catch {
    categories.value = []
  }
}

function openAdd() {
  formTitle.value = '新增SLA配置'
  form.value = {
    id: 0,
    priority: 'medium',
    category_id: null,
    response_minutes: 60,
    resolution_minutes: 480,
    warning_threshold: 0.8,
    enabled: true,
    created_at: '',
  }
  dialogVisible.value = true
}

function openEdit(row: SLAConfig) {
  formTitle.value = '编辑SLA配置'
  form.value = {
    id: row.id,
    priority: row.priority,
    category_id: row.category_id,
    response_minutes: row.response_minutes,
    resolution_minutes: row.resolution_minutes,
    warning_threshold: row.warning_threshold,
    enabled: row.enabled,
    created_at: row.created_at,
  }
  dialogVisible.value = true
}

async function handleSave() {
  try {
    if (form.value.id) {
      await request.put(`/sla-configs/${form.value.id}`, form.value)
      ElMessage.success('更新成功')
    } else {
      await request.post('/sla-configs', form.value)
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
    await ElMessageBox.confirm('确认删除该SLA配置吗？', '提示')
    await request.delete(`/sla-configs/${id}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch {
    // cancelled or error
  }
}

function priorityType(priority: string): 'primary' | 'success' | 'warning' | 'info' | 'danger' {
  const map: Record<string, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = { low: 'info', medium: 'warning', high: 'danger', urgent: 'danger' }
  return map[priority] || 'info'
}

function priorityLabel(priority: string): string {
  const map: Record<string, string> = { low: '低', medium: '中', high: '高', urgent: '紧急' }
  return map[priority] || priority
}

onMounted(() => {
  fetchData()
  fetchCategories()
})
</script>

<template>
  <div>
    <div class="page-header">
      <h2>SLA配置管理</h2>
      <el-button type="primary" @click="openAdd">新增SLA配置</el-button>
    </div>

    <el-card>
      <el-table :data="slas" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="priority" label="优先级" width="100">
          <template #default="{ row }">
            <el-tag :type="priorityType(row.priority)">{{ priorityLabel(row.priority) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="category_name" label="分类" min-width="120">
          <template #default="{ row }">
            {{ row.category_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="response_minutes" label="响应时间(分)" width="130" />
        <el-table-column prop="resolution_minutes" label="解决时间(分)" width="130" />
        <el-table-column prop="warning_threshold" label="预警阈值" width="100">
          <template #default="{ row }">
            {{ (row.warning_threshold * 100).toFixed(0) }}%
          </template>
        </el-table-column>
        <el-table-column prop="enabled" label="启用" width="80">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="160" fixed="right">
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
      <el-form :model="form" label-width="140px">
        <el-form-item label="优先级">
          <el-select v-model="form.priority">
            <el-option label="低" value="low" />
            <el-option label="中" value="medium" />
            <el-option label="高" value="high" />
            <el-option label="紧急" value="urgent" />
          </el-select>
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="form.category_id" clearable placeholder="请选择分类">
            <el-option v-for="cat in categories" :key="cat.id" :label="cat.name" :value="cat.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="响应时间(分钟)">
          <el-input-number v-model="form.response_minutes" :min="1" />
        </el-form-item>
        <el-form-item label="解决时间(分钟)">
          <el-input-number v-model="form.resolution_minutes" :min="1" />
        </el-form-item>
        <el-form-item label="预警阈值">
          <el-input-number v-model="form.warning_threshold" :min="0.1" :max="1" :step="0.1" />
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
