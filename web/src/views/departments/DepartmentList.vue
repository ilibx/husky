<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

interface Department {
  id: number
  name: string
  code: string
  parent_id: number | null
  manager_id: number | null
  manager?: User | null
  status: number
  created_at: string
}

interface User {
  id: number
  username: string
  email: string
}

const departments = ref<Department[]>([])
const parentOptions = ref<Department[]>([])
const userOptions = ref<User[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const dialogVisible = ref(false)
const formTitle = ref('新增部门')
const form = ref({ id: 0, name: '', code: '', parent_id: null as number | null, manager_id: null as number | null, status: 1 })

async function fetchData() {
  loading.value = true
  try {
    const res: any = await request.get('/departments', { params: { page: page.value, page_size: pageSize.value } })
    departments.value = res.data?.data || res.data || []
    total.value = res.data?.total || 0
  } catch {
    departments.value = []
  } finally {
    loading.value = false
  }
}

async function fetchParentOptions() {
  try {
    const res: any = await request.get('/departments', { params: { page: 1, page_size: 200 } })
    parentOptions.value = res.data?.data || res.data || []
  } catch {
    parentOptions.value = []
  }
}

async function fetchUserOptions() {
  try {
    const res: any = await request.get('/users', { params: { page: 1, page_size: 200 } })
    userOptions.value = res.data?.data || []
  } catch {
    userOptions.value = []
  }
}

function openAdd() {
  formTitle.value = '新增部门'
  form.value = { id: 0, name: '', code: '', parent_id: null, manager_id: null, status: 1 }
  fetchParentOptions()
  fetchUserOptions()
  dialogVisible.value = true
}

function openEdit(row: Department) {
  formTitle.value = '编辑部门'
  form.value = { id: row.id, name: row.name, code: row.code, parent_id: row.parent_id, manager_id: row.manager_id, status: row.status }
  fetchParentOptions()
  fetchUserOptions()
  dialogVisible.value = true
}

async function handleSave() {
  try {
    if (form.value.id) {
      await request.put(`/departments/${form.value.id}`, form.value)
      ElMessage.success('更新成功')
    } else {
      await request.post('/departments', form.value)
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
    await ElMessageBox.confirm('确认删除该部门吗？', '提示')
    await request.delete(`/departments/${id}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch {
    // cancelled or error
  }
}

onMounted(fetchData)
</script>

<template>
  <div class="page">
    <div class="list-view">
      <div class="page-header">
        <h2>部门管理</h2>
      </div>

      <el-card shadow="never" class="list-card">
        <div class="table-toolbar">
          <div class="toolbar-left" />
          <div class="toolbar-right">
            <el-button type="primary" @click="openAdd">新增部门</el-button>
          </div>
        </div>

        <el-table :data="departments" v-loading="loading" class="beauty-table" style="width:100%">
          <el-table-column prop="id" label="ID" width="60" align="center" />
          <el-table-column prop="name" label="名称" min-width="120" />
          <el-table-column prop="code" label="编码" width="120" />
          <el-table-column label="负责人" width="140" align="center">
            <template #default="{ row }">
              {{ row.manager?.username || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'warning'" size="small" effect="plain">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="180" align="center" />
          <el-table-column label="操作" width="150" fixed="right" align="center">
            <template #default="{ row }">
              <el-button size="small" @click="openEdit(row)">编辑</el-button>
              <el-button size="small" type="danger" @click="handleDelete(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <div v-if="total > 0" class="pagination-wrap">
          <el-pagination
            v-model:current-page="page"
            v-model:page-size="pageSize"
            :total="total"
            layout="total, prev, pager, next"
            @current-change="fetchData"
          />
        </div>
      </el-card>

      <el-dialog v-model="dialogVisible" :title="formTitle" width="500px">
        <el-form :model="form" label-width="80px">
          <el-form-item label="名称">
            <el-input v-model="form.name" />
          </el-form-item>
          <el-form-item label="编码">
            <el-input v-model="form.code" placeholder="系统内部唯一编码，如 ops、hr" />
            <div class="form-help">用于部门唯一标识和层级路径，建议使用英文、数字或短横线</div>
          </el-form-item>
          <el-form-item label="上级部门">
            <el-select v-model="form.parent_id" clearable placeholder="请选择">
              <el-option v-for="d in parentOptions" :key="d.id" :label="d.name" :value="d.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="负责人">
            <el-select v-model="form.manager_id" clearable filterable placeholder="请选择负责人" style="width: 100%">
              <el-option
                v-for="u in userOptions"
                :key="u.id"
                :label="u.username + (u.email ? `（${u.email}）` : '')"
                :value="u.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSave">保存</el-button>
        </template>
      </el-dialog>
    </div>
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
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 600;
  color: #1d2129;
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
  display: flex;
  align-items: center;
  gap: 8px;
}
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
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
.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
  flex-shrink: 0;
}
.form-help {
  margin-top: 4px;
  color: #909399;
  font-size: 12px;
  line-height: 18px;
}
</style>
