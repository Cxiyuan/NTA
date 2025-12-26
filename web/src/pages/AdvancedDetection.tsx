import { useState } from 'react'
import { Card, Form, Input, Button, Table, Tag, Tabs, Space, message } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import apiClient from '../utils/apiClient'

export default function AdvancedDetection() {
  const [dgaForm] = Form.useForm()
  const [dnsTunnelForm] = Form.useForm()
  const [c2Form] = Form.useForm()
  const [webshellForm] = Form.useForm()

  const [dgaResult, setDgaResult] = useState<any>(null)
  const [dnsTunnelResult, setDnsTunnelResult] = useState<any>(null)
  const [c2Result, setC2Result] = useState<any>(null)
  const [webshellResult, setWebshellResult] = useState<any>(null)

  const [loading, setLoading] = useState(false)

  const checkDGA = async (values: any) => {
    setLoading(true)
    try {
      const res = await apiClient.post('/api/v1/detection/dga', values)
      setDgaResult(res.data)
      if (res.data.is_dga) {
        message.warning(`检测到DGA域名！置信度: ${(res.data.confidence * 100).toFixed(1)}%`)
      } else {
        message.success('未检测到DGA特征')
      }
    } catch (error) {
      message.error('检测失败')
    } finally {
      setLoading(false)
    }
  }

  const checkDNSTunnel = async (values: any) => {
    setLoading(true)
    try {
      const res = await apiClient.post('/api/v1/detection/dns-tunnel', values)
      setDnsTunnelResult(res.data)
      if (res.data.is_tunnel) {
        message.warning(`检测到DNS隧道！置信度: ${(res.data.confidence * 100).toFixed(1)}%`)
      } else {
        message.success('未检测到DNS隧道特征')
      }
    } catch (error) {
      message.error('检测失败')
    } finally {
      setLoading(false)
    }
  }

  const checkC2 = async (values: any) => {
    setLoading(true)
    try {
      const res = await apiClient.post('/api/v1/detection/c2', values)
      setC2Result(res.data)
      if (res.data.is_c2) {
        message.warning(`检测到C2通信！类型: ${res.data.c2_type}，置信度: ${(res.data.confidence * 100).toFixed(1)}%`)
      } else {
        message.success('未检测到C2通信特征')
      }
    } catch (error) {
      message.error('检测失败')
    } finally {
      setLoading(false)
    }
  }

  const checkWebShell = async (values: any) => {
    setLoading(true)
    try {
      const res = await apiClient.post('/api/v1/detection/webshell', values)
      setWebshellResult(res.data)
      if (res.data.is_webshell) {
        message.error(`检测到WebShell！置信度: ${(res.data.confidence * 100).toFixed(1)}%`)
      } else {
        message.success('未检测到WebShell特征')
      }
    } catch (error) {
      message.error('检测失败')
    } finally {
      setLoading(false)
    }
  }

  const dgaTab = (
    <Card title="DGA 域名检测">
      <Form form={dgaForm} onFinish={checkDGA} layout="inline">
        <Form.Item name="domain" label="域名" rules={[{ required: true, message: '请输入域名' }]}>
          <Input placeholder="example.com" style={{ width: 300 }} />
        </Form.Item>
        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading} icon={<SearchOutlined />}>
            检测
          </Button>
        </Form.Item>
      </Form>

      {dgaResult && (
        <Card style={{ marginTop: 16 }} size="small">
          <p>
            <strong>检测结果: </strong>
            {dgaResult.is_dga ? (
              <Tag color="red">疑似DGA域名</Tag>
            ) : (
              <Tag color="green">正常域名</Tag>
            )}
          </p>
          <p><strong>置信度: </strong>{(dgaResult.confidence * 100).toFixed(1)}%</p>
          <p><strong>域名熵值: </strong>{dgaResult.entropy?.toFixed(2)}</p>
          <p><strong>检测指标:</strong></p>
          <ul>
            <li>元音比例: {dgaResult.vowel_ratio?.toFixed(2)}</li>
            <li>数字比例: {dgaResult.digit_ratio?.toFixed(2)}</li>
            <li>域名长度: {dgaResult.length}</li>
          </ul>
        </Card>
      )}
    </Card>
  )

  const dnsTunnelTab = (
    <Card title="DNS 隧道检测">
      <Form form={dnsTunnelForm} onFinish={checkDNSTunnel} layout="vertical">
        <Form.Item name="src_ip" label="源IP地址" rules={[{ required: true }]}>
          <Input placeholder="192.168.1.100" />
        </Form.Item>
        <Form.Item name="time_window" label="检测时间窗口（秒）" initialValue={300}>
          <Input type="number" />
        </Form.Item>
        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading} icon={<SearchOutlined />}>
            检测
          </Button>
        </Form.Item>
      </Form>

      {dnsTunnelResult && (
        <Card style={{ marginTop: 16 }} size="small">
          <p>
            <strong>检测结果: </strong>
            {dnsTunnelResult.is_tunnel ? (
              <Tag color="red">检测到DNS隧道</Tag>
            ) : (
              <Tag color="green">未发现异常</Tag>
            )}
          </p>
          <p><strong>置信度: </strong>{(dnsTunnelResult.confidence * 100).toFixed(1)}%</p>
          <p><strong>检测统计:</strong></p>
          <ul>
            <li>DNS查询数量: {dnsTunnelResult.query_count}</li>
            <li>平均域名长度: {dnsTunnelResult.avg_length?.toFixed(1)}</li>
            <li>请求频率: {dnsTunnelResult.request_rate?.toFixed(2)} 次/秒</li>
            <li>唯一域名数: {dnsTunnelResult.unique_domains}</li>
          </ul>
        </Card>
      )}
    </Card>
  )

  const c2Tab = (
    <Card title="C2 通信检测">
      <Form form={c2Form} onFinish={checkC2} layout="vertical">
        <Form.Item name="src_ip" label="源IP" rules={[{ required: true }]}>
          <Input placeholder="192.168.1.100" />
        </Form.Item>
        <Form.Item name="dst_ip" label="目标IP" rules={[{ required: true }]}>
          <Input placeholder="1.2.3.4" />
        </Form.Item>
        <Form.Item name="dst_port" label="目标端口" rules={[{ required: true }]}>
          <Input type="number" placeholder="443" />
        </Form.Item>
        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading} icon={<SearchOutlined />}>
            检测
          </Button>
        </Form.Item>
      </Form>

      {c2Result && (
        <Card style={{ marginTop: 16 }} size="small">
          <p>
            <strong>检测结果: </strong>
            {c2Result.is_c2 ? (
              <Tag color="red">疑似C2通信</Tag>
            ) : (
              <Tag color="green">未发现异常</Tag>
            )}
          </p>
          <p><strong>C2类型: </strong>{c2Result.c2_type}</p>
          <p><strong>置信度: </strong>{(c2Result.confidence * 100).toFixed(1)}%</p>
          <p><strong>连接特征:</strong></p>
          <ul>
            <li>连接时长: {c2Result.duration} 秒</li>
            <li>上传字节: {c2Result.orig_bytes}</li>
            <li>下载字节: {c2Result.resp_bytes}</li>
            <li>流量比: {c2Result.traffic_ratio?.toFixed(2)}</li>
          </ul>
        </Card>
      )}
    </Card>
  )

  const webshellTab = (
    <Card title="WebShell 检测">
      <Form form={webshellForm} onFinish={checkWebShell} layout="vertical">
        <Form.Item name="url" label="URL" rules={[{ required: true }]}>
          <Input placeholder="http://example.com/shell.php" />
        </Form.Item>
        <Form.Item name="payload" label="请求内容">
          <Input.TextArea rows={4} placeholder="GET/POST参数或请求体" />
        </Form.Item>
        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading} icon={<SearchOutlined />}>
            检测
          </Button>
        </Form.Item>
      </Form>

      {webshellResult && (
        <Card style={{ marginTop: 16 }} size="small">
          <p>
            <strong>检测结果: </strong>
            {webshellResult.is_webshell ? (
              <Tag color="red">检测到WebShell特征</Tag>
            ) : (
              <Tag color="green">未发现异常</Tag>
            )}
          </p>
          <p><strong>置信度: </strong>{(webshellResult.confidence * 100).toFixed(1)}%</p>
          <p><strong>匹配的危险函数:</strong></p>
          <ul>
            {webshellResult.matched_patterns?.map((pattern: string, idx: number) => (
              <li key={idx}><Tag color="red">{pattern}</Tag></li>
            ))}
          </ul>
        </Card>
      )}
    </Card>
  )

  const items = [
    { key: 'dga', label: 'DGA域名检测', children: dgaTab },
    { key: 'dns-tunnel', label: 'DNS隧道检测', children: dnsTunnelTab },
    { key: 'c2', label: 'C2通信检测', children: c2Tab },
    { key: 'webshell', label: 'WebShell检测', children: webshellTab },
  ]

  return (
    <div>
      <h2>高级威胁检测</h2>
      <Tabs items={items} />
    </div>
  )
}
