import { useState } from 'react'
import { Card, Form, Input, Button, Table, Tag, DatePicker, Space, message, Modal } from 'antd'
import { SearchOutlined, DownloadOutlined, EyeOutlined } from '@ant-design/icons'
import { pcapAPI } from '../services/api'
import dayjs from 'dayjs'

export default function PcapAnalysis() {
  const [form] = Form.useForm()
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)
  const [detailVisible, setDetailVisible] = useState(false)
  const [selectedSession, setSelectedSession] = useState<any>(null)

  const handleSearch = async (values: any) => {
    setLoading(true)
    try {
      const params: any = {
        src_ip: values.src_ip,
        dst_ip: values.dst_ip,
        limit: 100,
      }

      if (values.dateRange) {
        params.start_time = values.dateRange[0].toISOString()
        params.end_time = values.dateRange[1].toISOString()
      }

      const res = await pcapAPI.search(params)
      setData(res)
      message.success(`找到 ${res.length} 个会话`)
    } catch (error) {
      message.error('搜索失败')
    } finally {
      setLoading(false)
    }
  }

  const handleDownload = async (sessionId: string) => {
    try {
      const blob = await pcapAPI.download(sessionId)
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${sessionId}.pcap`
      a.click()
      message.success('下载成功')
    } catch (error) {
      message.error('下载失败')
    }
  }

  const showDetail = (record: any) => {
    setSelectedSession(record)
    setDetailVisible(true)
  }

  const columns = [
    {
      title: '会话ID',
      dataIndex: 'session_id',
      key: 'session_id',
      ellipsis: true,
      width: 250,
    },
    {
      title: '源IP',
      dataIndex: 'src_ip',
      key: 'src_ip',
    },
    {
      title: '源端口',
      dataIndex: 'src_port',
      key: 'src_port',
      width: 100,
    },
    {
      title: '目标IP',
      dataIndex: 'dst_ip',
      key: 'dst_ip',
    },
    {
      title: '目标端口',
      dataIndex: 'dst_port',
      key: 'dst_port',
      width: 100,
    },
    {
      title: '协议',
      dataIndex: 'protocol',
      key: 'protocol',
      render: (text: string) => <Tag>{text}</Tag>,
      width: 80,
    },
    {
      title: '开始时间',
      dataIndex: 'start_time',
      key: 'start_time',
      render: (text: string) => dayjs(text).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '结束时间',
      dataIndex: 'end_time',
      key: 'end_time',
      render: (text: string) => text ? dayjs(text).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '数据包数',
      dataIndex: 'packet_count',
      key: 'packet_count',
      width: 100,
    },
    {
      title: '总字节数',
      dataIndex: 'bytes_total',
      key: 'bytes_total',
      render: (bytes: number) => {
        if (bytes < 1024) return `${bytes} B`
        if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
        return `${(bytes / 1024 / 1024).toFixed(1)} MB`
      },
      width: 120,
    },
    {
      title: '操作',
      key: 'action',
      width: 180,
      render: (record: any) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => showDetail(record)}
          >
            详情
          </Button>
          <Button
            type="link"
            size="small"
            icon={<DownloadOutlined />}
            onClick={() => handleDownload(record.session_id)}
          >
            下载
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Card title="PCAP 流量回溯" style={{ marginBottom: 16 }}>
        <Form form={form} onFinish={handleSearch} layout="inline">
          <Form.Item name="src_ip" label="源IP">
            <Input placeholder="192.168.1.100" style={{ width: 150 }} />
          </Form.Item>
          <Form.Item name="dst_ip" label="目标IP">
            <Input placeholder="1.2.3.4" style={{ width: 150 }} />
          </Form.Item>
          <Form.Item name="dateRange" label="时间范围">
            <DatePicker.RangePicker showTime />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} icon={<SearchOutlined />}>
              搜索
            </Button>
          </Form.Item>
        </Form>
      </Card>

      <Card title="会话列表">
        <Table
          columns={columns}
          dataSource={data}
          loading={loading}
          rowKey="id"
          scroll={{ x: 1400 }}
          pagination={{
            pageSize: 20,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
        />
      </Card>

      <Modal
        title="会话详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailVisible(false)}>
            关闭
          </Button>,
          <Button
            key="download"
            type="primary"
            icon={<DownloadOutlined />}
            onClick={() => selectedSession && handleDownload(selectedSession.session_id)}
          >
            下载PCAP
          </Button>,
        ]}
        width={800}
      >
        {selectedSession && (
          <div>
            <p><strong>会话ID:</strong> {selectedSession.session_id}</p>
            <p><strong>五元组:</strong></p>
            <ul>
              <li>源IP: {selectedSession.src_ip}</li>
              <li>源端口: {selectedSession.src_port}</li>
              <li>目标IP: {selectedSession.dst_ip}</li>
              <li>目标端口: {selectedSession.dst_port}</li>
              <li>协议: {selectedSession.protocol}</li>
            </ul>
            <p><strong>时间信息:</strong></p>
            <ul>
              <li>开始时间: {dayjs(selectedSession.start_time).format('YYYY-MM-DD HH:mm:ss')}</li>
              <li>结束时间: {selectedSession.end_time ? dayjs(selectedSession.end_time).format('YYYY-MM-DD HH:mm:ss') : '进行中'}</li>
              <li>持续时长: {selectedSession.end_time ? 
                Math.floor((new Date(selectedSession.end_time).getTime() - new Date(selectedSession.start_time).getTime()) / 1000) + ' 秒' 
                : '未结束'}
              </li>
            </ul>
            <p><strong>流量统计:</strong></p>
            <ul>
              <li>数据包数量: {selectedSession.packet_count}</li>
              <li>总字节数: {selectedSession.bytes_total} bytes</li>
              <li>PCAP文件路径: <code>{selectedSession.file_path}</code></li>
            </ul>
            <div style={{ marginTop: 16, padding: 12, background: '#f0f2f5', borderRadius: 4 }}>
              <p style={{ margin: 0, fontSize: 12, color: '#666' }}>
                💡 提示：下载的PCAP文件可使用 Wireshark 或 tcpdump 进行深度分析
              </p>
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}
