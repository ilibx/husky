<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'
import MarkdownEditor from '@/components/MarkdownEditor.vue'

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

const currentView = ref<'list' | 'edit'>('list')
const isEdit = ref(false)
const activeTab = ref('params')
const form = ref({
  id: 0,
  name: '',
  description: '',
  category: '',
  enabled: true,
})

const fetchData = async () => {
  loading.value = true
  try {
    const params: any = { page: page.value, page_size: pageSize.value }
    if (keyword.value) params.keyword = keyword.value
    const res: any = await request.get('/skills', { params })
    list.value = res.data?.data || res.data || []
    total.value = res.data?.total || 0
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  isEdit.value = false
  form.value = { id: 0, name: '', description: '', category: '', enabled: true }
  activeTab.value = 'params'
  currentView.value = 'edit'
}

function openEdit(row: Skill) {
  isEdit.value = true
  form.value = { id: row.id, name: row.name, description: row.description, category: row.category, enabled: row.enabled }
  activeTab.value = 'params'
  currentView.value = 'edit'
}

function goBack() {
  currentView.value = 'list'
  fetchData()
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
    goBack()
  } catch {
    // handled by interceptor
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确认删除该 Skill 吗？', '提示')
    await request.delete(`/skills/${id}`)
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
    <template v-if="currentView === 'list'">
      <div class="page-header">
        <h2>Skill 管理</h2>
        <el-button type="primary" @click="openAdd">新增 Skill</el-button>
      </div>
      <el-card>
        <div style="margin-bottom: 16px">
          <el-input v-model="keyword" placeholder="搜索技能名称..." style="width: 300px" clearable @input="fetchData" />
        </div>
        <el-table :data="list" v-loading="loading" stripe border style="width: 100%">
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="name" label="名称" min-width="160" />
          <el-table-column prop="description" label="描述" min-width="300" show-overflow-tooltip />
          <el-table-column prop="category" label="分类" width="120" />
          <el-table-column prop="enabled" label="启用" width="70">
            <template #default="{ row }">
              <el-tag :type="row.enabled ? 'success' : 'info'" size="small">{{ row.enabled ? '是' : '否' }}</el-tag>
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
          <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="fetchData" />
        </div>
      </el-card>
    </template>

    <template v-else>
      <div class="page-header">
        <el-button text @click="goBack">
          <el-icon><ArrowLeft /></el-icon> 返回
        </el-button>
        <h2>{{ isEdit ? '编辑 Skill' : '新增 Skill' }}</h2>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </div>
      <el-card>
        <el-tabs v-model="activeTab">
          <el-tab-pane label="基本参数" name="params">
            <el-form :model="form" label-width="120px" class="edit-form">
              <el-form-item label="名称">
                <el-input v-model="form.name" placeholder="如 网络故障排查" />
              </el-form-item>
              <el-form-item label="分类">
                <el-input v-model="form.category" placeholder="如 硬件、软件、网络" />
              </el-form-item>
              <el-form-item label="启用">
                <el-switch v-model="form.enabled" />
              </el-form-item>
            </el-form>
          </el-tab-pane>
          <el-tab-pane label="内容编辑" name="content">
            <el-form :model="form" label-width="120px" class="edit-form">
              <el-form-item label="描述">
                <MarkdownEditor v-model="form.description" placeholder="技能详细描述" />
              </el-form-item>
            </el-form>
          </el-tab-pane>
        </el-tabs>
      </el-card>
    </template>
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
.edit-form {
  max-width: 960px;
  margin-top: 16px;
}
</style>
