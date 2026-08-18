import { useState, useEffect } from 'react'
import { Drawer } from 'antd'
import { Layout } from '@/components/Layout/Layout'
import AgentAccess from '@/pages/app/agents/access'
import useAccessStore from '@/stores/access'
import Sidebar from './components/Sidebar'

export default function AppLayout() {
  const [isDrawerOpen, setIsDrawerOpen] = useState(false)
  const { setAccessIds, clearAccessIds } = useAccessStore()

  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      const data = event.data

      if (!data || typeof data !== 'object') return

      if (data.type === 'OPEN_ACCESS_SETTINGS') {
        setAccessIds(data.botId || '', data.spaceId || '')
        setIsDrawerOpen(true)
      }
    }

    window.addEventListener('message', handleMessage)

    return () => {
      window.removeEventListener('message', handleMessage)
    }
  }, [setAccessIds])

  const handleClose = () => {
    clearAccessIds()
    setIsDrawerOpen(false)
  }

  return (
    <>
      <Layout>
        <Sidebar />
      </Layout>
      <Drawer
        title='访问设置'
        open={isDrawerOpen}
        onClose={handleClose}
        width={926}
        destroyOnClose={true}
      >
        {isDrawerOpen && <AgentAccess />}
      </Drawer>
    </>
  )
}
