import axios from 'axios'
import { ElMessage } from 'element-plus'

const request = axios.create({
  baseURL: '/api',
  timeout: 30000
})

request.interceptors.request.use(
  config => {
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

request.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    ElMessage.error(error.message || '请求失败')
    return Promise.reject(error)
  }
)

export const alertAPI = {
  getAlerts: (params) => request.get('/alerts', { params }),
  getAlertDetail: (id) => request.get(`/alerts/${id}`),
  handleAlert: (id, action) => request.post(`/alerts/${id}/handle`, { action })
}

export const statsAPI = {
  getStats: () => request.get('/stats'),
  getTrend: (range) => request.get('/stats/trend', { params: { range } })
}

export const configAPI = {
  getConfig: () => request.get('/config'),
  updateConfig: (data) => request.put('/config', data),
  getWhitelist: () => request.get('/config/whitelist'),
  updateWhitelist: (data) => request.put('/config/whitelist', data)
}

export const threatIntelAPI = {
  getIOCs: (params) => request.get('/threat-intel/iocs', { params }),
  addIOC: (data) => request.post('/threat-intel/iocs', data),
  deleteIOC: (id) => request.delete(`/threat-intel/iocs/${id}`),
  updateFeeds: () => request.post('/threat-intel/update')
}

export const topologyAPI = {
  getGraph: () => request.get('/topology/graph'),
  getAnomalies: () => request.get('/topology/anomalies')
}

export const reportAPI = {
  getReports: (params) => request.get('/reports', { params }),
  generateReport: (params) => request.post('/reports/generate', params),
  downloadReport: (id) => request.get(`/reports/${id}/download`, { responseType: 'blob' })
}

export default request
