import { useEffect, useState } from 'react'
import { Table, Tag, message } from 'antd'
import { probeAPI } from '../services/api'
import dayjs from 'dayjs'

export default function Probes() {
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const res = await probeAPI.list()
      setData(res)
    } catch (error) {
      message.error('加载失败')
    } finally {
      setLoading(false)
    }
  }

  const columns = [
    {
      title: '探针ID',
      dataIndex: 'probe_id',
    },
    {
      title: '主机名',
      dataIndex: 'hostname',
    },
    {
      title: 'IP地址',
      dataIndex: 'ip_address',
    },
    {
      title: '版本',
      dataIndex: 'version',
    },
    {
      title: '状态',
      dataIndex: 'status',
      render: (text: string) => {
        const colors: any = {
          online: 'green',
          offline: 'red',
          error: 'orange',
        }
        return <Tag color={colors[text]}>{text.toUpperCase()}</Tag>
      },
    },
    {
      title: '最后心跳',
      dataIndex: 'last_heartbeat',
      render: (text: string) => dayjs(text).format('YYYY-MM-DD HH:mm:ss'),
    },
  ]

  return (
    <div>
      <Table columns={columns} dataSource={data} loading={loading} rowKey="id" />
    </div>
  )
}
