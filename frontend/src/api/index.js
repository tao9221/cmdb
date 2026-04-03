import axios from 'axios'
import router from '../router'

const api = axios.create({ baseURL: '/api' })

api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

api.interceptors.response.use(
  res => res.data,
  err => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      router.push('/login')
    }
    return Promise.reject(err)
  }
)

export const login = (data) => api.post('/login', data)
export const getMe = () => api.get('/me')
export const getSSHKey = () => api.get('/sshkey')
export const saveSSHKey = (data) => api.post('/sshkey', data)
export const getStats = () => api.get('/stats')
export const getOverview = () => api.get('/overview')
export const getDataCenters = () => api.get('/datacenters')
export const getDataCenter = (id) => api.get(`/datacenters/${id}`)
export const createDataCenter = (data) => api.post('/datacenters', data)
export const deleteDataCenter = (id) => api.delete(`/datacenters/${id}`)
export const getCabinets = (dcId) => api.get(`/datacenters/${dcId}/cabinets`)
export const createCabinet = (data) => api.post('/cabinets', data)
export const deleteCabinet = (id) => api.delete(`/cabinets/${id}`)
export const getCabinetServers = (id) => api.get(`/cabinets/${id}/servers`)
export const updateCabinetPositions = (data) => api.put('/cabinets/positions', data)
export const getServers = (params) => api.get('/servers', { params })
export const getUnassignedServers = () => api.get('/servers', { params: { cabinet_id: 'unassigned' } })
export const assignServerCabinet = (id, cabinetId, slot = 0) => api.put(`/servers/${id}`, { cabinet_id: cabinetId, slot })
export const createServer = (data) => api.post('/servers', data)
export const getServer = (id) => api.get(`/servers/${id}`)
export const updateServer = (id, data) => api.put(`/servers/${id}`, data)
export const deleteServer = (id) => api.delete(`/servers/${id}`)
export const batchDeleteServers = (ids) => api.delete('/servers', { data: { ids } })

export const downloadRDP = (ip, username) => {
  const token = localStorage.getItem('token')
  const url = `/api/rdp/connect?ip=${encodeURIComponent(ip)}&username=${encodeURIComponent(username)}&token=${token}`
  const a = document.createElement('a')
  a.href = url
  a.download = `connect-${ip}.rdp`
  a.click()
}

// 用户管理（管理员）
export const listUsers = () => api.get('/admin/users')
export const createUser = (data) => api.post('/admin/users', data)
export const updateUser = (id, data) => api.put(`/admin/users/${id}`, data)
export const deleteUser = (id) => api.delete(`/admin/users/${id}`)
export const getUserAccess = (id) => api.get(`/admin/users/${id}/access`)
export const setUserAccess = (id, data) => api.put(`/admin/users/${id}/access`, data)

// 系统设置
export const getSettings = () => api.get('/admin/settings')
export const saveSettings = (data) => api.post('/admin/settings', data)
export const testMail = () => api.post('/admin/settings/test-mail')
