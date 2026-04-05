import React, { useMemo, useState } from 'react'
import { ArrowUp, Loader2, Paperclip, Zap, Brain, GraduationCap, Rocket } from 'lucide-react'
import { useTask } from '@/hooks'
import { calculateMode } from '@/store'
import { Button, DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuTrigger, Textarea } from '@/components/ui'

interface TaskRunFormProps {
  onSuccess?: (taskId: string) => void
}

type ModeKey = 'flash' | 'standard' | 'pro' | 'ultra'

export function TaskRunForm({ onSuccess }: TaskRunFormProps) {
  const [userInput, setUserInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const { runTask, modeConfig, setModeConfig } = useTask()

  const modeOptions = useMemo(() => ([
    {
      key: 'flash' as const,
      label: '闪速',
      description: '快速高效的完成任务，但可能不够精准',
      icon: Zap,
      modeConfig: { thinking: false, plan: false, subagent: false },
    },
    {
      key: 'standard' as const,
      label: '思考',
      description: '思考后再行动，在时间与准确性之间取得平衡',
      icon: Brain,
      modeConfig: { thinking: true, plan: false, subagent: false },
    },
    {
      key: 'pro' as const,
      label: 'Pro',
      description: '思考、计划两步行，获得更精准的结果',
      icon: GraduationCap,
      modeConfig: { thinking: false, plan: true, subagent: false },
    },
    {
      key: 'ultra' as const,
      label: 'Ultra',
      description: '终极 Pro 模式，可调用子代理工作，适合复杂多步骤任务',
      icon: Rocket,
      modeConfig: { thinking: false, plan: false, subagent: true },
    },
  ]), [])

  const currentMode = calculateMode(modeConfig) as ModeKey
  const currentModeOption = modeOptions.find((item) => item.key === currentMode) ?? modeOptions[0]

  const handleModeSelect = (mode: ModeKey) => {
    const target = modeOptions.find((item) => item.key === mode)
    if (!target) return
    setModeConfig(target.modeConfig)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!userInput.trim()) return

    setLoading(true)
    setError(null)

    try {
      const taskId = await runTask(userInput)
      setUserInput('')
      onSuccess?.(taskId)
    } catch (err) {
      setError(err instanceof Error ? err.message : '提交失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-3">
      <div className="rounded-[24px] border border-border/80 bg-card px-5 py-4 shadow-[0_2px_8px_rgba(0,0,0,0.04)]">
        <Textarea
          value={userInput}
          onChange={(e) => setUserInput(e.target.value)}
          placeholder="今天我能为你做些什么？"
          rows={4}
          className="w-full resize-none border-none bg-transparent px-1 py-2 text-[15px] leading-7 text-foreground placeholder:text-muted-foreground focus:outline-none"
        />
        <div className="mt-2 flex items-center justify-between border-t border-border/60 pt-3">
          <div className="flex items-center gap-2">
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="h-8 w-8 rounded-full text-muted-foreground hover:bg-muted"
            >
              <Paperclip className="h-4 w-4" />
            </Button>
            <div className="relative">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button type="button" variant="ghost" className="h-8 rounded-full px-3 text-sm font-medium text-foreground hover:bg-muted">
                    <currentModeOption.icon className="mr-1 h-4 w-4" />
                    {currentModeOption.label}
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start" className="w-80 p-2">
                  <DropdownMenuLabel className="text-xs text-muted-foreground">模式</DropdownMenuLabel>
                  {modeOptions.map((item) => (
                    <DropdownMenuItem
                      key={item.key}
                      onSelect={() => handleModeSelect(item.key)}
                      className={`rounded-lg p-2 ${currentMode === item.key ? 'bg-muted' : ''}`}
                    >
                      <div className="flex items-start gap-2">
                        <item.icon className="mt-0.5 h-4 w-4 shrink-0 text-foreground" />
                        <div>
                          <p className="text-sm font-medium text-foreground">{item.label}</p>
                          <p className="mt-1 text-xs text-muted-foreground">{item.description}</p>
                        </div>
                      </div>
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <span className="text-sm text-muted-foreground">DeepSeek v3.2</span>
            <Button
              type="submit"
              disabled={loading || !userInput.trim()}
              size="icon"
              variant="outline"
              className="h-9 w-9 rounded-full border-border/80 bg-card hover:bg-muted"
            >
              {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <ArrowUp className="h-4 w-4" />}
            </Button>
          </div>
        </div>
      </div>

      {error && (
        <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
          {error}
        </div>
      )}
    </form>
  )
}
