<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

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

const treeData = ref<MenuItem[]>([])
const flatList = ref<MenuItem[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const form = ref({
  name: '',
  path: '',
  icon: '',
  parent_id: null as number | null,
  sort: 0,
  roles: '',
  hidden: false,
  external: false,
  iframe: false,
})
const roleOptions = ['admin', 'agent', 'user']

const parentOptions = computed(() => {
  return flatList.value
    .filter(m => !m.path || m.path === '')
    .map(m => ({ label: m.name, value: m.id }))
})

function toggleRole(role: string) {
  const roles = form.value.roles ? form.value.roles.split(',') : []
  const idx = roles.indexOf(role)
  if (idx >= 0) {
    roles.splice(idx, 1)
  } else {
    roles.push(role)
  }
  form.value.roles = roles.join(',')
}

function hasRole(role: string) {
  return form.value.roles.split(',').includes(role)
}

async function fetchMenus() {
  loading.value = true
  try {
    const res: any = await request.get('/menus/all')
    treeData.value = res.data?.data || []
    flatList.value = res.data?.list || []
  } catch {
    treeData.value = []
    flatList.value = []
  } finally {
    loading.value = false
  }
}

function openAdd(parentId: number | null) {
  editingId.value = null
  form.value = { name: '', path: '', icon: '', parent_id: parentId, sort: 0, roles: '', hidden: false, external: false, iframe: false }
  dialogVisible.value = true
}

function openEdit(item: MenuItem) {
  editingId.value = item.id
  form.value = {
    name: item.name,
    path: item.path || '',
    icon: item.icon || '',
    parent_id: item.parent_id,
    sort: item.sort,
    roles: item.roles || '',
    hidden: item.hidden,
    external: item.external || false,
    iframe: item.iframe || false,
  }
  dialogVisible.value = true
}

async function handleSave() {
  if (!form.value.name) {
    ElMessage.warning('请输入菜单名称')
    return
  }
  try {
    if (editingId.value) {
      await request.put(`/menus/${editingId.value}`, form.value)
      ElMessage.success('更新成功')
    } else {
      await request.post('/menus', form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchMenus()
  } catch {
    // handled by interceptor
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确认删除该菜单及其子菜单？', '提示')
    await request.delete(`/menus/${id}`)
    ElMessage.success('删除成功')
    fetchMenus()
  } catch {
    // cancelled or error
  }
}

onMounted(fetchMenus)
</script>

<template>
  <div class="page">
    <div class="list-view">
      <div class="page-header">
        <h2>菜单管理</h2>
      </div>

      <el-card shadow="never" class="list-card">
        <div class="table-toolbar">
          <div class="toolbar-left" />
          <div class="toolbar-right">
            <el-button type="primary" @click="openAdd(null)">添加根菜单</el-button>
          </div>
        </div>

        <div v-loading="loading" class="table-wrapper">
          <template v-if="treeData.length === 0 && !loading">
            <el-empty description="暂无菜单，请先添加" />
          </template>

          <el-table v-else :data="treeData" row-key="id" class="beauty-table" default-expand-all style="width:100%">
            <el-table-column prop="name" label="名称" min-width="160">
              <template #default="{ row }">
                <span class="menu-name">{{ row.name }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="path" label="路径/链接" width="220" show-overflow-tooltip />
            <el-table-column prop="icon" label="图标" width="100" align="center">
              <template #default="{ row }">
                <el-tag v-if="row.icon" size="small" effect="plain">{{ row.icon }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="sort" label="排序" width="70" align="center" />
            <el-table-column prop="roles" label="可见角色" min-width="200">
              <template #default="{ row }">
                <el-tag v-if="!row.roles" size="small" type="success" effect="plain">全部</el-tag>
                <el-tag v-else v-for="r in row.roles.split(',')" :key="r" size="small" effect="plain" style="margin-right:4px">{{ r }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="hidden" label="隐藏" width="70" align="center">
              <template #default="{ row }">
                <el-tag v-if="row.hidden" type="warning" size="small" effect="plain">是</el-tag>
                <el-tag v-else type="info" size="small" effect="plain">否</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="打开方式" width="100" align="center">
              <template #default="{ row }">
                <el-tag v-if="row.external && row.iframe" size="small" type="success" effect="plain">内嵌</el-tag>
                <el-tag v-else-if="row.external" size="small" type="warning" effect="plain">外链</el-tag>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="220" fixed="right" align="center">
              <template #default="{ row }">
                <el-button size="small" @click="openAdd(row.id)">添加子菜单</el-button>
                <el-button size="small" @click="openEdit(row)">编辑</el-button>
                <el-button size="small" type="danger" @click="handleDelete(row.id)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-card>

      <el-dialog v-model="dialogVisible" :title="editingId ? '编辑菜单' : '添加菜单'" width="550px">
        <el-form :model="form" label-width="100px">
          <el-form-item label="名称">
            <el-input v-model="form.name" placeholder="菜单显示名称" />
          </el-form-item>
          <el-form-item label="路径/链接">
            <el-input v-model="form.path" placeholder="内部路由 /example，或外部链接 https://example.com" />
          </el-form-item>
          <el-form-item label="上级菜单">
            <el-select v-model="form.parent_id" placeholder="无（顶级菜单）" clearable style="width:100%">
              <el-option v-for="p in parentOptions" :key="p.value" :label="p.label" :value="p.value" />
            </el-select>
          </el-form-item>
          <el-form-item label="图标">
            <el-input v-model="form.icon" placeholder="Element Plus 图标名，如 Ticket" />
          </el-form-item>
          <el-form-item label="排序">
            <el-input-number v-model="form.sort" :min="0" />
          </el-form-item>
          <el-form-item label="可见角色">
            <el-checkbox-group>
              <el-checkbox v-for="role in roleOptions" :key="role" :checked="hasRole(role)" @change="toggleRole(role)">
                {{ role === 'admin' ? '管理员' : role === 'agent' ? '客服' : '用户' }}
              </el-checkbox>
            </el-checkbox-group>
            <div v-if="form.roles === ''" style="color:#999;font-size:12px;margin-top:4px">未选择时所有角色可见</div>
          </el-form-item>
          <el-form-item label="外部链接">
            <el-switch v-model="form.external" />
            <span style="color:#999;font-size:12px;margin-left:8px">开启后路径按外部 URL 处理</span>
          </el-form-item>
          <el-form-item v-if="form.external" label="内嵌访问">
            <el-switch v-model="form.iframe" />
            <span style="color:#999;font-size:12px;margin-left:8px">开启后在系统页面中 iframe 打开</span>
          </el-form-item>
          <el-form-item label="隐藏">
            <el-switch v-model="form.hidden" />
            <span style="color:#999;font-size:12px;margin-left:8px">隐藏后不显示在侧边栏</span>
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
.table-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
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
.menu-name {
  font-weight: 500;
}
</style>
