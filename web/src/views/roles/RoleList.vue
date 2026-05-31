<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

interface Role {
  id: number
  name: string
  description: string
  permissions: string[]
  status: number
  created_at: string
}

const roles = ref<Role[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const dialogVisible = ref(false)
const formTitle = ref('新增角色')
const form = ref({ id: 0, name: '', description: '', permissions: '', status: 1 })

async function fetchData() {
  loading.value = true
  try {
    const res: any = await request.get('/roles', { params: { page: page.value, page_size: pageSize.value } })
    roles.value = res.data?.data || res.data || []
    total.value = res.data?.total || 0
  } catch {
    roles.value = []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  formTitle.value = '新增角色'
  form.value = { id: 0, name: '', description: '', permissions: '', status: 1 }
  dialogVisible.value = true
}

function openEdit(row: Role) {
  formTitle.value = '编辑角色'
  form.value = { id: row.id, name: row.name, description: row.description, permissions: JSON.stringify(row.permissions), status: row.status }
  dialogVisible.value = true
}

async function handleSave() {
  try {
    const payload = {
      ...form.value,
      permissions: form.value.permissions ? JSON.parse(form.value.permissions) : []
    }
    if (form.value.id) {
      await request.put(`/roles/${form.value.id}`, payload)
      ElMessage.success('更新成功')
    } else {
      await request.post('/roles', payload)
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
    await ElMessageBox.confirm('确认删除该角色吗？', '提示')
    await request.delete(`/roles/${id}`)
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
      <h2>角色管理</h2>
      <el-button type="primary" @click="openAdd">新增角色</el-button>
    </div>

    <el-card>
      <el-table :data="roles" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="description" label="描述" min-width="160" show-overflow-tooltip />
        <el-table-column label="权限" min-width="240">
          <template #default="{ row }">
            {{ Array.isArray(row.permissions) ? row.permissions.join(', ') : row.permissions }}
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
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="权限">
          <el-input v-model="form.permissions" type="textarea" :rows="4" placeholder='["ticket:create","ticket:read"]' />
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
