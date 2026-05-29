<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

interface BotConfig {
  channel: string
  welcome_msg: string
  signature: string
  enabled: number
}

const channel = ref('lark')
const config = ref<BotConfig>({ channel: 'lark', welcome_msg: '', signature: '', enabled: 1 })
const loading = ref(false)
const saving = ref(false)
const changed = ref(false)

async function fetchConfig(ch: string) {
  loading.value = true
  try {
    const res: any = await request.get(`/bot-config/${ch}`)
    config.value = res.data || { channel: ch, welcome_msg: '', signature: '', enabled: 1 }
    changed.value = false
  } catch {
    config.value = { channel: ch, welcome_msg: '', signature: '', enabled: 1 }
  } finally {
    loading.value = false
  }
}

watch(channel, (val) => {
  fetchConfig(val)
})

async function handleSave() {
  saving.value = true
  try {
    await request.post('/bot-config', config.value)
    ElMessage.success('保存成功')
    changed.value = false
  } catch {
    // handled by interceptor
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div>
    <div class="page-header">
      <h2>机器人配置</h2>
    </div>

    <el-card>
      <div class="config-section">
        <el-form :model="config" label-width="120px">
          <el-form-item label="通道">
            <el-select v-model="channel" style="width: 300px">
              <el-option label="飞书" value="lark" />
              <el-option label="钉钉" value="dingtalk" />
              <el-option label="企业微信" value="wecom" />
            </el-select>
          </el-form-item>
        </el-form>
      </div>

      <el-divider />

      <div v-loading="loading" class="config-section">
        <el-form :model="config" label-width="120px">
          <el-form-item label="通道">
            <el-select v-model="config.channel" disabled style="width: 300px">
              <el-option label="飞书" value="lark" />
              <el-option label="钉钉" value="dingtalk" />
              <el-option label="企业微信" value="wecom" />
            </el-select>
          </el-form-item>
          <el-form-item label="欢迎消息">
            <el-input v-model="config.welcome_msg" type="textarea" :rows="4" placeholder="欢迎消息内容" />
          </el-form-item>
          <el-form-item label="签名">
            <el-input v-model="config.signature" type="textarea" :rows="3" placeholder="签名内容" />
          </el-form-item>
          <el-form-item label="启用">
            <el-switch v-model="config.enabled" :active-value="1" :inactive-value="0" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="saving" @click="handleSave">保存配置</el-button>
          </el-form-item>
        </el-form>
      </div>
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
.config-section {
  max-width: 600px;
}
</style>
