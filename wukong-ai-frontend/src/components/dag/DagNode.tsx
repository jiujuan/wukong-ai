import React from 'react'
import type { DagNode as DagNodeType, DagNodeStatus } from '@/types'

interface DagNodeProps {
  node: DagNodeType
  isCurrent?: boolean
  isHighlighted?: boolean
  position?: { x: number; y: number }
  onClick?: (nodeId: string) => void
}

// 节点状态样式
const statusStyles: Record<DagNodeStatus, { bg: string; border: string; text: string }> = {
  pending: {
    bg: 'bg-gray-100',
    border: 'border-gray-300',
    text: 'text-gray-600',
  },
  running: {
    bg: 'bg-blue-100',
    border: 'border-blue-400',
    text: 'text-blue-700',
  },
  success: {
    bg: 'bg-green-100',
    border: 'border-green-400',
    text: 'text-green-700',
  },
  failed: {
    bg: 'bg-red-100',
    border: 'border-red-400',
    text: 'text-red-700',
  },
}

// 节点图标
const nodeIcons: Record<string, string> = {
  user_input: 'U',
  intention_analysis: 'I',
  plan_generation: 'P',
  task_execution: 'T',
  subagent: 'S',
  synthesis: 'Y',
  memory: 'M',
  tools: 'X',
  default: 'N',
}

/**
 * DAG 节点组件
 */
export function DagNode({ node, isCurrent = false, isHighlighted = false, position, onClick }: DagNodeProps) {
  const style = statusStyles[node.status]
  const icon = nodeIcons[node.type] || nodeIcons.default

  return (
    <div
      className={`
        absolute flex flex-col items-center rounded-lg border-2 p-3 shadow-sm transition-all
        ${style.bg} ${style.border}
        ${isCurrent ? 'ring-2 ring-offset-2 ring-indigo-500' : ''}
        ${isHighlighted ? 'scale-105 shadow-md' : ''}
        ${onClick ? 'cursor-pointer hover:shadow-md' : ''}
      `}
      style={{
        left: position?.x ?? 0,
        top: position?.y ?? 0,
        minWidth: '120px',
      }}
      onClick={() => onClick?.(node.id)}
    >
      {/* 图标 */}
      <div
        className={`mb-2 flex h-10 w-10 items-center justify-center rounded-full ${style.bg} border ${style.border} ${style.text} font-bold`}
      >
        {icon}
      </div>

      {/* 标签 */}
      <span className={`text-sm font-medium ${style.text}`}>{node.label}</span>

      {/* 运行动画 */}
      {node.status === 'running' && (
        <div className="mt-2 h-1 w-full overflow-hidden rounded-full bg-gray-200">
          <div className="h-full animate-pulse bg-blue-500"></div>
        </div>
      )}

      {/* 状态图标 */}
      {node.status === 'success' && (
        <span className="absolute -right-1 -top-1 flex h-5 w-5 items-center justify-center rounded-full bg-green-500 text-white">
          <svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
          </svg>
        </span>
      )}
      {node.status === 'failed' && (
        <span className="absolute -right-1 -top-1 flex h-5 w-5 items-center justify-center rounded-full bg-red-500 text-white">
          <svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </span>
      )}
    </div>
  )
}
