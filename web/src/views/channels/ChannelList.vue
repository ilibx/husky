<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

interface Channel {
  id: number
  name: string
  type: string
  config: string
  enabled: number
  status: string
  created_at: string
}

interface BotConfig {
  channel: string
  welcome_msg: string
  signature: string
  enabled: number
}

const channels = ref<Channel[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const formTitle = ref('新增通道')
const form = ref<Channel>({ id: 0, name: '', type: 'lark', config: '', enabled: 1, status: '', created_at: '' })
const botForm = ref<BotConfig>({ channel: '', welcome_msg: '', signature: '', enabled: 1 })

async function fetchData() {
  loading.value = true
  try {
    const res: any = await request.get('/channels')
    channels.value = res.data || []
  } catch {
    channels.value = []
  } finally {
    loading.value = false
  }
}

async function openAdd() {
  formTitle.value = '新增通道'
  form.value = { id: 0, name: '', type: 'lark', config: '', enabled: 1, status: '', created_at: '' }
  botForm.value = { channel: 'lark', welcome_msg: '', signature: '', enabled: 1 }
  dialogVisible.value = true
}

async function openEdit(row: Channel) {
  formTitle.value = '编辑通道'
  form.value = { ...row }
  botForm.value = { channel: row.type, welcome_msg: '', signature: '', enabled: 1 }
  // Load existing bot config if any
  try {
    const res: any = await request.get(`/bot-config/${row.type}`)
    if (res.data) {
      botForm.value = res.data
    }
  } catch {
    // use defaults
  }
  dialogVisible.value = true
}

async function handleSave() {
  try {
    const payload = { name: form.value.name, type: form.value.type, config: form.value.config, enabled: form.value.enabled }
    if (form.value.id) {
      await request.put(`/channels/${form.value.id}`, payload)
    } else {
      await request.post('/channels', payload)
    }
    // Save bot config
    botForm.value.channel = form.value.type
    await request.post('/bot-config', botForm.value)
    ElMessage.success('保存成功')
    dialogVisible.value = false
    fetchData()
  } catch {
    // handled by interceptor
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确认删除该通道吗？', '提示')
    await request.delete(`/channels/${id}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch {
    // cancelled or error
  }
}

const typeTag = (type: string) => {
  const map: Record<string, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = { lark: 'primary', dingtalk: 'success', wecom: 'warning' }
  return map[type] || 'info'
}

const typeLabel = (type: string) => {
  const map: Record<string, string> = { lark: '飞书', dingtalk: '钉钉', wecom: '企业微信' }
  return map[type] || type
}

onMounted(fetchData)
</script>

<template>
  <div>
    <div class="page-header">
      <h2>渠道配置</h2>
      <el-button type="primary" @click="openAdd">新增渠道</el-button>
    </div>

    <el-card>
      <el-table :data="channels" v-loading="loading" stripe border style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag :type="typeTag(row.type)">{{ typeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="enabled" label="启用" width="80">
          <template #default="{ row }">
            <el-tag :type="row.enabled === 1 ? 'success' : 'warning'">{{ row.enabled === 1 ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" />
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="formTitle" width="600px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="form.type">
            <el-option label="飞书" value="lark" />
            <el-option label="钉钉" value="dingtalk" />
            <el-option label="企业微信" value="wecom" />
          </el-select>
        </el-form-item>
        <el-form-item label="配置">
          <el-input v-model="form.config" type="textarea" :rows="4" placeholder="JSON 格式配置（可选）" />
        </el-form-item>
        <el-divider>机器人设置</el-divider>
        <el-form-item label="欢迎消息">
          <el-input v-model="botForm.welcome_msg" type="textarea" :rows="3" placeholder="新工单创建时自动发送的欢迎消息" />
        </el-form-item>
        <el-form-item label="签名">
          <el-input v-model="botForm.signature" type="textarea" :rows="2" placeholder="消息末尾自动附加的签名" />
        </el-form-item>
        <el-form-item label="启用">
          <div class="switch-group">
            <span>渠道：</span>
            <el-switch v-model="form.enabled" :active-value="1" :inactive-value="0" />
            <span style="margin-left: 16px;">机器人：</span>
            <el-switch v-model="botForm.enabled" :active-value="1" :inactive-value="0" />
          </div>
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
.switch-group {
  display: flex;
  align-items: center;
}
</style>
