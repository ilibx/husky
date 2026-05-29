<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

interface WebhookConfig {
  id: number
  name: string
  url: string
  secret: string
  events: string
  enabled: boolean
  created_at: string
}

const webhookConfigs = ref<WebhookConfig[]>([])
const loading = ref(false)

const dialogVisible = ref(false)
const formTitle = ref('')
const form = ref<{ id: number; name: string; url: string; secret: string; events: string[]; enabled: boolean }>({
  id: 0,
  name: '',
  url: '',
  secret: '',
  events: [],
  enabled: true,
})

const eventOptions = [
  { label: 'SLA 违约', value: 'sla_breach' },
  { label: 'SLA 警告', value: 'sla_warning' },
  { label: 'SLA 恢复', value: 'sla_restored' },
  { label: '工单创建', value: 'ticket_created' },
  { label: '工单更新', value: 'ticket_updated' },
]

function parseEvents(events: string): string[] {
  try {
    const parsed = JSON.parse(events)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

function formatEventLabel(value: string): string {
  const opt = eventOptions.find(o => o.value === value)
  return opt ? opt.label : value
}

async function fetchData() {
  loading.value = true
  try {
    const res: any = await request.get('/webhook-configs')
    webhookConfigs.value = res.data?.list || res.data || []
  } catch {
    webhookConfigs.value = []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  formTitle.value = '新增 Webhook'
  form.value = { id: 0, name: '', url: '', secret: '', events: [], enabled: true }
  dialogVisible.value = true
}

function openEdit(row: WebhookConfig) {
  formTitle.value = '编辑 Webhook'
  form.value = {
    id: row.id,
    name: row.name,
    url: row.url,
    secret: row.secret,
    events: parseEvents(row.events),
    enabled: row.enabled,
  }
  dialogVisible.value = true
}

async function handleSave() {
  try {
    const body = {
      name: form.value.name,
      url: form.value.url,
      secret: form.value.secret,
      events: JSON.stringify(form.value.events),
      enabled: form.value.enabled,
    }
    if (form.value.id) {
      await request.put(`/webhook-configs/${form.value.id}`, body)
      ElMessage.success('更新成功')
    } else {
      await request.post('/webhook-configs', body)
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
    await ElMessageBox.confirm('确认删除该 Webhook 吗？', '提示')
    await request.delete(`/webhook-configs/${id}`)
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
      <h2>Webhook 配置</h2>
      <el-button type="primary" @click="openAdd">新增 Webhook</el-button>
    </div>

    <el-card>
      <el-table :data="webhookConfigs" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="url" label="URL" min-width="200">
          <template #default="{ row }">
            <el-tooltip :content="row.url" placement="top">
              <span>{{ row.url.length > 40 ? row.url.slice(0, 40) + '...' : row.url }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="events" label="事件" min-width="240">
          <template #default="{ row }">
            <el-tag v-for="evt in parseEvents(row.events)" :key="evt" size="small" style="margin-right: 4px; margin-bottom: 2px;">
              {{ formatEventLabel(evt) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="enabled" label="启用" width="70">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'" size="small">
              {{ row.enabled ? '是' : '否' }}
            </el-tag>
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
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="form.name" placeholder="Webhook 名称" />
        </el-form-item>
        <el-form-item label="URL">
          <el-input v-model="form.url" placeholder="https://example.com/webhook" />
        </el-form-item>
        <el-form-item label="密钥">
          <el-input v-model="form.secret" type="password" show-password placeholder="可选签名密钥" />
        </el-form-item>
        <el-form-item label="事件">
          <el-checkbox-group v-model="form.events">
            <el-checkbox v-for="opt in eventOptions" :key="opt.value" :label="opt.value" :value="opt.value">
              {{ opt.label }}
            </el-checkbox>
          </el-checkbox-group>
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
</style>
