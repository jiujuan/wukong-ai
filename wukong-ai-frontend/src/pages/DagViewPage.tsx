import React, { useEffect } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { DagGraph } from '@/components/dag'
import { LoadingSpinner, ErrorAlert } from '@/components/common'
import { useDag, useTask, useTaskList, useTaskStream } from '@/hooks'

/**
 * DAG 视图页面
 */
export function DagViewPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const navigate = useNavigate()

  const taskId = searchParams.get('task_id') || ''

  const { dagData, loading, error, loadDag, resetDag } = useDag()
  const { currentTask, getTask, clearEvents } = useTask()
  const { tasks, loading: taskListLoading, loadTasks } = useTaskList()

  useEffect(() => {
    loadTasks(1)
  }, [loadTasks])

  useEffect(() => {
    if (!taskId && tasks.length > 0) {
      setSearchParams({ task_id: tasks[0].task_id })
    }
  }, [taskId, tasks, setSearchParams])

  useEffect(() => {
    if (taskId) {
      clearEvents()
      getTask(taskId)
      loadDag(taskId)
    }
    return () => {
      resetDag()
    }
  }, [taskId, loadDag, resetDag, getTask, clearEvents])

  useTaskStream({
    taskId: taskId || '',
    enabled: !!taskId && currentTask?.status !== 'success' && currentTask?.status !== 'failed',
  })

  if (loading && !dagData) {
    return (
      <div className="flex h-64 items-center justify-center">
        <LoadingSpinner text="加载 DAG 数据..." />
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-4">
        <div className="flex items-center gap-3 rounded-lg border border-gray-200 bg-white p-4">
          <span className="text-sm text-gray-600">切换任务：</span>
          <select
            value={taskId}
            onChange={(event) => setSearchParams({ task_id: event.target.value })}
            className="min-w-[320px] rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 outline-none focus:border-indigo-500"
          >
            {tasks.map((task) => (
              <option key={task.task_id} value={task.task_id}>
                {task.task_id} ｜ {task.status} ｜ {task.mode}
              </option>
            ))}
          </select>
        </div>
        <ErrorAlert message={error} />
        <button
          onClick={() => navigate('/tasks')}
          className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
        >
          返回列表
        </button>
      </div>
    )
  }

  if (!taskId && taskListLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <LoadingSpinner text="加载任务列表..." />
      </div>
    )
  }

  if (tasks.length === 0 && !taskListLoading) {
    return (
      <div className="flex h-64 flex-col items-center justify-center gap-4">
        <p className="text-gray-500">暂无任务，创建任务后可在此查看 DAG 视图</p>
        <button
          onClick={() => navigate('/tasks')}
          className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
        >
          前往任务管理
        </button>
      </div>
    )
  }

  if (!taskId) {
    return (
      <div className="flex h-64 items-center justify-center">
        <LoadingSpinner text="正在切换任务..." />
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate(`/tasks/${taskId}`)}
          className="flex h-9 w-9 items-center justify-center rounded-lg border border-gray-300 bg-white text-gray-700 hover:bg-gray-50"
        >
          <ArrowLeft className="h-4 w-4" />
        </button>
        <div>
          <h2 className="text-lg font-semibold text-gray-900">工作流执行图</h2>
          <p className="text-sm text-gray-500">任务 ID: {taskId}</p>
        </div>
      </div>

      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <div className="flex flex-col gap-2 md:flex-row md:items-center">
          <span className="text-sm text-gray-600">选择任务</span>
          <select
            value={taskId}
            onChange={(event) => setSearchParams({ task_id: event.target.value })}
            className="rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 outline-none focus:border-indigo-500 md:min-w-[520px]"
          >
            {tasks.map((task) => (
              <option key={task.task_id} value={task.task_id}>
                {task.task_id} ｜ {task.status} ｜ {task.mode} ｜ {task.user_input.slice(0, 24)}
              </option>
            ))}
          </select>
        </div>
      </div>

      <DagGraph
        data={dagData}
        height={600}
        onNodeClick={(nodeId) => {
          console.log('Selected node:', nodeId)
        }}
      />

      {dagData && dagData.nodes.length > 0 && (
        <div className="rounded-lg bg-blue-50 p-4">
          <p className="text-sm text-blue-700">
            提示：拖拽可以移动视图，滚轮可以缩放。点击节点查看详情。
          </p>
        </div>
      )}
    </div>
  )
}
