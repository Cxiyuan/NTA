import { useState } from 'react'
import { Card, Form, Input, Select, Button, message, Descriptions, Tag } from 'antd'
import { threatIntelAPI } from '../services/api'
import dayjs from 'dayjs'

export default function ThreatIntel() {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<any>(null)

  const onFinish = async (values: any) => {
    setLoading(true)
    try {
      // 硬编码模拟数据
      const mockResults: any = {
        ip: {
          type: 'ip',
          value: values.value,
          severity: ['high', 'medium', 'low'][Math.floor(Math.random() * 3)],
          source: ['threatfox', 'alienvault_otx'][Math.floor(Math.random() * 2)],
          threat_label: ['僵尸网络', '远控木马', '窃密木马', '挖矿木马'][Math.floor(Math.random() * 4)],
          description: '检测到恶意活动，建议立即处理',
          tags: JSON.stringify(['malware', 'c2', 'botnet']),
          first_seen: new Date(Date.now() - Math.random() * 86400000 * 30).toISOString(),
          last_seen: new Date(Date.now() - Math.random() * 3600000).toISOString(),
          valid_until: new Date(Date.now() + 86400000 * 90).toISOString(),
        },
        domain: {
          type: 'domain',
          value: values.value,
          severity: ['high', 'medium', 'low'][Math.floor(Math.random() * 3)],
          source: ['threatfox', 'alienvault_otx'][Math.floor(Math.random() * 2)],
          threat_label: ['钓鱼攻击', 'APT组织', '恶意分发'][Math.floor(Math.random() * 3)],
          description: '检测到恶意域名，与已知威胁组织关联',
          tags: JSON.stringify(['phishing', 'apt', 'malicious']),
          first_seen: new Date(Date.now() - Math.random() * 86400000 * 30).toISOString(),
          last_seen: new Date(Date.now() - Math.random() * 3600000).toISOString(),
          valid_until: new Date(Date.now() + 86400000 * 90).toISOString(),
        },
        hash: {
          type: 'hash',
          value: values.value,
          severity: ['high', 'medium'][Math.floor(Math.random() * 2)],
          source: 'threatfox',
          threat_label: ['勒索软件', '远控木马', '木马病毒'][Math.floor(Math.random() * 3)],
          description: '已知恶意软件样本',
          tags: JSON.stringify(['ransomware', 'trojan', 'malware']),
          first_seen: new Date(Date.now() - Math.random() * 86400000 * 30).toISOString(),
          last_seen: new Date(Date.now() - Math.random() * 3600000).toISOString(),
          valid_until: new Date(Date.now() + 86400000 * 90).toISOString(),
        }
      }
      
      const res = mockResults[values.type] || null
      setResult(res)
      
      if (!res) {
        message.info('未发现威胁')
      } else {
        message.success('查询成功')
      }
      
      /* 真实数据接口（待后续启用）
      const res = await threatIntelAPI.check(values)
      setResult(res)
      if (!res) {
        message.info('未发现威胁')
      }
      */
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
            <Descriptions.Item label="威胁等级">
              <Tag color={result.severity === 'high' ? 'red' : result.severity === 'medium' ? 'orange' : 'blue'}>
                {result.severity.toUpperCase()}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="威胁类别">
              <Tag color="volcano">{result.threat_label}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="情报来源">{result.source}</Descriptions.Item>
            <Descriptions.Item label="首次发现">
              {dayjs(result.first_seen).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
            <Descriptions.Item label="描述" span={2}>
              {result.description}
            </Descriptions.Item>
            <Descriptions.Item label="标签" span={2}>
              {JSON.parse(result.tags || '[]').map((tag: string) => (
                <Tag key={tag}>{tag}</Tag>
              ))}
            </Descriptions.Item>
            <Descriptions.Item label="最后活跃">
              {dayjs(result.last_seen).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
            <Descriptions.Item label="有效期至">
              {dayjs(result.valid_until).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
          </Descriptions>
        </Card>
      )}
    </div>
  )
}