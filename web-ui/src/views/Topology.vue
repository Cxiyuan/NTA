<template>
  <div class="topology-page">
    <el-card shadow="hover">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span class="card-title">网络拓扑图</span>
          <div>
            <el-button-group>
              <el-button :type="viewMode === 'graph' ? 'primary' : ''" size="small" @click="viewMode = 'graph'">
                图形视图
              </el-button>
              <el-button :type="viewMode === 'table' ? 'primary' : ''" size="small" @click="viewMode = 'table'">
                列表视图
              </el-button>
            </el-button-group>
            <el-button type="primary" size="small" :icon="Refresh" @click="refreshTopology" style="margin-left: 12px">
              刷新
            </el-button>
          </div>
        </div>
      </template>
      
      <div v-if="viewMode === 'graph'" id="topology-graph" style="height: 600px"></div>
      
      <el-table v-else :data="connections" style="width: 100%">
        <el-table-column prop="source" label="源IP" width="150" />
        <el-table-column prop="target" label="目标IP" width="150" />
        <el-table-column prop="protocol" label="协议" width="100" />
        <el-table-column prop="count" label="连接次数" width="120" sortable />
        <el-table-column prop="firstSeen" label="首次发现" width="180" />
        <el-table-column prop="lastSeen" label="最后活动" width="180" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.anomaly ? 'danger' : 'success'" size="small">
              {{ row.anomaly ? '异常' : '正常' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    
    <el-row :gutter="24" style="margin-top: 20px">
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <span class="card-title">异常扇出检测</span>
          </template>
          <el-table :data="anomalies.fanout" height="300">
            <el-table-column prop="node" label="节点IP" />
            <el-table-column prop="targetCount" label="目标数量" sortable />
            <el-table-column prop="score" label="异常分数">
              <template #default="{ row }">
                <el-progress :percentage="row.score * 100" :color="getScoreColor(row.score)" />
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <span class="card-title">多跳链路</span>
          </template>
          <el-table :data="anomalies.chains" height="300">
            <el-table-column prop="path" label="路径" show-overflow-tooltip />
            <el-table-column prop="length" label="跳数" width="80" />
            <el-table-column prop="score" label="风险分数" width="150">
              <template #default="{ row }">
                <el-progress :percentage="row.score" :color="getScoreColor(row.score / 100)" />
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import * as echarts from 'echarts'

const viewMode = ref('graph')
let chartInstance = null

const connections = ref([
  {
    source: '192.168.1.100',
    target: '10.0.1.50',
    protocol: 'SMB',
    count: 125,
    firstSeen: '2025-12-22 09:00:00',
    lastSeen: '2025-12-22 10:30:00',
    anomaly: true
  }
])

const anomalies = ref({
  fanout: [
    { node: '192.168.1.100', targetCount: 25, score: 0.85 }
  ],
  chains: [
    { path: '192.168.1.100 → 10.0.1.50 → 10.0.1.51 → 10.0.1.52', length: 3, score: 75 }
  ]
})

const getScoreColor = (score) => {
  if (score >= 0.8) return '#f56c6c'
  if (score >= 0.6) return '#e6a23c'
  return '#409eff'
}

const initGraph = () => {
  const dom = document.getElementById('topology-graph')
  if (!dom) return
  
  chartInstance = echarts.init(dom)
  
  const option = {
    tooltip: {},
    series: [{
      type: 'graph',
      layout: 'force',
      data: [
        { name: '192.168.1.100', symbolSize: 50, itemStyle: { color: '#f56c6c' } },
        { name: '10.0.1.50', symbolSize: 30 },
        { name: '10.0.1.51', symbolSize: 30 },
        { name: '10.0.1.52', symbolSize: 30 }
      ],
      links: [
        { source: '192.168.1.100', target: '10.0.1.50', lineStyle: { color: '#f56c6c', width: 2 } },
        { source: '192.168.1.100', target: '10.0.1.51' },
        { source: '10.0.1.50', target: '10.0.1.52' }
      ],
      roam: true,
      label: {
        show: true,
        position: 'right'
      },
      force: {
        repulsion: 100
      }
    }]
  }
  
  chartInstance.setOption(option)
}

const refreshTopology = () => {
  if (viewMode.value === 'graph' && chartInstance) {
    initGraph()
  }
}

onMounted(() => {
  setTimeout(initGraph, 100)
})
</script>

<style lang="scss" scoped>
.topology-page {
  .card-title {
    font-size: 16px;
    font-weight: 600;
  }
}
</style>
