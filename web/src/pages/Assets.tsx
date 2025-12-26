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
      // 硬编码模拟数据
      const mockData = Array.from({ length: 30 }, (_, i) => ({
        id: i + 1,
        ip: `192.168.${Math.floor(i / 255)}.${(i % 255) + 1}`,
        mac: `00:${Math.floor(Math.random() * 256).toString(16).padStart(2, '0')}:${Math.floor(Math.random() * 256).toString(16).padStart(2, '0')}:${Math.floor(Math.random() * 256).toString(16).padStart(2, '0')}:${Math.floor(Math.random() * 256).toString(16).padStart(2, '0')}:${Math.floor(Math.random() * 256).toString(16).padStart(2, '0')}`,
        hostname: ['server-web', 'server-db', 'workstation', 'printer', 'nas', 'router'][Math.floor(Math.random() * 6)] + `-${i + 1}`,
        vendor: ['Dell', 'HP', 'Cisco', 'Lenovo', 'Apple', 'Unknown'][Math.floor(Math.random() * 6)],
        os: ['Windows 10', 'Windows Server 2019', 'Ubuntu 20.04', 'CentOS 7', 'macOS', 'Unknown'][Math.floor(Math.random() * 6)],
        services: JSON.stringify(['HTTP:80', 'HTTPS:443', 'SSH:22', 'RDP:3389', 'SMB:445'].slice(0, Math.floor(Math.random() * 3) + 1)),
        first_seen: new Date(Date.now() - Math.random() * 86400000 * 30).toISOString(),
        last_seen: new Date(Date.now() - Math.random() * 3600000).toISOString(),
      }))
      
      setData(mockData)
      
      /* 真实数据接口（待后续启用）
      const res = await assetAPI.list()
      setData(res)
      */
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