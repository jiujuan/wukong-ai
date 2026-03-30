import { create } from 'zustand'
import type { DagData, DagNode, DagNodeStatus } from '@/types'
import { dagApi } from '@/api'

// DAG 状态
interface DagStore {
  dagData: DagData | null
  loading: boolean
  error: string | null

  // 当前节点状态
  currentNodeId: string | null
  currentNodeStatus: DagNodeStatus | null

  // SSE 连接状态
  sseConnected: boolean

  // Actions
  fetchDag: (taskId: string) => Promise<void>
  updateNodeStatus: (nodeId: string, status: DagNodeStatus) => void
  setSseConnected: (connected: boolean) => void
  reset: () => void
}

export const useDagStore = create<DagStore>((set, get) => ({
  // 初始状态
  dagData: null,
  loading: false,
  error: null,

  currentNodeId: null,
  currentNodeStatus: null,

  sseConnected: false,

  // 获取 DAG 数据
  fetchDag: async (taskId: string) => {
    set({ loading: true, error: null })
    try {
      const data = await dagApi.getDag(taskId)
      set({ dagData: data, loading: false })
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : 'Failed to fetch DAG',
        loading: false,
      })
    }
  },

  // 更新节点状态
  updateNodeStatus: (nodeId: string, status: DagNodeStatus) => {
    const { dagData } = get()
    if (!dagData) return

    const updatedNodes = dagData.nodes.map((node) =>
      node.id === nodeId ? { ...node, status } : node
    )

    set({
      dagData: { ...dagData, nodes: updatedNodes },
      currentNodeId: nodeId,
      currentNodeStatus: status,
    })
  },

  // 设置 SSE 连接状态
  setSseConnected: (connected: boolean) => {
    set({ sseConnected: connected })
  },

  // 重置状态
  reset: () => {
    set({
      dagData: null,
      loading: false,
      error: null,
      currentNodeId: null,
      currentNodeStatus: null,
      sseConnected: false,
    })
  },
}))
