import React, { useEffect, useState } from 'react'
import { Plus, Filter, RefreshCw } from 'lucide-react'
import { useLocation, useNavigate } from 'react-router-dom'
import { TaskCard } from '@/components/task'
import { TaskRunForm } from '@/components/task'
import { LoadingSpinner, EmptyState } from '@/components/common'
import { useTaskList } from '@/hooks'
import { Button } from '@/components/ui'

type StatusFilter = 'all' | 'pending' | 'queued' | 'running' | 'success' | 'failed'

/**
 * 任务列表页面
 */
export function TaskListPage() {
  const location = useLocation()
  const navigate = useNavigate()
  const isNewTaskPage = location.pathname === '/tasks/new'
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')

  const {
    tasks,
    tasksWithMode,
    total,
    currentPage,
    totalPages,
    loading,
    error,
    loadTasks,
    goToPage,
    filterByStatus,
  } = useTaskList()

  // 加载任务列表
  useEffect(() => {
    if (!isNewTaskPage) {
      loadTasks(1, statusFilter === 'all' ? undefined : statusFilter)
    }
  }, [])

  // 刷新
  const handleRefresh = () => {
    loadTasks(currentPage, statusFilter === 'all' ? undefined : statusFilter)
  }

  // 切换状态筛选
  const handleStatusFilter = (status: StatusFilter) => {
    setStatusFilter(status)
    filterByStatus(status === 'all' ? undefined : status)
  }

  // 运行成功回调
  const handleRunSuccess = (taskId: string) => {
    navigate(`/tasks/${taskId}`)
  }

  const statusOptions: { value: StatusFilter; label: string }[] = [
    { value: 'all', label: '全部' },
    { value: 'pending', label: '待处理' },
    { value: 'queued', label: '排队中' },
    { value: 'running', label: '运行中' },
    { value: 'success', label: '成功' },
    { value: 'failed', label: '失败' },
  ]

  if (isNewTaskPage) {
    return (
      <div className="flex min-h-[70vh] items-center justify-center">
        <div className="w-full max-w-3xl">
          <TaskRunForm onSuccess={handleRunSuccess} />
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* 头部 */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold text-foreground">任务管理</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            共 {total} 个任务
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Button
            onClick={handleRefresh}
            disabled={loading}
            variant="outline"
            className="gap-2"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            刷新
          </Button>
          <Button
            onClick={() => navigate(isNewTaskPage ? '/tasks' : '/tasks/new')}
            className="gap-2"
          >
            <Plus className="h-4 w-4" />
            {isNewTaskPage ? '收起新建' : '新建任务'}
          </Button>
        </div>
      </div>

      {/* 筛选 */}
      <div className="flex items-center gap-2">
        <Filter className="h-4 w-4 text-muted-foreground" />
        <div className="flex gap-2">
          {statusOptions.map((option) => (
            <Button
              key={option.value}
              onClick={() => handleStatusFilter(option.value)}
              variant={statusFilter === option.value ? 'secondary' : 'ghost'}
              className={`h-8 px-3 py-1.5 text-sm font-medium ${
                statusFilter === option.value
                  ? 'bg-primary/10 text-primary'
                  : 'text-muted-foreground'
              }`}
            >
              {option.label}
            </Button>
          ))}
        </div>
      </div>

      {/* 错误 */}
      {error && (
        <div className="rounded-lg bg-destructive/10 p-4 text-sm text-destructive">
          {error}
        </div>
      )}

      {/* 加载状态 */}
      {loading && tasks.length === 0 ? (
        <div className="flex h-64 items-center justify-center">
          <LoadingSpinner text="加载中..." />
        </div>
      ) : tasks.length === 0 ? (
        <EmptyState
          title="暂无任务"
          description="创建您的第一个任务，开始使用悟空 AI"
          action={
            <Button
              onClick={() => navigate('/tasks/new')}
              size="sm"
            >
              创建任务
            </Button>
          }
        />
      ) : (
        <>
          {/* 任务列表 */}
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {tasksWithMode.map((task) => (
              <TaskCard key={task.task_id} task={task} />
            ))}
          </div>

          {/* 分页 */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between">
              <p className="text-sm text-muted-foreground">
                第 {currentPage} / {totalPages} 页，共 {total} 条
              </p>
              <div className="flex gap-2">
                <Button
                  onClick={() => goToPage(currentPage - 1)}
                  disabled={currentPage <= 1}
                  variant="outline"
                  size="sm"
                >
                  上一页
                </Button>
                <Button
                  onClick={() => goToPage(currentPage + 1)}
                  disabled={currentPage >= totalPages}
                  variant="outline"
                  size="sm"
                >
                  下一页
                </Button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}
