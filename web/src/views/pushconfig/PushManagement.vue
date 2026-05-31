<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

interface PushMethod {
  name: string
  type: string
  level: string
  template: string
  enabled: boolean
  params: Record<string, string>
}

const methods = ref<PushMethod[]>([])
const loading = ref(false)
const dialog = ref(false)
const editing = ref(false)
const form = ref<PushMethod>({
  name: '', type: 'webhook', level: 'warning', template: '', enabled: true, params: {},
})

const isWebhook = computed(() => form.value.type === 'webhook')

const typeOptions = [
  { label: 'Webhook', value: 'webhook' },
  { label: '邮件', value: 'email' },
]

const levelOptions = ref<{ name: string; key: string; color: string }[]>([])

async function fetchMethods() {
  loading.value = true
  try {
    const res: any = await request.get('/system-config', { params: { page: 1, page_size: 200 } })
    const list: any[] = res.data?.data || []
    methods.value = list
      .filter((c: any) => c.category === 'push' && c.key.startsWith('method:'))
      .map((c: any) => {
        try { return JSON.parse(c.value) } catch { return null }
      })
      .filter(Boolean)
  } catch { methods.value = [] } finally { loading.value = false }
}

function openAdd() {
  editing.value = false
  form.value = { name: '', type: 'webhook', level: 'warning', template: '', enabled: true, params: {} }
  dialog.value = true
}

function openEdit(m: PushMethod) {
  editing.value = true
  form.value = { ...m, params: { ...m.params } }
  dialog.value = true
}

async function save() {
  if (!form.value.name) { ElMessage.warning('请输入推送方式名称'); return }
  try {
    const value = JSON.stringify(form.value)
    await request.post('/system-config/upsert', { category: 'push', key: `method:${form.value.name}`, value })
    ElMessage.success(editing.value ? '已更新' : '已创建')
    dialog.value = false
    fetchMethods()
  } catch { /* handled */ }
}

async function remove(name: string) {
  try {
    await ElMessageBox.confirm(`确认删除推送方式 "${name}"？`, '提示')
    const res: any = await request.get('/system-config', { params: { page: 1, page_size: 200 } })
    const list: any[] = res.data?.data || []
    const cfg = list.find((c: any) => c.key === `method:${name}`)
    if (cfg) await request.delete(`/system-config/${cfg.id}`)
    ElMessage.success('已删除')
    fetchMethods()
  } catch { /* cancelled */ }
}

async function toggle(m: PushMethod) {
  try {
    const value = JSON.stringify(m)
    await request.post('/system-config/upsert', { category: 'push', key: `method:${m.name}`, value })
  } catch { m.enabled = !m.enabled }
}

async function fetchLevels() {
  try {
    const res: any = await request.get('/system-config/lookup', { params: { category: 'push', key: 'levels' } })
    if (res.data?.value) levelOptions.value = JSON.parse(res.data.value)
  } catch { levelOptions.value = [] }
}

onMounted(() => {
  fetchMethods()
  fetchLevels()
})
</script>

<template>
  <div>
    <div class="page-header">
      <h2><el-icon><Bell /></el-icon> 推送管理</h2>
    </div>

    <el-card>
      <div class="tab-toolbar">
        <el-button type="primary" @click="openAdd">添加推送方式</el-button>
      </div>

      <el-table :data="methods" v-loading="loading" stripe style="width:100%">
        <el-table-column prop="name" label="推送方式名称" width="150" />
        <el-table-column prop="type" label="方式" width="100">
          <template #default="{ row }">
            <el-tag>{{ typeOptions.find(t => t.value === row.type)?.label || row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="通知等级" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.level" :color="levelOptions.find(l => l.key === row.level)?.color || '#909399'" style="color:#fff" size="small">
              {{ levelOptions.find(l => l.key === row.level)?.name || row.level }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="template" label="推送模板" min-width="240" show-overflow-tooltip />
        <el-table-column prop="enabled" label="启用" width="80">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="toggle(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="remove(row.name)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-dialog v-model="dialog" :title="editing ? '编辑推送方式' : '添加推送方式'" width="600px">
        <el-form :model="form" label-width="100px">
          <el-form-item label="推送方式名称">
            <el-input v-model="form.name" placeholder="如 SLA 预警推送" :disabled="editing" />
          </el-form-item>
          <el-form-item label="方式">
            <el-select v-model="form.type" style="width:100%">
              <el-option v-for="t in typeOptions" :key="t.value" :label="t.label" :value="t.value" />
            </el-select>
          </el-form-item>
          <el-form-item label="通知等级">
            <el-select v-model="form.level" style="width:100%">
              <el-option v-for="l in levelOptions" :key="l.key" :label="l.name" :value="l.key" />
            </el-select>
          </el-form-item>
          <el-form-item label="推送模板">
            <el-input v-model="form.template" type="textarea" :rows="3" placeholder="推送消息模板，支持 {ticket_id} {title} {status} 等变量" />
          </el-form-item>

          <!-- Webhook 参数 -->
          <template v-if="isWebhook">
            <el-form-item label="Webhook URL">
              <el-input v-model="form.params.url" placeholder="https://example.com/webhook" />
            </el-form-item>
            <el-form-item label="Secret">
              <el-input v-model="form.params.secret" type="password" show-password placeholder="签名密钥（可选）" />
            </el-form-item>
          </template>

          <!-- 邮件参数 -->
          <template v-else>
            <el-form-item label="SMTP 服务器">
              <el-input v-model="form.params.host" placeholder="smtp.example.com" />
            </el-form-item>
            <el-form-item label="端口">
              <el-input v-model="form.params.port" placeholder="587" />
            </el-form-item>
            <el-form-item label="用户名">
              <el-input v-model="form.params.user" placeholder="notification@example.com" />
            </el-form-item>
            <el-form-item label="密码">
              <el-input v-model="form.params.password" type="password" show-password />
            </el-form-item>
            <el-form-item label="发件地址">
              <el-input v-model="form.params.from" placeholder="notification@example.com" />
            </el-form-item>
            <el-form-item label="使用 TLS">
              <el-switch v-model="form.params.use_tls" :active-value="'true'" :inactive-value="'false'" />
            </el-form-item>
          </template>

          <el-form-item label="启用">
            <el-switch v-model="form.enabled" />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="dialog = false">取消</el-button>
          <el-button type="primary" @click="save">保存</el-button>
        </template>
      </el-dialog>
    </el-card>
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
.tab-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
</style>
