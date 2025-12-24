import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export const alertAPI = {
  list: (params: any) => api.get('/alerts', { params }),
  get: (id: number) => api.get(`/alerts/${id}`),
  update: (id: number, data: any) => api.put(`/alerts/${id}`, data),
}

export const assetAPI = {
  list: () => api.get('/assets'),
  get: (ip: string) => api.get(`/assets/${ip}`),
}

export const threatIntelAPI = {
  check: (params: { type: string; value: string }) => 
    api.get('/threat-intel/check', { params }),
  update: () => api.post('/threat-intel/update'),
}

export const probeAPI = {
  list: () => api.get('/probes'),
  get: (id: string) => api.get(`/probes/${id}`),
  register: (data: any) => api.post('/probes/register', data),
  heartbeat: (id: string) => api.post(`/probes/${id}/heartbeat`),
}

export const reportAPI = {
  list: (params: any) => api.get('/reports', { params }),
  generate: (data: any) => api.post('/reports/generate', data),
  download: (id: number) => api.get(`/reports/${id}/download`, { responseType: 'blob' }),
}

export const notificationAPI = {
  list: () => api.get('/notifications/config'),
  update: (data: any) => api.put('/notifications/config', data),
  test: (channel: string) => api.post('/notifications/test', { channel }),
}

export const pcapAPI = {
  list: (params: any) => api.get('/pcap/sessions', { params }),
  download: (sessionId: string) => api.get(`/pcap/${sessionId}/download`, { responseType: 'blob' }),
  search: (params: any) => api.post('/pcap/search', params),
}

export default api
