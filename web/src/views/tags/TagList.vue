<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

interface Tag {
  id: number
  name: string
  color: string
  created_at: string
}

const tags = ref<Tag[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const dialogVisible = ref(false)
const formTitle = ref('新增标签')
const form = ref({ id: 0, name: '', color: '#409EFF' })

async function fetchData() {
  loading.value = true
  try {
    const res: any = await request.get('/tags', { params: { page: page.value, page_size: pageSize.value } })
    tags.value = res.data?.data || res.data || []
    total.value = res.data?.total || 0
  } catch {
    tags.value = []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  formTitle.value = '新增标签'
  form.value = { id: 0, name: '', color: '#409EFF' }
  dialogVisible.value = true
}

function openEdit(row: Tag) {
  formTitle.value = '编辑标签'
  form.value = { id: row.id, name: row.name, color: row.color }
  dialogVisible.value = true
}

async function handleSave() {
  try {
    if (form.value.id) {
      await request.put(`/tags/${form.value.id}`, { name: form.value.name, color: form.value.color })
      ElMessage.success('更新成功')
    } else {
      await request.post('/tags', { name: form.value.name, color: form.value.color })
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
    await ElMessageBox.confirm('确认删除该标签吗？', '提示')
    await request.delete(`/tags/${id}`)
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
      <div>
        <h2><el-icon><CollectionTag /></el-icon> 标签管理</h2>
        <p>维护系统标签颜色和名称，用于工单筛选、标记和展示。</p>
      </div>
      <el-button type="primary" @click="openAdd">新增标签</el-button>
    </div>

    <el-card class="tag-card" shadow="never">
      <el-table :data="tags" v-loading="loading" class="tag-table" row-key="id" style="width: 100%">
        <el-table-column label="标签" min-width="220">
          <template #default="{ row }">
            <div class="tag-name-cell">
              <span class="color-dot" :style="{ backgroundColor: row.color }"></span>
              <div>
                <div class="tag-name">{{ row.name }}</div>
                <div class="tag-id">ID: {{ row.id }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="颜色" width="180" align="center">
          <template #default="{ row }">
            <div class="color-cell">
              <span class="color-block" :style="{ backgroundColor: row.color }"></span>
              <span>{{ row.color || '-' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="预览" width="160" align="center">
          <template #default="{ row }">
            <el-tag :style="{ backgroundColor: row.color, borderColor: row.color, color: '#fff' }" effect="dark">
              {{ row.name }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" width="140" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="fetchData"
        />
      </div>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="formTitle" width="500px">
      <el-form :model="form" label-width="80px" class="tag-form">
        <el-form-item label="名称">
          <el-input v-model="form.name" placeholder="请输入标签名称" />
        </el-form-item>
        <el-form-item label="颜色">
          <div class="form-color-row">
            <el-color-picker v-model="form.color" />
            <el-input v-model="form.color" placeholder="#409EFF" />
            <el-tag :style="{ backgroundColor: form.color, borderColor: form.color, color: '#fff' }">{{ form.name || '标签预览' }}</el-tag>
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
.tag-card {
  border-radius: 12px;
}
.tag-table :deep(.el-table__header th) {
  background: #f7f9fc;
  color: #606266;
  font-weight: 600;
}
.tag-table :deep(.el-table__row) {
  height: 64px;
}
.tag-name-cell {
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
.tag-name {
  font-weight: 600;
  color: #303133;
}
.tag-id {
  margin-top: 4px;
  color: #a8abb2;
  font-size: 12px;
}
.color-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #909399;
  font-size: 12px;
}
.color-block {
  width: 28px;
  height: 18px;
  border-radius: 6px;
  box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.08);
}
.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
.form-color-row {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 12px;
  width: 100%;
}
.tag-form :deep(.el-form-item:last-child) {
  margin-bottom: 0;
}
</style>
