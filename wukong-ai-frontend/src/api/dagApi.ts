import apiClient from './client'
import type { DagData } from '@/types'

// DAG API
export const dagApi = {
  // 获取 DAG 数据
  getDag: async (taskId: string): Promise<DagData> => {
    return apiClient.get(`/api/task/dag?task_id=${taskId}`)
  },
}

// 导出
export default dagApi
