<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

interface Tag {
  id: number
  name: string
  color: string
  created_at: string
}

const tags = ref<Tag[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const dialogVisible = ref(false)
const formTitle = ref('新增标签')
const form = ref({ id: 0, name: '', color: '#409EFF' })

async function fetchData() {
  loading.value = true
  try {
    const res: any = await request.get('/tags', { params: { page: page.value, page_size: pageSize.value } })
    tags.value = res.data?.data || res.data || []
    total.value = res.data?.total || 0
  } catch {
    tags.value = []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  formTitle.value = '新增标签'
  form.value = { id: 0, name: '', color: '#409EFF' }
  dialogVisible.value = true
}

function openEdit(row: Tag) {
  formTitle.value = '编辑标签'
  form.value = { id: row.id, name: row.name, color: row.color }
  dialogVisible.value = true
}

async function handleSave() {
  try {
    if (form.value.id) {
      await request.put(`/tags/${form.value.id}`, { name: form.value.name, color: form.value.color })
      ElMessage.success('更新成功')
    } else {
      await request.post('/tags', { name: form.value.name, color: form.value.color })
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
    await ElMessageBox.confirm('确认删除该标签吗？', '提示')
    await request.delete(`/tags/${id}`)
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
      <h2>标签管理</h2>
      <el-button type="primary" @click="openAdd">新增标签</el-button>
    </div>

    <el-card>
      <el-table :data="tags" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column label="颜色" width="120">
          <template #default="{ row }">
            <el-tag :style="{ backgroundColor: row.color, borderColor: row.color, color: '#fff' }">{{ row.color }}</el-tag>
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
        <el-form-item label="颜色">
          <el-color-picker v-model="form.color" />
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
