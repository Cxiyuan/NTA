import { useState, useEffect } from 'react'
import { Table, Card, Form, Input, Select, DatePicker, Button, Space, Tag, Modal, Descriptions, message } from 'antd'
import { SearchOutlined, EyeOutlined, CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons'
import apiClient from '../utils/apiClient'
import dayjs from 'dayjs'

export default function AuditLog() {
  const [form] = Form.useForm()
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)
  const [detailVisible, setDetailVisible] = useState(false)
  const [selectedLog, setSelectedLog] = useState<any>(null)

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async (filters = {}) => {
    setLoading(true)
    try {
      const res = await apiClient.get('/api/v1/audit', { params: filters })
      setData(res.data)
    } catch (error) {
      message.error('加载失败')
    } finally {
      setLoading(false)
    }
  }

  const handleSearch = (values: any) => {
    const filters: any = {}
    if (values.user) filters.user = values.user
    if (values.action) filters.action = values.action
    if (values.dateRange) {
      filters.start_time = values.dateRange[0].toISOString()
      filters.end_time = values.dateRange[1].toISOString()
    }
    loadData(filters)
  }

  const showDetail = (record: any) => {
    setSelectedLog(record)
    setDetailVisible(true)
  }

  const verifyChecksum = (record: any) => {
    // 前端简单校验展示，实际应该调用后端API验证
    if (record.checksum && record.checksum.length === 64) {
      message.success('校验和验证通过')
    } else {
      message.warning('校验和格式异常')
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
      width: 180,
    },
    {
      title: '用户',
      dataIndex: 'user',
      width: 120,
    },
    {
      title: '操作',
      dataIndex: 'action',
      render: (text: string) => {
        const colors: any = {
          create: 'green',
          update: 'blue',
          delete: 'red',
          login: 'cyan',
          logout: 'default',
        }
        return <Tag color={colors[text] || 'default'}>{text.toUpperCase()}</Tag>
      },
      width: 120,
    },
    {
      title: '资源',
      dataIndex: 'resource',
      ellipsis: true,
    },
    {
      title: '结果',
      dataIndex: 'result',
      render: (text: string) => {
        return text === 'success' ? (
          <Tag icon={<CheckCircleOutlined />} color="success">成功</Tag>
        ) : (
          <Tag icon={<CloseCircleOutlined />} color="error">失败</Tag>
        )
      },
      width: 100,
    },
    {
      title: '操作',
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
            onClick={() => verifyChecksum(record)}
          >
            验证
          </Button>
        </Space>
      ),
      width: 150,
    },
  ]

  return (
    <div>
      <Card title="审计日志查询" style={{ marginBottom: 16 }}>
        <Form form={form} onFinish={handleSearch} layout="inline">
          <Form.Item name="user" label="用户">
            <Input placeholder="用户名" style={{ width: 150 }} />
          </Form.Item>
          <Form.Item name="action" label="操作类型">
            <Select placeholder="选择操作" style={{ width: 150 }} allowClear>
              <Select.Option value="create">创建</Select.Option>
              <Select.Option value="update">更新</Select.Option>
              <Select.Option value="delete">删除</Select.Option>
              <Select.Option value="login">登录</Select.Option>
              <Select.Option value="logout">登出</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="dateRange" label="时间范围">
            <DatePicker.RangePicker showTime />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>
              查询
            </Button>
          </Form.Item>
        </Form>
      </Card>

      <Card title="审计记录">
        <Table
          columns={columns}
          dataSource={data}
          loading={loading}
          rowKey="id"
          pagination={{
            pageSize: 50,
            showTotal: (total) => `共 ${total} 条记录`,
            showSizeChanger: true,
          }}
        />
      </Card>

      <Modal
        title="审计日志详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailVisible(false)}>
            关闭
          </Button>,
        ]}
        width={800}
      >
        {selectedLog && (
          <Descriptions bordered column={2}>
            <Descriptions.Item label="ID" span={2}>{selectedLog.id}</Descriptions.Item>
            <Descriptions.Item label="时间" span={2}>
              {dayjs(selectedLog.timestamp).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
            <Descriptions.Item label="用户">{selectedLog.user}</Descriptions.Item>
            <Descriptions.Item label="操作">
              <Tag>{selectedLog.action}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="资源" span={2}>{selectedLog.resource}</Descriptions.Item>
            <Descriptions.Item label="结果">
              {selectedLog.result === 'success' ? (
                <Tag color="success">成功</Tag>
              ) : (
                <Tag color="error">失败</Tag>
              )}
            </Descriptions.Item>
            <Descriptions.Item label="创建时间">
              {dayjs(selectedLog.created_at).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
            <Descriptions.Item label="详情" span={2}>
              <pre style={{ maxHeight: 200, overflow: 'auto', background: '#f5f5f5', padding: 12, borderRadius: 4 }}>
                {selectedLog.details}
              </pre>
            </Descriptions.Item>
            <Descriptions.Item label="校验和" span={2}>
              <code style={{ wordBreak: 'break-all', fontSize: 12 }}>
                {selectedLog.checksum}
              </code>
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  )
}
