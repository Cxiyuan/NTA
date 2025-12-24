import { useState, useEffect } from 'react'
import { Card, Button, Table, Form, Select, DatePicker, message } from 'antd'
import { reportAPI } from '../services/api'
import dayjs from 'dayjs'

export default function Reports() {
  const [form] = Form.useForm()
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)
  const [generating, setGenerating] = useState(false)

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const res = await reportAPI.list({})
      setData(res.data || [])
    } catch (error) {
      message.error('加载失败')
    } finally {
      setLoading(false)
    }
  }

  const handleGenerate = async (values: any) => {
    setGenerating(true)
    try {
      await reportAPI.generate({
        type: values.type,
        start_time: values.dateRange[0].toISOString(),
        end_time: values.dateRange[1].toISOString(),
      })
      message.success('报表生成中，请稍后查看')
      loadData()
    } catch (error) {
      message.error('生成失败')
    } finally {
      setGenerating(false)
    }
  }

  const handleDownload = async (id: number) => {
    try {
      const blob = await reportAPI.download(id)
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `report_${id}.pdf`
      a.click()
    } catch (error) {
      message.error('下载失败')
    }
  }

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
    },
    {
      title: '报表类型',
      dataIndex: 'type',
    },
    {
      title: '时间范围',
      dataIndex: 'time_range',
    },
    {
      title: '生成时间',
      dataIndex: 'created_at',
      render: (text: string) => dayjs(text).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '状态',
      dataIndex: 'status',
    },
    {
      title: '操作',
      render: (record: any) => (
        <Button type="link" onClick={() => handleDownload(record.id)}>
          下载
        </Button>
      ),
    },
  ]

  return (
    <div>
      <Card title="生成报表" style={{ marginBottom: 16 }}>
        <Form form={form} onFinish={handleGenerate} layout="inline">
          <Form.Item name="type" label="报表类型" rules={[{ required: true }]}>
            <Select style={{ width: 150 }}>
              <Select.Option value="daily">日报</Select.Option>
              <Select.Option value="weekly">周报</Select.Option>
              <Select.Option value="monthly">月报</Select.Option>
              <Select.Option value="custom">自定义</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="dateRange" label="时间范围" rules={[{ required: true }]}>
            <DatePicker.RangePicker />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={generating}>
              生成报表
            </Button>
          </Form.Item>
        </Form>
      </Card>

      <Card title="历史报表">
        <Table columns={columns} dataSource={data} loading={loading} rowKey="id" />
      </Card>
    </div>
  )
}
