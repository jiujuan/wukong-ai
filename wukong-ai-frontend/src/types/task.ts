// 任务状态
export type TaskStatus = 'pending' | 'queued' | 'running' | 'success' | 'failed'

// 执行模式
export type TaskMode = 'flash' | 'standard' | 'pro' | 'ultra'

// 任务详情
export interface TaskDetail {
  task_id: string
  status: TaskStatus
  mode: TaskMode
  user_input: string
  intention?: string
  plan?: string
  tasks?: string[]
  sub_results?: string[]
  final_output?: string
  last_node?: string
  create_time: string
  finish_time?: string
  error_msg?: string
}

// 运行任务请求
export interface RunTaskRequest {
  user_input: string
  thinking_enabled?: boolean
  plan_enabled?: boolean
  subagent_enabled?: boolean
  llm_provider?: string
  max_sub_agents?: number
  timeout_seconds?: number
}

// 运行任务响应
export interface RunTaskResponse {
  task_id: string
  status: TaskStatus
  mode: TaskMode
  stream_url?: string
  create_time: string
}

// 任务列表响应
export interface TaskListResponse {
  total: number
  page: number
  size: number
  tasks: TaskListItem[]
}

// 任务列表项
export interface TaskListItem {
  task_id: string
  status: TaskStatus
  mode: TaskMode
  user_input: string
  create_time: string
  finish_time?: string
}

// 续跑请求
export interface ResumeTaskRequest {
  task_id: string
}

// 续跑响应
export interface ResumeTaskResponse {
  task_id: string
  status: TaskStatus
  resumed_from: string
  message: string
}

// 取消任务请求
export interface CancelTaskRequest {
  task_id: string
}

// 取消任务响应
export interface CancelTaskResponse {
  task_id: string
  status: string
  message: string
}

// 模式配置
export interface ModeConfig {
  thinking: boolean
  plan: boolean
  subagent: boolean
}

// 进度事件
export interface ProgressEvent {
  type: 'node_start' | 'node_done' | 'sub_agent_update' | 'task_done' | 'task_failed'
  node?: string
  status?: string
  progress?: number
  total?: number
  done?: number
  latest?: string
  final_output?: string
  timestamp: string
}
