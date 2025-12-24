import { Card, Form, Input, Button, Switch, message, Tabs, InputNumber, Select } from 'antd'
import { useState, useEffect } from 'react'
import { notificationAPI } from '../services/api'
import axios from 'axios'

export default function Settings() {
  const [notifForm] = Form.useForm()
  const [detectionForm] = Form.useForm()
  const [backupForm] = Form.useForm()
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    loadSettings()
  }, [])

  const loadSettings = async () => {
    try {
      const [notifRes, configRes] = await Promise.all([
        axios.get('/api/v1/notifications/config'),
        axios.get('/api/v1/config'),
      ])
      notifForm.setFieldsValue(notifRes.data)
      detectionForm.setFieldsValue(configRes.data.detection)
      backupForm.setFieldsValue(configRes.data.backup)
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
      await axios.put('/api/v1/config/detection', values)
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
      await axios.put('/api/v1/config/backup', values)
      message.success('保存成功')
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
