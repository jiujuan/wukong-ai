import { useEffect, useRef, useCallback } from 'react'

interface UsePollingOptions {
  enabled?: boolean
  interval?: number
  onPoll: () => void | Promise<void>
}

/**
 * 轮询 Hook
 * @param options.enabled - 是否启用轮询
 * @param options.interval - 轮询间隔（毫秒）
 * @param options.onPoll - 轮询回调函数
 */
export function usePolling({ enabled = true, interval = 2000, onPoll }: UsePollingOptions) {
  const intervalRef = useRef<NodeJS.Timeout | null>(null)
  const onPollRef = useRef(onPoll)

  // 更新回调引用
  useEffect(() => {
    onPollRef.current = onPoll
  }, [onPoll])

  // 开始/停止轮询
  useEffect(() => {
    if (!enabled) {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
        intervalRef.current = null
      }
      return
    }

    // 立即执行一次
    onPollRef.current()

    // 设置定时器
    intervalRef.current = setInterval(() => {
      onPollRef.current()
    }, interval)

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
        intervalRef.current = null
      }
    }
  }, [enabled, interval])

  const stop = useCallback(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current)
      intervalRef.current = null
    }
  }, [])

  const start = useCallback(() => {
    if (!intervalRef.current) {
      onPollRef.current()
      intervalRef.current = setInterval(() => {
        onPollRef.current()
      }, interval)
    }
  }, [interval])

  return { stop, start }
}
