<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

interface TicketField {
  id: number
  name: string
  field_key: string
  field_type: string
  options: string
  required: number
  sort_order: number
  placeholder: string
  enabled: number
  created_at: string
}

const fields = ref<TicketField[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const formTitle = ref('新增字段')
const form = ref<TicketField>({
  id: 0, name: '', field_key: '', field_type: 'text', options: '',
  required: 0, sort_order: 0, placeholder: '', enabled: 1, created_at: ''
})

async function fetchData() {
  loading.value = true
  try {
    const res: any = await request.get('/tickets/fields/definitions')
    fields.value = res.data || []
  } catch {
    fields.value = []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  formTitle.value = '新增字段'
  form.value = {
    id: 0, name: '', field_key: '', field_type: 'text', options: '',
    required: 0, sort_order: 0, placeholder: '', enabled: 1, created_at: ''
  }
  dialogVisible.value = true
}

function openEdit(row: TicketField) {
  formTitle.value = '编辑字段'
  form.value = { ...row }
  dialogVisible.value = true
}

async function handleSave() {
  try {
    const payload = {
      name: form.value.name, field_key: form.value.field_key,
      field_type: form.value.field_type, options: form.value.options,
      required: form.value.required, sort_order: form.value.sort_order,
      placeholder: form.value.placeholder, enabled: form.value.enabled
    }
    if (form.value.id) {
      await request.put(`/tickets/fields/definitions/${form.value.id}`, payload)
      ElMessage.success('更新成功')
    } else {
      await request.post('/tickets/fields/definitions', payload)
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
    await ElMessageBox.confirm('确认删除该字段吗？', '提示')
    await request.delete(`/tickets/fields/definitions/${id}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch {
    // cancelled or error
  }
}

const fieldTypeTag = (type: string) => {
  const map: Record<string, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = { text: 'info', textarea: 'success', select: 'warning', number: 'info', date: 'primary', boolean: 'danger' }
  return map[type] || 'info'
}

onMounted(fetchData)
</script>

<template>
  <div>
    <div class="page-header">
      <h2>工单字段管理</h2>
      <el-button type="primary" @click="openAdd">新增字段</el-button>
    </div>

    <el-card>
      <el-table :data="fields" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="field_key" label="字段标识" min-width="130" />
        <el-table-column prop="field_type" label="字段类型" width="110">
          <template #default="{ row }">
            <el-tag :type="fieldTypeTag(row.field_type)">{{ row.field_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="required" label="必填" width="70">
          <template #default="{ row }">
            <el-tag :type="row.required === 1 ? 'danger' : 'info'">{{ row.required === 1 ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sort_order" label="排序" width="70" />
        <el-table-column prop="enabled" label="启用" width="70">
          <template #default="{ row }">
            <el-tag :type="row.enabled === 1 ? 'success' : 'warning'">{{ row.enabled === 1 ? '是' : '否' }}</el-tag>
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
    </el-card>

    <el-dialog v-model="dialogVisible" :title="formTitle" width="550px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="字段标识">
          <el-input v-model="form.field_key" />
        </el-form-item>
        <el-form-item label="字段类型">
          <el-select v-model="form.field_type" style="width: 100%">
            <el-option label="文本" value="text" />
            <el-option label="多行文本" value="textarea" />
            <el-option label="下拉选择" value="select" />
            <el-option label="数字" value="number" />
            <el-option label="日期" value="date" />
            <el-option label="布尔" value="boolean" />
          </el-select>
        </el-form-item>
        <el-form-item label="选项" v-if="form.field_type === 'select'">
          <el-input v-model="form.options" type="textarea" :rows="4" placeholder="每行一个选项" />
        </el-form-item>
        <el-form-item label="占位文本">
          <el-input v-model="form.placeholder" />
        </el-form-item>
        <el-form-item label="排序号">
          <el-input-number v-model="form.sort_order" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="必填">
          <el-switch v-model="form.required" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" :active-value="1" :inactive-value="0" />
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
