import { FC, Fragment } from 'react'
import { Link } from 'react-router-dom'
import { Divider } from 'antd'
import EmptyState from '../../../../search/components/EmptyState'
import { Title } from '../../Title'
import { FileInSearchResult } from '../../searchType'

type Content = {
  value?: (FileInSearchResult & {
    agent_type: 'role_play' | 'prompt' | 'knowledge'
  })[]
}
const AgentContent: FC<Content> = (props) => {
  const { value } = props
  if (!value || value.length === 0)
    return <EmptyState message='暂未查询到相关内容～' />
  return (
    <>
      {value.map((file, index) => {
        const {
          highlighted_agent_name,
          highlighted_description,
          agent_type,
          id,
          image_url,
        } = file
        const type = agent_type === 'prompt' ? 'prompt' : 'role'
        const url = `/agents/detail/${type}/${id}`
        return (
          <div key={url}>
            <Link
              to={url}
              className='text-[unset] block hover:bg-[#EFF0F6] rounded-[10px] p-2.5 transition-colors'
              target='_blank'
            >
              <Title
                image={image_url!}
                name={highlighted_agent_name!}
                desc={highlighted_description}
              ></Title>
            </Link>
            {/* {index < value.length - 1 && (
              <Divider className='my-3 border-gray-100' />
            )} */}
          </div>
        )
      })}
    </>
  )
}
export default AgentContent
