<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'
import MarkdownEditor from '@/components/MarkdownEditor.vue'

interface SOP {
  id: number
  name: string
  description: string
  version: string
  trigger_type: string
  trigger_config: string
  status: number
  created_at: string
  risk_level: string
  notification_config: string
}

const list = ref<SOP[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const currentView = ref<'list' | 'edit'>('list')
const isEdit = ref(false)
const activeTab = ref('params')
const form = ref({
  id: 0,
  name: '',
  description: '',
  version: '1.0',
  trigger_type: 'ticket_created',
  trigger_config: '',
  status: 1,
  risk_level: 'low',
  notification_config: '',
})

async function fetchData() {
  loading.value = true
  try {
    const res: any = await request.get('/sops', { params: { page: page.value, page_size: pageSize.value } })
    list.value = res.data?.data || res.data || []
    total.value = res.data?.total || 0
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  isEdit.value = false
  form.value = { id: 0, name: '', description: '', version: '1.0', trigger_type: 'ticket_created', trigger_config: '', status: 1, risk_level: 'low', notification_config: '' }
  activeTab.value = 'params'
  currentView.value = 'edit'
}

function openEdit(row: SOP) {
  isEdit.value = true
  form.value = { id: row.id, name: row.name, description: row.description, version: row.version, trigger_type: row.trigger_type, trigger_config: row.trigger_config || '', status: row.status, risk_level: row.risk_level || 'low', notification_config: row.notification_config || '' }
  activeTab.value = 'params'
  currentView.value = 'edit'
}

function goBack() {
  currentView.value = 'list'
  fetchData()
}

async function handleSave() {
  try {
    if (isEdit.value) {
      await request.put(`/sops/${form.value.id}`, form.value)
      ElMessage.success('更新成功')
    } else {
      await request.post('/sops', form.value)
      ElMessage.success('创建成功')
    }
    goBack()
  } catch {
    // handled by interceptor
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm('确认删除该 SOP 吗？', '提示')
    await request.delete(`/sops/${id}`)
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
    <template v-if="currentView === 'list'">
      <div class="page-header">
        <h2>SOP 管理</h2>
        <el-button type="primary" @click="openAdd">新增 SOP</el-button>
      </div>
      <el-card>
        <el-table :data="list" v-loading="loading" stripe border style="width: 100%">
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="name" label="名称" min-width="140" />
          <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
          <el-table-column prop="version" label="版本" width="80" />
          <el-table-column prop="risk_level" label="风险等级" width="100">
            <template #default="{ row }">
              <el-tag :type="row.risk_level === 'critical' ? 'danger' : row.risk_level === 'high' ? 'warning' : row.risk_level === 'medium' ? 'warning' : 'info'" size="small">
                {{ ({ low: '低', medium: '中', high: '高', critical: '严重' } as Record<string, string>)[row.risk_level] || row.risk_level }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="trigger_type" label="触发类型" width="120" />
          <el-table-column prop="status" label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : row.status === 0 ? 'info' : 'warning'">
                {{ row.status === 1 ? '激活' : row.status === 0 ? '草稿' : '停用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="180" />
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button size="small" @click="openEdit(row)">编辑</el-button>
              <el-button size="small" type="danger" @click="handleDelete(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="pagination-wrap">
          <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="fetchData" />
        </div>
      </el-card>
    </template>

    <template v-else>
      <div class="page-header">
        <el-button text @click="goBack">
          <el-icon><ArrowLeft /></el-icon> 返回
        </el-button>
        <h2>{{ isEdit ? '编辑 SOP' : '新增 SOP' }}</h2>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </div>
      <el-card>
        <el-tabs v-model="activeTab">
          <el-tab-pane label="基本参数" name="params">
            <el-form :model="form" label-width="120px" class="edit-form">
              <el-form-item label="名称">
                <el-input v-model="form.name" />
              </el-form-item>
              <el-row :gutter="20">
                <el-col :span="8">
                  <el-form-item label="版本">
                    <el-input v-model="form.version" />
                  </el-form-item>
                </el-col>
                <el-col :span="8">
                  <el-form-item label="触发类型">
                    <el-select v-model="form.trigger_type">
                      <el-option label="工单创建" value="ticket_created" />
                      <el-option label="事件" value="event" />
                      <el-option label="手动" value="manual" />
                    </el-select>
                  </el-form-item>
                </el-col>
                <el-col :span="8">
                  <el-form-item label="风险等级">
                    <el-select v-model="form.risk_level">
                      <el-option label="低" value="low" />
                      <el-option label="中" value="medium" />
                      <el-option label="高" value="high" />
                      <el-option label="严重" value="critical" />
                    </el-select>
                  </el-form-item>
                </el-col>
              </el-row>
              <el-form-item label="状态">
                <el-radio-group v-model="form.status">
                  <el-radio :value="0">草稿</el-radio>
                  <el-radio :value="1">激活</el-radio>
                  <el-radio :value="2">停用</el-radio>
                </el-radio-group>
              </el-form-item>
              <el-form-item label="通知配置">
                <el-input v-model="form.notification_config" type="textarea" :rows="2" placeholder='JSON, 如 {"channels":["lark","dingtalk"],"events":["on_complete","on_failure"]}' />
              </el-form-item>
              <el-form-item label="触发配置">
                <el-input v-model="form.trigger_config" type="textarea" :rows="2" placeholder="触发配置 JSON（可选）" />
              </el-form-item>
            </el-form>
          </el-tab-pane>
          <el-tab-pane label="内容编辑" name="content">
            <el-form :model="form" label-width="120px" class="edit-form">
              <el-form-item label="描述">
                <MarkdownEditor v-model="form.description" />
              </el-form-item>
            </el-form>
          </el-tab-pane>
        </el-tabs>
      </el-card>
    </template>
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
.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
.edit-form {
  max-width: 960px;
  margin-top: 16px;
}
</style>
