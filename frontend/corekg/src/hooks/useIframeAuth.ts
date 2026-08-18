import { useState, useEffect } from 'react'
import { useParams, useSearchParams } from 'react-router-dom'
import { useLocalStorageState } from 'ahooks'
import { iframeAuth } from '@/api/iframe'

const generateClientId = (): string => {
  return Math.random().toString(36).substring(2, 12)
}

export const useIframeAuth = () => {
  const { id } = useParams<{ id: string }>()
  const [searchParams] = useSearchParams()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [clientId, setClientId] =
    useLocalStorageState<string>('iframe_client_id')

  useEffect(() => {
    if (!clientId) {
      const newClientId = generateClientId()
      setClientId(newClientId)
    }
  }, [])

  const regenerateClientId = () => {
    const newClientId = generateClientId()
    setClientId(newClientId)
    iframeAuth.setClientId(newClientId)
    return newClientId
  }

  const clearAuth = () => {
    iframeAuth.clear()
    setClientId('')
    setError(null)
  }

  useEffect(() => {
    const initAuth = async () => {
      try {
        setLoading(true)
        const agentId = searchParams.get('access_token') || id

        if (!agentId) {
          throw new Error('缺少 agent_id')
        }

        if (!clientId) {
          throw new Error('clientId 生成失败')
        }

        await iframeAuth.init(agentId, clientId)
        setError(null)
      } catch (err) {
        console.error('iframe认证失败:', err)
        setError(err instanceof Error ? err.message : '认证失败')
      } finally {
        setLoading(false)
      }
    }

    if (clientId) {
      initAuth()
    }
  }, [id, searchParams, clientId])

  return {
    loading,
    error,
    authenticated: iframeAuth.isAuthenticated(),
    clientId: clientId || '',
    token: iframeAuth.getToken(),
    regenerateClientId,
    clearAuth,
  }
}
