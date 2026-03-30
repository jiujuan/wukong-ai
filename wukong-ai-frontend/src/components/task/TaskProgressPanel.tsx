import { Activity, Loader2 } from 'lucide-react'
import { useDagStore, useTaskStore } from '@/store'
import type { ProgressEvent } from '@/types'
import { formatTimestamp } from '@/utils'

/**
 * 任务进度面板
 */
export function TaskProgressPanel() {
  const events = useTaskStore((state) => state.events)
  const sseConnected = useDagStore((state) => state.sseConnected)

  const getEventMessage = (event: ProgressEvent) => {
    if (event.type === 'node_start') return `节点 ${event.node || '-'} 开始执行`
    if (event.type === 'node_done') return `节点 ${event.node || '-'} 执行${event.status === 'failed' ? '失败' : '完成'}`
    if (event.type === 'sub_agent_update') return event.latest || '子任务进度更新'
    if (event.type === 'task_done') return '任务执行完成'
    if (event.type === 'task_failed') return '任务执行失败'
    return '收到进度更新'
  }

  return (
    <div className="rounded-lg border border-gray-200 bg-white">
      {/* 头部 */}
      <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
        <div className="flex items-center gap-2">
          <Activity className="h-5 w-5 text-indigo-500" />
          <h3 className="font-medium text-gray-900">实时进度</h3>
        </div>
        <div className="flex items-center gap-2">
          {sseConnected ? (
            <span className="flex items-center gap-1.5 text-sm text-green-600">
              <span className="h-2 w-2 rounded-full bg-green-500"></span>
              已连接
            </span>
          ) : (
            <span className="flex items-center gap-1.5 text-sm text-gray-500">
              <Loader2 className="h-4 w-4 animate-spin" />
              连接中...
            </span>
          )}
        </div>
      </div>

      {/* 事件列表 */}
      <div className="max-h-80 overflow-y-auto p-4">
        {events.length === 0 ? (
          <div className="py-8 text-center text-sm text-gray-500">
            等待任务开始...
          </div>
        ) : (
          <div className="space-y-2">
            {events.map((event, index) => (
              <div
                key={index}
                className="flex items-start gap-3 rounded-lg bg-gray-50 p-3"
              >
                <div className="mt-0.5 h-2 w-2 flex-shrink-0 rounded-full bg-indigo-500" />
                <div className="min-w-0 flex-1">
                  <p className="text-sm text-gray-700">{getEventMessage(event)}</p>
                  {event.node && (
                    <p className="mt-1 text-xs text-gray-400">
                      节点: {event.node}
                    </p>
                  )}
                  <p className="text-xs text-gray-400">
                    {formatTimestamp(event.timestamp)}
                  </p>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
