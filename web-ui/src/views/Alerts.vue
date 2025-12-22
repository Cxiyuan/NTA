<template>
  <div class="alerts-page">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span class="card-title">告警管理</span>
          <div class="header-actions">
            <el-input
              v-model="searchText"
              placeholder="搜索告警"
              :prefix-icon="Search"
              style="width: 240px"
              clearable
            />
            <el-select v-model="severityFilter" placeholder="严重级别" clearable style="width: 120px">
              <el-option label="严重" value="CRITICAL" />
              <el-option label="高危" value="HIGH" />
              <el-option label="中危" value="MEDIUM" />
              <el-option label="低危" value="LOW" />
            </el-select>
            <el-button type="primary" :icon="Refresh" @click="refreshAlerts">刷新</el-button>
          </div>
        </div>
      </template>
      
      <el-table
        :data="filteredAlerts"
        style="width: 100%"
        v-loading="loading"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="timestamp" label="时间" width="180" sortable />
        <el-table-column prop="severity" label="级别" width="100" sortable>
          <template #default="{ row }">
            <el-tag :type="getSeverityType(row.severity)" effect="dark">
              {{ row.severity }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="攻击类型" width="150" />
        <el-table-column prop="source" label="源IP" width="140" />
        <el-table-column prop="target" label="目标IP" width="140" />
        <el-table-column prop="confidence" label="置信度" width="100">
          <template #default="{ row }">
            <el-progress :percentage="row.confidence * 100" :color="getConfidenceColor(row.confidence)" />
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" show-overflow-tooltip />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" link @click="viewDetail(row)">详情</el-button>
            <el-button type="success" size="small" link @click="handleAlert(row, 'confirm')">确认</el-button>
            <el-button type="danger" size="small" link @click="handleAlert(row, 'block')">阻断</el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>
    
    <!-- 详情对话框 -->
    <el-dialog v-model="detailVisible" title="告警详情" width="800px">
      <el-descriptions :column="2" border v-if="selectedAlert">
        <el-descriptions-item label="告警ID">{{ selectedAlert.id }}</el-descriptions-item>
        <el-descriptions-item label="时间">{{ selectedAlert.timestamp }}</el-descriptions-item>
        <el-descriptions-item label="严重级别">
          <el-tag :type="getSeverityType(selectedAlert.severity)">{{ selectedAlert.severity }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="攻击类型">{{ selectedAlert.type }}</el-descriptions-item>
        <el-descriptions-item label="源IP">{{ selectedAlert.source }}</el-descriptions-item>
        <el-descriptions-item label="目标IP">{{ selectedAlert.target }}</el-descriptions-item>
        <el-descriptions-item label="置信度">{{ (selectedAlert.confidence * 100).toFixed(1) }}%</el-descriptions-item>
        <el-descriptions-item label="检测模块">{{ selectedAlert.detector }}</el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">{{ selectedAlert.description }}</el-descriptions-item>
        <el-descriptions-item label="证据" :span="2">
          <pre>{{ selectedAlert.evidence }}</pre>
        </el-descriptions-item>
      </el-descriptions>
      
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
        <el-button type="primary" @click="exportDetail">导出</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Search, Refresh } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const loading = ref(false)
const searchText = ref('')
const severityFilter = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const detailVisible = ref(false)
const selectedAlert = ref(null)
const selectedAlerts = ref([])

const alerts = ref([
  {
    id: 1001,
    timestamp: '2025-12-22 10:30:15',
    severity: 'CRITICAL',
    type: 'PTH攻击',
    source: '192.168.1.100',
    target: '10.0.1.50',
    confidence: 0.95,
    description: 'Pass-the-Hash攻击检测',
    detector: 'lateral-auth.zeek',
    evidence: 'NTLM Hash重用于3台主机'
  }
])

total.value = alerts.value.length

const filteredAlerts = computed(() => {
  let result = alerts.value
  
  if (searchText.value) {
    result = result.filter(item => 
      item.description.includes(searchText.value) ||
      item.source.includes(searchText.value) ||
      item.target.includes(searchText.value)
    )
  }
  
  if (severityFilter.value) {
    result = result.filter(item => item.severity === severityFilter.value)
  }
  
  return result
})

const getSeverityType = (severity) => {
  const types = {
    'CRITICAL': 'danger',
    'HIGH': 'warning',
    'MEDIUM': 'info',
    'LOW': 'success'
  }
  return types[severity] || 'info'
}

const getConfidenceColor = (confidence) => {
  if (confidence >= 0.9) return '#67c23a'
  if (confidence >= 0.7) return '#e6a23c'
  return '#f56c6c'
}

const refreshAlerts = () => {
  loading.value = true
  setTimeout(() => {
    loading.value = false
    ElMessage.success('刷新成功')
  }, 500)
}

const viewDetail = (row) => {
  selectedAlert.value = row
  detailVisible.value = true
}

const handleAlert = (row, action) => {
  ElMessage.success(`${action === 'confirm' ? '确认' : '阻断'}告警：${row.type}`)
}

const handleSelectionChange = (selection) => {
  selectedAlerts.value = selection
}

const exportDetail = () => {
  ElMessage.success('导出成功')
  detailVisible.value = false
}
</script>

<style lang="scss" scoped>
.alerts-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    
    .card-title {
      font-size: 16px;
      font-weight: 600;
    }
    
    .header-actions {
      display: flex;
      gap: 12px;
    }
  }
  
  pre {
    background: #f5f7fa;
    padding: 12px;
    border-radius: 4px;
    font-size: 12px;
  }
}
</style>
