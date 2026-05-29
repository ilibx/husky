<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

interface SOP {
  id: number
  name: string
  description: string
  version: string
  trigger_type: string
  status: number
  created_at: string
  steps: string
}

const list = ref<SOP[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const dialogVisible = ref(false)
const isEdit = ref(false)
const form = ref({
  id: 0,
  name: '',
  description: '',
  version: '1.0',
  trigger_type: 'ticket_created',
  trigger_config: '',
  steps: '[]',
  status: 1,
})

const stepEditorVisible = ref(false)
const editingSteps = ref('[]')

async function fetchData() {
  loading.value = true
  try {
    const res: any = await request.get('/sops', { params: { page: page.value, page_size: pageSize.value } })
    list.value = res.data?.list || res.data || []
    total.value = res.data?.total || res.total || 0
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  isEdit.value = false
  form.value = { id: 0, name: '', description: '', version: '1.0', trigger_type: 'ticket_created', trigger_config: '', steps: '[]', status: 1 }
  editingSteps.value = '[]'
  dialogVisible.value = true
}

function openEdit(row: SOP) {
  isEdit.value = true
  form.value = { id: row.id, name: row.name, description: row.description, version: row.version, trigger_type: row.trigger_type, trigger_config: '', steps: row.steps, status: row.status }
  editingSteps.value = row.steps
  dialogVisible.value = true
}

function editSteps() {
  editingSteps.value = form.value.steps
  stepEditorVisible.value = true
}

function saveSteps() {
  try {
    JSON.parse(editingSteps.value)
    form.value.steps = editingSteps.value
    stepEditorVisible.value = false
  } catch {
    ElMessage.error('步骤格式不是合法的 JSON')
  }
}

async function handleSave() {
  try {
    if (isEdit.value) {
      await request.put(`/sops/${form.value.id}`, form.value)
      ElMessage.success('更新成功')
    } else {
      await request.post('/sops', form.value)
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
    await ElMessageBox.confirm('确认删除该 SOP 吗？', '提示')
    await request.delete(`/sops/${id}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch {
    // cancelled or error
  }
}

function formatSteps(steps: string): string {
  try {
    const arr = JSON.parse(steps)
    return arr.map((s: any) => s.name || s.type).join(' → ')
  } catch {
    return steps
  }
}

onMounted(fetchData)
</script>

<template>
  <div>
    <div class="page-header">
      <h2>SOP 管理</h2>
      <el-button type="primary" @click="openAdd">新增 SOP</el-button>
    </div>

    <el-card>
      <el-table :data="list" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="名称" min-width="140" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="version" label="版本" width="80" />
        <el-table-column label="步骤" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">{{ formatSteps(row.steps) }}</template>
        </el-table-column>
        <el-table-column prop="trigger_type" label="触发类型" width="120" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : row.status === 0 ? 'info' : 'warning'">
              {{ row.status === 1 ? '激活' : row.status === 0 ? '草稿' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑 SOP' : '新增 SOP'" width="700px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="版本">
              <el-input v-model="form.version" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="触发类型">
              <el-select v-model="form.trigger_type">
                <el-option label="工单创建" value="ticket_created" />
                <el-option label="事件" value="event" />
                <el-option label="手动" value="manual" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="0">草稿</el-radio>
            <el-radio :value="1">激活</el-radio>
            <el-radio :value="2">停用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="步骤配置">
          <el-input v-model="form.steps" type="textarea" :rows="4" />
          <el-button size="small" style="margin-top: 8px" @click="editSteps">编辑步骤</el-button>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="stepEditorVisible" title="编辑步骤" width="600px">
      <el-input v-model="editingSteps" type="textarea" :rows="12" />
      <template #footer>
        <el-button @click="stepEditorVisible = false">取消</el-button>
        <el-button type="primary" @click="saveSteps">保存步骤</el-button>
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
