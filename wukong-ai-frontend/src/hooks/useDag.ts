import { useCallback } from 'react'
import { useDagStore } from '@/store'
import { usePolling } from './usePolling'
import type { DagNodeStatus } from '@/types'

/**
 * DAG Hook
 */
export function useDag() {
  const dagData = useDagStore((state) => state.dagData)
  const loading = useDagStore((state) => state.loading)
  const error = useDagStore((state) => state.error)
  const currentNodeId = useDagStore((state) => state.currentNodeId)
  const currentNodeStatus = useDagStore((state) => state.currentNodeStatus)
  const sseConnected = useDagStore((state) => state.sseConnected)

  const fetchDag = useDagStore((state) => state.fetchDag)
  const updateNodeStatus = useDagStore((state) => state.updateNodeStatus)
  const setSseConnected = useDagStore((state) => state.setSseConnected)
  const reset = useDagStore((state) => state.reset)

  // 获取 DAG 数据
  const loadDag = useCallback(async (taskId: string) => {
    return fetchDag(taskId)
  }, [fetchDag])

  // 更新节点状态
  const setNodeStatus = useCallback((nodeId: string, status: DagNodeStatus) => {
    updateNodeStatus(nodeId, status)
  }, [updateNodeStatus])

  // 加载并轮询 DAG 状态
  const loadDagWithPolling = useCallback((taskId: string, interval = 3000) => {
    loadDag(taskId)

    // 使用轮询获取更新（作为 SSE 的后备方案）
    const pollFn = () => fetchDag(taskId)

    // 这里不使用自动轮询，因为有 SSE
    return { pollFn }
  }, [loadDag, fetchDag])

  // 获取节点样式
  const getNodeStyle = useCallback((nodeId: string) => {
    const isCurrent = currentNodeId === nodeId
    const status = isCurrent ? currentNodeStatus : undefined
    return { isCurrent, status }
  }, [currentNodeId, currentNodeStatus])

  // 获取当前节点
  const getCurrentNode = useCallback(() => {
    if (!dagData || !currentNodeId) return null
    return dagData.nodes.find((node) => node.id === currentNodeId) || null
  }, [dagData, currentNodeId])

  // 重置 DAG 状态
  const resetDag = useCallback(() => {
    reset()
  }, [reset])

  return {
    // 状态
    dagData,
    loading,
    error,
    currentNodeId,
    currentNodeStatus,
    sseConnected,
    // 方法
    loadDag,
    loadDagWithPolling,
    setNodeStatus,
    setSseConnected,
    resetDag,
    getNodeStyle,
    getCurrentNode,
  }
}
