<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

interface ChannelConfig {
  id: number
  name: string
  type: string
  enabled: boolean
  status: number
  app_id: string
  webhook_url: string
  welcome_msg: string
  signature: string
  category_id: number | null
  created_at: string
}

const configs = ref<ChannelConfig[]>([])
const configLoading = ref(false)
const configDialog = ref(false)
const isEditConfig = ref(false)
const emptyForm = () => ({
  name: '', type: 'lark', app_id: '', app_secret: '', agent_id: '',
  webhook_url: '', verify_token: '', encrypt_key: '',
  welcome_msg: '', signature: '', category_id: null, enabled: true
})
const configForm = ref<any>(emptyForm())

const isBot = computed(() => configForm.value.type !== 'email')

async function fetchConfigs() {
  configLoading.value = true
  try {
    const res: any = await request.get('/channels')
    configs.value = res.data?.data || []
  } catch { configs.value = [] }
  finally { configLoading.value = false }
}

function openAddConfig() {
  isEditConfig.value = false
  configForm.value = emptyForm()
  configDialog.value = true
}

function openEditConfig(row: ChannelConfig) {
  isEditConfig.value = true
  configForm.value = { ...row }
  configDialog.value = true
}

async function saveConfig() {
  try {
    if (isEditConfig.value) {
      await request.put(`/channels/${configForm.value.id}`, configForm.value)
      ElMessage.success('更新成功')
    } else {
      await request.post('/channels', configForm.value)
      ElMessage.success('创建成功')
    }
    configDialog.value = false
    fetchConfigs()
  } catch { /* handled */ }
}

async function deleteConfig(id: number) {
  try {
    await ElMessageBox.confirm('确认删除？', '提示')
    await request.delete(`/channels/${id}`)
    ElMessage.success('删除成功')
    fetchConfigs()
  } catch { /* cancelled */ }
}

const typeMap: Record<string, string> = { lark: '飞书', dingtalk: '钉钉', wecom: '企微', email: '邮件' }

const categories = ref<{ id: number; name: string }[]>([])
async function fetchCategories() {
  try {
    const res: any = await request.get('/categories')
    categories.value = res.data?.data || []
  } catch { categories.value = [] }
}

onMounted(() => {
  fetchConfigs()
  fetchCategories()
})
</script>

<template>
  <div>
    <div class="page-header">
      <h2><el-icon><Connection /></el-icon> 渠道列表</h2>
    </div>

    <el-card>
      <div class="tab-toolbar">
        <el-button type="primary" @click="openAddConfig">新增渠道</el-button>
      </div>
      <el-table :data="configs" v-loading="configLoading" stripe style="width: 100%">
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag>{{ typeMap[row.type] || row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="app_id" label="AppID" min-width="160" />
        <el-table-column prop="webhook_url" label="Webhook" min-width="200" show-overflow-tooltip />
        <el-table-column prop="enabled" label="启用" width="80">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170" />
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEditConfig(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="deleteConfig(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-dialog v-model="configDialog" :title="isEditConfig ? '编辑渠道' : '新增渠道'" width="600px">
        <el-form :model="configForm" label-width="120px">
          <el-form-item label="名称">
            <el-input v-model="configForm.name" />
          </el-form-item>
          <el-form-item label="类型">
            <el-select v-model="configForm.type">
              <el-option label="飞书" value="lark" />
              <el-option label="钉钉" value="dingtalk" />
              <el-option label="企微" value="wecom" />
              <el-option label="邮件" value="email" />
            </el-select>
          </el-form-item>
          <el-form-item label="启用">
            <el-switch v-model="configForm.enabled" />
          </el-form-item>
          <template v-if="isBot">
            <el-form-item label="职责分类">
              <el-select v-model="configForm.category_id" placeholder="选择分类（可选）" clearable>
                <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.id" />
              </el-select>
              <div class="form-help">关联分类后，该渠道只处理对应分类的工单</div>
            </el-form-item>
            <el-form-item label="欢迎语">
              <el-input v-model="configForm.welcome_msg" type="textarea" :rows="3" placeholder="新对话时自动发送的欢迎消息" />
            </el-form-item>
            <el-form-item label="签名">
              <el-input v-model="configForm.signature" type="textarea" :rows="2" placeholder="消息末尾附带的签名" />
            </el-form-item>
            <el-form-item label="AppID">
              <el-input v-model="configForm.app_id" />
            </el-form-item>
            <el-form-item label="AppSecret">
              <el-input v-model="configForm.app_secret" type="password" show-password />
            </el-form-item>
            <el-form-item label="AgentID">
              <el-input v-model="configForm.agent_id" />
            </el-form-item>
            <el-form-item label="Webhook URL">
              <el-input v-model="configForm.webhook_url" />
            </el-form-item>
            <el-form-item label="VerifyToken">
              <el-input v-model="configForm.verify_token" />
            </el-form-item>
            <el-form-item label="EncryptKey">
              <el-input v-model="configForm.encrypt_key" type="password" show-password />
            </el-form-item>
          </template>
          <template v-else>
            <el-form-item label="分类">
              <el-select v-model="configForm.category_id" placeholder="选择分类（可选）" clearable>
                <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="签名">
              <el-input v-model="configForm.signature" type="textarea" :rows="2" placeholder="邮件末尾附带的签名" />
            </el-form-item>
          </template>
        </el-form>
        <template #footer>
          <el-button @click="configDialog = false">取消</el-button>
          <el-button type="primary" @click="saveConfig">保存</el-button>
        </template>
      </el-dialog>
    </el-card>
  </div>
</template>

<style scoped>
.tab-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.form-help {
  font-size: 12px;
  color: var(--text-muted, #94a3b8);
  margin-top: 4px;
}
</style>
