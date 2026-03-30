import React, { useRef, useEffect, useState, useCallback } from 'react'
import { DagNode } from './DagNode'
import { DagEdge } from './DagEdge'
import { DagLegend } from './DagLegend'
import { useDag } from '@/hooks'
import type { DagData, DagNode as DagNodeType, DagEdge as DagEdgeType } from '@/types'
import { computeDagLayout } from '@/utils'

interface DagGraphProps {
  data: DagData | null
  selectedNodeId?: string | null
  onNodeClick?: (nodeId: string) => void
  height?: number
}

const NODE_WIDTH = 140
const NODE_HEIGHT = 100
const HORIZONTAL_GAP = 80
const VERTICAL_GAP = 60

/**
 * DAG 图组件
 */
export function DagGraph({ data, selectedNodeId, onNodeClick, height = 600 }: DagGraphProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const [scale, setScale] = useState(1)
  const [pan, setPan] = useState({ x: 0, y: 0 })
  const [isDragging, setIsDragging] = useState(false)
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 })

  const { currentNodeId } = useDag()

  // 计算布局
  const layout = data ? computeDagLayout(data.nodes, data.edges, {
    nodeWidth: NODE_WIDTH,
    nodeHeight: NODE_HEIGHT,
    horizontalGap: HORIZONTAL_GAP,
    verticalGap: VERTICAL_GAP,
  }) : { nodes: [], edges: [], width: 0, height: 0 }

  // 缩放
  const handleZoom = useCallback((delta: number) => {
    setScale((prev) => Math.min(Math.max(prev + delta, 0.5), 2))
  }, [])

  // 拖拽开始
  const handleMouseDown = (e: React.MouseEvent) => {
    if (e.target === containerRef.current || (e.target as HTMLElement).tagName === 'svg') {
      setIsDragging(true)
      setDragStart({ x: e.clientX - pan.x, y: e.clientY - pan.y })
    }
  }

  // 拖拽移动
  const handleMouseMove = (e: React.MouseEvent) => {
    if (isDragging) {
      setPan({
        x: e.clientX - dragStart.x,
        y: e.clientY - dragStart.y,
      })
    }
  }

  // 拖拽结束
  const handleMouseUp = () => {
    setIsDragging(false)
  }

  // 重置视图
  const handleReset = () => {
    setScale(1)
    setPan({ x: 0, y: 0 })
  }

  if (!data) {
    return (
      <div className="flex h-96 items-center justify-center rounded-lg border border-gray-200 bg-gray-50">
        <p className="text-gray-500">暂无 DAG 数据</p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* 工具栏 */}
      <div className="flex items-center justify-between rounded-lg border border-gray-200 bg-white px-4 py-2">
        <div className="flex items-center gap-2">
          <button
            onClick={() => handleZoom(0.1)}
            className="rounded px-2 py-1 text-sm hover:bg-gray-100"
          >
            +
          </button>
          <span className="text-sm text-gray-600">{Math.round(scale * 100)}%</span>
          <button
            onClick={() => handleZoom(-0.1)}
            className="rounded px-2 py-1 text-sm hover:bg-gray-100"
          >
            -
          </button>
          <button
            onClick={handleReset}
            className="ml-2 rounded px-3 py-1 text-sm text-gray-600 hover:bg-gray-100"
          >
            重置
          </button>
        </div>
        <DagLegend compact />
      </div>

      {/* 图 */}
      <div
        ref={containerRef}
        className="relative overflow-hidden rounded-lg border border-gray-200 bg-gray-50"
        style={{ height }}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
      >
        <svg
          className="pointer-events-none absolute inset-0"
          width="100%"
          height="100%"
          style={{ cursor: isDragging ? 'grabbing' : 'grab' }}
        >
          {/* 箭头定义 */}
          <defs>
            <marker id="arrow-gray" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">
              <path d="M0,0 L0,6 L9,3 z" fill="#d1d5db" />
            </marker>
            <marker id="arrow-blue" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">
              <path d="M0,0 L0,6 L9,3 z" fill="#3b82f6" />
            </marker>
            <marker id="arrow-green" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">
              <path d="M0,0 L0,6 L9,3 z" fill="#22c55e" />
            </marker>
            <marker id="arrow-red" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">
              <path d="M0,0 L0,6 L9,3 z" fill="#ef4444" />
            </marker>
          </defs>

          <g
            style={{
              transform: `translate(${pan.x}px, ${pan.y}px) scale(${scale})`,
              transformOrigin: '0 0',
            }}
          >
            {layout.edges.map((edge, index) => {
              const sourceNode = layout.nodes.find((n) => n.node.id === edge.source)
              const targetNode = layout.nodes.find((n) => n.node.id === edge.target)

              if (!sourceNode || !targetNode) return null

              return (
                <DagEdge
                  key={`edge-${index}`}
                  sourceX={sourceNode.x + NODE_WIDTH}
                  sourceY={sourceNode.y + NODE_HEIGHT / 2}
                  targetX={targetNode.x}
                  targetY={targetNode.y + NODE_HEIGHT / 2}
                  status={edge.status}
                  isHighlighted={edge.source === currentNodeId || edge.target === currentNodeId}
                />
              )
            })}
          </g>
        </svg>

        {/* 内容 */}
        <div
          className="absolute"
          style={{
            transform: `translate(${pan.x}px, ${pan.y}px) scale(${scale})`,
            transformOrigin: '0 0',
          }}
        >
          {/* 节点 */}
          {layout.nodes.map((item) => (
            <DagNode
              key={item.node.id}
              node={item.node}
              position={{ x: item.x, y: item.y }}
              isCurrent={item.node.id === currentNodeId}
              isHighlighted={item.node.id === selectedNodeId}
              onClick={onNodeClick}
            />
          ))}
        </div>
      </div>
    </div>
  )
}
