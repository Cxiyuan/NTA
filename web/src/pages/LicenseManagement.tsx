import { useState, useEffect } from 'react'
import { Card, Descriptions, Tag, Button, Upload, message, Progress, Alert, Space } from 'antd'
import { UploadOutlined, CheckCircleOutlined, CloseCircleOutlined, InfoCircleOutlined } from '@ant-design/icons'
import type { UploadProps } from 'antd'
import axios from 'axios'
import dayjs from 'dayjs'

export default function LicenseManagement() {
  const [licenseInfo, setLicenseInfo] = useState<any>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    loadLicenseInfo()
  }, [])

  const loadLicenseInfo = async () => {
    setLoading(true)
    try {
      const res = await axios.get('/api/v1/license')
      setLicenseInfo(res.data)
    } catch (error) {
      message.error('获取License信息失败')
    } finally {
      setLoading(false)
    }
  }

  const uploadProps: UploadProps = {
    name: 'file',
    action: '/api/v1/license/upload',
    headers: {
      Authorization: `Bearer ${localStorage.getItem('token')}`,
    },
    onChange(info) {
      if (info.file.status === 'done') {
        message.success('License上传成功')
        loadLicenseInfo()
      } else if (info.file.status === 'error') {
        message.error('License上传失败')
      }
    },
    accept: '.key,.lic,.license',
  }

  const getRemainingDays = () => {
    if (!licenseInfo?.expiry_date) return 0
    const now = dayjs()
    const expiry = dayjs(licenseInfo.expiry_date)
    return expiry.diff(now, 'day')
  }

  const getStatusTag = () => {
    const days = getRemainingDays()
    if (days < 0) {
      return <Tag icon={<CloseCircleOutlined />} color="error">已过期</Tag>
    } else if (days < 30) {
      return <Tag icon={<InfoCircleOutlined />} color="warning">即将过期</Tag>
    } else {
      return <Tag icon={<CheckCircleOutlined />} color="success">正常</Tag>
    }
  }

  const getUsagePercent = (current: number, max: number) => {
    if (!max) return 0
    return Math.round((current / max) * 100)
  }

  if (loading) {
    return <Card loading />
  }

  if (!licenseInfo) {
    return (
      <Card title="License管理">
        <Alert
          message="未找到有效License"
          description="请上传License文件以激活系统功能"
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
        />
        <Upload {...uploadProps}>
          <Button icon={<UploadOutlined />} type="primary">
            上传License文件
          </Button>
        </Upload>
      </Card>
    )
  }

  const remainingDays = getRemainingDays()

  return (
    <div>
      <Card
        title="License信息"
        extra={
          <Upload {...uploadProps}>
            <Button icon={<UploadOutlined />}>更新License</Button>
          </Upload>
        }
      >
        {remainingDays < 30 && remainingDays >= 0 && (
          <Alert
            message={`License即将过期，剩余 ${remainingDays} 天`}
            type="warning"
            showIcon
            closable
            style={{ marginBottom: 16 }}
          />
        )}
        {remainingDays < 0 && (
          <Alert
            message="License已过期"
            description="系统功能可能受限，请联系厂商更新License"
            type="error"
            showIcon
            style={{ marginBottom: 16 }}
          />
        )}

        <Descriptions bordered column={2}>
          <Descriptions.Item label="客户名称" span={2}>
            {licenseInfo.customer || '-'}
          </Descriptions.Item>
          <Descriptions.Item label="产品版本" span={2}>
            {licenseInfo.product || 'NTA Enterprise'}
          </Descriptions.Item>
          <Descriptions.Item label="License状态" span={2}>
            {getStatusTag()}
          </Descriptions.Item>
          <Descriptions.Item label="签发日期">
            {licenseInfo.issue_date ? dayjs(licenseInfo.issue_date).format('YYYY-MM-DD') : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="过期日期">
            {licenseInfo.expiry_date ? (
              <Space>
                {dayjs(licenseInfo.expiry_date).format('YYYY-MM-DD')}
                <Tag color={remainingDays > 30 ? 'green' : 'orange'}>
                  剩余 {remainingDays} 天
                </Tag>
              </Space>
            ) : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="最大探针数">
            {licenseInfo.max_probes || '无限制'}
          </Descriptions.Item>
          <Descriptions.Item label="最大带宽 (Mbps)">
            {licenseInfo.max_bandwidth_mbps || '无限制'}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title="资源配额使用情况" style={{ marginTop: 16 }}>
        <div style={{ marginBottom: 24 }}>
          <p style={{ marginBottom: 8 }}>探针数量使用情况</p>
          <Progress
            percent={getUsagePercent(licenseInfo.current_probes || 0, licenseInfo.max_probes)}
            status={getUsagePercent(licenseInfo.current_probes || 0, licenseInfo.max_probes) >= 90 ? 'exception' : 'active'}
            format={(percent) => `${licenseInfo.current_probes || 0} / ${licenseInfo.max_probes || '∞'} (${percent}%)`}
          />
        </div>

        <div>
          <p style={{ marginBottom: 8 }}>资产数量使用情况</p>
          <Progress
            percent={getUsagePercent(licenseInfo.current_assets || 0, licenseInfo.max_assets)}
            status={getUsagePercent(licenseInfo.current_assets || 0, licenseInfo.max_assets) >= 90 ? 'exception' : 'active'}
            format={(percent) => `${licenseInfo.current_assets || 0} / ${licenseInfo.max_assets || '∞'} (${percent}%)`}
          />
        </div>
      </Card>

      <Card title="已授权功能" style={{ marginTop: 16 }}>
        <Space wrap>
          {licenseInfo.features?.map((feature: string) => (
            <Tag key={feature} color="blue" icon={<CheckCircleOutlined />}>
              {feature}
            </Tag>
          )) || <Tag>基础功能</Tag>}
        </Space>

        <div style={{ marginTop: 16, padding: 12, background: '#f0f2f5', borderRadius: 4 }}>
          <p style={{ margin: 0, fontSize: 12, color: '#666' }}>
            <strong>可用功能模块：</strong><br />
            {licenseInfo.features?.includes('threat_intel') && '✅ 威胁情报集成'}<br />
            {licenseInfo.features?.includes('apt_detection') && '✅ APT检测'}<br />
            {licenseInfo.features?.includes('encryption_analysis') && '✅ 加密流量分析'}<br />
            {licenseInfo.features?.includes('ml_detection') && '✅ 机器学习检测'}<br />
            {licenseInfo.features?.includes('multi_tenant') && '✅ 多租户支持'}<br />
            {!licenseInfo.features?.length && '基础流量分析功能'}
          </p>
        </div>
      </Card>
    </div>
  )
}
