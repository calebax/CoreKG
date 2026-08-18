import { useEffect, useRef, useState } from 'react'

interface UsePollingProps {
  callback: () => void
  interval: number
  enabled: boolean
  dependencies?: any[]
}

export const usePolling = ({
  callback,
  interval,
  enabled,
  dependencies = [],
}: UsePollingProps) => {
  const pollingTimerRef = useRef<number | null>(null)
  const [pollingEnabled, setPollingEnabled] = useState<boolean>(enabled)

  // 设置轮询机制
  useEffect(() => {
    if (pollingEnabled) {
      pollingTimerRef.current = window.setInterval(() => {
        callback()
      }, interval)
    }

    // 清理函数 - 组件卸载时清除定时器
    return () => {
      if (pollingTimerRef.current !== null) {
        clearInterval(pollingTimerRef.current)
        pollingTimerRef.current = null
      }
    }
  }, [pollingEnabled, interval, ...dependencies])

  return {
    pollingEnabled,
    setPollingEnabled,
  }
}
