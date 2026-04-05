import React, { useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { LoadingSpinner, ErrorAlert } from '@/components/common'
import { TaskDetail } from '@/components/task'
import { useTask, useTaskStream } from '@/hooks'

/**
 * 任务详情页面
 */
export function TaskDetailPage() {
  const { taskId } = useParams<{ taskId: string }>()
  const navigate = useNavigate()

  const { currentTask, currentTaskLoading, currentTaskError, getTask, clearEvents } = useTask()

  // 加载任务详情
  useEffect(() => {
    if (taskId) {
      clearEvents()
      getTask(taskId)
    }
  }, [taskId, getTask, clearEvents, currentTask?.task_id])

  // 监听任务进度
  useTaskStream({
    taskId: taskId || '',
    enabled: !!taskId && currentTask?.status !== 'success' && currentTask?.status !== 'failed',
    onTaskComplete: () => {
      if (taskId) {
        getTask(taskId)
      }
    },
  })

  // 加载中
  if (currentTaskLoading && !currentTask) {
    return (
      <div className="flex h-64 items-center justify-center">
        <LoadingSpinner text="加载任务详情..." />
      </div>
    )
  }

  // 错误
  if (currentTaskError) {
    return (
      <div className="space-y-4">
        <ErrorAlert
          message={currentTaskError}
          onDismiss={() => navigate('/tasks')}
        />
        <button
          onClick={() => navigate('/tasks')}
          className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
        >
          返回列表
        </button>
      </div>
    )
  }

  // 无任务
  if (!currentTask) {
    return (
      <div className="flex h-64 flex-col items-center justify-center gap-4">
        <p className="text-gray-500">未找到任务</p>
        <button
          onClick={() => navigate('/tasks')}
          className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
        >
          返回列表
        </button>
      </div>
    )
  }

  return (
    <div>
      <TaskDetail task={currentTask} />
    </div>
  )
}
