import React, { useState } from 'react'
import { Play, Loader2 } from 'lucide-react'
import { useTask } from '@/hooks'

interface TaskRunFormProps {
  onSuccess?: (taskId: string) => void
}

/**
 * 任务运行表单
 */
export function TaskRunForm({ onSuccess }: TaskRunFormProps) {
  const [userInput, setUserInput] = useState('')
  const [intention, setIntention] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const { runTask } = useTask()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!userInput.trim()) return

    setLoading(true)
    setError(null)

    try {
      const taskId = await runTask(userInput, intention || undefined)
      setUserInput('')
      setIntention('')
      onSuccess?.(taskId)
    } catch (err) {
      setError(err instanceof Error ? err.message : '提交失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {/* 用户输入 */}
      <div>
        <label className="mb-2 block text-sm font-medium text-gray-700">
          输入任务描述
        </label>
        <textarea
          value={userInput}
          onChange={(e) => setUserInput(e.target.value)}
          placeholder="请描述您想要完成的任务..."
          rows={4}
          className="w-full rounded-lg border border-gray-300 p-3 text-sm placeholder-gray-400 focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
        />
      </div>

      {/* 意图分析（可选） */}
      <div>
        <label className="mb-2 block text-sm font-medium text-gray-700">
          意图分析 <span className="text-gray-400">(可选)</span>
        </label>
        <input
          type="text"
          value={intention}
          onChange={(e) => setIntention(e.target.value)}
          placeholder="指定任务的具体意图..."
          className="w-full rounded-lg border border-gray-300 p-3 text-sm placeholder-gray-400 focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
        />
      </div>

      {/* 错误提示 */}
      {error && (
        <div className="rounded-lg bg-red-50 p-3 text-sm text-red-600">
          {error}
        </div>
      )}

      {/* 提交按钮 */}
      <button
        type="submit"
        disabled={loading || !userInput.trim()}
        className="flex w-full items-center justify-center gap-2 rounded-lg bg-indigo-600 px-4 py-3 text-sm font-medium text-white hover:bg-indigo-700 disabled:cursor-not-allowed disabled:opacity-50"
      >
        {loading ? (
          <>
            <Loader2 className="h-4 w-4 animate-spin" />
            运行中...
          </>
        ) : (
          <>
            <Play className="h-4 w-4" />
            开始运行
          </>
        )}
      </button>
    </form>
  )
}
