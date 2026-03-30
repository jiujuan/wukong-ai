import React from 'react'
import { Loader2 } from 'lucide-react'
import type { TaskStatus } from '@/types'

interface TaskStatusBadgeProps {
  status: TaskStatus
  size?: 'sm' | 'md'
}

const statusConfig: Record<TaskStatus, { label: string; className: string }> = {
  pending: {
    label: '待处理',
    className: 'bg-gray-100 text-gray-700',
  },
  queued: {
    label: '排队中',
    className: 'bg-amber-100 text-amber-700',
  },
  running: {
    label: '运行中',
    className: 'bg-green-100 text-green-700',
  },
  success: {
    label: '成功',
    className: 'bg-green-100 text-green-700',
  },
  failed: {
    label: '失败',
    className: 'bg-red-100 text-red-700',
  },
}

/**
 * 任务状态徽章
 */
export function TaskStatusBadge({ status, size = 'sm' }: TaskStatusBadgeProps) {
  const config = statusConfig[status]

  const sizeClasses = size === 'sm' ? 'px-2 py-0.5 text-xs' : 'px-3 py-1 text-sm'

  return (
    <span
      className={`inline-flex items-center rounded-full font-medium ${config.className} ${sizeClasses}`}
    >
      {status === 'running' && (
        <Loader2 className="mr-1.5 h-3 w-3 animate-spin" />
      )}
      {config.label}
    </span>
  )
}
