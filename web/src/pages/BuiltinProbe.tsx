import { useEffect, useState } from 'react'
import { Card, Tabs, Button, Table, Tag, Switch, Modal, Form, Input, Select, message, Space, Statistic, Row, Col } from 'antd'
import { PlayCircleOutlined, PauseCircleOutlined, ReloadOutlined, SettingOutlined, FileTextOutlined, ApiOutlined } from '@ant-design/icons'
import { builtinProbeAPI } from '../services/api'
import dayjs from 'dayjs'

export default function BuiltinProbe() {
  const [probe, setProbe] = useState<any>(null)
  const [status, setStatus] = useState<any>(null)
  const [scripts, setScripts] = useState<any[]>([])
  const [logs, setLogs] = useState<any[]>([])
  const [logStats, setLogStats] = useState<any>({})
  const [interfaces, setInterfaces] = useState<string[]>([])
  const [loading, setLoading] = useState(false)
  const [configModalVisible, setConfigModalVisible] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    loadAll()
    const interval = setInterval(() => {
      loadStatus()
      loadLogStats()
    }, 10000)
    return () => clearInterval(interval)
  }, [])

  const loadAll = async () => {
    await Promise.all([
      loadProbe(),
      loadStatus(),
      loadScripts(),
      loadLogs(),
      loadLogStats(),
      loadInterfaces(),
    ])
  }

  const loadProbe = async () => {
    try {
      const res = await builtinProbeAPI.get()
      setProbe(res)
    } catch (error: any) {
      if (error.response?.status !== 404) {
        message.error('加载探针信息失败')
      }
    }
  }

  const loadStatus = async () => {
    try {
      const res = await builtinProbeAPI.getStatus()
      setStatus(res)
    } catch (error) {
      console.error('Failed to load status', error)
    }
  }

  const loadScripts = async () => {
    try {
      const res = await builtinProbeAPI.getScripts()
      setScripts(res.scripts || [])
    } catch (error) {
      message.error('加载脚本列表失败')
    }
  }

  const loadLogs = async () => {
    try {
      const res = await builtinProbeAPI.getLogs({ limit: 100 })
      setLogs(res.logs || [])
    } catch (error) {
      console.error('Failed to load logs', error)
    }
  }

  const loadLogStats = async () => {
    try {
      const res = await builtinProbeAPI.getLogStats()
      setLogStats(res.stats || {})
    } catch (error) {
      console.error('Failed to load log stats', error)
    }
  }

  const loadInterfaces = async () => {
    try {
      const res = await builtinProbeAPI.getInterfaces()
      setInterfaces(res.interfaces || [])
    } catch (error) {
      console.error('Failed to load interfaces', error)
    }
  }

  const handleStart = async () => {
    setLoading(true)
    try {
      await builtinProbeAPI.start()
      message.success('探针已启动')
      await loadStatus()
    } catch (error) {
      message.error('启动失败')
    } finally {
      setLoading(false)
    }
  }

  const handleStop = async () => {
    setLoading(true)
    try {
      await builtinProbeAPI.stop()
      message.success('探针已停止')
      await loadStatus()
    } catch (error) {
      message.error('停止失败')
    } finally {
      setLoading(false)
    }
  }

  const handleRestart = async () => {
    setLoading(true)
    try {
      await builtinProbeAPI.restart()
      message.success('探针已重启')
      await loadStatus()
    } catch (error) {
      message.error('重启失败')
    } finally {
      setLoading(false)
    }
  }

  const handleToggleScript = async (scriptName: string, enabled: boolean) => {
    try {
      if (enabled) {
        await builtinProbeAPI.enableScript(scriptName)
        message.success('脚本已启用')
      } else {
        await builtinProbeAPI.disableScript(scriptName)
        message.success('脚本已禁用')
      }
      await loadScripts()
    } catch (error) {
      message.error('操作失败')
    }
  }

  const showConfigModal = () => {
    form.setFieldsValue({
      interface: probe?.interface || 'eth0',
      bpf_filter: probe?.bpf_filter || '',
    })
    setConfigModalVisible(true)
  }

  const handleConfigSubmit = async () => {
    try {
      const values = await form.validateFields()
      await builtinProbeAPI.updateConfig(values)
      message.success('配置已更新')
      setConfigModalVisible(false)
      await loadProbe()
    } catch (error) {
      message.error('更新配置失败')
    }
  }

  const scriptColumns = [
    {
      title: '脚本名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
    },
    {
      title: '文件',
      dataIndex: 'file',
      key: 'file',
      render: (text: string) => <Tag>{text}</Tag>,
    },
    {
      title: '状态',
      key: 'enabled',
      render: (record: any) => (
        <Switch
          checked={record.enabled === 'true'}
          onChange={(checked) => handleToggleScript(record.name, checked)}
        />
      ),
    },
  ]

  const logColumns = [
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      render: (text: string) => dayjs(text).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '类型',
      dataIndex: 'log_type',
      key: 'log_type',
      render: (text: string) => <Tag color="blue">{text}</Tag>,
    },
    {
      title: '源IP',
      dataIndex: 'src_ip',
      key: 'src_ip',
    },
    {
      title: '目的IP',
      dataIndex: 'dst_ip',
      key: 'dst_ip',
    },
    {
      title: '协议',
      dataIndex: 'protocol',
      key: 'protocol',
    },
    {
      title: 'UID',
      dataIndex: 'uid',
      key: 'uid',
      ellipsis: true,
    },
  ]

  const statusColor = status?.status === 'running' ? 'green' : status?.status === 'stopped' ? 'orange' : 'red'

  return (
    <div>
      <Card
        title="内置探针管理"
        extra={
          <Space>
            <Button
              icon={<SettingOutlined />}
              onClick={showConfigModal}
            >
              配置
            </Button>
            <Button
              icon={<PlayCircleOutlined />}
              type="primary"
              onClick={handleStart}
              loading={loading}
              disabled={status?.status === 'running'}
            >
              启动
            </Button>
            <Button
              icon={<PauseCircleOutlined />}
              danger
              onClick={handleStop}
              loading={loading}
              disabled={status?.status !== 'running'}
            >
              停止
            </Button>
            <Button
              icon={<ReloadOutlined />}
              onClick={handleRestart}
              loading={loading}
            >
              重启
            </Button>
          </Space>
        }
      >
        <Row gutter={16} style={{ marginBottom: 24 }}>
          <Col span={6}>
            <Card>
              <Statistic
                title="状态"
                value={status?.status || '-'}
                valueStyle={{ color: statusColor }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="监听接口"
                value={probe?.interface || '-'}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="捕获包数"
                value={probe?.packets_captured || 0}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="捕获字节"
                value={probe?.bytes_captured || 0}
                suffix="B"
              />
            </Card>
          </Col>
        </Row>

        <Tabs
          items={[
            {
              key: 'scripts',
              label: (
                <span>
                  <ApiOutlined />
                  检测脚本
                </span>
              ),
              children: (
                <Table
                  columns={scriptColumns}
                  dataSource={scripts}
                  rowKey="name"
                  pagination={false}
                />
              ),
            },
            {
              key: 'logs',
              label: (
                <span>
                  <FileTextOutlined />
                  日志 ({logs.length})
                </span>
              ),
              children: (
                <div>
                  <Row gutter={16} style={{ marginBottom: 16 }}>
                    {Object.entries(logStats).map(([type, count]) => (
                      <Col span={4} key={type}>
                        <Card size="small">
                          <Statistic title={type} value={count as number} />
                        </Card>
                      </Col>
                    ))}
                  </Row>
                  <Table
                    columns={logColumns}
                    dataSource={logs}
                    rowKey="id"
                    pagination={{ pageSize: 20 }}
                  />
                </div>
              ),
            },
          ]}
        />
      </Card>

      <Modal
        title="探针配置"
        open={configModalVisible}
        onOk={handleConfigSubmit}
        onCancel={() => setConfigModalVisible(false)}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            label="监听接口"
            name="interface"
            rules={[{ required: true, message: '请选择监听接口' }]}
          >
            <Select>
              {interfaces.map((iface) => (
                <Select.Option key={iface} value={iface}>
                  {iface}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item
            label="BPF过滤器"
            name="bpf_filter"
            help="例如: tcp port 80 or udp port 53"
          >
            <Input.TextArea rows={3} placeholder="留空表示不过滤" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
