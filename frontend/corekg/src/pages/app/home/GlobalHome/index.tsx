import { FC, useEffect, useState } from 'react'
import { useLocation } from 'react-router-dom'
import Project from '@/pages/project'

const GlobalHome: FC = () => {
  return <Project useCustomIds={useCustomIds} type='global' />
}

export default GlobalHome

const useCustomIds = () => {
  const location = useLocation()
  const [session_id, setSessionId] = useState<number>()

  useEffect(() => {
    if (location.pathname === '/global') {
      setSessionId(undefined)
    }
  }, [location.pathname])

  useEffect(() => {
    const startGlobalSession = () => setSessionId(undefined)
    window.addEventListener('startGlobalSession', startGlobalSession)
    return () => {
      window.removeEventListener('startGlobalSession', startGlobalSession)
    }
  }, [])

  return {
    project_id: 0,
    session_id,
    setSessionId: (id?: number, stopNavigate?: boolean) => {
      if (id && !stopNavigate) {
        const pathname = `/project/0/${id}`
        window.history.replaceState(null, '', pathname)
        window.dispatchEvent(
          new CustomEvent('globalSessionRouteChange', {
            detail: { pathname, sessionId: id },
          }),
        )
      }
      setSessionId(id)
    },
  }
}
