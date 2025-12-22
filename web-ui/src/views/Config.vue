<template>
  <div class="config-page">
    <el-tabs v-model="activeTab" type="border-card">
      <!-- 检测配置 -->
      <el-tab-pane label="检测配置" name="detection">
        <el-form :model="config.detection" label-width="150px">
          <el-divider content-position="left">横向扫描检测</el-divider>
          <el-form-item label="扫描阈值">
            <el-input-number v-model="config.detection.scan.threshold" :min="1" :max="100" />
            <span class="form-tip">单IP在时间窗口内扫描主机数量</span>
          </el-form-item>
          <el-form-item label="时间窗口(秒)">
            <el-input-number v-model="config.detection.scan.time_window" :min="60" :max="3600" />
          </el-form-item>
          <el-form-item label="最小失败率">
            <el-slider v-model="config.detection.scan.min_fail_rate" :min="0" :max="1" :step="0.1" :format-tooltip="formatPercent" />
          </el-form-item>
          
          <el-divider content-position="left">认证异常检测</el-divider>
          <el-form-item label="失败阈值">
            <el-input-number v-model="config.detection.authentication.fail_threshold" :min="1" :max="50" />
          </el-form-item>
          <el-form-item label="PTH检测窗口(秒)">
            <el-input-number v-model="config.detection.authentication.pth_window" :min="300" :max="7200" />
          </el-form-item>
          
          <el-form-item>
            <el-button type="primary" @click="saveConfig">保存配置</el-button>
            <el-button @click="resetConfig">重置</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
      
      <!-- 白名单管理 -->
      <el-tab-pane label="白名单管理" name="whitelist">
        <el-card shadow="never" style="margin-bottom: 20px">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center">
              <span>监控系统白名单</span>
              <el-button type="primary" size="small" :icon="Plus" @click="addWhitelist('monitoring')">
                添加
              </el-button>
            </div>
          </template>
          <el-tag
            v-for="(ip, index) in config.whitelist.monitoring_systems"
            :key="index"
            closable
            @close="removeWhitelist('monitoring', index)"
            style="margin: 4px"
          >
            {{ ip }}
          </el-tag>
        </el-card>
        
        <el-card shadow="never" style="margin-bottom: 20px">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center">
              <span>运维管理网段</span>
              <el-button type="primary" size="small" :icon="Plus" @click="addWhitelist('admin')">
                添加
              </el-button>
            </div>
          </template>
          <el-tag
            v-for="(subnet, index) in config.whitelist.admin_workstations"
            :key="index"
            closable
            @close="removeWhitelist('admin', index)"
            style="margin: 4px"
          >
            {{ subnet }}
          </el-tag>
        </el-card>
        
        <el-card shadow="never">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center">
              <span>自动化服务器</span>
              <el-button type="primary" size="small" :icon="Plus" @click="addWhitelist('automation')">
                添加
              </el-button>
            </div>
          </template>
          <el-tag
            v-for="(subnet, index) in config.whitelist.automation_servers"
            :key="index"
            closable
            @close="removeWhitelist('automation', index)"
            style="margin: 4px"
          >
            {{ subnet }}
          </el-tag>
        </el-card>
        
        <div style="margin-top: 20px; text-align: right">
          <el-button type="primary" @click="saveWhitelist">保存白名单</el-button>
        </div>
      </el-tab-pane>
      
      <!-- 决策引擎配置 -->
      <el-tab-pane label="决策引擎" name="decision">
        <el-form :model="config.decision_engine" label-width="150px">
          <el-divider content-position="left">告警阈值</el-divider>
          <el-form-item label="自动阻断">
            <el-input-number v-model="config.decision_engine.thresholds.auto_block" :min="0.9" :max="1" :step="0.0001" :precision="4" />
            <span class="form-tip">置信度 ≥ 此值将自动阻断</span>
          </el-form-item>
          <el-form-item label="紧急告警">
            <el-input-number v-model="config.decision_engine.thresholds.urgent_alert" :min="0.9" :max="1" :step="0.01" :precision="2" />
          </el-form-item>
          <el-form-item label="高危告警">
            <el-input-number v-model="config.decision_engine.thresholds.high_alert" :min="0.8" :max="1" :step="0.01" :precision="2" />
          </el-form-item>
          <el-form-item label="普通告警">
            <el-input-number v-model="config.decision_engine.thresholds.normal_alert" :min="0.7" :max="1" :step="0.01" :precision="2" />
          </el-form-item>
          
          <el-divider content-position="left">业务规则</el-divider>
          <el-form-item label="非工作时间倍数">
            <el-input-number v-model="config.decision_engine.business_rules.off_hours_multiplier" :min="1" :max="2" :step="0.05" :precision="2" />
          </el-form-item>
          <el-form-item label="重复告警倍数">
            <el-input-number v-model="config.decision_engine.business_rules.repeat_offender_multiplier" :min="1" :max="2" :step="0.05" :precision="2" />
          </el-form-item>
          
          <el-form-item>
            <el-button type="primary" @click="saveConfig">保存配置</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
      
      <!-- ML模型配置 -->
      <el-tab-pane label="ML模型" name="ml">
        <el-form :model="config.ml_model" label-width="150px">
          <el-form-item label="启用ML检测">
            <el-switch v-model="config.ml_model.enabled" />
          </el-form-item>
          <el-form-item label="训练间隔">
            <el-select v-model="config.ml_model.training_interval">
              <el-option label="每天" value="1 day" />
              <el-option label="每周" value="7 days" />
              <el-option label="每月" value="30 days" />
            </el-select>
          </el-form-item>
          <el-form-item label="异常比例">
            <el-slider v-model="config.ml_model.contamination" :min="0.001" :max="0.1" :step="0.001" :format-tooltip="formatPercent" />
          </el-form-item>
          
          <el-form-item>
            <el-button type="primary" @click="trainModel">立即训练模型</el-button>
            <el-button @click="exportModel">导出模型</el-button>
          </el-form-item>
        </el-form>
        
        <el-divider />
        
        <el-descriptions title="模型状态" :column="2" border>
          <el-descriptions-item label="模型版本">v2.0.1</el-descriptions-item>
          <el-descriptions-item label="训练时间">2025-12-22 08:00:00</el-descriptions-item>
          <el-descriptions-item label="样本数量">125,000</el-descriptions-item>
          <el-descriptions-item label="准确率">85.3%</el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>
      
      <!-- 系统设置 -->
      <el-tab-pane label="系统设置" name="system">
        <el-form label-width="150px">
          <el-divider content-position="left">Zeek配置</el-divider>
          <el-form-item label="监听接口">
            <el-input v-model="systemConfig.zeek.interface" placeholder="eth0" />
          </el-form-item>
          <el-form-item label="日志目录">
            <el-input v-model="systemConfig.zeek.log_dir" />
          </el-form-item>
          <el-form-item label="日志保留(天)">
            <el-input-number v-model="systemConfig.zeek.retention_days" :min="1" :max="365" />
          </el-form-item>
          
          <el-divider content-position="left">性能优化</el-divider>
          <el-form-item label="Worker数量">
            <el-input-number v-model="systemConfig.performance.workers" :min="1" :max="32" />
          </el-form-item>
          <el-form-item label="最大跟踪主机数">
            <el-input-number v-model="systemConfig.performance.max_tracked_hosts" :min="10000" :max="1000000" :step="10000" />
          </el-form-item>
          <el-form-item label="内存限制(MB)">
            <el-input-number v-model="systemConfig.performance.memory_limit_mb" :min="1024" :max="32768" :step="1024" />
          </el-form-item>
          
          <el-form-item>
            <el-button type="primary" @click="saveSystemConfig">保存系统配置</el-button>
            <el-button type="warning" @click="restartZeek">重启Zeek</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
    </el-tabs>
    
    <!-- 添加白名单对话框 -->
    <el-dialog v-model="addDialogVisible" title="添加白名单" width="500px">
      <el-form :model="newWhitelist" label-width="100px">
        <el-form-item label="IP/网段">
          <el-input v-model="newWhitelist.value" placeholder="192.168.1.100 或 10.0.0.0/24" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="newWhitelist.comment" placeholder="可选" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmAddWhitelist">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const activeTab = ref('detection')
