import { FC, ReactNode } from 'react'
import { Outlet } from 'react-router-dom'
import { ConfigProvider } from 'antd'

interface EmbedLayoutProps {
  children?: ReactNode
}

const EmbedLayout: FC<EmbedLayoutProps> = ({ children }) => {
  return (
    <ConfigProvider
      theme={{
        token: {
          colorPrimary: '#1890ff',
        },
      }}
    >
      <div className='h-screen w-full overflow-hidden bg-white'>
        {children || <Outlet />}
      </div>
    </ConfigProvider>
  )
}

export default EmbedLayout
