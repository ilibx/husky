<script setup lang="ts">
import { ref } from 'vue'
import request from '@/api/request'

const queryText = ref('')
const queryLoading = ref(false)
const queryResults = ref<any[]>([])
const queryCategory = ref('')
const queryLimit = ref(5)

async function handleQuery() {
  if (!queryText.value.trim()) return
  queryLoading.value = true
  try {
    const res: any = await request.post('/knowledge/search', {
      query: queryText.value,
      category: queryCategory.value || undefined,
      limit: queryLimit.value,
    })
    queryResults.value = res.data?.data || res.data?.results || []
  } catch {
    queryResults.value = []
  } finally {
    queryLoading.value = false
  }
}

const queryQR = ref('')
const queryQALoading = ref(false)
const queryQAResult = ref('')

async function handleAsk() {
  if (!queryQR.value.trim()) return
  queryQALoading.value = true
  queryQAResult.value = ''
  try {
    const res: any = await request.post('/knowledge/ask', { query: queryQR.value })
    queryQAResult.value = res.data?.answer || res.data?.content || JSON.stringify(res.data)
  } catch {
    queryQAResult.value = '查询失败'
  } finally {
    queryQALoading.value = false
  }
}
</script>

<template>
  <div>
    <div class="page-header">
      <h2>资料检索</h2>
    </div>

    <el-card>
      <div style="display:flex;flex-direction:column;gap:24px">
        <div>
          <h3 style="margin:0 0 12px 0">语义搜索</h3>
          <div style="display:flex;gap:12px;margin-bottom:12px">
            <el-input v-model="queryText" placeholder="输入搜索关键词..." style="flex:1" clearable @keyup.enter="handleQuery" />
            <el-input-number v-model="queryLimit" :min="1" :max="20" style="width:100px" />
            <el-button type="primary" :loading="queryLoading" @click="handleQuery">搜索</el-button>
          </div>

          <div v-if="queryResults.length" style="display:flex;flex-direction:column;gap:12px">
            <el-card v-for="(item, idx) in queryResults" :key="idx" shadow="hover">
              <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:8px">
                <strong>{{ item.title || item.name || '未知' }}</strong>
                <el-tag v-if="item.score !== undefined" size="small">{{ (item.score * 100).toFixed(1) }}%</el-tag>
              </div>
              <p style="color:#666;font-size:13px;margin:0;line-height:1.6;white-space:pre-wrap">{{ item.content?.slice(0, 300) }}</p>
            </el-card>
          </div>
          <el-empty v-else-if="queryText && !queryLoading" description="无结果" />
        </div>

        <el-divider />

        <div>
          <h3 style="margin:0 0 12px 0">知识问答</h3>
          <div style="display:flex;gap:12px;margin-bottom:12px">
            <el-input v-model="queryQR" placeholder="输入问题..." style="flex:1" clearable @keyup.enter="handleAsk" />
            <el-button type="success" :loading="queryQALoading" @click="handleAsk">提问</el-button>
          </div>
          <div v-if="queryQAResult" class="qa-result">
            <div class="qa-result-content">{{ queryQAResult }}</div>
          </div>
        </div>
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
.qa-result {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 16px;
  line-height: 1.8;
}
.qa-result-content {
  white-space: pre-wrap;
}
</style>
