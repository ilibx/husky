<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

interface Role {
  id: number
  name: string
  description: string
  permissions: string | string[]
  status: number
  created_at: string
}

interface MenuItem {
  id: number
  name: string
  path: string
  icon: string
  parent_id: number | null
  sort: number
  roles: string
  hidden: boolean
  external: boolean
  iframe: boolean
  children?: MenuItem[]
}

const roles = ref<Role[]>([])
const menuTree = ref<MenuItem[]>([])
const menuList = ref<MenuItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const menuTreeRef = ref<any>()

const dialogVisible = ref(false)
const permissionDialogVisible = ref(false)
const formTitle = ref('新增角色')
const form = ref({ id: 0, name: '', description: '', permissions: [] as string[], status: 1 })
const permissionRole = ref<Role | null>(null)

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

async function fetchMenus() {
  const res: any = await request.get('/menus/all')
  menuTree.value = res.data?.data || []
  menuList.value = res.data?.list || []
}

function normalizePermissions(permissions: string | string[] | null | undefined): string[] {
  if (Array.isArray(permissions)) return permissions
  if (!permissions) return []
  try {
    const parsed = JSON.parse(permissions)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

function menuHasRole(menu: MenuItem, roleName: string) {
  if (!menu.roles) return true
  return menu.roles.split(',').map(r => r.trim()).includes(roleName)
}

function getRoleMenuCount(roleName: string) {
  return menuList.value.filter(menu => menuHasRole(menu, roleName)).length
}

function getCheckedMenuIds(roleName: string) {
  return menuList.value.filter(menu => menuHasRole(menu, roleName)).map(menu => menu.id)
}

function formatDate(date: string) {
  if (!date) return '-'
  return date.slice(0, 10)
}

function openAdd() {
  formTitle.value = '新增角色'
  form.value = { id: 0, name: '', description: '', permissions: [], status: 1 }
  dialogVisible.value = true
}

function openEdit(row: Role) {
  formTitle.value = '编辑角色'
  form.value = { id: row.id, name: row.name, description: row.description, permissions: normalizePermissions(row.permissions), status: row.status }
  dialogVisible.value = true
}

async function openPermission(row: Role) {
  permissionRole.value = row
  await fetchMenus()
  permissionDialogVisible.value = true
  await nextTick()
  menuTreeRef.value?.setCheckedKeys(getCheckedMenuIds(row.name))
}

async function handleSave() {
  try {
    const payload = { ...form.value, permissions: JSON.stringify(form.value.permissions) }
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

async function handlePermissionSave() {
  if (!permissionRole.value || permissionRole.value.name === 'admin') return
  const roleName = permissionRole.value.name
  const menuIds = menuTreeRef.value?.getCheckedKeys(false) || []
  try {
    await request.put(`/menus/roles/${roleName}/permissions`, { menu_ids: menuIds })
    ElMessage.success('权限更新成功')
    permissionDialogVisible.value = false
    await fetchMenus()
  } catch {
    // handled by interceptor
  }
}

async function handleDelete(row: Role) {
  if (row.name === 'admin') return
  try {
    await ElMessageBox.confirm('确认删除该角色吗？', '提示')
    await request.delete(`/roles/${row.id}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch {
    // cancelled or error
  }
}

onMounted(async () => {
  await fetchData()
  await fetchMenus()
})
</script>

<template>
  <div class="page">
    <div class="list-view">
      <div class="page-header">
        <h2>角色管理</h2>
      </div>

      <el-card shadow="never" class="list-card">
        <div class="table-toolbar">
          <div class="toolbar-left" />
          <div class="toolbar-right">
            <el-button type="primary" @click="openAdd">新增角色</el-button>
          </div>
        </div>

        <el-table :data="roles" v-loading="loading" class="beauty-table" style="width:100%">
          <el-table-column prop="id" label="ID" width="60" align="center" />
          <el-table-column prop="name" label="名称" min-width="120" />
          <el-table-column prop="description" label="描述" min-width="160" show-overflow-tooltip />
          <el-table-column label="菜单权限" min-width="140" align="center">
            <template #default="{ row }">
              <el-tag size="small" effect="plain">{{ getRoleMenuCount(row.name) }} 个菜单</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'warning'" size="small" effect="plain">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="创建日期" width="120" align="center">
            <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="240" fixed="right" align="center">
            <template #default="{ row }">
              <el-button size="small" @click="openEdit(row)">编辑</el-button>
              <el-button size="small" type="primary" :disabled="row.name === 'admin'" @click="openPermission(row)">权限管理</el-button>
              <el-button size="small" type="danger" :disabled="row.name === 'admin'" @click="handleDelete(row)">删除</el-button>
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
          <el-form-item label="描述">
            <el-input v-model="form.description" type="textarea" :rows="3" />
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

      <el-dialog v-model="permissionDialogVisible" :title="`权限管理 - ${permissionRole?.name || ''}`" width="680px">
        <el-tree
          ref="menuTreeRef"
          :data="menuTree"
          node-key="id"
          show-checkbox
          default-expand-all
          check-strictly
          :props="{ label: 'name', children: 'children' }"
        />
        <template #footer>
          <el-button @click="permissionDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handlePermissionSave">保存</el-button>
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
