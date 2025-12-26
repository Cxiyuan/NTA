import { useState, useEffect } from 'react'
import { Layout as AntLayout, Menu, theme } from 'antd'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import {
  DashboardOutlined,
  AlertOutlined,
  DatabaseOutlined,
  GlobalOutlined,
  FileTextOutlined,
  ApiOutlined,
  SettingOutlined,
  LogoutOutlined,
  SafetyOutlined,
  FileSearchOutlined,
  AuditOutlined,
  TeamOutlined,
  UserOutlined,
  SafetyCertificateOutlined,
  ClusterOutlined,
} from '@ant-design/icons'

const { Header, Sider, Content } = AntLayout

export default function Layout() {
  const navigate = useNavigate()
  const location = useLocation()
  const {
    token: { colorBgContainer },
  } = theme.useToken()

  const menuItems = [
    { key: '/dashboard', icon: <DashboardOutlined />, label: '态势大屏' },
    { key: '/alerts', icon: <AlertOutlined />, label: '安全告警' },
    { key: '/assets', icon: <DatabaseOutlined />, label: '资产管理' },
    { key: '/threat-intel', icon: <GlobalOutlined />, label: '威胁情报' },
    { key: '/advanced-detection', icon: <SafetyOutlined />, label: '高级检测' },
    { key: '/pcap-analysis', icon: <FileSearchOutlined />, label: 'PCAP回溯' },
    { key: '/reports', icon: <FileTextOutlined />, label: '报表中心' },
    { 
      key: 'probes',
      icon: <ApiOutlined />,
      label: '探针管理',
      children: [
        { key: '/builtin-probe', icon: <ApiOutlined />, label: '内置探针' },
        { key: '/probes', icon: <ClusterOutlined />, label: '外部探针' },
      ]
    },
    { 
      key: 'system',
      icon: <SettingOutlined />,
      label: '系统管理',
      children: [
        { key: '/users', icon: <UserOutlined />, label: '用户管理' },
        { key: '/roles', icon: <TeamOutlined />, label: '角色管理' },
        { key: '/tenants', icon: <ClusterOutlined />, label: '租户管理' },
        { key: '/audit', icon: <AuditOutlined />, label: '审计日志' },
        { key: '/license', icon: <SafetyCertificateOutlined />, label: 'License' },
        { key: '/settings', icon: <SettingOutlined />, label: '系统设置' },
      ]
    },
  ]

  const handleLogout = () => {
    localStorage.removeItem('token')
    navigate('/login')
  }

  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      <Header style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div style={{ color: 'white', fontSize: 20, fontWeight: 'bold' }}>
          NTA - 网络流量分析系统
        </div>
        <LogoutOutlined
          style={{ color: 'white', fontSize: 18, cursor: 'pointer' }}
          onClick={handleLogout}
        />
      </Header>
      <AntLayout>
        <Sider width={200} style={{ background: colorBgContainer }}>
          <Menu
            mode="inline"
            selectedKeys={[location.pathname]}
            items={menuItems}
            onClick={({ key }) => navigate(key)}
            style={{ height: '100%', borderRight: 0 }}
          />
        </Sider>
        <AntLayout style={{ padding: '24px' }}>
          <Content
            style={{
              padding: 24,
              margin: 0,
              minHeight: 280,
              background: colorBgContainer,
            }}
          >
            <Outlet />
          </Content>
        </AntLayout>
      </AntLayout>
    </AntLayout>
  )
}