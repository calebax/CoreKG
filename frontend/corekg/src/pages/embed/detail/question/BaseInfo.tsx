export default function BaseInfo({ agentDetail }) {
  return (
    <div className='flex-none w-full py-2.5'>
      <div className='flex items-center gap-1'>
        <img
          src={agentDetail.avatar}
          alt='avatar'
          className='flex-none w-12.5 h-12.5 rounded-full object-cover'
        />
        <div className='flex-grow w-full'>
          <div className='flex items-center gap-1'>
            <h1 className='text-title font-medium'>{agentDetail.showName}</h1>
          </div>
        </div>
      </div>
      <h2 className='mt-2 text-description text-sm'>
        {agentDetail.description}
      </h2>
    </div>
  )
}
