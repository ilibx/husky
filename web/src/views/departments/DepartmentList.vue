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
  status: number
  created_at: string
}

const departments = ref<Department[]>([])
const parentOptions = ref<Department[]>([])
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
    const res: any = await request.get('/departments')
    parentOptions.value = res.data?.data || res.data || []
  } catch {
    parentOptions.value = []
  }
}

function openAdd() {
  formTitle.value = '新增部门'
  form.value = { id: 0, name: '', code: '', parent_id: null, manager_id: null, status: 1 }
  fetchParentOptions()
  dialogVisible.value = true
}

function openEdit(row: Department) {
  formTitle.value = '编辑部门'
  form.value = { id: row.id, name: row.name, code: row.code, parent_id: row.parent_id, manager_id: row.manager_id, status: row.status }
  fetchParentOptions()
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
  <div>
    <div class="page-header">
      <h2>部门管理</h2>
      <el-button type="primary" @click="openAdd">新增部门</el-button>
    </div>

    <el-card>
      <el-table :data="departments" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="code" label="编码" width="120" />
        <el-table-column prop="manager_id" label="负责人ID" width="100">
          <template #default="{ row }">
            {{ row.manager_id ?? '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'warning'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
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

    <el-dialog v-model="dialogVisible" :title="formTitle" width="500px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="编码">
          <el-input v-model="form.code" />
        </el-form-item>
        <el-form-item label="上级部门">
          <el-select v-model="form.parent_id" clearable placeholder="请选择">
            <el-option v-for="d in parentOptions" :key="d.id" :label="d.name" :value="d.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="负责人ID">
          <el-input-number v-model="form.manager_id" :min="0" controls-position="right" style="width: 100%" />
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
