import axios, { type AxiosInstance, type InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

const service: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 30000,
})

service.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

service.interceptors.response.use(
  (res) => {
    const data = res.data
    const code = data.code ?? res.status
    if (code !== 0 && (code < 200 || code >= 300)) {
      ElMessage.error(data.message || '请求失败')
      if (data.code === 401) {
        localStorage.removeItem('token')
        router.push('/login')
      }
      return Promise.reject(new Error(data.message))
    }
    return data
  },
  (err) => {
    const data = err.response?.data
    const message = typeof data === 'string'
      ? data
      : data?.message || data?.error || data?.msg || err.message || '网络错误'
    if (err.response?.status === 401 || data?.code === 401) {
      localStorage.removeItem('token')
      router.push('/login')
    }
    ElMessage.error(message)
    return Promise.reject(new Error(message))
  },
)

export default service
