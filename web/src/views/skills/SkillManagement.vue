<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

interface Skill {
  id: number
  name: string
  description: string
  category: string
  enabled: boolean
  created_at: string
}

const list = ref<Skill[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const keyword = ref('')

const dialogVisible = ref(false)
const formTitle = ref('新增 Skill')
const form = ref({
  id: 0,
  name: '',
  description: '',
  category: '',
  enabled: true,
})

async function fetchData() {
  loading.value = true
  try {
    const params: any = { page: page.value, page_size: pageSize.value }
    if (keyword.value) params.keyword = keyword.value
    const res: any = await request.get('/skills', { params })
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
  formTitle.value = '新增 Skill'
  form.value = { id: 0, name: '', description: '', category: '', enabled: true }
  dialogVisible.value = true
}

function openEdit(row: Skill) {
  formTitle.value = '编辑 Skill'
  form.value = { ...row }
  dialogVisible.value = true
}

async function handleSave() {
  try {
    if (form.value.id) {
      await request.put(`/skills/${form.value.id}`, form.value)
      ElMessage.success('更新成功')
    } else {
      await request.post('/skills', form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch { /* handled */ }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确认删除该 Skill？', '提示')
    await request.delete(`/skills/${id}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch { /* cancelled */ }
}

onMounted(fetchData)
</script>

<template>
  <div>
    <div class="page-header">
      <h2>Skill 管理</h2>
      <el-button type="primary" @click="openAdd">新增 Skill</el-button>
    </div>

    <el-card>
      <div style="margin-bottom:16px;display:flex;gap:12px">
        <el-input v-model="keyword" placeholder="搜索 Skill" style="width:300px" clearable @keyup.enter="onSearch" />
        <el-button type="primary" @click="onSearch">搜索</el-button>
      </div>

      <el-table :data="list" v-loading="loading" stripe border style="width:100%">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="名称" min-width="150" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="category" label="分类" width="120" />
        <el-table-column prop="enabled" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'warning'">{{ row.enabled ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="180" fixed="right">
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

    <el-dialog v-model="dialogVisible" :title="formTitle" width="550px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="form.name" placeholder="如 网络故障排查" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="技能详细描述" />
        </el-form-item>
        <el-form-item label="分类">
          <el-input v-model="form.category" placeholder="如 硬件、软件、网络" />
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
