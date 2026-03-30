import apiClient from './client'
import type {
  TaskDetail,
  RunTaskRequest,
  RunTaskResponse,
  TaskListResponse,
  ResumeTaskResponse,
  CancelTaskResponse,
} from '@/types'

// 任务 API
export const taskApi = {
  // 运行新任务
  run: async (req: RunTaskRequest): Promise<RunTaskResponse> => {
    return apiClient.post('/api/run', req)
  },

  // 续跑任务
  resume: async (taskId: string): Promise<ResumeTaskResponse> => {
    return apiClient.post('/api/resume', { task_id: taskId })
  },

  // 获取任务详情
  getTask: async (taskId: string): Promise<TaskDetail> => {
    return apiClient.get(`/api/task?task_id=${taskId}`)
  },

  // 获取任务列表
  listTasks: async (
    page = 1,
    size = 10,
    status?: string
  ): Promise<TaskListResponse> => {
    const params = new URLSearchParams({
      page: String(page),
      size: String(size),
    })
    if (status) {
      params.append('status', status)
    }
    return apiClient.get(`/api/list?${params.toString()}`)
  },

  // 取消任务
  cancel: async (taskId: string): Promise<CancelTaskResponse> => {
    return apiClient.post('/api/task/cancel', { task_id: taskId })
  },

  // 健康检查
  health: async (): Promise<{ status: string }> => {
    return apiClient.get('/health')
  },
}

// 导出
export default taskApi
