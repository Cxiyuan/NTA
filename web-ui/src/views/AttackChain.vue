<template>
  <div class="attack-chain-page">
    <el-card shadow="hover" style="margin-bottom: 20px">
      <template #header>
        <span class="card-title">APT攻击活动</span>
      </template>
      
      <el-table :data="aptCampaigns" style="width: 100%">
        <el-table-column prop="id" label="活动ID" width="100" />
        <el-table-column prop="attacker" label="攻击者IP" width="150" />
        <el-table-column prop="firstSeen" label="首次发现" width="180" />
        <el-table-column prop="stages" label="攻击阶段" width="300">
          <template #default="{ row }">
            <el-tag v-for="stage in row.stages" :key="stage" size="small" style="margin-right: 8px">
              {{ stage }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="victims" label="受害主机" width="100" />
        <el-table-column prop="riskScore" label="风险评分" width="120">
          <template #default="{ row }">
            <el-progress :percentage="row.riskScore" :color="getRiskColor(row.riskScore)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="viewChain(row)">查看攻击链</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    
    <!-- 攻击链可视化 -->
    <el-card shadow="hover" v-if="selectedCampaign">
      <template #header>
        <div style="display: flex; justify-content: space-between">
          <span class="card-title">攻击链时间线 - {{ selectedCampaign.attacker }}</span>
          <el-button size="small" @click="selectedCampaign = null">关闭</el-button>
        </div>
      </template>
      
      <el-timeline>
        <el-timeline-item
          v-for="(event, index) in attackEvents"
          :key="index"
          :timestamp="event.timestamp"
          :type="getSeverityType(event.severity)"
          placement="top"
        >
          <el-card>
            <h4>{{ event.stage }} - {{ event.type }}</h4>
            <p>源: {{ event.source }} → 目标: {{ event.target }}</p>
            <p>{{ event.description }}</p>
            <el-tag :type="getSeverityType(event.severity)" size="small">{{ event.severity }}</el-tag>
          </el-card>
        </el-timeline-item>
      </el-timeline>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const selectedCampaign = ref(null)

const aptCampaigns = ref([
  {
    id: 'APT-001',
    attacker: '192.168.1.100',
    firstSeen: '2025-12-22 09:30:00',
    stages: ['侦察', '凭证获取', '横向移动', 'C2通信'],
    victims: 5,
    riskScore: 95
  }
])

const attackEvents = ref([
  {
    timestamp: '2025-12-22 09:30:00',
    stage: '阶段1',
    type: '横向扫描',
    source: '192.168.1.100',
    target: '10.0.1.0/24',
    description: '扫描25台内网主机',
    severity: 'HIGH'
  },
  {
    timestamp: '2025-12-22 09:45:00',
    stage: '阶段2',
    type: 'Pass-the-Hash',
    source: '192.168.1.100',
    target: '10.0.1.50',
    description: 'NTLM Hash重用攻击',
    severity: 'CRITICAL'
  },
  {
    timestamp: '2025-12-22 10:00:00',
    stage: '阶段3',
    type: 'PSExec执行',
    source: '192.168.1.100',
    target: '10.0.1.51',
    description: 'PSExec远程执行',
    severity: 'CRITICAL'
  },
  {
    timestamp: '2025-12-22 10:15:00',
    stage: '阶段4',
    type: 'C2 Beacon',
    source: '10.0.1.51',
    target: '8.8.8.8:443',
    description: '规律性心跳通信',
    severity: 'CRITICAL'
  }
])

const getRiskColor = (score) => {
  if (score >= 80) return '#f56c6c'
  if (score >= 60) return '#e6a23c'
  return '#409eff'
}

const getSeverityType = (severity) => {
  const types = {
    'CRITICAL': 'danger',
    'HIGH': 'warning',
    'MEDIUM': 'info',
    'LOW': 'success'
  }
  return types[severity] || 'info'
}

const viewChain = (campaign) => {
  selectedCampaign.value = campaign
}
</script>

<style lang="scss" scoped>
.attack-chain-page {
  .card-title {
    font-size: 16px;
    font-weight: 600;
  }
}
</style>
