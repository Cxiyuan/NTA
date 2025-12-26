import { useState, useEffect } from 'react'
import { Table, Card, Button, Space, Modal, Form, Input, InputNumber, Select, message, Tag, Popconfirm } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, TeamOutlined } from '@ant-design/icons'
import apiClient from '../utils/apiClient'
import dayjs from 'dayjs'

export default function TenantManagement() {
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [usersModalVisible, setUsersModalVisible] = useState(false)
  const [editingTenant, setEditingTenant] = useState<any>(null)
  const [selectedTenant, setSelectedTenant] = useState<any>(null)
  const [tenantUsers, setTenantUsers] = useState([])
  const [form] = Form.useForm()

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const res = await apiClient.get('/api/v1/tenants')
      setData(res.data)
    } catch (error) {
      message.error('加载失败')
    } finally {
      setLoading(false)
    }
  }

  const loadTenantUsers = async (tenantId: string) => {
    try {
      const res = await apiClient.get(`/api/v1/tenants/${tenantId}/users`)
      setTenantUsers(res.data)
    } catch (error) {
      message.error('获取租户用户失败')
    }
  }

  const handleAdd = () => {
    setEditingTenant(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (record: any) => {
    setEditingTenant(record)
    form.setFieldsValue(record)
    setModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await apiClient.delete(`/api/v1/tenants/${id}`)
      message.success('删除成功')
      loadData()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleViewUsers = async (record: any) => {
    setSelectedTenant(record)
    await loadTenantUsers(record.tenant_id)
    setUsersModalVisible(true)
  }

  const handleSubmit = async (values: any) => {
    try {
      if (editingTenant) {
        await apiClient.put(`/api/v1/tenants/${editingTenant.id}`, values)
        message.success('更新成功')
      } else {
        await apiClient.post('/api/v1/tenants', values)
        message.success('创建成功')
      }
      setModalVisible(false)
      loadData()
    } catch (error) {
      message.error(editingTenant ? '更新失败' : '创建失败')
    }
  }

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
    },
    {
      title: '租户ID',
      dataIndex: 'tenant_id',
    },
    {
      title: '名称',
      dataIndex: 'name',
    },
    {
      title: '描述',
      dataIndex: 'description',
      ellipsis: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      render: (status: string) => {
        const colors: any = {
          active: 'green',
          inactive: 'default',
          suspended: 'red',
        }
        return <Tag color={colors[status]}>{status.toUpperCase()}</Tag>
      },
    },
    {
      title: '最大探针数',
      dataIndex: 'max_probes',
      render: (val: number) => val || '无限制',
    },
    {
      title: '最大资产数',
      dataIndex: 'max_assets',
      render: (val: number) => val || '无限制',
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      render: (text: string) => dayjs(text).format('YYYY-MM-DD'),
    },
    {
      title: '操作',
      render: (record: any) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<TeamOutlined />}
            onClick={() => handleViewUsers(record)}
          >
            用户
          </Button>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定删除此租户吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  const userColumns = [
    {
      title: '用户名',
      dataIndex: 'username',
    },
    {
      title: '邮箱',
      dataIndex: 'email',
    },
    {
      title: '状态',
      dataIndex: 'status',
      render: (status: string) => <Tag color={status === 'active' ? 'green' : 'default'}>{status}</Tag>,
    },
  ]

  return (
    <div>
      <Card
        title="租户管理"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
            新增租户
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={data}
          loading={loading}
          rowKey="id"
        />
      </Card>

      <Modal
        title={editingTenant ? '编辑租户' : '新增租户'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
        width={600}
      >
        <Form form={form} onFinish={handleSubmit} layout="vertical">
          <Form.Item
            name="tenant_id"
            label="租户ID"
            rules={[{ required: true, message: '请输入租户ID' }]}
          >
            <Input placeholder="tenant_001" disabled={!!editingTenant} />
          </Form.Item>

          <Form.Item
            name="name"
            label="租户名称"
            rules={[{ required: true, message: '请输入租户名称' }]}
          >
            <Input placeholder="企业名称" />
          </Form.Item>

          <Form.Item
            name="description"
            label="描述"
          >
            <Input.TextArea rows={3} placeholder="租户描述信息" />
          </Form.Item>

          <Form.Item
            name="status"
            label="状态"
            initialValue="active"
            rules={[{ required: true }]}
          >
            <Select>
              <Select.Option value="active">激活</Select.Option>
              <Select.Option value="inactive">未激活</Select.Option>
              <Select.Option value="suspended">已停用</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="max_probes"
            label="最大探针数"
            initialValue={0}
            tooltip="0表示无限制"
          >
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            name="max_assets"
            label="最大资产数"
            initialValue={0}
            tooltip="0表示无限制"
          >
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={`租户用户列表 - ${selectedTenant?.name}`}
        open={usersModalVisible}
        onCancel={() => setUsersModalVisible(false)}
        footer={null}
        width={800}
      >
        <Table
          columns={userColumns}
          dataSource={tenantUsers}
          rowKey="id"
          pagination={{ pageSize: 10 }}
        />
      </Modal>
    </div>
  )
}
