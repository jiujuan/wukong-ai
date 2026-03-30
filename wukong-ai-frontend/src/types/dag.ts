// DAG 节点状态
export type DagNodeStatus = 'pending' | 'running' | 'success' | 'failed'

// DAG 节点
export interface DagNode {
  id: string
  label: string
  status: DagNodeStatus
}

// DAG 边
export interface DagEdge {
  source: string
  target: string
}

// DAG 数据
export interface DagData {
  task_id: string
  mode: string
  nodes: DagNode[]
  edges: DagEdge[]
}

// DAG 节点位置（用于渲染）
export interface DagNodePosition {
  id: string
  x: number
  y: number
  width: number
  height: number
}

// DAG 布局配置
export interface DagLayoutConfig {
  nodeWidth: number
  nodeHeight: number
  horizontalGap: number
  verticalGap: number
  startX: number
  startY: number
}

// 节点状态样式映射
export const DAG_NODE_STYLES: Record<DagNodeStatus, { bg: string; border: string; text: string; ring: string }> = {
  pending: {
    bg: 'bg-gray-50',
    border: 'border-gray-300',
    text: 'text-gray-500',
    ring: '',
  },
  running: {
    bg: 'bg-blue-50',
    border: 'border-blue-500',
    text: 'text-blue-700',
    ring: 'ring-2 ring-blue-500 animate-pulse',
  },
  success: {
    bg: 'bg-green-50',
    border: 'border-green-500',
    text: 'text-green-700',
    ring: '',
  },
  failed: {
    bg: 'bg-red-50',
    border: 'border-red-500',
    text: 'text-red-700',
    ring: '',
  },
}

// 节点状态图标映射
export const DAG_NODE_ICONS: Record<DagNodeStatus, string> = {
  pending: '⏳',
  running: '🔄',
  success: '✅',
  failed: '❌',
}

// 模式对应的节点配置
export const MODE_NODE_CONFIG: Record<string, { nodes: string[]; edges: [string, string][] }> = {
  flash: {
    nodes: ['coordinator'],
    edges: [],
  },
  standard: {
    nodes: ['coordinator', 'background', 'researcher', 'reporter'],
    edges: [
      ['coordinator', 'background'],
      ['background', 'researcher'],
      ['researcher', 'reporter'],
    ],
  },
  pro: {
    nodes: ['coordinator', 'planner', 'researcher', 'reporter'],
    edges: [
      ['coordinator', 'planner'],
      ['planner', 'researcher'],
      ['researcher', 'reporter'],
    ],
  },
  ultra: {
    nodes: ['coordinator', 'planner', 'subagentmanager', 'reporter'],
    edges: [
      ['coordinator', 'planner'],
      ['planner', 'subagentmanager'],
      ['subagentmanager', 'reporter'],
    ],
  },
}
