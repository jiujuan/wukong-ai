import { useCallback } from 'react'
import { useTaskStore } from '@/store'
import type { RunTaskRequest } from '@/types'

/**
 * 任务操作 Hook
 */
export function useTask() {
  const currentTask = useTaskStore((state) => state.currentTask)
  const currentTaskLoading = useTaskStore((state) => state.currentTaskLoading)
  const currentTaskError = useTaskStore((state) => state.currentTaskError)
  const modeConfig = useTaskStore((state) => state.modeConfig)
  const events = useTaskStore((state) => state.events)

  const getTask = useTaskStore((state) => state.getTask)
  const runTask = useTaskStore((state) => state.runTask)
  const resumeTask = useTaskStore((state) => state.resumeTask)
  const cancelTask = useTaskStore((state) => state.cancelTask)
  const setModeConfig = useTaskStore((state) => state.setModeConfig)
  const clearEvents = useTaskStore((state) => state.clearEvents)
  const setCurrentTask = useTaskStore((state) => state.setCurrentTask)

  // 运行新任务
  const handleRunTask = useCallback(async (userInput: string, intention?: string) => {
    const request: RunTaskRequest = {
      user_input: userInput,
    }
    if (intention) {
      request.intention = intention
    }
    return runTask(request)
  }, [runTask])

  // 续跑任务
  const handleResumeTask = useCallback(async (taskId: string) => {
    return resumeTask(taskId)
  }, [resumeTask])

  // 取消任务
  const handleCancelTask = useCallback(async (taskId: string) => {
    return cancelTask(taskId)
  }, [cancelTask])

  // 获取任务详情
  const handleGetTask = useCallback(async (taskId: string) => {
    return getTask(taskId)
  }, [getTask])

  // 设置模式
  const handleSetModeConfig = useCallback((config: Partial<{ thinking: boolean; plan: boolean; subagent: boolean }>) => {
    setModeConfig(config)
  }, [setModeConfig])

  // 切换模式
  const toggleThinking = useCallback(() => {
    setModeConfig({ thinking: !modeConfig.thinking })
  }, [modeConfig.thinking, setModeConfig])

  const togglePlan = useCallback(() => {
    setModeConfig({ plan: !modeConfig.plan })
  }, [modeConfig.plan, setModeConfig])

  const toggleSubagent = useCallback(() => {
    setModeConfig({ subagent: !modeConfig.subagent })
  }, [modeConfig.subagent, setModeConfig])

  return {
    // 状态
    currentTask,
    currentTaskLoading,
    currentTaskError,
    modeConfig,
    events,
    // 方法
    runTask: handleRunTask,
    resumeTask: handleResumeTask,
    cancelTask: handleCancelTask,
    getTask: handleGetTask,
    setModeConfig: handleSetModeConfig,
    toggleThinking,
    togglePlan,
    toggleSubagent,
    clearEvents,
    setCurrentTask,
  }
}
