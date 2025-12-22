<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="24" class="stats-row">
      <el-col :xs="24" :sm="12" :lg="6">
        <el-card shadow="hover" class="stat-card critical">
          <div class="stat-content">
            <div class="stat-icon">
              <el-icon :size="40"><Warning /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-label">严重告警</div>
              <div class="stat-value">{{ stats.critical }}</div>
              <div class="stat-trend">
                <el-icon color="#f56c6c"><CaretTop /></el-icon>
                <span>+15%</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :sm="12" :lg="6">
        <el-card shadow="hover" class="stat-card high">
          <div class="stat-content">
            <div class="stat-icon">
              <el-icon :size="40"><WarnTriangleFilled /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-label">高危告警</div>
              <div class="stat-value">{{ stats.high }}</div>
              <div class="stat-trend">
                <el-icon color="#e6a23c"><CaretBottom /></el-icon>
                <span>-8%</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :sm="12" :lg="6">
        <el-card shadow="hover" class="stat-card apt">
          <div class="stat-content">
            <div class="stat-icon">
              <el-icon :size="40"><Connection /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-label">APT活动</div>
              <div class="stat-value">{{ stats.apt }}</div>
              <div class="stat-trend">
                <el-icon color="#909399"><Minus /></el-icon>
                <span>0%</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :sm="12" :lg="6">
        <el-card shadow="hover" class="stat-card traffic">
          <div class="stat-content">
            <div class="stat-icon">
              <el-icon :size="40"><Monitor /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-label">流量处理</div>
              <div class="stat-value">{{ stats.traffic }}</div>
              <div class="stat-trend success">
                <el-icon color="#67c23a"><Checked /></el-icon>
                <span>正常</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 图表行 -->
    <el-row :gutter="24" class="chart-row">
      <!-- 告警趋势图 -->
      <el-col :xs="24" :lg="16">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span class="card-title">告警趋势分析</span>
              <el-radio-group v-model="timeRange" size="small">
                <el-radio-button label="1h">1小时</el-radio-button>
                <el-radio-button label="24h">24小时</el-radio-button>
                <el-radio-button label="7d">7天</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <v-chart :option="alertTrendOption" :autoresize="true" style="height: 350px" />
        </el-card>
      </el-col>
      
      <!-- 攻击类型分布 -->
      <el-col :xs="24" :lg="8">
        <el-card shadow="hover">
          <template #header>
            <span class="card-title">攻击类型分布</span>
          </template>
          <v-chart :option="attackTypeOption" :autoresize="true" style="height: 350px" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 实时告警流 -->
    <el-row :gutter="24">
      <el-col :span="24">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span class="card-title">实时告警流</span>
              <el-button type="primary" size="small" @click="pauseStream">
                {{ isPaused ? '继续' : '暂停' }}
              </el-button>
            </div>
          </template>
          
          <el-table
            :data="realtimeAlerts"
            style="width: 100%"
            :height="400"
            stripe
            v-loading="loading"
          >
            <el-table-column prop="timestamp" label="时间" width="180" />
            <el-table-column prop="severity" label="级别" width="100">
              <template #default="{ row }">
                <el-tag :type="getSeverityType(row.severity)" size="small">
                  {{ row.severity }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="type" label="类型" width="180" />
            <el-table-column prop="source" label="源IP" width="150" />
            <el-table-column prop="target" label="目标IP" width="150" />
            <el-table-column prop="description" label="描述" show-overflow-tooltip />
            <el-table-column label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" size="small" link @click="viewDetail(row)">
                  详情
                </el-button>
                <el-button type="warning" size="small" link @click="handleAlert(row)">
                  处置
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart } from 'echarts/charts'
import { 
  TitleComponent, 
  TooltipComponent, 
  LegendComponent,
  GridComponent 
} from 'echarts/components'
import VChart from 'vue-echarts'
import { ElMessage } from 'element-plus'
import { 
  Warning, WarnTriangleFilled, Connection, Monitor,
  CaretTop, CaretBottom, Minus, Checked
} from '@element-plus/icons-vue'

use([
  CanvasRenderer,
  LineChart,
  PieChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
])

const stats = reactive({
  critical: 15,
  high: 42,
  apt: 3,
  traffic: '8.5GB'
})

const timeRange = ref('24h')
const isPaused = ref(false)
const loading = ref(false)

const realtimeAlerts = ref([
  {
    timestamp: '2025-12-22 10:30:15',
    severity: 'CRITICAL',
    type: 'PTH攻击',
    source: '192.168.1.100',
    target: '10.0.1.50',
    description: 'Pass-the-Hash攻击检测'
  },
  {
    timestamp: '2025-12-22 10:28:42',
    severity: 'HIGH',
    type: '横向扫描',
    source: '192.168.1.100',
    target: '10.0.1.0/24',
    description: '扫描25台主机'
  },
  {
    timestamp: '2025-12-22 10:25:18',
    severity: 'CRITICAL',
    type: 'PSExec执行',
    source: '192.168.1.101',
    target: '10.0.2.30',
    description: 'PSExec远程执行检测'
  }
])

