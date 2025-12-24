import { useState } from 'react'
import { Card, Form, Input, Select, Button, message, Descriptions } from 'antd'
import { threatIntelAPI } from '../services/api'

export default function ThreatIntel() {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<any>(null)

  const onFinish = async (values: any) => {
    setLoading(true)
    try {
      const res = await threatIntelAPI.check(values)
      setResult(res)
      if (!res) {
        message.info('未发现威胁')
      }
    } catch (error) {
      message.error('查询失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <Card title="威胁情报查询" style={{ marginBottom: 16 }}>
        <Form form={form} onFinish={onFinish} layout="inline">
          <Form.Item name="type" label="类型" rules={[{ required: true }]}>
            <Select style={{ width: 120 }}>
              <Select.Option value="ip">IP地址</Select.Option>
              <Select.Option value="domain">域名</Select.Option>
              <Select.Option value="hash">文件哈希</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="value" label="IOC值" rules={[{ required: true }]}>
            <Input placeholder="输入IP/域名/哈希" style={{ width: 300 }} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading}>
              查询
            </Button>
          </Form.Item>
        </Form>
      </Card>

      {result && (
        <Card title="查询结果">
          <Descriptions bordered column={2}>
            <Descriptions.Item label="类型">{result.type}</Descriptions.Item>
            <Descriptions.Item label="值">{result.value}</Descriptions.Item>
            <Descriptions.Item label="威胁等级">{result.severity}</Descriptions.Item>
            <Descriptions.Item label="来源">{result.source}</Descriptions.Item>
            <Descriptions.Item label="标签" span={2}>
              {result.tags}
            </Descriptions.Item>
            <Descriptions.Item label="有效期至" span={2}>
              {result.valid_until}
            </Descriptions.Item>
          </Descriptions>
        </Card>
      )}
    </div>
  )
}
