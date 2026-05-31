<script setup lang="ts">
import { ref, onMounted } from 'vue'
import request from '@/api/request'

interface Stats {
  userCount: number
  ticketCount: number
  sopCount: number
  todayActive: number
}

const stats = ref<Stats>({ userCount: 0, ticketCount: 0, sopCount: 0, todayActive: 0 })

onMounted(async () => {
  try {
    const res: any = await request.get('/stats/overview')
    stats.value = res.data || res
  } catch {
    // API not available, show zeros
  }
})
</script>

<template>
  <div class="dashboard">
    <div class="page-header">
      <h2><el-icon><DataAnalysis /></el-icon> 仪表盘</h2>
    </div>
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="dashboard-stat">
            <div class="dashboard-stat-icon" style="background: linear-gradient(135deg, #6366f1, #8b5cf6);">
              <el-icon :size="24"><Ticket /></el-icon>
            </div>
            <div class="dashboard-stat-info">
              <div class="dashboard-stat-value">{{ stats.ticketCount }}</div>
              <div class="dashboard-stat-label">工单总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="dashboard-stat">
            <div class="dashboard-stat-icon" style="background: linear-gradient(135deg, #06b6d4, #0891b2);">
              <el-icon :size="24"><User /></el-icon>
            </div>
            <div class="dashboard-stat-info">
              <div class="dashboard-stat-value">{{ stats.userCount }}</div>
              <div class="dashboard-stat-label">用户总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="dashboard-stat">
            <div class="dashboard-stat-icon" style="background: linear-gradient(135deg, #f59e0b, #d97706);">
              <el-icon :size="24"><Document /></el-icon>
            </div>
            <div class="dashboard-stat-info">
              <div class="dashboard-stat-value">{{ stats.sopCount }}</div>
              <div class="dashboard-stat-label">SOP 数量</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="dashboard-stat">
            <div class="dashboard-stat-icon" style="background: linear-gradient(135deg, #22c55e, #16a34a);">
              <el-icon :size="24"><TrendCharts /></el-icon>
            </div>
            <div class="dashboard-stat-info">
              <div class="dashboard-stat-value">{{ stats.todayActive }}</div>
              <div class="dashboard-stat-label">今日活跃</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.stats-row {
  margin-bottom: 24px;
}

.dashboard-stat {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 8px 0;
}

.dashboard-stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}

.dashboard-stat-info {
  flex: 1;
  min-width: 0;
}

.dashboard-stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.2;
}

.dashboard-stat-label {
  margin-top: 2px;
  font-size: 13px;
  color: var(--text-secondary);
}
</style>
