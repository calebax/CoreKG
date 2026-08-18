import { FC } from 'react'
import { CommonResultItem } from '../../CommonResultItem'

const AgentItem: FC<{ value?: any }> = (props) => {
  const { value } = props
  const {
    id,
    agent_type,
    user_name,
    avatar_url,
    created_at,
    highlighted_agent_name,
    highlighted_description,
  } = value
  const type = agent_type === 'prompt' ? 'prompt' : agent_type === 'role_play' ? 'role' : 'question'
  return (
    <CommonResultItem
      className='w-full'
      type='agent'
      creator={user_name}
      creatorAvatar={avatar_url}
      time={created_at}
      title_html={highlighted_agent_name}
      to={`/agents/detail/${type}/${id}`}
    >
      <div
        className='text-[#3C4149]'
        dangerouslySetInnerHTML={{ __html: highlighted_description }}
      ></div>
    </CommonResultItem>
  )
}
export default AgentItem
