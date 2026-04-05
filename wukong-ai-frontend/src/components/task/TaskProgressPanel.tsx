import { Activity, Loader2 } from 'lucide-react'
import { useDagStore, useTaskStore } from '@/store'
import type { ProgressEvent } from '@/types'
import { formatTimestamp } from '@/utils'
import { Badge, Card, CardContent, CardHeader, CardTitle } from '@/components/ui'

/**
 * 任务进度面板
 */
export function TaskProgressPanel() {
  const events = useTaskStore((state) => state.events)
  const sseConnected = useDagStore((state) => state.sseConnected)
  const displayEvents = events.filter((event, index, list) => {
    if (event.type !== 'sub_agent_update') {
      return true
    }
    const isProgressEvent = typeof event.done === 'number' && typeof event.total === 'number' && event.total > 0
    const key = isProgressEvent
      ? `${event.type}:${event.node || ''}:${event.done}/${event.total}`
      : `${event.type}:${event.node || ''}:${event.status || ''}`
    const firstIndex = list.findIndex((item) => {
      if (item.type !== 'sub_agent_update') {
        return false
      }
      const itemIsProgressEvent = typeof item.done === 'number' && typeof item.total === 'number' && item.total > 0
      const itemKey = itemIsProgressEvent
        ? `${item.type}:${item.node || ''}:${item.done}/${item.total}`
        : `${item.type}:${item.node || ''}:${item.status || ''}`
      return itemKey === key
    })
    return firstIndex === index
  })
  const isConnected = sseConnected || displayEvents.length > 0

  const getEventMessage = (event: ProgressEvent) => {
    if (event.type === 'node_start') return `节点 ${event.node || '-'} 开始执行`
    if (event.type === 'node_done') return `节点 ${event.node || '-'} 执行${event.status === 'failed' ? '失败' : '完成'}`
    if (event.type === 'sub_agent_update') {
      if (typeof event.done === 'number' && typeof event.total === 'number' && event.total > 0) {
        return `子任务进度 ${event.done}/${event.total}`
      }
      if (event.node) {
        return `节点 ${event.node} 运行中`
      }
      return '子任务进度更新'
    }
    if (event.type === 'task_done') return '任务执行完成'
    if (event.type === 'task_failed') return '任务执行失败'
    return '收到进度更新'
  }

  return (
    <Card>
      <CardHeader className="flex-row items-center justify-between space-y-0">
        <div className="flex items-center gap-2">
          <Activity className="h-5 w-5 text-primary" />
          <CardTitle className="text-base">实时进度</CardTitle>
        </div>
        <div>
          {isConnected ? (
            <Badge variant="secondary" className="bg-emerald-100 text-emerald-700">已连接</Badge>
          ) : (
            <span className="flex items-center gap-1.5 text-sm text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" />
              连接中...
            </span>
          )}
        </div>
      </CardHeader>
      <CardContent className="max-h-80 overflow-y-auto">
        {displayEvents.length === 0 ? (
          <div className="py-8 text-center text-sm text-muted-foreground">
            等待任务开始...
          </div>
        ) : (
          <div className="space-y-2">
            {displayEvents.map((event, index) => (
              <div
                key={index}
                className="flex items-start gap-3 rounded-lg bg-muted/50 p-3"
              >
                <div className="mt-0.5 h-2 w-2 flex-shrink-0 rounded-full bg-primary" />
                <div className="min-w-0 flex-1">
                  <p className="text-sm text-foreground">{getEventMessage(event)}</p>
                  {event.node && (
                    <p className="mt-1 text-xs text-muted-foreground">
                      节点: {event.node}
                    </p>
                  )}
                  <p className="text-xs text-muted-foreground">
                    {formatTimestamp(event.timestamp)}
                  </p>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
