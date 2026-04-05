import React, { useEffect } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { DagGraph } from '@/components/dag'
import { LoadingSpinner, ErrorAlert } from '@/components/common'
import { useDag, useTask, useTaskList, useTaskStream } from '@/hooks'
import { Button, Card, CardContent, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui'

/**
 * DAG 视图页面
 */
export function DagViewPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const navigate = useNavigate()

  const taskId = searchParams.get('task_id') || ''

  const { dagData, loading, error, loadDag, resetDag } = useDag()
  const { currentTask, getTask, clearEvents } = useTask()
  const { tasks, loading: taskListLoading, loadTasks } = useTaskList()

  useEffect(() => {
    loadTasks(1)
  }, [loadTasks])

  useEffect(() => {
    if (!taskId && tasks.length > 0) {
      setSearchParams({ task_id: tasks[0].task_id })
    }
  }, [taskId, tasks, setSearchParams])

  useEffect(() => {
    if (taskId) {
      clearEvents()
      getTask(taskId)
      loadDag(taskId)
    }
    return () => {
      resetDag()
    }
  }, [taskId, loadDag, resetDag, getTask, clearEvents])

  useTaskStream({
    taskId: taskId || '',
    enabled: !!taskId && currentTask?.status !== 'success' && currentTask?.status !== 'failed',
  })

  if (loading && !dagData) {
    return (
      <div className="flex h-64 items-center justify-center">
        <LoadingSpinner text="加载 DAG 数据..." />
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-4">
        <Card className="p-4">
          <div className="flex items-center gap-3">
            <span className="text-sm text-muted-foreground">切换任务：</span>
            <Select value={taskId} onValueChange={(value) => setSearchParams({ task_id: value })}>
              <SelectTrigger className="min-w-[320px]">
                <SelectValue placeholder="请选择任务" />
              </SelectTrigger>
              <SelectContent>
                {tasks.map((task) => (
                  <SelectItem key={task.task_id} value={task.task_id}>
                    {task.task_id} ｜ {task.status} ｜ {task.mode}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </Card>
        <ErrorAlert message={error} />
        <Button
          variant="outline"
          onClick={() => navigate('/tasks')}
        >
          返回列表
        </Button>
      </div>
    )
  }

  if (!taskId && taskListLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <LoadingSpinner text="加载任务列表..." />
      </div>
    )
  }

  if (tasks.length === 0 && !taskListLoading) {
    return (
      <div className="flex h-64 flex-col items-center justify-center gap-4">
        <p className="text-muted-foreground">暂无任务，创建任务后可在此查看 DAG 视图</p>
        <Button
          variant="outline"
          onClick={() => navigate('/tasks')}
        >
          前往任务管理
        </Button>
      </div>
    )
  }

  if (!taskId) {
    return (
      <div className="flex h-64 items-center justify-center">
        <LoadingSpinner text="正在切换任务..." />
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-4">
        <Button
          variant="outline"
          size="icon"
          onClick={() => navigate(`/tasks/${taskId}`)}
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div>
          <h2 className="text-lg font-semibold text-foreground">工作流执行图</h2>
          <p className="text-sm text-muted-foreground">任务 ID: {taskId}</p>
        </div>
      </div>

      <Card className="p-4">
        <div className="flex flex-col gap-2 md:flex-row md:items-center">
          <span className="text-sm text-muted-foreground">选择任务</span>
          <Select value={taskId} onValueChange={(value) => setSearchParams({ task_id: value })}>
            <SelectTrigger className="md:min-w-[520px]">
              <SelectValue placeholder="请选择任务" />
            </SelectTrigger>
            <SelectContent>
              {tasks.map((task) => (
                <SelectItem key={task.task_id} value={task.task_id}>
                  {task.task_id} ｜ {task.status} ｜ {task.mode} ｜ {task.user_input.slice(0, 24)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </Card>

      <DagGraph
        data={dagData}
        height={600}
        onNodeClick={(nodeId) => {
          console.log('Selected node:', nodeId)
        }}
      />

      {dagData && dagData.nodes.length > 0 && (
        <Card className="border-primary/20 bg-primary/5">
          <CardContent className="p-4">
            <p className="text-sm text-primary">
              提示：拖拽可以移动视图，滚轮可以缩放。点击节点查看详情。
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
