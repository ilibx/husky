<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

const loading = ref(false)
const result = ref<{ total: number; created: number; errors: number } | null>(null)

async function handleSync() {
  loading.value = true
  result.value = null
  try {
    const res: any = await request.post('/users/sync-ldap')
    result.value = res.data || res
    ElMessage.success('同步完成')
  } catch {
    result.value = null
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <div class="page-header">
      <h2>LDAP 同步</h2>
    </div>

    <el-card class="config-card">
      <template #header>
        <span>环境变量说明</span>
      </template>
      <div class="env-hints">
        <p>系统通过以下环境变量连接 LDAP 服务器：</p>
        <ul>
          <li><code>LDAP_HOST</code> — LDAP 服务器地址</li>
          <li><code>LDAP_BIND_DN</code> — 绑定 DN</li>
          <li><code>LDAP_BASE_DN</code> — 搜索基 DN</li>
        </ul>
        <p class="hint-note">请确保这些环境变量已在服务端正确配置。</p>
      </div>
    </el-card>

    <el-card class="sync-card">
      <div class="sync-section">
        <el-button type="primary" size="large" :loading="loading" @click="handleSync" style="padding: 16px 48px; font-size: 18px;">
          {{ loading ? '同步中...' : '立即同步' }}
        </el-button>
      </div>

      <div v-if="result" class="result-section">
        <el-divider />
        <h3>同步结果</h3>
        <el-row :gutter="24">
          <el-col :span="8">
            <el-statistic title="总用户" :value="result.total" />
          </el-col>
          <el-col :span="8">
            <el-statistic title="新增" :value="result.created" />
          </el-col>
          <el-col :span="8">
            <el-statistic title="错误" :value="result.errors" />
          </el-col>
        </el-row>
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
.config-card {
  margin-bottom: 16px;
}
.env-hints ul {
  padding-left: 20px;
}
.env-hints li {
  margin-bottom: 8px;
}
.hint-note {
  color: #909399;
  font-size: 13px;
  margin-top: 12px;
}
.sync-section {
  display: flex;
  justify-content: center;
  padding: 32px 0;
}
.result-section {
  text-align: center;
}
</style>
