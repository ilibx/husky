<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

interface User {
  id: number
  username: string
  email: string
  role: string
  status: number
  created_at: string
}

const users = ref<User[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const dialogVisible = ref(false)
const formTitle = ref('新增用户')
const form = ref({ id: 0, username: '', email: '', role: '', password: '' })
const roleOptions = ref<{ name: string; description: string }[]>([])

async function fetchRoles() {
  try {
    const res: any = await request.get('/roles')
    roleOptions.value = res.data?.data || []
  } catch {
    roleOptions.value = []
  }
}

async function fetchData() {
  loading.value = true
  try {
    const res: any = await request.get('/users', { params: { page: page.value, page_size: pageSize.value } })
    users.value = res.data?.data || res.data || []
    total.value = res.data?.total || 0
  } catch {
    users.value = []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  formTitle.value = '新增用户'
  form.value = { id: 0, username: '', email: '', role: roleOptions.value[0]?.name || '', password: '' }
  dialogVisible.value = true
}

function openEdit(row: User) {
  formTitle.value = '编辑用户'
  form.value = { id: row.id, username: row.username, email: row.email, role: row.role, password: '' }
  dialogVisible.value = true
}

async function handleSave() {
  try {
    if (form.value.id) {
      await request.put(`/users/${form.value.id}`, form.value)
      ElMessage.success('更新成功')
    } else {
      await request.post('/users', form.value)
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
    await ElMessageBox.confirm('确认删除该用户吗？', '提示')
    await request.delete(`/users/${id}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch {
    // cancelled or error
  }
}

onMounted(async () => {
  await fetchRoles()
  await fetchData()
})
</script>

<template>
  <div class="page">
    <div class="list-view">
      <div class="page-header">
        <h2>用户管理</h2>
      </div>

      <el-card shadow="never" class="list-card">
        <div class="table-toolbar">
          <div class="toolbar-left" />
          <div class="toolbar-right">
            <el-button type="primary" @click="openAdd">新增用户</el-button>
          </div>
        </div>

        <el-table :data="users" v-loading="loading" class="beauty-table" style="width:100%">
          <el-table-column prop="id" label="ID" width="60" align="center" />
          <el-table-column prop="username" label="用户名" min-width="120" />
          <el-table-column prop="email" label="邮箱" min-width="180" />
          <el-table-column prop="role" label="角色" width="100" align="center">
            <template #default="{ row }">
              <el-tag :type="row.role === 'admin' ? 'danger' : 'info'" size="small" effect="plain">{{ row.role }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'warning'" size="small" effect="plain">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="180" align="center" />
          <el-table-column label="操作" width="160" fixed="right" align="center">
            <template #default="{ row }">
              <el-button size="small" @click="openEdit(row)">编辑</el-button>
              <el-button size="small" type="danger" :disabled="row.role === 'admin' || row.username === 'admin'" @click="handleDelete(row.id)">删除</el-button>
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
          <el-form-item label="用户名">
            <el-input v-model="form.username" />
          </el-form-item>
          <el-form-item label="邮箱">
            <el-input v-model="form.email" />
          </el-form-item>
          <el-form-item label="角色">
            <el-select v-model="form.role" placeholder="请选择角色">
              <el-option v-for="r in roleOptions" :key="r.name" :label="r.description ? `${r.name}（${r.description}）` : r.name" :value="r.name" />
            </el-select>
          </el-form-item>
          <el-form-item label="密码" v-if="!form.id">
            <el-input v-model="form.password" type="password" show-password />
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
</style>
