<template>
  <div class="reports-page">
    <el-card shadow="hover">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span class="card-title">报告中心</span>
          <el-button type="primary" :icon="Document" @click="generateDialogVisible = true">
            生成报告
          </el-button>
        </div>
      </template>
      
      <el-table :data="reports" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="title" label="报告标题" width="300" />
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag>{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="time_range" label="时间范围" width="200" />
        <el-table-column prop="alerts_count" label="告警数" width="100" />
        <el-table-column prop="created_at" label="生成时间" width="180" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === '已完成' ? 'success' : 'info'">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" link @click="viewReport(row)">查看</el-button>
            <el-button type="success" size="small" link @click="downloadReport(row)">下载</el-button>
            <el-button type="danger" size="small" link @click="deleteReport(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    
    <!-- 生成报告对话框 -->
    <el-dialog v-model="generateDialogVisible" title="生成报告" width="600px">
      <el-form :model="reportForm" label-width="100px">
        <el-form-item label="报告类型">
          <el-select v-model="reportForm.type">
            <el-option label="日报" value="daily" />
            <el-option label="周报" value="weekly" />
            <el-option label="月报" value="monthly" />
            <el-option label="自定义" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围" v-if="reportForm.type === 'custom'">
          <el-date-picker
            v-model="reportForm.dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
          />
        </el-form-item>
        <el-form-item label="包含内容">
          <el-checkbox-group v-model="reportForm.includes">
            <el-checkbox label="summary">执行摘要</el-checkbox>
            <el-checkbox label="alerts">告警详情</el-checkbox>
            <el-checkbox label="statistics">统计分析</el-checkbox>
            <el-checkbox label="topology">网络拓扑</el-checkbox>
            <el-checkbox label="apt">APT活动</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="输出格式">
          <el-radio-group v-model="reportForm.format">
            <el-radio label="html">HTML</el-radio>
            <el-radio label="pdf">PDF</el-radio>
            <el-radio label="json">JSON</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="generateDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="generateReport">生成</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { Document } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const generateDialogVisible = ref(false)

const reports = ref([
  {
    id: 1,
    title: '2025-12-22 安全检测日报',
    type: '日报',
    time_range: '2025-12-22 00:00 - 23:59',
    alerts_count: 145,
    created_at: '2025-12-22 23:30:00',
    status: '已完成'
  }
])

const reportForm = reactive({
  type: 'daily',
  dateRange: [],
  includes: ['summary', 'alerts', 'statistics'],
  format: 'html'
})

const generateReport = () => {
  ElMessage.success('报告生成中，请稍候...')
  generateDialogVisible.value = false
}

const viewReport = (row) => {
  window.open(`/api/reports/${row.id}/view`, '_blank')
}

const downloadReport = (row) => {
  ElMessage.success('开始下载报告...')
}

const deleteReport = (row) => {
  ElMessage.success('报告删除成功')
}
</script>

<style lang="scss" scoped>
.reports-page {
  .card-title {
    font-size: 16px;
    font-weight: 600;
  }
}
</style>
