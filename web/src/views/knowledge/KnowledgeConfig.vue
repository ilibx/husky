<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

type VectorProvider = 'postgresql' | 'neo4j'

interface VectorSource {
  provider: VectorProvider
  host: string
  port: string
  user: string
  password: string
  database: string
  sslmode: string
  label: string
}

const sources = ref<VectorSource[]>([])
const sourceDialog = ref(false)
const editingSource = ref(false)
const sourceForm = reactive<VectorSource>({
  provider: 'postgresql', host: 'localhost', port: '5432',
  user: 'postgres', password: '', database: 'husky', sslmode: 'disable',
  label: '',
})

const vectorKeyMap: Record<string, string> = {
  provider: 'vector_provider', host: 'vector_host', port: 'vector_port',
  user: 'vector_user', password: 'vector_password', database: 'vector_database', sslmode: 'vector_sslmode',
}

async function fetchSources() {
  try {
    const res: any = await request.get('/system-config/vector')
    if (res.data) {
      sourceForm.provider = res.data.provider || 'postgresql'
      sourceForm.host = res.data.host || 'localhost'
      sourceForm.port = res.data.port || '5432'
      sourceForm.user = res.data.user || 'postgres'
      sourceForm.password = res.data.password || ''
      sourceForm.database = res.data.database || 'husky'
      sourceForm.sslmode = res.data.sslmode || 'disable'
    }
  } catch { /* default */ }
  try {
    const res: any = await request.get('/system-config', { params: { page: 1, page_size: 100 } })
    const list: any[] = res.data?.data || []
    sources.value = list
      .filter((c: any) => c.category === 'vector' && c.key.startsWith('source:'))
      .map((c: any) => {
        try { return { label: c.key.replace('source:', ''), ...JSON.parse(c.value) } }
        catch { return null }
      })
      .filter(Boolean)
  } catch { /* ignore */ }
}

async function handleSaveSource() {
  try {
    const promises = Object.entries(vectorKeyMap).map(([field, key]) =>
      request.post('/system-config/upsert', {
        category: 'vector', key, value: String((sourceForm as any)[field]),
      })
    )
    await Promise.all(promises)
    ElMessage.success('向量库配置已保存')
    sourceDialog.value = false
    fetchSources()
  } catch { /* handled */ }
}

onMounted(fetchSources)
</script>

<template>
  <div>
    <div class="page-header">
      <h2>向量库配置</h2>
    </div>

    <el-card>
      <div class="config-section">
        <h3>当前向量库</h3>
        <el-form :model="sourceForm" label-width="120px">
          <el-form-item label="类型">
            <el-select v-model="sourceForm.provider" style="width:300px">
              <el-option label="PGVector" value="postgresql" />
              <el-option label="Neo4j" value="neo4j" />
            </el-select>
          </el-form-item>
          <el-form-item label="主机"><el-input v-model="sourceForm.host" style="width:300px" placeholder="localhost" /></el-form-item>
          <el-form-item label="端口"><el-input v-model="sourceForm.port" style="width:200px" placeholder="5432" /></el-form-item>
          <el-form-item label="用户名"><el-input v-model="sourceForm.user" style="width:300px" placeholder="postgres" /></el-form-item>
          <el-form-item label="密码"><el-input v-model="sourceForm.password" type="password" show-password style="width:300px" placeholder="数据库密码" /></el-form-item>
          <el-form-item label="数据库名"><el-input v-model="sourceForm.database" style="width:300px" placeholder="husky" /></el-form-item>
          <el-form-item label="SSL 模式">
            <el-select v-model="sourceForm.sslmode" style="width:200px">
              <el-option label="disable" value="disable" />
              <el-option label="require" value="require" />
              <el-option label="verify-ca" value="verify-ca" />
              <el-option label="verify-full" value="verify-full" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSaveSource">保存配置</el-button>
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
  max-width: 640px;
}
.config-section h3 {
  margin: 0 0 16px 0;
  font-weight: 500;
  color: #333;
}
</style>
