import { Card, Form, Input, Button, Switch, message, Tabs, InputNumber, Select, Statistic, Row, Col, Typography } from 'antd'
import { useState, useEffect } from 'react'
import { notificationAPI } from '../services/api'
import apiClient from '../utils/apiClient'

const { Text } = Typography

export default function Settings() {
  const [notifForm] = Form.useForm()
  const [detectionForm] = Form.useForm()
  const [backupForm] = Form.useForm()
  const [threatIntelForm] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [threatIntelStats, setThreatIntelStats] = useState<any>({})

  useEffect(() => {
    loadSettings()
  }, [])

  const loadSettings = async () => {
    try {
      // 硬编码模拟数据
      const mockNotifData = {
        email_enabled: false,
        email_smtp: 'smtp.example.com:587',
        email_username: '',
        email_password: '',
        email_recipients: '',
        webhook_enabled: false,
        webhook_url: '',
        dingtalk_enabled: false,
        dingtalk_webhook: '',
      }
      
      const mockConfigData = {
        detection: {
          scan: { threshold: 20, time_window: 300, min_fail_rate: 0.6 },
          auth: { fail_threshold: 5, pth_window: 3600 },
          ml: { enabled: true, contamination: 0.01 },
        },
        backup: {
          enabled: true,
          backup_dir: '/opt/nta-probe/backups',
          interval_hours: 24,
          retention_days: 7,
        },
      }
      
      const mockThreatIntelData = {
        update_interval_hours: 24,
        update_hour: 2,
        enable_local_db: true,
        sources: [
          { name: 'ThreatFox', enabled: true },
          { name: 'AlienVault OTX', enabled: true },
        ],
        last_sync_time: new Date(Date.now() - Math.random() * 3600000).toISOString(),
        total_iocs: Math.floor(Math.random() * 50000) + 10000,
      }
      
      notifForm.setFieldsValue(mockNotifData)
      detectionForm.setFieldsValue(mockConfigData.detection)
      backupForm.setFieldsValue(mockConfigData.backup)
      threatIntelForm.setFieldsValue({
        update_interval_hours: mockThreatIntelData.update_interval_hours,
        update_hour: mockThreatIntelData.update_hour,
      })
      setThreatIntelStats(mockThreatIntelData)
      
      /* 真实数据接口（待后续启用）
      const [notifRes, configRes, threatIntelRes] = await Promise.all([
        apiClient.get('/api/v1/notifications/config'),
        apiClient.get('/api/v1/config'),
        apiClient.get('/api/v1/config/threat-intel'),
      ])
      notifForm.setFieldsValue(notifRes.data)
      detectionForm.setFieldsValue(configRes.data.detection)
      backupForm.setFieldsValue(configRes.data.backup)
      threatIntelForm.setFieldsValue({
        update_interval_hours: threatIntelRes.data.update_interval_hours,
        update_hour: threatIntelRes.data.update_hour,
      })
      setThreatIntelStats(threatIntelRes.data)
      */
    } catch (error) {
      console.error('Failed to load settings')
    }
  }

  const handleNotifSubmit = async (values: any) => {
    setLoading(true)
    try {
      await notificationAPI.update(values)
      message.success('保存成功')
    } catch (error) {
      message.error('保存失败')
    } finally {
      setLoading(false)
    }
  }

  const handleDetectionSubmit = async (values: any) => {
    setLoading(true)
    try {
      await apiClient.put('/api/v1/config/detection', values)
      message.success('保存成功')
    } catch (error) {
      message.error('保存失败')
    } finally {
      setLoading(false)
    }
  }

  const handleBackupSubmit = async (values: any) => {
    setLoading(true)
    try {
      await apiClient.put('/api/v1/config/backup', values)
      message.success('保存成功')
    } catch (error) {
      message.error('保存失败')
    } finally {
      setLoading(false)
    }
  }

  const handleThreatIntelSubmit = async (values: any) => {
    setLoading(true)
    try {
      await apiClient.put('/api/v1/config/threat-intel', values)
      message.success('威胁情报配置已更新')
      loadSettings()
    } catch (error) {
      message.error('保存失败')
    } finally {
      setLoading(false)
    }
  }

  const notificationTab = (
    <Form form={notifForm} layout="vertical" onFinish={handleNotifSubmit}>
      <Form.Item name="email_enabled" label="邮件通知" valuePropName="checked">
        <Switch />
      </Form.Item>
      <Form.Item name="email_smtp" label="SMTP服务器">
        <Input placeholder="smtp.example.com:587" />
      </Form.Item>
      <Form.Item name="email_username" label="邮箱用户名">
        <Input />
      </Form.Item>
      <Form.Item name="email_password" label="邮箱密码">
        <Input.Password />
      </Form.Item>
      <Form.Item name="email_recipients" label="接收人">
        <Input placeholder="多个邮箱用逗号分隔" />
      </Form.Item>

      <Form.Item name="webhook_enabled" label="Webhook通知" valuePropName="checked">
        <Switch />
      </Form.Item>
      <Form.Item name="webhook_url" label="Webhook URL">
        <Input placeholder="https://example.com/webhook" />
      </Form.Item>

      <Form.Item name="dingtalk_enabled" label="钉钉通知" valuePropName="checked">
        <Switch />
      </Form.Item>
      <Form.Item name="dingtalk_webhook" label="钉钉Webhook">
        <Input placeholder="https://oapi.dingtalk.com/robot/send?access_token=xxx" />
      </Form.Item>

      <Form.Item>
        <Button type="primary" htmlType="submit" loading={loading}>
          保存设置
        </Button>
      </Form.Item>
    </Form>
  )

  const detectionTab = (
    <Form form={detectionForm} layout="vertical" onFinish={handleDetectionSubmit}>
      <h4>端口扫描检测</h4>
      <Form.Item name={['scan', 'threshold']} label="扫描阈值">
        <InputNumber min={1} addonAfter="次" />
      </Form.Item>
      <Form.Item name={['scan', 'time_window']} label="时间窗口">
        <InputNumber min={1} addonAfter="秒" />
      </Form.Item>
      <Form.Item name={['scan', 'min_fail_rate']} label="最小失败率">
        <InputNumber min={0} max={1} step={0.1} />
      </Form.Item>

      <h4>认证攻击检测</h4>
      <Form.Item name={['auth', 'fail_threshold']} label="失败次数阈值">
        <InputNumber min={1} addonAfter="次" />
      </Form.Item>
      <Form.Item name={['auth', 'pth_window']} label="PTH检测时间窗口">
        <InputNumber min={1} addonAfter="秒" />
      </Form.Item>

      <h4>机器学习检测</h4>
      <Form.Item name={['ml', 'enabled']} label="启用ML检测" valuePropName="checked">
        <Switch />
      </Form.Item>
      <Form.Item name={['ml', 'contamination']} label="异常比例">
        <InputNumber min={0} max={1} step={0.01} />
      </Form.Item>

      <Form.Item>
        <Button type="primary" htmlType="submit" loading={loading}>
          保存设置
        </Button>
      </Form.Item>
    </Form>
  )

  const backupTab = (
    <Form form={backupForm} layout="vertical" onFinish={handleBackupSubmit}>
      <Form.Item name="enabled" label="启用自动备份" valuePropName="checked">
        <Switch />
      </Form.Item>
      <Form.Item name="backup_dir" label="备份目录">
        <Input placeholder="/opt/nta-probe/backups" />
      </Form.Item>
      <Form.Item name="interval_hours" label="备份间隔">
        <InputNumber min={1} addonAfter="小时" />
      </Form.Item>
      <Form.Item name="retention_days" label="保留天数">
        <InputNumber min={1} addonAfter="天" />
      </Form.Item>

      <Form.Item>
        <Button type="primary" htmlType="submit" loading={loading}>
          保存设置
        </Button>
      </Form.Item>
    </Form>
  )

  const threatIntelTab = (
    <div>
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={8}>
          <Card>
            <Statistic title="威胁情报总数" value={threatIntelStats.total_iocs || 0} />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic title="上次同步时间" value={threatIntelStats.last_sync_time || '未同步'} valueStyle={{ fontSize: 16 }} />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Text strong>数据源状态</Text>
            <div style={{ marginTop: 8 }}>
              {threatIntelStats.sources?.map((src: any) => (
                <div key={src.name}>
                  <Text>{src.name}: </Text>
                  <Text type={src.enabled ? 'success' : 'secondary'}>
                    {src.enabled ? '已启用' : '已禁用'}
                  </Text>
                </div>
              ))}
            </div>
          </Card>
        </Col>
      </Row>

      <Form form={threatIntelForm} layout="vertical" onFinish={handleThreatIntelSubmit}>
        <Form.Item 
          name="update_interval_hours" 
          label="更新间隔" 
          rules={[{ required: true, message: '请输入更新间隔' }]}
          tooltip="威胁情报自动同步的间隔时间，单位：小时"
        >
          <InputNumber min={1} max={720} addonAfter="小时" style={{ width: 200 }} />
        </Form.Item>

        <Form.Item 
          name="update_hour" 
          label="更新时间点" 
          rules={[{ required: true, message: '请选择更新时间点' }]}
          tooltip="每天自动同步威胁情报的时间点（24小时制）"
        >
          <Select style={{ width: 200 }}>
            {Array.from({ length: 24 }, (_, i) => (
              <Select.Option key={i} value={i}>
                {i.toString().padStart(2, '0')}:00
              </Select.Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading}>
            保存设置
          </Button>
          <Text type="secondary" style={{ marginLeft: 16 }}>
            配置将在下次定时任务时生效
          </Text>
        </Form.Item>
      </Form>
    </div>
  )

  const items = [
    {
      key: 'notification',
      label: '通知设置',
      children: notificationTab,
    },
    {
      key: 'detection',
      label: '检测规则',
      children: detectionTab,
    },
    {
      key: 'threat-intel',
      label: '威胁情报',
      children: threatIntelTab,
    },
    {
      key: 'backup',
      label: '备份策略',
      children: backupTab,
    },
  ]

  return (
    <Card title="系统设置">
      <Tabs items={items} />
    </Card>
  )
}