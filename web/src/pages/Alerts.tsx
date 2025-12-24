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
      const res = await alertAPI.list({
        page: pagination.current,
        page_size: pagination.pageSize,
        ...filters,
      })
      setData(res.data || [])
      setPagination({ ...pagination, total: res.total })
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
      render: (text: string) => {
        const colors: any = {
          critical: 'red',
          high: 'orange',
          medium: 'gold',
          low: 'blue',
        }
        return <Tag color={colors[text]}>{text.toUpperCase()}</Tag>
      },
    },
    {
      title: '类型',
      dataIndex: 'type',
    },
    {
      title: '源IP',
      dataIndex: 'src_ip',
    },
    {
      title: '目标IP',
      dataIndex: 'dst_ip',
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
          <Select.Option value="medium">中危</Select.Option>
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
            <p><strong>源IP:</strong> {selectedAlert.src_ip}:{selectedAlert.src_port}</p>
            <p><strong>目标IP:</strong> {selectedAlert.dst_ip}:{selectedAlert.dst_port}</p>
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
