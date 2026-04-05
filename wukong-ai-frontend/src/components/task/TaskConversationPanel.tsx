import { useCallback, useEffect, useMemo, useState } from 'react'
import { ArrowUp, CheckCircle2, ChevronDown, ChevronRight, Loader2, MessageSquare, Zap, Brain, GraduationCap, Rocket } from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import type { Components } from 'react-markdown'
import type { ReactNode } from 'react'
import { conversationApi } from '@/api'
import { useTask, useTaskStream } from '@/hooks'
import type { ConversationTurn, ProgressEvent, TaskDetail } from '@/types'
import { calculateMode } from '@/store'
import { Button, Card, CardContent, CardHeader, CardTitle, DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuTrigger, Textarea } from '@/components/ui'

type ModeKey = 'flash' | 'standard' | 'pro' | 'ultra'

interface TaskConversationPanelProps {
  task: TaskDetail
}

function getStepMessage(event: ProgressEvent) {
  if (event.type === 'node_start') return `节点 ${event.node || '-'} 开始执行`
  if (event.type === 'node_done') return `节点 ${event.node || '-'} 执行${event.status === 'failed' ? '失败' : '完成'}`
  if (event.type === 'sub_agent_update') {
    if (typeof event.done === 'number' && typeof event.total === 'number' && event.total > 0) {
      return `子任务进度 ${event.done}/${event.total}`
    }
    if (event.node) return `节点 ${event.node} 运行中`
    return '子任务进度更新'
  }
  if (event.type === 'task_done') return '任务执行完成'
  if (event.type === 'task_failed') return '任务执行失败'
  return '收到进度更新'
}

function renderInlineBreaks(children: ReactNode): ReactNode {
  if (typeof children === 'string') {
    const parts = children.split(/<br\s*\/?>/gi)
    if (parts.length === 1) return children
    return parts.flatMap((part, index) => (index < parts.length - 1 ? [part, <br key={`conv-br-${index}`} />] : [part]))
  }
  if (Array.isArray(children)) {
    return children.map((child, index) => <span key={index}>{renderInlineBreaks(child)}</span>)
  }
  return children
}

