<template>
  <div class="threat-intel-page">
    <el-row :gutter="24" style="margin-bottom: 20px">
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="恶意IP" :value="stats.malicious_ips" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="恶意域名" :value="stats.malicious_domains" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="文件哈希" :value="stats.malicious_hashes" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="JA3指纹" :value="stats.ja3_fingerprints" />
        </el-card>
      </el-col>
    </el-row>
    
    <el-card shadow="hover">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span class="card-title">威胁情报库</span>
          <div>
            <el-button type="success" :icon="Download" @click="updateFeeds">更新情报源</el-button>
            <el-button type="primary" :icon="Plus" @click="addDialogVisible = true">添加IOC</el-button>
          </div>
        </div>
      </template>
      
      <el-tabs v-model="activeTab">
        <el-tab-pane label="恶意IP" name="ip">
          <el-table :data="iocs.ips" style="width: 100%">
            <el-table-column prop="value" label="IP地址" width="150" />
            <el-table-column prop="source" label="来源" width="150" />
            <el-table-column prop="category" label="分类" width="150">
              <template #default="{ row }">
                <el-tag type="danger" size="small">{{ row.category }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="confidence" label="置信度" width="150">
              <template #default="{ row }">
                <el-progress :percentage="row.confidence * 100" />
              </template>
            </el-table-column>
            <el-table-column prop="first_seen" label="首次发现" width="180" />
            <el-table-column prop="description" label="描述" show-overflow-tooltip />
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button type="danger" size="small" link @click="deleteIOC(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
        
        <el-tab-pane label="恶意域名" name="domain">
          <el-table :data="iocs.domains" style="width: 100%">
            <el-table-column prop="value" label="域名" width="250" />
            <el-table-column prop="source" label="来源" width="150" />
            <el-table-column prop="category" label="分类" width="150">
              <template #default="{ row }">
                <el-tag type="danger" size="small">{{ row.category }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="confidence" label="置信度" width="150">
              <template #default="{ row }">
                <el-progress :percentage="row.confidence * 100" />
              </template>
            </el-table-column>
            <el-table-column prop="first_seen" label="首次发现" width="180" />
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button type="danger" size="small" link @click="deleteIOC(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
        
        <el-tab-pane label="JA3指纹" name="ja3">
          <el-table :data="iocs.ja3" style="width: 100%">
            <el-table-column prop="value" label="JA3哈希" width="350" />
            <el-table-column prop="tool_name" label="工具名称" width="180">
              <template #default="{ row }">
                <el-tag type="warning">{{ row.tool_name }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="tool_type" label="类型" width="180" />
            <el-table-column prop="severity" label="严重性" width="120">
              <template #default="{ row }">
                <el-tag :type="getSeverityType(row.severity)">{{ row.severity }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="description" label="描述" show-overflow-tooltip />
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button type="danger" size="small" link @click="deleteIOC(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>
    
    <!-- 添加IOC对话框 -->
    <el-dialog v-model="addDialogVisible" title="添加IOC" width="600px">
      <el-form :model="newIOC" label-width="100px">
        <el-form-item label="类型">
          <el-select v-model="newIOC.type" placeholder="选择IOC类型">
            <el-option label="恶意IP" value="ip" />
            <el-option label="恶意域名" value="domain" />
            <el-option label="文件哈希" value="hash" />
            <el-option label="JA3指纹" value="ja3" />
          </el-select>
        </el-form-item>
        <el-form-item label="值">
          <el-input v-model="newIOC.value" placeholder="输入IOC值" />
        </el-form-item>
        <el-form-item label="来源">
          <el-input v-model="newIOC.source" placeholder="手动添加" />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="newIOC.category">
            <el-option label="C2服务器" value="C2" />
            <el-option label="恶意软件" value="Malware" />
            <el-option label="钓鱼" value="Phishing" />
            <el-option label="僵尸网络" value="Botnet" />
          </el-select>
        </el-form-item>
        <el-form-item label="置信度">
          <el-slider v-model="newIOC.confidence" :min="0" :max="1" :step="0.1" :format-tooltip="formatPercent" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="newIOC.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="addIOC">添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { Plus, Download } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const activeTab = ref('ip')
const addDialogVisible = ref(false)

const stats = reactive({
  malicious_ips: 1245,
  malicious_domains: 3890,
  malicious_hashes: 5621,
  ja3_fingerprints: 45
})

const iocs = reactive({
  ips: [
    {
      value: '8.8.8.8',
      source: 'abuse.ch',
      category: 'C2服务器',
      confidence: 0.95,
      first_seen: '2025-12-20 10:00:00',
      description: 'Cobalt Strike C2服务器'
    }
  ],
  domains: [
    {
      value: 'evil.example.com',
      source: 'VirusTotal',
      category: '恶意软件',
      confidence: 0.98,
      first_seen: '2025-12-19 08:30:00'
    }
  ],
  ja3: [
    {
      value: 'a0e9f5d64349fb13191bc781f81f42e1',
      tool_name: 'Metasploit',
      tool_type: 'C2框架',
      severity: 'CRITICAL',
      description: 'Metasploit Framework默认TLS指纹'
    },
    {
      value: '51c64c77e60f3980eea90869b68c58a8',
      tool_name: 'Cobalt Strike',
      tool_type: 'C2框架',
      severity: 'CRITICAL',
      description: 'Cobalt Strike默认TLS指纹'
    }
  ]
})

const newIOC = reactive({
  type: 'ip',
  value: '',
  source: '手动添加',
  category: 'C2',
  confidence: 0.8,
  description: ''
})

const formatPercent = (val) => `${(val * 100).toFixed(0)}%`

const getSeverityType = (severity) => {
  const types = {
    'CRITICAL': 'danger',
    'HIGH': 'warning',
    'MEDIUM': 'info'
  }
  return types[severity] || 'info'
}

const updateFeeds = () => {
  ElMessage.success('正在更新威胁情报源...')
  setTimeout(() => {
    ElMessage.success('情报源更新完成')
  }, 2000)
}

const addIOC = () => {
  if (!newIOC.value) {
    ElMessage.warning('请输入IOC值')
    return
  }
  
  ElMessage.success('IOC添加成功')
  addDialogVisible.value = false
  
  newIOC.value = ''
  newIOC.description = ''
}

const deleteIOC = (row) => {
  ElMessage.success('IOC删除成功')
}
</script>

<style lang="scss" scoped>
.threat-intel-page {
  .card-title {
    font-size: 16px;
    font-weight: 600;
  }
}
</style>
