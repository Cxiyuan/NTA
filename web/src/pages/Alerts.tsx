import { useEffect, useState } from 'react'
import { Table, Tag, Button, Select, Space, Modal, Form, Input, message } from 'antd'
import { alertAPI } from '../services/api'
import dayjs from 'dayjs'

export default function Alerts() {
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)
  const [pagination, setPagination] = useState({ current: 1, pageSize: 50, total: 0 })
  const [filters, setFilters] = useState<any>({})
  const [modalVisible, setModalVisible] = useState(false)
  const [selectedAlert, setSelectedAlert] = useState<any>(null)

  useEffect(() => {
    loadData()
  }, [pagination.current, filters])

  const loadData = async () => {
    setLoading(true)
    try {
      // 硬编码模拟数据
      const mockData = Array.from({ length: 50 }, (_, i) => {
        const eventType = ['threat_intel_match', 'port_scan', 'c2_communication', 'data_exfiltration', 'dga_domain', 'webshell'][Math.floor(Math.random() * 6)]
        const isPortScan = eventType === 'port_scan'
        const isDGA = eventType === 'dga_domain'
        const hasDomain = Math.random() > 0.3 // 70%的请求有域名
        
        return {
          id: i + 1,
          timestamp: new Date(Date.now() - Math.random() * 86400000 * 7).toISOString(),
          severity: isPortScan ? ['medium', 'low'][Math.floor(Math.random() * 2)] : ['critical', 'high', 'medium', 'low'][Math.floor(Math.random() * 4)],
          type: eventType,
          src_ip: isPortScan ? `192.168.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}` : `${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}`,
          dst_ip: `${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}`,
          src_port: Math.floor(Math.random() * 60000) + 1024,
          dst_port: [80, 443, 22, 3389, 445, 3306, 1433, 8080][Math.floor(Math.random() * 8)],
          protocol: ['TCP', 'UDP', 'HTTP', 'HTTPS'][Math.floor(Math.random() * 4)],
          domain: hasDomain 
            ? (isDGA 
                ? `${Math.random().toString(36).substring(2, 15)}.${['com', 'net', 'org', 'ru'][Math.floor(Math.random() * 4)]}`
                : ['www.example.com', 'api.malicious.net', 'cdn.suspicious.org', 'mail.evil.ru'][Math.floor(Math.random() * 4)])
            : null,
          description: isPortScan ? '检测到端口扫描行为' : '检测到威胁活动',
          threat_label: isPortScan ? null : ['僵尸网络', '远控木马', '窃密木马', '挖矿木马', '勒索软件', 'APT组织', '钓鱼攻击', '渗透工具', '恶意分发', '高危威胁', null][Math.floor(Math.random() * 11)],
          threat_source: (eventType === 'threat_intel_match') ? ['threatfox', 'alienvault_otx'][Math.floor(Math.random() * 2)] : null,
          confidence: Math.random() * 0.3 + 0.7,
          status: ['new', 'investigating', 'resolved', 'false_positive'][Math.floor(Math.random() * 4)],
          details: 'Mock alert details',
        }
      })
      
      setData(mockData)
      setPagination({ ...pagination, total: 500 })
      
      /* 真实数据接口（待后续启用）
      const res = await alertAPI.list({
        page: pagination.current,
        page_size: pagination.pageSize,
        ...filters,
      })
      setData(res.data || [])
      setPagination({ ...pagination, total: res.total })
      */
    } catch (error) {
      message.error('加载失败')
    } finally {
      setLoading(false)
    }
  }

  const handleStatusChange = async (id: number, status: string) => {
    try {
      await alertAPI.update(id, { status })
      message.success('更新成功')
      loadData()
    } catch (error) {
      message.error('更新失败')
    }
  }

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
    },
    {
      title: '时间',
      dataIndex: 'timestamp',
      render: (text: string) => dayjs(text).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '等级',
      dataIndex: 'severity',
      width: 90,
      render: (text: string) => {
        const config: any = {
          critical: { label: '严重', color: 'red' },
          high: { label: '高危', color: 'orange' },
          medium: { label: '中等', color: 'gold' },
          low: { label: '低危', color: 'blue' },
        }
        return <Tag color={config[text]?.color || 'default'}>{config[text]?.label || text}</Tag>
      },
    },
    {
      title: '类型',
      dataIndex: 'type',
    },
    {
      title: '威胁类别',
      dataIndex: 'threat_label',
      render: (text: string) => {
        if (!text) return '-'
        const colors: any = {
          '僵尸网络': 'red',
          '远控木马': 'volcano',
          '窃密木马': 'orange',
          '挖矿木马': 'gold',
          '勒索软件': 'magenta',
          '木马病毒': 'red',
          '后门木马': 'red',
          '钓鱼攻击': 'orange',
          'APT组织': 'purple',
          '间谍木马': 'geekblue',
          '渗透工具': 'cyan',
          '恶意分发': 'volcano',
          '高危威胁': 'red',
          '恶意地址': 'orange',
          '可疑地址': 'gold',
        }
        return <Tag color={colors[text] || 'default'}>{text}</Tag>
      },
    },
    {
      title: '源IP',
      dataIndex: 'src_ip',
      width: 140,
    },
    {
      title: '源端口',
      dataIndex: 'src_port',
      width: 90,
    },
    {
      title: '目标IP',
      dataIndex: 'dst_ip',
      width: 140,
    },
    {
      title: '目标端口',
      dataIndex: 'dst_port',
      width: 90,
    },
    {
      title: '域名',
      dataIndex: 'domain',
      width: 180,
      ellipsis: true,
      render: (text: string) => text || '-',
    },
    {
      title: '描述',
      dataIndex: 'description',
      ellipsis: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      render: (text: string, record: any) => (
        <Select
          value={text}
          style={{ width: 120 }}
          onChange={(value) => handleStatusChange(record.id, value)}
        >
          <Select.Option value="new">新告警</Select.Option>
          <Select.Option value="investigating">处理中</Select.Option>
          <Select.Option value="resolved">已解决</Select.Option>
          <Select.Option value="false_positive">误报</Select.Option>
        </Select>
      ),
    },
    {
      title: '操作',
      render: (record: any) => (
        <Button
          type="link"
          onClick={() => {
            setSelectedAlert(record)
            setModalVisible(true)
          }}
        >
          详情
        </Button>
      ),
    },
  ]

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Select
          placeholder="等级筛选"
          style={{ width: 120 }}
          allowClear
          onChange={(value) => setFilters({ ...filters, severity: value })}
        >
          <Select.Option value="critical">严重</Select.Option>
          <Select.Option value="high">高危</Select.Option>
          <Select.Option value="medium">中等</Select.Option>
          <Select.Option value="low">低危</Select.Option>
        </Select>
        <Select
          placeholder="状态筛选"
          style={{ width: 120 }}
          allowClear
          onChange={(value) => setFilters({ ...filters, status: value })}
        >
          <Select.Option value="new">新告警</Select.Option>
          <Select.Option value="investigating">处理中</Select.Option>
          <Select.Option value="resolved">已解决</Select.Option>
          <Select.Option value="false_positive">误报</Select.Option>
        </Select>
      </Space>

      <Table
        columns={columns}
        dataSource={data}
        loading={loading}
        rowKey="id"
        pagination={{
          ...pagination,
          onChange: (page) => setPagination({ ...pagination, current: page }),
        }}
      />

      <Modal
        title="告警详情"
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={800}
      >
        {selectedAlert && (
          <div>
            <p><strong>ID:</strong> {selectedAlert.id}</p>
            <p><strong>时间:</strong> {dayjs(selectedAlert.timestamp).format('YYYY-MM-DD HH:mm:ss')}</p>
            <p><strong>等级:</strong> {selectedAlert.severity}</p>
            <p><strong>类型:</strong> {selectedAlert.type}</p>
            {selectedAlert.threat_label && (
              <p><strong>威胁类别:</strong> <Tag>{selectedAlert.threat_label}</Tag></p>
            )}
            {selectedAlert.threat_source && (
              <p><strong>情报来源:</strong> {selectedAlert.threat_source}</p>
            )}
            <p><strong>源IP:</strong> {selectedAlert.src_ip}:{selectedAlert.src_port}</p>
            <p><strong>目标IP:</strong> {selectedAlert.dst_ip}:{selectedAlert.dst_port}</p>
            {selectedAlert.domain && (
              <p><strong>域名:</strong> {selectedAlert.domain}</p>
            )}
            <p><strong>协议:</strong> {selectedAlert.protocol}</p>
            <p><strong>描述:</strong> {selectedAlert.description}</p>
            <p><strong>置信度:</strong> {(selectedAlert.confidence * 100).toFixed(0)}%</p>
            <p><strong>详情:</strong></p>
            <pre>{selectedAlert.details}</pre>
          </div>
        )}
      </Modal>
    </div>
  )
}