function normalizeStreamingMarkdown(content: string) {
  const normalizedBr = content
    .split('\n')
    .map((line) => {
      if (line.includes('|')) return line
      return line.replace(/<br\s*\/?>/gi, '\n')
    })
    .join('\n')

  const normalized = normalizedBr
    .replace(/\r\n/g, '\n')
    .replace(/([。！？.!?:：\)])\s*(#{1,6}\s)/g, '$1\n$2')

  const fenceCount = (normalized.match(/```/g) || []).length
  if (fenceCount % 2 === 1) return `${normalized}\n\`\`\``

  const boldCount = (normalized.match(/\*\*/g) || []).length
  if (boldCount % 2 === 1) return `${normalized}**`

  return normalized
}

const markdownComponents: Components = {
  h1: ({ children }) => <h1 className="mb-2 mt-3 text-lg font-semibold text-foreground">{children}</h1>,
  h2: ({ children }) => <h2 className="mb-2 mt-3 text-base font-semibold text-foreground">{children}</h2>,
  h3: ({ children }) => <h3 className="mb-1 mt-2 text-sm font-semibold text-foreground">{children}</h3>,
  p: ({ children }) => <p className="text-sm leading-7 text-foreground">{renderInlineBreaks(children)}</p>,
  strong: ({ children }) => <strong className="font-semibold text-foreground">{children}</strong>,
  em: ({ children }) => <em className="italic text-foreground">{children}</em>,
  a: ({ children, href }) => (
    <a href={href} target="_blank" rel="noreferrer" className="text-primary underline">
      {children}
    </a>
  ),
  ul: ({ children }) => <ul className="list-disc space-y-1 pl-5 text-sm leading-7 text-foreground">{children}</ul>,
  ol: ({ children }) => <ol className="list-decimal space-y-1 pl-5 text-sm leading-7 text-foreground">{children}</ol>,
  code: ({ children }) => <code className="rounded bg-muted px-1 py-0.5 text-xs text-foreground">{children}</code>,
  pre: ({ children }) => <pre className="overflow-x-auto rounded-lg bg-muted p-3 text-xs text-foreground">{children}</pre>,
  blockquote: ({ children }) => <blockquote className="border-l-2 border-border pl-3 text-sm text-muted-foreground">{children}</blockquote>,
  table: ({ children }) => <table className="w-full border-collapse text-sm text-foreground">{children}</table>,
  th: ({ children }) => <th className="border border-border px-2 py-1 text-left font-medium">{children}</th>,
  td: ({ children }) => <td className="border border-border px-2 py-1 align-top">{children}</td>,
}

export function TaskConversationPanel({ task }: TaskConversationPanelProps) {
  const { modeConfig, setModeConfig } = useTask()
  const [conversationId, setConversationId] = useState<string | null>(null)
  const [turns, setTurns] = useState<ConversationTurn[]>([])
  const [input, setInput] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [resolving, setResolving] = useState(false)
  const [activeTaskId, setActiveTaskId] = useState<string>('')
  const [activeTaskMode, setActiveTaskMode] = useState<ModeKey>('flash')
  const [liveOutput, setLiveOutput] = useState('')
  const [stepEvents, setStepEvents] = useState<ProgressEvent[]>([])
  const [hideSteps, setHideSteps] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const modeOptions = useMemo(() => ([
    { key: 'flash' as const, label: '闪速', icon: Zap, modeConfig: { thinking: false, plan: false, subagent: false } },
    { key: 'standard' as const, label: '思考', icon: Brain, modeConfig: { thinking: true, plan: false, subagent: false } },
    { key: 'pro' as const, label: 'Pro', icon: GraduationCap, modeConfig: { thinking: false, plan: true, subagent: false } },
    { key: 'ultra' as const, label: 'Ultra', icon: Rocket, modeConfig: { thinking: false, plan: false, subagent: true } },
  ]), [])

  const currentMode = calculateMode(modeConfig) as ModeKey
  const currentModeOption = modeOptions.find((item) => item.key === currentMode) ?? modeOptions[0]

  const refreshTurns = useCallback(async (convId: string) => {
    const detail = await conversationApi.getDetail(convId)
    setTurns(detail.turns)
  }, [])

  const resolveConversationId = useCallback(async () => {
    if (conversationId) return conversationId
    setResolving(true)
    try {
      const created = await conversationApi.create({ title: task.user_input.slice(0, 20) || '任务对话' })
      const convId = created.conversation_id
      setConversationId(convId)
      setTurns([])
      return convId
    } finally {
      setResolving(false)
    }
  }, [conversationId, task.user_input])

  useEffect(() => {
    setConversationId(null)
    setTurns([])
    setInput('')
    setActiveTaskId('')
    setSubmitting(false)
    setLiveOutput('')
    setStepEvents([])
    setHideSteps(false)
    setError(null)
  }, [task.task_id])

  useEffect(() => {
    let cancelled = false

    const loadExistingConversation = async () => {
      setResolving(true)
      try {
        const convId = await conversationApi.findConversationIdByTask(task.task_id)
        if (!convId || cancelled) return
        setConversationId(convId)
        await refreshTurns(convId)
      } finally {
        if (!cancelled) {
          setResolving(false)
        }
      }
    }

    void loadExistingConversation()

    return () => {
      cancelled = true
    }
  }, [task.task_id, refreshTurns])

  const handleTaskComplete = useCallback(async () => {
    if (!conversationId) return
    setSubmitting(false)
    setActiveTaskId('')
    await refreshTurns(conversationId)
  }, [conversationId, refreshTurns])

  useTaskStream({
    taskId: activeTaskId,
    enabled: !!activeTaskId,
    onTaskComplete: () => {
      void handleTaskComplete()
    },
    onEvent: (event) => {
      setStepEvents((prev) => [...prev, event])
      if (event.type !== 'sub_agent_update' || !event.latest) return
      if (activeTaskMode === 'flash' && event.node === 'coordinator') {
        setLiveOutput((prev) => prev + event.latest)
        return
      }
      if (activeTaskMode !== 'flash' && event.node === 'reporter') {
        setLiveOutput((prev) => prev + event.latest)
      }
    },
  })

  const displayStepEvents = useMemo(() => {
    const result: ProgressEvent[] = []
    let lastMessage = ''
    for (const event of stepEvents) {
      const message = getStepMessage(event)
      if (message === lastMessage) {
        continue
      }
      result.push(event)
      lastMessage = message
    }
    return result
  }, [stepEvents])

  const normalizedLiveOutput = useMemo(() => normalizeStreamingMarkdown(liveOutput), [liveOutput])

  const handleModeSelect = (mode: ModeKey) => {
    const target = modeOptions.find((item) => item.key === mode)
    if (!target) return
    setModeConfig(target.modeConfig)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!input.trim()) return
    setError(null)
    try {
      const convId = await resolveConversationId()
      const response = await conversationApi.run(convId, {
        user_input: input.trim(),
        thinking_enabled: modeConfig.thinking,
        plan_enabled: modeConfig.plan,
        subagent_enabled: modeConfig.subagent,
      })
      setInput('')
      setLiveOutput('')
      setStepEvents([])
      setSubmitting(true)
      setActiveTaskId(response.task_id)
      setActiveTaskMode((response.mode as ModeKey) || currentMode)
    } catch (err) {
      setSubmitting(false)
      setActiveTaskId('')
      setError(err instanceof Error ? err.message : '发送失败')
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <MessageSquare className="h-4 w-4 text-primary" />
          多轮对话
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="max-h-[420px] space-y-3 overflow-y-auto rounded-lg border border-border/60 bg-background p-4">
          {turns.length === 0 && !liveOutput && (
            <p className="text-sm text-muted-foreground">暂无对话内容，输入问题开始多轮对话。</p>
          )}
          {turns.map((turn) => (
            <div key={turn.id} className={turn.role === 'user' ? 'text-right' : ''}>
              <div className={`inline-block max-w-[90%] rounded-2xl px-3 py-2 text-sm whitespace-pre-wrap ${
                turn.role === 'user'
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-card border border-border/60 text-foreground'
              }`}>
                {turn.role === 'assistant' ? (
                  <ReactMarkdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
                    {normalizeStreamingMarkdown(turn.full_output || turn.content)}
                  </ReactMarkdown>
                ) : turn.content}
              </div>
            </div>
          ))}
          {liveOutput && (
            <div>
              <div className="inline-block max-w-[90%] rounded-2xl border border-border/60 bg-card px-3 py-2 text-sm whitespace-pre-wrap text-foreground">
                <ReactMarkdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
                  {normalizedLiveOutput}
                </ReactMarkdown>
              </div>
            </div>
          )}
        </div>

        {(submitting || displayStepEvents.length > 0) && (
          <div className="rounded-lg border border-border/60 bg-background p-3">
            <div className="mb-2 flex items-center justify-between">
              <Button variant="ghost" size="sm" className="h-7 px-1 text-muted-foreground" onClick={() => setHideSteps((prev) => !prev)}>
                {hideSteps ? <ChevronRight className="mr-1 h-4 w-4" /> : <ChevronDown className="mr-1 h-4 w-4" />}
                {hideSteps ? '显示步骤' : '隐藏步骤'}
              </Button>
              <span className="text-xs text-emerald-600">已连接</span>
            </div>
            {!hideSteps && (
              <div className="space-y-1">
                {displayStepEvents.map((event, index) => (
                  <div key={`${event.type}-${index}`} className="flex items-center gap-2 text-sm text-muted-foreground">
                    <CheckCircle2 className="h-3.5 w-3.5 text-primary" />
                    <span>{getStepMessage(event)}</span>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-3">
          <div className="rounded-[24px] border border-border/80 bg-card px-5 py-4 shadow-[0_2px_8px_rgba(0,0,0,0.04)]">
            <Textarea
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder={resolving ? '初始化对话中...' : '继续追问，支持多轮上下文...'}
              rows={3}
              disabled={resolving || submitting}
              className="w-full resize-none border-none bg-transparent px-1 py-2 text-[15px] leading-7 text-foreground placeholder:text-muted-foreground focus:outline-none"
            />
            <div className="mt-2 flex items-center justify-between border-t border-border/60 pt-3">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button type="button" variant="ghost" className="h-8 rounded-full px-3 text-sm font-medium text-foreground hover:bg-muted">
                    <currentModeOption.icon className="mr-1 h-4 w-4" />
                    {currentModeOption.label}
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start" className="w-60 p-2">
                  <DropdownMenuLabel className="text-xs text-muted-foreground">执行模式</DropdownMenuLabel>
                  {modeOptions.map((item) => (
                    <DropdownMenuItem key={item.key} onSelect={() => handleModeSelect(item.key)} className={`rounded-lg p-2 ${currentMode === item.key ? 'bg-muted' : ''}`}>
                      <item.icon className="mr-2 h-4 w-4" />
                      {item.label}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
              <Button type="submit" disabled={!input.trim() || resolving || submitting} size="icon" variant="outline" className="h-9 w-9 rounded-full border-border/80 bg-card hover:bg-muted">
                {submitting ? <Loader2 className="h-4 w-4 animate-spin" /> : <ArrowUp className="h-4 w-4" />}
              </Button>
            </div>
          </div>
        </form>
        {error && <p className="text-sm text-destructive">{error}</p>}
      </CardContent>
    </Card>
  )
}
