<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

interface Level {
  name: string
  key: string
  color: string
  enabled: boolean
}

const levels = ref<Level[]>([])
const loading = ref(false)
const saving = ref(false)

const defaultLevels: Level[] = [
  { name: '紧急', key: 'critical', color: '#F56C6C', enabled: true },
  { name: '警告', key: 'warning', color: '#E6A23C', enabled: true },
  { name: '信息', key: 'info', color: '#409EFF', enabled: true },
]

async function fetchLevels() {
  loading.value = true
  try {
    const res: any = await request.get('/system-config/lookup', { params: { category: 'push', key: 'levels' } })
    if (res.data?.value) {
      levels.value = JSON.parse(res.data.value)
    } else {
      levels.value = defaultLevels.map(l => ({ ...l }))
    }
  } catch {
    levels.value = defaultLevels.map(l => ({ ...l }))
  } finally { loading.value = false }
}

async function save() {
  saving.value = true
  try {
    await request.post('/system-config/upsert', {
      category: 'push', key: 'levels', value: JSON.stringify(levels.value),
    })
    ElMessage.success('等级配置已保存')
  } catch { /* handled */ } finally { saving.value = false }
}

function add() {
  const n = levels.value.length + 1
  levels.value.push({ name: '', key: `level_${n}`, color: '#909399', enabled: true })
}

function autoKey(row: { name: string; key: string }) {
  if (!row.key) row.key = row.name.toLowerCase().replace(/\s+/g, '_')
}

function remove(i: number) {
  levels.value.splice(i, 1)
}

onMounted(fetchLevels)
</script>

<template>
  <div>
    <div class="page-header">
      <h2><el-icon><Bell /></el-icon> 等级管理</h2>
    </div>

    <el-card>
      <div class="tab-toolbar">
        <el-button size="small" @click="add">添加等级</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </div>

      <el-table :data="levels" v-loading="loading" stripe style="width:100%">
        <el-table-column label="等级名称" width="200">
          <template #default="{ row, $index }">
            <el-input v-model="row.name" placeholder="如 紧急" size="small" @input="autoKey(row)" />
          </template>
        </el-table-column>
        <el-table-column label="标识" width="200">
          <template #default="{ row, $index }">
            <el-input v-model="row.key" placeholder="如 critical" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="颜色" width="120">
          <template #default="{ row, $index }">
            <el-color-picker v-model="row.color" />
          </template>
        </el-table-column>
        <el-table-column label="启用" width="100">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row, $index }">
            <el-button size="small" type="danger" :disabled="levels.length <= 1" @click="remove($index)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
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
