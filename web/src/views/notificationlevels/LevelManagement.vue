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
      <div>
        <h2><el-icon><Bell /></el-icon> 等级管理</h2>
        <p>配置通知预警等级、颜色和启用状态，用于 SLA 与推送规则匹配。</p>
      </div>
      <div class="header-actions">
        <el-button @click="add">添加等级</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存配置</el-button>
      </div>
    </div>

    <el-card class="level-card" shadow="never">
      <el-table
        :data="levels"
        v-loading="loading"
        row-key="key"
        class="level-table"
        style="width:100%"
      >
        <el-table-column label="等级" min-width="240">
          <template #default="{ row }">
            <div class="level-name-cell">
              <span class="color-dot" :style="{ backgroundColor: row.color }"></span>
              <el-input v-model="row.name" placeholder="如 紧急" @input="autoKey(row)" />
            </div>
          </template>
        </el-table-column>
        <el-table-column label="标识" min-width="220">
          <template #default="{ row }">
            <el-input v-model="row.key" placeholder="如 critical" />
          </template>
        </el-table-column>
        <el-table-column label="颜色" width="180" align="center">
          <template #default="{ row }">
            <div class="color-cell">
              <el-color-picker v-model="row.color" />
              <span>{{ row.color }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="预览" width="140" align="center">
          <template #default="{ row }">
            <el-tag :color="row.color" class="preview-tag" effect="dark">{{ row.name || row.key || '未命名' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="120" align="center">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" active-text="启用" inactive-text="停用" inline-prompt />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="110" align="center" fixed="right">
          <template #default="{ $index }">
            <el-button link type="danger" :disabled="levels.length <= 1" @click="remove($index)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<style scoped>
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}
.page-header h2 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
}
.page-header p {
  margin: 8px 0 0;
  color: #8c8c8c;
  font-size: 13px;
}
.header-actions {
  display: flex;
  gap: 10px;
  flex-shrink: 0;
}
.level-card {
  border-radius: 12px;
}
.level-table :deep(.el-table__header th) {
  background: #f7f9fc;
  color: #606266;
  font-weight: 600;
}
.level-table :deep(.el-table__row) {
  height: 64px;
}
.level-table :deep(.el-input__wrapper) {
  box-shadow: none;
  background: #f7f9fc;
}
.level-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}
.color-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  box-shadow: 0 0 0 4px rgba(64, 158, 255, 0.08);
  flex-shrink: 0;
}
.color-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #909399;
  font-size: 12px;
}
.preview-tag {
  border: none;
  min-width: 64px;
}
</style>
