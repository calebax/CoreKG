import { getAvatar } from '@/pages/app/agents/utils/getAvatar'

export default function BaseInfo(props: any) {
  const { agentDetail } = props
  console.log('agentDetail', agentDetail)
  return (
    <div className='flex-none w-full py-2.5'>
      <div className='flex items-center gap-1'>
        <img
          src={getAvatar(
            agentDetail.agent_info.avatar_url,
            agentDetail.agent_info.agent_type,
          )}
          alt='avatar'
          className='flex-none w-12.5 h-12.5 rounded-full object-cover'
        />
        <h1 className='text-title font-medium'>
          {agentDetail.agent_info.show_name}
        </h1>
      </div>
      <h2 className='mt-2 text-description text-sm'>
        {agentDetail.description}
      </h2>
    </div>
  )
}
