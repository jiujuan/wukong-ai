import axios, { AxiosInstance, AxiosError } from 'axios'

// API 基础配置
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

// 创建 axios 实例
const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
apiClient.interceptors.request.use(
  (config) => {
    // 可以在这里添加认证 token 等
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
apiClient.interceptors.response.use(
  (response) => response.data,
  (error: AxiosError<{ error: string }>) => {
    if (error.response) {
      // 服务器返回错误
      const message = error.response.data?.error || 'An error occurred'
      return Promise.reject(new Error(message))
    } else if (error.request) {
      // 请求发出但没有收到响应
      return Promise.reject(new Error('Network error, please check your connection'))
    } else {
      // 请求配置出错
      return Promise.reject(new Error(error.message))
    }
  }
)

export default apiClient

// 导出基础配置
export { API_BASE_URL }
