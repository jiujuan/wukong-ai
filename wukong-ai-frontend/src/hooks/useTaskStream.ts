import { useCallback, useRef, useEffect } from 'react'
import { useTaskStore } from '@/store'
import { useDagStore } from '@/store'
import type { ProgressEvent } from '@/types'

interface UseTaskStreamOptions {
  taskId: string
  enabled?: boolean
  onNodeStatusChange?: (nodeId: string, status: string) => void
  onTaskComplete?: () => void
}

/**
 * 任务流 Hook - 处理 SSE 事件
 */
export function useTaskStream({ taskId, enabled = true, onNodeStatusChange, onTaskComplete }: UseTaskStreamOptions) {
  const eventSourceRef = useRef<EventSource | null>(null)
  const onNodeStatusChangeRef = useRef<typeof onNodeStatusChange>(onNodeStatusChange)
  const onTaskCompleteRef = useRef<typeof onTaskComplete>(onTaskComplete)
  const addEvent = useTaskStore((state) => state.addEvent)
  const updateNodeStatus = useDagStore((state) => state.updateNodeStatus)
  const setSseConnected = useDagStore((state) => state.setSseConnected)

  useEffect(() => {
    onNodeStatusChangeRef.current = onNodeStatusChange
  }, [onNodeStatusChange])

  useEffect(() => {
    onTaskCompleteRef.current = onTaskComplete
  }, [onTaskComplete])

  const connect = useCallback(() => {
    if (!enabled || eventSourceRef.current) return

    const url = `/api/task/stream?task_id=${taskId}`
    const eventSource = new EventSource(url)
    eventSourceRef.current = eventSource

    // 连接成功
    eventSource.onopen = () => {
      setSseConnected(true)
    }

    const handleIncomingEvent = (event: MessageEvent) => {
      try {
        const parsed = JSON.parse(event.data)
        const data: ProgressEvent =
          typeof parsed === 'string' ? JSON.parse(parsed) : parsed
        addEvent(data)

        const nodeId = data.node
        if (nodeId) {
          const status = data.status as 'pending' | 'running' | 'success' | 'failed'
          updateNodeStatus(nodeId, status)
          onNodeStatusChangeRef.current?.(nodeId, status)
        }

        if (data.type === 'task_done' || data.type === 'task_failed') {
          onTaskCompleteRef.current?.()
          disconnect()
        }
      } catch (error) {
        console.error('Failed to parse SSE event:', error)
      }
    }

    eventSource.onmessage = handleIncomingEvent
    const eventTypes: ProgressEvent['type'][] = ['node_start', 'node_done', 'sub_agent_update', 'task_done', 'task_failed']
    eventTypes.forEach((eventType) => {
      eventSource.addEventListener(eventType, handleIncomingEvent as EventListener)
    })

    eventSource.onerror = () => {
      setSseConnected(false)
      disconnect()
    }
  }, [taskId, enabled, addEvent, updateNodeStatus, setSseConnected])

  const disconnect = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
      eventSourceRef.current = null
      setSseConnected(false)
    }
  }, [setSseConnected])

  // 连接
  useEffect(() => {
    if (enabled) {
      connect()
    }
    return () => {
      disconnect()
    }
  }, [enabled, connect, disconnect])

  return {
    connect,
    disconnect,
    isConnected: !!eventSourceRef.current,
  }
}
