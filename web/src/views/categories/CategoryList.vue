<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

interface Category {
  id: number
  name: string
  parent_id: number | null
  sort_order: number
  status: number
  description: string
  children?: Category[]
}

const flatList = ref<Category[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const form = ref({
  name: '',
  parent_id: null as number | null,
  sort_order: 0,
  status: 1,
  description: '',
})
const formSubtitle = ref('')

function buildTree(items: Category[]): Category[] {
  const map = new Map<number, Category>()
  const roots: Category[] = []
  for (const item of items) {
    map.set(item.id, { ...item, children: [] })
  }
  for (const item of items) {
    const node = map.get(item.id)!
    if (item.parent_id != null && map.has(item.parent_id)) {
      map.get(item.parent_id)!.children!.push(node)
    } else if (item.parent_id == null) {
      roots.push(node)
    }
  }
  function sortChildren(list: Category[]) {
    list.sort((a, b) => a.sort_order - b.sort_order)
    for (const n of list) {
      if (n.children?.length) sortChildren(n.children)
    }
  }
  sortChildren(roots)
  return roots
}

const treeData = computed(() => buildTree(flatList.value))

function getDescendantIds(id: number): Set<number> {
  const ids = new Set<number>()
  function walk(nodeId: number) {
    ids.add(nodeId)
    for (const item of flatList.value) {
      if (item.parent_id === nodeId) {
        walk(item.id)
      }
    }
  }
  walk(id)
  return ids
}

const categoryOptions = computed(() => {
  const exclude = editingId.value ? getDescendantIds(editingId.value) : new Set<number>()
  return flatList.value
    .filter(item => !exclude.has(item.id))
    .map(item => ({ label: item.name, value: item.id }))
})

function openAdd(parentId: number | null = null) {
  editingId.value = null
  const parent = parentId ? flatList.value.find(i => i.id === parentId) : null
  formSubtitle.value = parent ? `（父分类：${parent.name}）` : ''
  form.value = {
    name: '',
    parent_id: parentId,
    sort_order: 0,
    status: 1,
    description: '',
  }
  dialogVisible.value = true
}

function openEdit(row: Category) {
  editingId.value = row.id
  formSubtitle.value = ''
  form.value = {
    name: row.name,
    parent_id: row.parent_id,
    sort_order: row.sort_order,
    status: row.status,
    description: row.description,
  }
  dialogVisible.value = true
}

async function handleSave() {
  try {
    const body = { ...form.value }
    if (editingId.value) {
      await request.put(`/categories/${editingId.value}`, body)
      ElMessage.success('更新成功')
    } else {
      await request.post('/categories', body)
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
    await ElMessageBox.confirm('确认删除该分类吗？', '提示')
    await request.delete(`/categories/${id}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch {
    // cancelled or error
  }
}

async function fetchData() {
  loading.value = true
  try {
    const res: any = await request.get('/categories')
    flatList.value = (res.data?.data || []).map((item: Category) => ({
      ...item,
      parent_id: item.parent_id ?? null,
      description: item.description || '',
    }))
  } catch {
    flatList.value = []
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <div>
    <div class="page-header">
      <h2>分类管理</h2>
      <el-button type="primary" @click="openAdd(null)">新增根分类</el-button>
    </div>

    <el-card>
      <el-table :data="treeData" v-loading="loading" row-key="id" default-expand-all stripe border style="width: 100%">
        <el-table-column prop="name" label="名称" min-width="240" />
        <el-table-column prop="description" label="描述" min-width="220" show-overflow-tooltip />
        <el-table-column prop="sort_order" label="排序" width="80" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openAdd(row.id)">新增子分类</el-button>
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="(editingId ? '编辑' : '新增') + '分类' + formSubtitle" width="550px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="form.name" placeholder="请输入分类名称" />
        </el-form-item>
        <el-form-item label="父分类">
          <el-select v-model="form.parent_id" placeholder="无（根分类）" clearable style="width: 100%">
            <el-option v-for="opt in categoryOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序值">
          <el-input-number v-model="form.sort_order" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch
            v-model="form.status"
            :active-value="1"
            :inactive-value="0"
            active-text="启用"
            inactive-text="禁用"
          />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入描述" />
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
</style>
