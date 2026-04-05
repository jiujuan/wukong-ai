import React from 'react'
import { Loader2 } from 'lucide-react'
import type { TaskStatus } from '@/types'
import { Badge } from '@/components/ui'

interface TaskStatusBadgeProps {
  status: TaskStatus
  size?: 'sm' | 'md'
}

const statusConfig: Record<TaskStatus, { label: string; className: string; variant: 'secondary' | 'destructive' | 'default' }> = {
  pending: {
    label: '待处理',
    className: '',
    variant: 'secondary',
  },
  queued: {
    label: '排队中',
    className: 'bg-amber-100 text-amber-700 hover:bg-amber-100',
    variant: 'secondary',
  },
  running: {
    label: '运行中',
    className: 'bg-blue-100 text-blue-700 hover:bg-blue-100',
    variant: 'secondary',
  },
  success: {
    label: '成功',
    className: 'bg-emerald-100 text-emerald-700 hover:bg-emerald-100',
    variant: 'secondary',
  },
  failed: {
    label: '失败',
    className: '',
    variant: 'destructive',
  },
}

/**
 * 任务状态徽章
 */
export function TaskStatusBadge({ status, size = 'sm' }: TaskStatusBadgeProps) {
  const config = statusConfig[status]

  const sizeClasses = size === 'sm' ? 'px-2 py-0.5 text-xs' : 'px-3 py-1 text-sm'

  return (
    <Badge variant={config.variant} className={`${sizeClasses} ${config.className}`}>
      {status === 'running' && (
        <Loader2 className="mr-1.5 h-3 w-3 animate-spin" />
      )}
      {config.label}
    </Badge>
  )
}
