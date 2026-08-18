import { useMemoizedFn } from 'ahooks'

/**
 * 获取和设置查询参数session_id\
 * 提供一个key用于重置相关组件\
 * 也可以使用其他查询参数获取session_id
 */
export const useSessionId = (sessionIdKey: string = 'session_id') => {
  const [searchParams, setSearchParams] = useSearchParams()
  const originSessionId = useMemo(() => {
    const _session_id = parseInt(searchParams.get(sessionIdKey)!)
    if (Number.isInteger(_session_id)) return _session_id
    return undefined
  }, [searchParams, sessionIdKey])
  const [{ sessionId, key }, setSessionInfo] = useState({
    sessionId: originSessionId,
    key: 0,
  })

  // 如果下一次的sessionId与此值相同 则不会变化key
  const stableSessionId = useRef<number | undefined>(NaN)
  useEffect(() => {
    setSessionInfo((prev) => {
      const { key } = prev
      const shouldKeyChange = stableSessionId.current !== originSessionId
      stableSessionId.current = NaN
      return {
        sessionId: originSessionId,
        key: shouldKeyChange ? key + 1 : key,
      }
    })
  }, [originSessionId])

  const setSessionId = useMemoizedFn(
    (newSessionId: number | undefined, stable?: boolean) => {
      const checkedNewSessionId = (() => {
        if (newSessionId === undefined || Number.isInteger(newSessionId)) {
          return newSessionId
        }
        return undefined
      })()

      if (stable) {
        stableSessionId.current = checkedNewSessionId
      }

      setSearchParams((prev) => {
        const newSearchParams = new URLSearchParams(prev)
        if (checkedNewSessionId === undefined) {
          newSearchParams.delete(sessionIdKey)
        } else {
          newSearchParams.set(sessionIdKey, String(checkedNewSessionId))
        }
        return newSearchParams
      })
    },
  )
  return {
    sessionId,
    setSessionId,
    key,
  }
}
