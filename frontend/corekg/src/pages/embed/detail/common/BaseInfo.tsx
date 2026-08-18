import { Skeleton } from 'antd'

export default function BaseInfo({ loading = false, agentDetail }) {
  if (loading) {
    return (
      <div className='flex-none w-full py-2.5'>
        <div className='flex items-center gap-1 overflow-hidden'>
          <Skeleton.Avatar shape='circle' size={50} className='flex-none' />
          <div className='flex-grow'>
            <Skeleton.Input active className='block! w-full! h-6!' />
          </div>
        </div>
        <div className='w-full mt-2'>
          <Skeleton.Input active className='block! w-full! h-5!' />
        </div>
      </div>
    )
  }

  return (
    <div className='flex-none w-full py-2.5'>
      <div className='flex items-center gap-1'>
        <img
          src={agentDetail.avatar}
          alt='avatar'
          className='flex-none w-12.5 h-12.5 rounded-full object-cover'
        />
        <h1 className='text-title font-medium'>{agentDetail.showName}</h1>
      </div>
      <h2 className='mt-2 text-description text-sm'>
        {agentDetail.description}
      </h2>
    </div>
  )
}
