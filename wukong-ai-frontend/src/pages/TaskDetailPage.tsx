import React, { useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { LoadingSpinner, ErrorAlert } from '@/components/common'
import { TaskDetail } from '@/components/task'
import { useTask, useTaskStream } from '@/hooks'
import { Button } from '@/components/ui'

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
        <Button
          variant="outline"
          onClick={() => navigate('/tasks')}
        >
          返回列表
        </Button>
      </div>
    )
  }

  // 无任务
  if (!currentTask) {
    return (
      <div className="flex h-64 flex-col items-center justify-center gap-4">
        <p className="text-muted-foreground">未找到任务</p>
        <Button
          variant="outline"
          onClick={() => navigate('/tasks')}
        >
          返回列表
        </Button>
      </div>
    )
  }

  return (
    <div>
      <TaskDetail task={currentTask} />
    </div>
  )
}
