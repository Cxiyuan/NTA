import { useState, useEffect } from 'react'
import { Table, Card, Button, Space, Modal, Form, Input, message, Tag, Popconfirm, Drawer, Tree } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, SafetyOutlined } from '@ant-design/icons'
import apiClient from '../utils/apiClient'
import type { DataNode } from 'antd/es/tree'

export default function RoleManagement() {
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [permDrawerVisible, setPermDrawerVisible] = useState(false)
  const [editingRole, setEditingRole] = useState<any>(null)
  const [selectedRole, setSelectedRole] = useState<any>(null)
  const [form] = Form.useForm()
  const [permForm] = Form.useForm()

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const res = await apiClient.get('/api/v1/roles')
      setData(res.data)
    } catch (error) {
      message.error('加载失败')
    } finally {
      setLoading(false)
    }
  }

  const handleAdd = () => {
    setEditingRole(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (record: any) => {
    setEditingRole(record)
    form.setFieldsValue({
      name: record.name,
      description: record.description,
    })
    setModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await apiClient.delete(`/api/v1/roles/${id}`)
      message.success('删除成功')
      loadData()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handlePermission = (record: any) => {
    setSelectedRole(record)
    try {
      const perms = JSON.parse(record.permissions)
      permForm.setFieldsValue({ permissions: JSON.stringify(perms, null, 2) })
    } catch {
      permForm.setFieldsValue({ permissions: '[]' })
    }
    setPermDrawerVisible(true)
  }

  const handleSubmit = async (values: any) => {
    try {
      if (editingRole) {
        await apiClient.put(`/api/v1/roles/${editingRole.id}`, values)
        message.success('更新成功')
      } else {
        await apiClient.post('/api/v1/roles', values)
        message.success('创建成功')
      }
      setModalVisible(false)
      loadData()
    } catch (error) {
      message.error(editingRole ? '更新失败' : '创建失败')
    }
  }

  const handlePermSubmit = async (values: any) => {
    try {
      // 验证JSON格式
      JSON.parse(values.permissions)
      
      await apiClient.put(`/api/v1/roles/${selectedRole.id}/permissions`, {
        permissions: values.permissions,
      })
      message.success('权限更新成功')
      setPermDrawerVisible(false)
      loadData()
    } catch (error: any) {
      if (error instanceof SyntaxError) {
        message.error('权限JSON格式错误')
      } else {
        message.error('权限更新失败')
      }
    }
  }

  const permissionTreeData: DataNode[] = [
    {
      title: '告警管理',
      key: 'alerts',
      children: [
        { title: '查看告警', key: 'alerts:read' },
        { title: '更新告警', key: 'alerts:update' },
        { title: '删除告警', key: 'alerts:delete' },
      ],
    },
    {
      title: '资产管理',
      key: 'assets',
      children: [
        { title: '查看资产', key: 'assets:read' },
        { title: '更新资产', key: 'assets:update' },
      ],
    },
    {
      title: '威胁情报',
      key: 'threat_intel',
      children: [
        { title: '查询情报', key: 'threat_intel:read' },
        { title: '更新情报', key: 'threat_intel:update' },
      ],
    },
    {
      title: '探针管理',
      key: 'probes',
      children: [
        { title: '查看探针', key: 'probes:read' },
        { title: '注册探针', key: 'probes:create' },
        { title: '删除探针', key: 'probes:delete' },
      ],
    },
    {
      title: '审计日志',
      key: 'audit',
      children: [
        { title: '查看日志', key: 'audit:read' },
      ],
    },
    {
      title: '用户管理',
      key: 'users',
      children: [
        { title: '查看用户', key: 'users:read' },
        { title: '创建用户', key: 'users:create' },
        { title: '更新用户', key: 'users:update' },
        { title: '删除用户', key: 'users:delete' },
      ],
    },
  ]

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
    },
    {
      title: '角色名称',
      dataIndex: 'name',
      render: (text: string) => {
        const colors: any = {
          admin: 'red',
          analyst: 'blue',
          viewer: 'green',
        }
        return <Tag color={colors[text] || 'default'}>{text.toUpperCase()}</Tag>
      },
    },
    {
      title: '描述',
      dataIndex: 'description',
    },
    {
      title: '权限数量',
      dataIndex: 'permissions',
      render: (text: string) => {
        try {
          const perms = JSON.parse(text)
          return <Tag>{Array.isArray(perms) ? perms.length : 0} 项</Tag>
        } catch {
          return <Tag>0 项</Tag>
        }
      },
    },
    {
      title: '操作',
      render: (record: any) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<SafetyOutlined />}
            onClick={() => handlePermission(record)}
          >
            权限
          </Button>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          {!['admin', 'analyst', 'viewer'].includes(record.name) && (
            <Popconfirm
              title="确定删除此角色吗？"
              onConfirm={() => handleDelete(record.id)}
              okText="确定"
              cancelText="取消"
            >
              <Button type="link" size="small" danger icon={<DeleteOutlined />}>
                删除
              </Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Card
        title="角色管理"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
            新增角色
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
        title={editingRole ? '编辑角色' : '新增角色'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
      >
        <Form form={form} onFinish={handleSubmit} layout="vertical">
          <Form.Item
            name="name"
            label="角色名称"
            rules={[{ required: true, message: '请输入角色名称' }]}
          >
            <Input placeholder="custom_role" disabled={!!editingRole} />
          </Form.Item>

          <Form.Item
            name="description"
            label="描述"
            rules={[{ required: true, message: '请输入描述' }]}
          >
            <Input.TextArea rows={3} placeholder="角色描述" />
          </Form.Item>
        </Form>
      </Modal>

      <Drawer
        title={`配置权限 - ${selectedRole?.name}`}
        open={permDrawerVisible}
        onClose={() => setPermDrawerVisible(false)}
        width={720}
        extra={
          <Button type="primary" onClick={() => permForm.submit()}>
            保存
          </Button>
        }
      >
        <Form form={permForm} onFinish={handlePermSubmit} layout="vertical">
          <p style={{ marginBottom: 16, color: '#666' }}>
            权限配置（JSON格式）：
          </p>
          <Form.Item
            name="permissions"
            rules={[
              { required: true, message: '请输入权限配置' },
              {
                validator: (_, value) => {
                  try {
                    JSON.parse(value)
                    return Promise.resolve()
                  } catch {
                    return Promise.reject(new Error('JSON格式错误'))
                  }
                },
              },
            ]}
          >
            <Input.TextArea
              rows={20}
              placeholder={`[
  {"resource": "alerts", "action": "read"},
  {"resource": "alerts", "action": "update"}
]`}
              style={{ fontFamily: 'monospace', fontSize: 13 }}
            />
          </Form.Item>

          <div style={{ marginTop: 16, padding: 12, background: '#f0f2f5', borderRadius: 4 }}>
            <p style={{ margin: 0, fontWeight: 'bold' }}>可用权限参考：</p>
            <Tree
              showLine
              defaultExpandAll
              treeData={permissionTreeData}
              selectable={false}
              style={{ marginTop: 8 }}
            />
          </div>

          <div style={{ marginTop: 16, padding: 12, background: '#fff7e6', borderRadius: 4 }}>
            <p style={{ margin: 0, fontSize: 12, color: '#666' }}>
              <strong>格式说明：</strong><br />
              • resource: 资源名称（alerts/assets/probes等）<br />
              • action: 操作类型（read/create/update/delete）<br />
              • 使用 "*" 表示所有资源或操作<br />
              • 例: {`{"resource": "*", "action": "*"}`} 表示所有权限
            </p>
          </div>
        </Form>
      </Drawer>
    </div>
  )
}