const alertTrendOption = ref({
  tooltip: {
    trigger: 'axis'
  },
  legend: {
    data: ['严重', '高危', '中危', '低危']
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    containLabel: true
  },
  xAxis: {
    type: 'category',
    boundaryGap: false,
    data: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00', '24:00']
  },
  yAxis: {
    type: 'value'
  },
  series: [
    {
      name: '严重',
      type: 'line',
      stack: 'Total',
      data: [5, 8, 12, 15, 18, 12, 15],
      itemStyle: { color: '#f56c6c' }
    },
    {
      name: '高危',
      type: 'line',
      stack: 'Total',
      data: [15, 22, 35, 42, 38, 35, 42],
      itemStyle: { color: '#e6a23c' }
    },
    {
      name: '中危',
      type: 'line',
      stack: 'Total',
      data: [30, 42, 55, 48, 52, 45, 50],
      itemStyle: { color: '#409eff' }
    },
    {
      name: '低危',
      type: 'line',
      stack: 'Total',
      data: [50, 65, 72, 68, 75, 70, 72],
      itemStyle: { color: '#67c23a' }
    }
  ]
})

const attackTypeOption = ref({
  tooltip: {
    trigger: 'item',
    formatter: '{b}: {c} ({d}%)'
  },
  legend: {
    orient: 'vertical',
    left: 'left'
  },
  series: [
    {
      name: '攻击类型',
      type: 'pie',
      radius: ['40%', '70%'],
      avoidLabelOverlap: false,
      itemStyle: {
        borderRadius: 10,
        borderColor: '#fff',
        borderWidth: 2
      },
      label: {
        show: false,
        position: 'center'
      },
      emphasis: {
        label: {
          show: true,
          fontSize: 20,
          fontWeight: 'bold'
        }
      },
      data: [
        { value: 35, name: 'PTH攻击', itemStyle: { color: '#f56c6c' } },
        { value: 28, name: '横向扫描', itemStyle: { color: '#e6a23c' } },
        { value: 22, name: 'PSExec', itemStyle: { color: '#409eff' } },
        { value: 18, name: 'WMI执行', itemStyle: { color: '#67c23a' } },
        { value: 12, name: 'RDP跳板', itemStyle: { color: '#909399' } }
      ]
    }
  ]
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

const pauseStream = () => {
  isPaused.value = !isPaused.value
  ElMessage.success(isPaused.value ? '已暂停告警流' : '已恢复告警流')
}

const viewDetail = (row) => {
  ElMessage.info('查看详情：' + row.type)
}

const handleAlert = (row) => {
  ElMessage.warning('处置告警：' + row.type)
}

let updateInterval = null

onMounted(() => {
  updateInterval = setInterval(() => {
    if (!isPaused.value) {
      // 模拟新告警
      stats.critical = Math.floor(Math.random() * 20) + 10
      stats.high = Math.floor(Math.random() * 50) + 30
    }
  }, 5000)
})

onUnmounted(() => {
  if (updateInterval) {
    clearInterval(updateInterval)
  }
})
</script>

<style lang="scss" scoped>
.dashboard {
  .stats-row {
    margin-bottom: 24px;
  }
  
  .stat-card {
    border-radius: 8px;
    transition: all 0.3s;
    
    &:hover {
      transform: translateY(-4px);
    }
    
    .stat-content {
      display: flex;
      align-items: center;
      gap: 16px;
      
      .stat-icon {
        width: 60px;
        height: 60px;
        border-radius: 12px;
        display: flex;
        align-items: center;
        justify-content: center;
      }
      
      .stat-info {
        flex: 1;
        
        .stat-label {
          font-size: 14px;
          color: #909399;
          margin-bottom: 8px;
        }
        
        .stat-value {
          font-size: 28px;
          font-weight: 600;
          color: #303133;
          line-height: 1;
          margin-bottom: 8px;
        }
        
        .stat-trend {
          font-size: 12px;
          color: #606266;
          display: flex;
          align-items: center;
          gap: 4px;
          
          &.success {
            color: #67c23a;
          }
        }
      }
    }
    
    &.critical .stat-icon {
      background: linear-gradient(135deg, #ff6b6b 0%, #ff5252 100%);
      color: #fff;
    }
    
    &.high .stat-icon {
      background: linear-gradient(135deg, #ffa726 0%, #ff9800 100%);
      color: #fff;
    }
    
    &.apt .stat-icon {
      background: linear-gradient(135deg, #ab47bc 0%, #9c27b0 100%);
      color: #fff;
    }
    
    &.traffic .stat-icon {
      background: linear-gradient(135deg, #42a5f5 0%, #2196f3 100%);
      color: #fff;
    }
  }
  
  .chart-row {
    margin-bottom: 24px;
  }
  
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    
    .card-title {
      font-size: 16px;
      font-weight: 600;
      color: #303133;
    }
  }
  
  :deep(.el-card__body) {
    padding: 20px;
  }
}
</style>
