import React, { useEffect, useState } from 'react'
import { Plus, Filter, RefreshCw } from 'lucide-react'
import { useLocation, useNavigate } from 'react-router-dom'
import { TaskCard } from '@/components/task'
import { TaskRunForm } from '@/components/task'
import { ModeSelector } from '@/components/mode'
import { LoadingSpinner, EmptyState } from '@/components/common'
import { useTaskList } from '@/hooks'

type StatusFilter = 'all' | 'pending' | 'queued' | 'running' | 'success' | 'failed'

/**
 * 任务列表页面
 */
export function TaskListPage() {
  const location = useLocation()
  const navigate = useNavigate()
  const isNewTaskPage = location.pathname === '/tasks/new'
  const [showRunForm, setShowRunForm] = useState(isNewTaskPage)
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
    loadTasks(1, statusFilter === 'all' ? undefined : statusFilter)
  }, [])

  useEffect(() => {
    setShowRunForm(isNewTaskPage)
  }, [isNewTaskPage])

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
    setShowRunForm(false)
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

  return (
    <div className="space-y-6">
      {/* 头部 */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold text-gray-900">任务管理</h2>
          <p className="mt-1 text-sm text-gray-500">
            共 {total} 个任务
          </p>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={handleRefresh}
            disabled={loading}
            className="flex items-center gap-2 rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            刷新
          </button>
          <button
            onClick={() => navigate(isNewTaskPage ? '/tasks' : '/tasks/new')}
            className="flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700"
          >
            <Plus className="h-4 w-4" />
            {isNewTaskPage ? '收起新建' : '新建任务'}
          </button>
        </div>
      </div>

      {/* 新建任务表单 */}
      {showRunForm && (
        <div className="grid gap-6 lg:grid-cols-3">
          <div className="lg:col-span-2">
            <div className="rounded-lg border border-gray-200 bg-white p-6">
              <TaskRunForm onSuccess={handleRunSuccess} />
            </div>
          </div>
          <div>
            <ModeSelector />
          </div>
        </div>
      )}

      {/* 筛选 */}
      <div className="flex items-center gap-2">
        <Filter className="h-4 w-4 text-gray-400" />
        <div className="flex gap-2">
          {statusOptions.map((option) => (
            <button
              key={option.value}
              onClick={() => handleStatusFilter(option.value)}
              className={`rounded-lg px-3 py-1.5 text-sm font-medium transition-colors ${
                statusFilter === option.value
                  ? 'bg-indigo-100 text-indigo-700'
                  : 'text-gray-600 hover:bg-gray-100'
              }`}
            >
              {option.label}
            </button>
          ))}
        </div>
      </div>

      {/* 错误 */}
      {error && (
        <div className="rounded-lg bg-red-50 p-4 text-sm text-red-600">
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
            <button
              onClick={() => navigate('/tasks/new')}
              className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700"
            >
              创建任务
            </button>
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
              <p className="text-sm text-gray-500">
                第 {currentPage} / {totalPages} 页，共 {total} 条
              </p>
              <div className="flex gap-2">
                <button
                  onClick={() => goToPage(currentPage - 1)}
                  disabled={currentPage <= 1}
                  className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
                >
                  上一页
                </button>
                <button
                  onClick={() => goToPage(currentPage + 1)}
                  disabled={currentPage >= totalPages}
                  className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
                >
                  下一页
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}
