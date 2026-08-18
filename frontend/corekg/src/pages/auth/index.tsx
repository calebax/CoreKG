import { useState } from 'react'
import { cn } from '@/utils'
import ApplyAuth from './applyAuth'
import ShowInfo from './showInfo'

export default function Auth() {
  const [refreshTrigger, setRefreshTrigger] = useState(0)

  const handleAuthSuccess = () => {
    // 触发 ShowInfo 组件刷新
    setRefreshTrigger((prev) => prev + 1)
  }

  return (
    <div
      className={cn(
        'max-h-full pt-20 pb-4 overflow-auto',
        'flex flex-col items-center',
      )}
    >
      <div className='left-0 text-2xl font-semibold mb-4'>设备授权激活</div>
      <div className='flex gap-4 mt-4'>
        <ApplyAuth onAuthSuccess={handleAuthSuccess} />
        <ShowInfo refreshTrigger={refreshTrigger} />
      </div>
    </div>
  )
}