const addDialogVisible = ref(false)
const currentWhitelistType = ref('')

const config = reactive({
  detection: {
    scan: {
      threshold: 20,
      time_window: 300,
      min_fail_rate: 0.6
    },
    authentication: {
      fail_threshold: 5,
      pth_window: 3600
    }
  },
  whitelist: {
    monitoring_systems: ['192.168.1.100', '192.168.1.101'],
    admin_workstations: ['10.0.10.0/24'],
    automation_servers: ['10.0.20.0/24']
  },
  decision_engine: {
    thresholds: {
      auto_block: 0.9999,
      urgent_alert: 0.99,
      high_alert: 0.95,
      normal_alert: 0.90
    },
    business_rules: {
      off_hours_multiplier: 1.15,
      repeat_offender_multiplier: 1.2
    }
  },
  ml_model: {
    enabled: true,
    training_interval: '7 days',
    contamination: 0.01
  }
})

const systemConfig = reactive({
  zeek: {
    interface: 'eth0',
    log_dir: '/var/log/zeek',
    retention_days: 30
  },
  performance: {
    workers: 8,
    max_tracked_hosts: 100000,
    memory_limit_mb: 2048
  }
})

const newWhitelist = reactive({
  value: '',
  comment: ''
})

const formatPercent = (val) => `${(val * 100).toFixed(0)}%`

const saveConfig = () => {
  ElMessage.success('配置保存成功')
}

const resetConfig = () => {
  ElMessage.info('配置已重置')
}

const addWhitelist = (type) => {
  currentWhitelistType.value = type
  newWhitelist.value = ''
  newWhitelist.comment = ''
  addDialogVisible.value = true
}

const confirmAddWhitelist = () => {
  if (!newWhitelist.value) {
    ElMessage.warning('请输入IP或网段')
    return
  }
  
  const typeMap = {
    'monitoring': 'monitoring_systems',
    'admin': 'admin_workstations',
    'automation': 'automation_servers'
  }
  
  config.whitelist[typeMap[currentWhitelistType.value]].push(newWhitelist.value)
  addDialogVisible.value = false
  ElMessage.success('添加成功')
}

const removeWhitelist = (type, index) => {
  const typeMap = {
    'monitoring': 'monitoring_systems',
    'admin': 'admin_workstations',
    'automation': 'automation_servers'
  }
  
  config.whitelist[typeMap[type]].splice(index, 1)
  ElMessage.success('删除成功')
}

const saveWhitelist = () => {
  ElMessage.success('白名单保存成功')
}

const trainModel = () => {
  ElMessage.success('模型训练已启动，预计需要10分钟')
}

const exportModel = () => {
  ElMessage.success('模型导出成功')
}

const saveSystemConfig = () => {
  ElMessage.success('系统配置保存成功')
}

const restartZeek = () => {
  ElMessage.warning('正在重启Zeek...')
  setTimeout(() => {
    ElMessage.success('Zeek重启成功')
  }, 2000)
}
</script>

<style lang="scss" scoped>
.config-page {
  .form-tip {
    margin-left: 12px;
    font-size: 12px;
    color: #909399;
  }
}
</style>
