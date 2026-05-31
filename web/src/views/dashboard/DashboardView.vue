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
    <h2>仪表盘</h2>
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value">{{ stats.ticketCount }}</div>
            <div class="stat-label">工单总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value">{{ stats.userCount }}</div>
            <div class="stat-label">用户总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value">{{ stats.sopCount }}</div>
            <div class="stat-label">SOP 数量</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value">{{ stats.todayActive }}</div>
            <div class="stat-label">今日活跃</div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.dashboard h2 {
  margin-bottom: 20px;
}
.stats-row {
  margin-bottom: 24px;
}
.stat-card {
  text-align: center;
  padding: 20px 0;
}
.stat-value {
  font-size: 36px;
  font-weight: bold;
  color: #409eff;
}
.stat-label {
  margin-top: 8px;
  color: #666;
  font-size: 14px;
}
</style>
