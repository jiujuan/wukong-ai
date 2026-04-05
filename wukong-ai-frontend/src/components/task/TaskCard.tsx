import React from 'react'
import { Link } from 'react-router-dom'
import { Clock, ArrowRight, Zap, Brain, Map, Users } from 'lucide-react'
import { TaskStatusBadge } from './TaskStatusBadge'
import type { TaskListItem } from '@/types'
import { calculateMode } from '@/store'
import { Badge, Card, CardContent } from '@/components/ui'

interface TaskCardProps {
  task: TaskListItem
}

/**
 * 任务卡片
 */
export function TaskCard({ task }: TaskCardProps) {
  const mode = calculateMode({
    thinking: task.mode === 'standard' || task.mode === 'pro' || task.mode === 'ultra',
    plan: task.mode === 'pro' || task.mode === 'ultra',
    subagent: task.mode === 'ultra',
  })

  const modeIcons = {
    flash: <Zap className="h-3 w-3" />,
    standard: <Brain className="h-3 w-3" />,
    pro: <Map className="h-3 w-3" />,
    ultra: <Users className="h-3 w-3" />,
  }

  const modeLabels = {
    flash: '快速',
    standard: '标准',
    pro: '增强',
    ultra: '超级',
  }

  const formatTime = (time: string) => {
    const date = new Date(time)
    return date.toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  const truncateText = (text: string, maxLength: number = 100) => {
    if (text.length <= maxLength) return text
    return text.slice(0, maxLength) + '...'
  }

  return (
    <Link to={`/tasks/${task.task_id}`} className="block">
      <Card className="transition-colors hover:border-primary/30">
        <CardContent className="p-4">
          <div className="mb-3 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <TaskStatusBadge status={task.status} />
              <Badge variant="secondary" className="gap-1">
                {modeIcons[mode]}
                {modeLabels[mode]}
              </Badge>
            </div>
            <ArrowRight className="h-4 w-4 text-muted-foreground" />
          </div>
          <p className="mb-3 text-sm text-foreground">{truncateText(task.user_input)}</p>
          <div className="flex items-center justify-between text-xs text-muted-foreground">
            <div className="flex items-center gap-1">
              <Clock className="h-3 w-3" />
              {formatTime(task.create_time)}
            </div>
            {task.finish_time && (
              <span>
                耗时: {Math.round((new Date(task.finish_time).getTime() - new Date(task.create_time).getTime()) / 1000)}s
              </span>
            )}
          </div>
        </CardContent>
      </Card>
    </Link>
  )
}
