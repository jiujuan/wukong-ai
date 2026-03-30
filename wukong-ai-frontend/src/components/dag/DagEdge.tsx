import React from 'react'

interface DagEdgeProps {
  sourceX: number
  sourceY: number
  targetX: number
  targetY: number
  status?: 'pending' | 'running' | 'success' | 'failed'
  isHighlighted?: boolean
}

// 边的状态样式
const statusStyles = {
  pending: { stroke: '#d1d5db', strokeWidth: 2, markerEnd: 'url(#arrow-gray)' },
  running: { stroke: '#3b82f6', strokeWidth: 3, markerEnd: 'url(#arrow-blue)' },
  success: { stroke: '#22c55e', strokeWidth: 2, markerEnd: 'url(#arrow-green)' },
  failed: { stroke: '#ef4444', strokeWidth: 2, markerEnd: 'url(#arrow-red)' },
  highlighted: { stroke: '#6366f1', strokeWidth: 3 },
}

/**
 * DAG 边组件
 */
export function DagEdge({ sourceX, sourceY, targetX, targetY, status = 'pending', isHighlighted = false }: DagEdgeProps) {
  const controlOffset = Math.abs(targetY - sourceY) / 2

  // 水平连接
  if (Math.abs(sourceY - targetY) < 20) {
    return (
      <path
        d={`M ${sourceX} ${sourceY} L ${targetX} ${targetY}`}
        fill="none"
        {...(isHighlighted ? statusStyles.highlighted : statusStyles[status])}
        className="transition-all duration-300"
      />
    )
  }

  // 垂直或斜线连接
  return (
    <path
      d={`M ${sourceX} ${sourceY} C ${sourceX} ${sourceY + controlOffset}, ${targetX} ${targetY - controlOffset}, ${targetX} ${targetY}`}
      fill="none"
      {...(isHighlighted ? statusStyles.highlighted : statusStyles[status])}
      className="transition-all duration-300"
    />
  )
}
