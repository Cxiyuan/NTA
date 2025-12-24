import { useEffect, useState } from 'react'
import { Table, Tag, message } from 'antd'
import { assetAPI } from '../services/api'
import dayjs from 'dayjs'

export default function Assets() {
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const res = await assetAPI.list()
      setData(res)
    } catch (error) {
      message.error('加载失败')
    } finally {
      setLoading(false)
    }
  }

  const columns = [
    {
      title: 'IP地址',
      dataIndex: 'ip',
      key: 'ip',
    },
    {
      title: 'MAC地址',
      dataIndex: 'mac',
      key: 'mac',
    },
    {
      title: '主机名',
      dataIndex: 'hostname',
      key: 'hostname',
    },
    {
      title: '厂商',
      dataIndex: 'vendor',
      key: 'vendor',
    },
    {
      title: '操作系统',
      dataIndex: 'os',
      key: 'os',
    },
    {
      title: '服务',
      dataIndex: 'services',
      key: 'services',
      render: (services: string) => {
        try {
          const list = JSON.parse(services)
          return list.map((s: string) => <Tag key={s}>{s}</Tag>)
        } catch {
          return '-'
        }
      },
    },
    {
      title: '首次发现',
      dataIndex: 'first_seen',
      key: 'first_seen',
      render: (text: string) => dayjs(text).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '最后活跃',
      dataIndex: 'last_seen',
      key: 'last_seen',
      render: (text: string) => dayjs(text).format('YYYY-MM-DD HH:mm:ss'),
    },
  ]

  return (
    <div>
      <Table
        columns={columns}
        dataSource={data}
        loading={loading}
        rowKey="id"
      />
    </div>
  )
}
