import type {
  TaskDetail,
  RunTaskRequest,
  RunTaskResponse,
  TaskListResponse,
  ResumeTaskResponse,
  CancelTaskResponse,
  ProgressEvent,
  DagData,
} from './task'

// API 错误响应
export interface ApiError {
  error: string
}

// API 基础响应
export interface ApiResponse<T> {
  data?: T
  error?: string
}

// 导出所有类型
export type {
  TaskDetail,
  RunTaskRequest,
  RunTaskResponse,
  TaskListResponse,
  ResumeTaskResponse,
  CancelTaskResponse,
  ProgressEvent,
  DagData,
}
