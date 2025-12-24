import { Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Alerts from './pages/Alerts'
import Assets from './pages/Assets'
import ThreatIntel from './pages/ThreatIntel'
import Reports from './pages/Reports'
import Probes from './pages/Probes'
import Settings from './pages/Settings'
import AdvancedDetection from './pages/AdvancedDetection'
import PcapAnalysis from './pages/PcapAnalysis'
import AuditLog from './pages/AuditLog'
import UserManagement from './pages/UserManagement'
import RoleManagement from './pages/RoleManagement'
import LicenseManagement from './pages/LicenseManagement'
import TenantManagement from './pages/TenantManagement'
import Login from './pages/Login'

function App() {
  const isAuthenticated = !!localStorage.getItem('token')

  if (!isAuthenticated && window.location.pathname !== '/login') {
    return <Login />
  }

  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/" element={<Layout />}>
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<Dashboard />} />
        <Route path="alerts" element={<Alerts />} />
        <Route path="assets" element={<Assets />} />
        <Route path="threat-intel" element={<ThreatIntel />} />
        <Route path="advanced-detection" element={<AdvancedDetection />} />
        <Route path="pcap-analysis" element={<PcapAnalysis />} />
        <Route path="reports" element={<Reports />} />
        <Route path="probes" element={<Probes />} />
        <Route path="audit" element={<AuditLog />} />
        <Route path="users" element={<UserManagement />} />
        <Route path="roles" element={<RoleManagement />} />
        <Route path="tenants" element={<TenantManagement />} />
        <Route path="license" element={<LicenseManagement />} />
        <Route path="settings" element={<Settings />} />
      </Route>
    </Routes>
  )
}

export default App