import { create } from 'zustand'
import type { TaskDetail, TaskListItem, RunTaskRequest, ProgressEvent } from '@/types'
import { taskApi } from '@/api'

// 任务状态
interface TaskStore {
  // 任务列表
  tasks: TaskListItem[]
  total: number
  currentPage: number
  pageSize: number
  loading: boolean
  error: string | null

  // 当前任务
  currentTask: TaskDetail | null
  currentTaskLoading: boolean
  currentTaskError: string | null

  // 进度事件
  events: ProgressEvent[]

  // 模式配置
  modeConfig: {
    thinking: boolean
    plan: boolean
    subagent: boolean
  }

  // Actions
  fetchTasks: (page?: number, status?: string) => Promise<void>
  runTask: (req: RunTaskRequest) => Promise<string>
  resumeTask: (taskId: string) => Promise<void>
  cancelTask: (taskId: string) => Promise<void>
  getTask: (taskId: string) => Promise<void>
  setModeConfig: (config: Partial<{ thinking: boolean; plan: boolean; subagent: boolean }>) => void
  addEvent: (event: ProgressEvent) => void
  clearEvents: () => void
  setCurrentTask: (task: TaskDetail | null) => void
}

// 计算模式
const calculateMode = (config: { thinking: boolean; plan: boolean; subagent: boolean }) => {
  if (!config.thinking && !config.plan && !config.subagent) return 'flash'
  if (config.thinking && !config.plan && !config.subagent) return 'standard'
  if (config.thinking && config.plan && !config.subagent) return 'pro'
  return 'ultra'
}

export const useTaskStore = create<TaskStore>((set, get) => ({
  // 初始状态
  tasks: [],
  total: 0,
  currentPage: 1,
  pageSize: 10,
  loading: false,
  error: null,

  currentTask: null,
  currentTaskLoading: false,
  currentTaskError: null,

  events: [],

  modeConfig: {
    thinking: false,
    plan: false,
    subagent: false,
  },

  // 获取任务列表
  fetchTasks: async (page = 1, status?: string) => {
    set({ loading: true, error: null })
    try {
      const response = await taskApi.listTasks(page, get().pageSize, status)
      set({
        tasks: response.tasks,
        total: response.total,
        currentPage: response.page,
        loading: false,
      })
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : 'Failed to fetch tasks',
        loading: false,
      })
    }
  },

  // 运行任务
  runTask: async (req: RunTaskRequest) => {
    const { modeConfig } = get()
    const fullReq: RunTaskRequest = {
      ...req,
      thinking_enabled: modeConfig.thinking,
      plan_enabled: modeConfig.plan,
      subagent_enabled: modeConfig.subagent,
    }

    try {
      const response = await taskApi.run(fullReq)
      // 刷新任务列表
      get().fetchTasks()
      return response.task_id
    } catch (error) {
      throw error
    }
  },

  // 续跑任务
  resumeTask: async (taskId: string) => {
    try {
      await taskApi.resume(taskId)
      get().fetchTasks()
    } catch (error) {
      throw error
    }
  },

  // 取消任务
  cancelTask: async (taskId: string) => {
    try {
      await taskApi.cancel(taskId)
      get().fetchTasks()
    } catch (error) {
      throw error
    }
  },

  // 获取任务详情
  getTask: async (taskId: string) => {
    set({ currentTaskLoading: true, currentTaskError: null })
    try {
      const task = await taskApi.getTask(taskId)
      set({ currentTask: task, currentTaskLoading: false })
    } catch (error) {
      set({
        currentTaskError: error instanceof Error ? error.message : 'Failed to get task',
        currentTaskLoading: false,
      })
    }
  },

  // 设置模式配置
  setModeConfig: (config) => {
    set((state) => ({
      modeConfig: { ...state.modeConfig, ...config },
    }))
  },

  // 添加进度事件
  addEvent: (event) => {
    set((state) => ({
      events: [...state.events, event],
    }))
  },

  // 清空事件
  clearEvents: () => {
    set({ events: [] })
  },

  // 设置当前任务
  setCurrentTask: (task) => {
    set({ currentTask: task })
  },
}))

// 导出计算模式函数
export { calculateMode }
