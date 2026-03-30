import { useCallback, useMemo } from 'react'
import { useTaskStore } from '@/store'
import { calculateMode } from '@/store'

/**
 * 任务列表 Hook
 */
export function useTaskList() {
  const tasks = useTaskStore((state) => state.tasks)
  const total = useTaskStore((state) => state.total)
  const currentPage = useTaskStore((state) => state.currentPage)
  const pageSize = useTaskStore((state) => state.pageSize)
  const loading = useTaskStore((state) => state.loading)
  const error = useTaskStore((state) => state.error)

  const fetchTasks = useTaskStore((state) => state.fetchTasks)

  // 计算总页数
  const totalPages = useMemo(() => {
    return Math.ceil(total / pageSize)
  }, [total, pageSize])

  // 加载任务列表
  const loadTasks = useCallback(async (page?: number, status?: string) => {
    await fetchTasks(page, status)
  }, [fetchTasks])

  // 刷新列表
  const refresh = useCallback(() => {
    return fetchTasks(currentPage)
  }, [currentPage, fetchTasks])

  // 切换页面
  const goToPage = useCallback((page: number) => {
    if (page >= 1 && page <= totalPages) {
      return fetchTasks(page)
    }
  }, [totalPages, fetchTasks])

  // 下一页
  const nextPage = useCallback(() => {
    if (currentPage < totalPages) {
      return fetchTasks(currentPage + 1)
    }
  }, [currentPage, totalPages, fetchTasks])

  // 上一页
  const prevPage = useCallback(() => {
    if (currentPage > 1) {
      return fetchTasks(currentPage - 1)
    }
  }, [currentPage, fetchTasks])

  // 按状态筛选
  const filterByStatus = useCallback((status: string | undefined) => {
    return fetchTasks(1, status)
  }, [fetchTasks])

  // 获取任务并计算模式
  const tasksWithMode = useMemo(() => {
    return tasks.map((task) => ({
      ...task,
      mode: calculateMode({
        thinking: task.mode === 'standard' || task.mode === 'pro' || task.mode === 'ultra',
        plan: task.mode === 'pro' || task.mode === 'ultra',
        subagent: task.mode === 'ultra',
      }),
    }))
  }, [tasks])

  return {
    // 状态
    tasks,
    tasksWithMode,
    total,
    currentPage,
    pageSize,
    totalPages,
    loading,
    error,
    // 方法
    loadTasks,
    refresh,
    goToPage,
    nextPage,
    prevPage,
    filterByStatus,
  }
}
