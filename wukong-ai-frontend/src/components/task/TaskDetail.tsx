import React from 'react'
import { ArrowLeft, RefreshCw, XCircle, Clock, AlertCircle, Loader2 } from 'lucide-react'
import { Link } from 'react-router-dom'
import { TaskStatusBadge } from './TaskStatusBadge'
import { TaskResultPanel } from './TaskResultPanel'
import { TaskProgressPanel } from './TaskProgressPanel'
import { useTask } from '@/hooks'
import type { TaskDetail as TaskDetailType } from '@/types'
import { calculateMode } from '@/store'

interface TaskDetailProps {
  task: TaskDetailType
}

/**
 * 任务详情组件
 */
export function TaskDetail({ task }: TaskDetailProps) {
  const { resumeTask, cancelTask, loading, events } = useTask()
  const normalizeMultilineText = (input?: string) => {
    if (!input) return ''
    const raw = input.trim()
    if (!raw) return ''
    try {
      if (raw.startsWith('"') && raw.endsWith('"')) {
        const decoded = JSON.parse(raw)
        if (typeof decoded === 'string') {
          return decoded
        }
      }
    } catch {
    }
    return raw
      .replace(/\\r\\n/g, '\n')
      .replace(/\\n/g, '\n')
      .replace(/\\t/g, '\t')
  }

  const mode = calculateMode({
    thinking: task.mode === 'standard' || task.mode === 'pro' || task.mode === 'ultra',
    plan: task.mode === 'pro' || task.mode === 'ultra',
    subagent: task.mode === 'ultra',
  })

  const modeLabels = {
    flash: '快速模式',
    standard: '标准模式',
    pro: '增强模式',
    ultra: '超级模式',
  }

  const handleResume = async () => {
    if (task.task_id) {
      await resumeTask(task.task_id)
    }
  }

  const handleCancel = async () => {
    if (task.task_id) {
      await cancelTask(task.task_id)
    }
  }

  const formatTime = (time?: string) => {
    if (!time) return '-'
    return new Date(time).toLocaleString('zh-CN')
  }

  const formatDuration = () => {
    if (!task.finish_time) return '-'
    const duration = new Date(task.finish_time).getTime() - new Date(task.create_time).getTime()
    const seconds = Math.round(duration / 1000)
    if (seconds < 60) return `${seconds}s`
    const minutes = Math.floor(seconds / 60)
    const remainingSeconds = seconds % 60
    return `${minutes}m ${remainingSeconds}s`
  }

  const liveFlashOutput = task.mode === 'flash'
    ? events
      .filter((event) => event.type === 'sub_agent_update' && event.node === 'coordinator' && !!event.latest)
      .map((event) => event.latest || '')
      .join('')
    : ''
  const livePlannerOutput = events
    .filter((event) => event.type === 'sub_agent_update' && event.node === 'planner' && !!event.latest)
    .map((event) => event.latest || '')
    .join('')
  const liveReporterOutput = events
    .filter((event) => event.type === 'sub_agent_update' && event.node === 'reporter' && !!event.latest)
    .map((event) => event.latest || '')
    .join('')

  const isTaskActive = task.status !== 'success' && task.status !== 'failed'
  const isFlashRunning = task.mode === 'flash' && isTaskActive
  const isFlashCompleted = task.mode === 'flash' && task.status === 'success'
  const normalizedPlan = normalizeMultilineText(livePlannerOutput || task.plan || '')
  const resultContent = task.mode === 'flash'
    ? (isFlashRunning ? (liveFlashOutput || task.final_output || '') : (task.final_output || liveFlashOutput))
    : (isTaskActive ? (liveReporterOutput || task.final_output || '') : (task.final_output || liveReporterOutput))

  return (
    <div className="space-y-6">
      {/* 头部 */}
      <div className="flex items-center gap-4">
        <Link
          to="/tasks"
          className="flex h-9 w-9 items-center justify-center rounded-lg border border-gray-300 bg-white text-gray-700 hover:bg-gray-50"
        >
          <ArrowLeft className="h-4 w-4" />
        </Link>
        <div>
          <h2 className="text-lg font-semibold text-gray-900">任务详情</h2>
          <p className="text-sm text-gray-500">ID: {task.task_id}</p>
        </div>
      </div>

      {/* 基本信息 */}
      <div className="rounded-lg border border-gray-200 bg-white p-6">
        <div className="mb-4 flex items-center justify-between">
          <h3 className="text-base font-medium text-gray-900">基本信息</h3>
          <div className="flex items-center gap-2">
            <TaskStatusBadge status={task.status} size="md" />
            <span className="rounded-full bg-indigo-100 px-3 py-1 text-sm font-medium text-indigo-700">
              {modeLabels[mode]}
            </span>
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <label className="mb-1 block text-sm text-gray-500">用户输入</label>
            <p className="text-sm text-gray-900">{task.user_input}</p>
          </div>
          {task.intention && (
            <div>
              <label className="mb-1 block text-sm text-gray-500">意图分析</label>
              <p className="text-sm text-gray-900">{task.intention}</p>
            </div>
          )}
          {normalizedPlan && (
            <div className="md:col-span-2">
              <label className="mb-1 block text-sm text-gray-500">执行计划</label>
              <p className="whitespace-pre-wrap text-sm text-gray-900">{normalizedPlan}</p>
            </div>
          )}
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-1 text-sm text-gray-500">
              <Clock className="h-4 w-4" />
              创建: {formatTime(task.create_time)}
            </div>
            {task.finish_time && (
              <div className="text-sm text-gray-500">耗时: {formatDuration()}</div>
            )}
          </div>
        </div>

        {/* 操作按钮 */}
        {task.status !== 'success' && task.status !== 'failed' && (
          <div className="mt-4 flex gap-3 border-t border-gray-100 pt-4">
            {task.status === 'running' && (
              <button
                onClick={handleCancel}
                disabled={loading}
                className="flex items-center gap-2 rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50"
              >
                <XCircle className="h-4 w-4" />
                取消任务
              </button>
            )}
            {(task.status === 'pending' || task.status === 'queued') && (
              <button
                onClick={handleCancel}
                disabled={loading}
                className="flex items-center gap-2 rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50"
              >
                <XCircle className="h-4 w-4" />
                取消任务
              </button>
            )}
          </div>
        )}

        {/* 续跑按钮 */}
        {(task.status === 'success' || task.status === 'failed') && (
          <div className="mt-4 flex gap-3 border-t border-gray-100 pt-4">
            <button
              onClick={handleResume}
              disabled={loading}
              className="flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
            >
              <RefreshCw className="h-4 w-4" />
              续跑任务
            </button>
          </div>
        )}

        {/* 错误信息 */}
        {task.status !== 'success' && task.error_msg && (
          <div className="mt-4 rounded-lg bg-red-50 p-4">
            <div className="flex items-start gap-3">
              <AlertCircle className="h-5 w-5 flex-shrink-0 text-red-500" />
              <div>
                <h4 className="text-sm font-medium text-red-800">错误信息</h4>
                <p className="mt-1 text-sm text-red-700">{task.error_msg}</p>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* 子任务 */}
      {task.tasks && task.tasks.length > 0 && (
        <div className="rounded-lg border border-gray-200 bg-white p-6">
          <h3 className="mb-4 text-base font-medium text-gray-900">
            子任务 ({task.tasks.length})
          </h3>
          <div className="space-y-2">
            {task.tasks.map((subTask, index) => (
              <div
                key={index}
                className="flex items-center gap-3 rounded-lg bg-gray-50 p-3 text-sm"
              >
                <span className="flex h-6 w-6 items-center justify-center rounded-full bg-indigo-100 text-xs font-medium text-indigo-700">
                  {index + 1}
                </span>
                <span className="text-gray-700">{subTask}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 执行结果 */}
      {resultContent && (
        <TaskResultPanel
          content={resultContent}
          title={isTaskActive ? '实时输出' : '执行结果'}
        />
      )}

      {isTaskActive && !resultContent && (
        <div className="rounded-lg border border-gray-200 bg-white p-6">
          <div className="flex items-center gap-2 text-sm text-gray-500">
            <Loader2 className="h-4 w-4 animate-spin text-gray-400" />
            <span>正在获取实时输出...</span>
          </div>
        </div>
      )}

      {isFlashCompleted && (
        <div className="rounded-lg border border-gray-200 bg-white p-6">
          <div className="flex items-center gap-2 text-sm text-gray-500">
            <Loader2 className="h-4 w-4 text-gray-400" />
            <span>回答完成✅</span>
          </div>
        </div>
      )}

      {isTaskActive && task.mode !== 'flash' && <TaskProgressPanel />}
    </div>
  )
}
