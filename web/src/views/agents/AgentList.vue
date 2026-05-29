<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

interface Agent {
  id: number
  name: string
  description: string
  type: string
  config: string
  model: string
  temperature: number
  max_tokens: number
  enabled: boolean
  created_at: string
}

const agents = ref<Agent[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const dialogVisible = ref(false)
const formTitle = ref('新增Agent')
const form = ref({
  id: 0,
  name: '',
  description: '',
  type: 'llm',
  config: '',
  model: '',
  temperature: 0.7,
  max_tokens: 2048,
  enabled: true,
})

async function fetchData() {
  loading.value = true
  try {
    const res: any = await request.get('/agents', { params: { page: page.value, page_size: pageSize.value } })
    agents.value = res.data?.list || res.data || []
    total.value = res.data?.total || res.total || 0
  } catch {
    agents.value = []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  formTitle.value = '新增Agent'
  form.value = { id: 0, name: '', description: '', type: 'llm', config: '', model: '', temperature: 0.7, max_tokens: 2048, enabled: true }
  dialogVisible.value = true
}

function openEdit(row: Agent) {
  formTitle.value = '编辑Agent'
  form.value = { id: row.id, name: row.name, description: row.description, type: row.type, config: row.config, model: row.model, temperature: row.temperature, max_tokens: row.max_tokens, enabled: row.enabled }
  dialogVisible.value = true
}

async function handleSave() {
  try {
    if (form.value.id) {
      await request.put(`/agents/${form.value.id}`, form.value)
      ElMessage.success('更新成功')
    } else {
      await request.post('/agents', form.value)
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
    await ElMessageBox.confirm('确认删除该Agent吗？', '提示')
    await request.delete(`/agents/${id}`)
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
      <h2>Agent管理</h2>
      <el-button type="primary" @click="openAdd">新增Agent</el-button>
    </div>

    <el-card>
      <el-table :data="agents" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.type === 'llm' ? 'primary' : row.type === 'rule' ? 'warning' : 'success'">{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="model" label="模型" min-width="140" />
        <el-table-column prop="temperature" label="温度" width="80" />
        <el-table-column prop="enabled" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'warning'">{{ row.enabled ? '启用' : '禁用' }}</el-tag>
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

    <el-dialog v-model="dialogVisible" :title="formTitle" width="600px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="form.type">
            <el-option label="LLM" value="llm" />
            <el-option label="规则" value="rule" />
            <el-option label="混合" value="hybrid" />
          </el-select>
        </el-form-item>
        <el-form-item label="配置">
          <el-input v-model="form.config" type="textarea" :rows="4" />
        </el-form-item>
        <el-form-item label="模型">
          <el-input v-model="form.model" />
        </el-form-item>
        <el-form-item label="温度">
          <el-input-number v-model="form.temperature" :min="0" :max="2" :step="0.1" />
        </el-form-item>
        <el-form-item label="最大Token">
          <el-input-number v-model="form.max_tokens" :min="1" :max="65536" />
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
