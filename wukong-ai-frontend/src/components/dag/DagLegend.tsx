import React from 'react'
import type { DagNodeStatus } from '@/types'

interface DagLegendProps {
  compact?: boolean
}

/**
 * DAG 图例
 */
export function DagLegend({ compact = false }: DagLegendProps) {
  const items: { status: DagNodeStatus; label: string; color: string }[] = [
    { status: 'pending', label: '待处理', color: 'bg-gray-100 border-gray-300 text-gray-600' },
    { status: 'running', label: '运行中', color: 'bg-blue-100 border-blue-400 text-blue-700' },
    { status: 'success', label: '成功', color: 'bg-green-100 border-green-400 text-green-700' },
    { status: 'failed', label: '失败', color: 'bg-red-100 border-red-400 text-red-700' },
  ]

  if (compact) {
    return (
      <div className="flex items-center gap-3 text-xs text-gray-500">
        {items.map((item) => (
          <div key={item.status} className="flex items-center gap-1">
            <div className={`h-3 w-3 rounded border ${item.color}`}></div>
            <span>{item.label}</span>
          </div>
        ))}
      </div>
    )
  }

  return (
    <div className="rounded-lg border border-gray-200 bg-white p-4">
      <h4 className="mb-3 text-sm font-medium text-gray-700">节点状态</h4>
      <div className="grid grid-cols-2 gap-3">
        {items.map((item) => (
          <div key={item.status} className="flex items-center gap-2">
            <div className={`h-4 w-4 rounded border ${item.color}`}></div>
            <span className="text-sm text-gray-600">{item.label}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
