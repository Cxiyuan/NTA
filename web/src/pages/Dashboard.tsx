import { useEffect, useState } from 'react'
import { Row, Col, Card, Statistic } from 'antd'
import { AlertOutlined, DatabaseOutlined, GlobalOutlined, ApiOutlined } from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import { alertAPI, assetAPI, probeAPI } from '../services/api'

export default function Dashboard() {
  const [stats, setStats] = useState({
    alerts: 0,
    assets: 0,
    probes: 0,
    threats: 0,
  })

  const [alertTrend, setAlertTrend] = useState<any>({})
  const [severityDist, setSeverityDist] = useState<any>({})

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    try {
      const [alertsRes, assetsRes, probesRes] = await Promise.all([
        alertAPI.list({ page: 1, page_size: 1 }),
        assetAPI.list(),
        probeAPI.list(),
      ])
      
      setStats({
        alerts: alertsRes.total || 0,
        assets: assetsRes.length || 0,
        probes: probesRes.length || 0,
        threats: 0,
      })
    } catch (error) {
      console.error('Failed to load dashboard data', error)
    }
  }

  const alertTrendOption = {
    title: { text: '告警趋势' },
    tooltip: { trigger: 'axis' },
    xAxis: {
      type: 'category',
      data: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00'],
    },
    yAxis: { type: 'value' },
    series: [
      {
        name: '告警数',
        type: 'line',
        data: [120, 200, 150, 80, 70, 110],
        smooth: true,
      },
    ],
  }

  const severityOption = {
    title: { text: '告警等级分布' },
    tooltip: { trigger: 'item' },
    series: [
      {
        type: 'pie',
        radius: '50%',
        data: [
          { value: 335, name: '严重' },
          { value: 234, name: '高危' },
          { value: 154, name: '中危' },
          { value: 135, name: '低危' },
        ],
      },
    ],
  }

  return (
    <div>
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总告警数"
              value={stats.alerts}
              prefix={<AlertOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="资产数量"
              value={stats.assets}
              prefix={<DatabaseOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="在线探针"
              value={stats.probes}
              prefix={<ApiOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="威胁情报"
              value={stats.threats}
              prefix={<GlobalOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col span={12}>
          <Card>
            <ReactECharts option={alertTrendOption} style={{ height: 400 }} />
          </Card>
        </Col>
        <Col span={12}>
          <Card>
            <ReactECharts option={severityOption} style={{ height: 400 }} />
          </Card>
        </Col>
      </Row>
    </div>
  )
}
