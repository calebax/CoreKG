export default function BaseInfo(props: any) {
  const { agentDetail } = props
  return (
    <div className='flex-none w-full py-2.5'>
      <div className='flex items-center gap-1'>
        <img
          src={agentDetail.avatar_url}
          alt='avatar'
          className='flex-none w-12.5 h-12.5 rounded-full object-cover'
        />
        <h1 className='text-title font-medium'>{agentDetail.show_name}</h1>
      </div>
      <h2 className='mt-2 text-description text-sm'>
        {agentDetail.description}
      </h2>
    </div>
  )
}
