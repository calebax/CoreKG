import { FC, PropsWithChildren, ReactNode } from 'react'
import { Outlet } from 'react-router-dom'

// import Bg from '/images/bg.jpg'

type Layout = PropsWithChildren & {
  header?: ReactNode
}
export const Layout: FC<Layout> = (props) => {
  const { children } = props
  return (
    <div className='w-full h-full flex overflow-hidden'>
      {children}
      <div
        className='flex-1 overflow-hidden'
        style={{
          backgroundColor: '#ffffff',
        }}
      >
        {/* <div className='w-full h-full overflow-hidden relative rounded'> */}
        <Outlet />
        {/* </div> */}
      </div>
    </div>
  )
}
