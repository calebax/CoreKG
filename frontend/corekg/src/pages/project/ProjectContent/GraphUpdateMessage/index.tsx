import { FC } from 'react'
import { useBoolean, useRequest } from 'ahooks'
import { cn } from '@/utils'
import { getFileInfo, getKnowledgeBaseDetail } from '@/api/knowledge'
import { SessionInfo } from '..'
import ArrowRight from './images/arrow-right.svg?react'
import Close from './images/close.svg?react'

export const GraphUpdateMessage: FC<
  Pick<SessionInfo, 'sessionStatus' | 'sessionConfig'>
> = (props) => {
  const { sessionStatus, sessionConfig } = props
  const [open, { toggle }] = useBoolean(true)
  const firstFileId = sessionConfig?.graphKnowledgeBase?.[0]?.id as
    | number
    | undefined
  const showMessage =
    open &&
    (sessionStatus === 'creating' || sessionStatus === 'created') &&
    firstFileId

  const { loading, data } = useRequest(
    async () => {
      const {
        Forest: { ID },
      } = await getFileInfo({ file_id: firstFileId! })
      const {
        data: { graph_status },
        graph_info: { ID: graph_id },
      } = await getKnowledgeBaseDetail({ id: ID })
      return {
        graph_id,
        graph_status,
      }
    },
    {
      ready: Boolean(showMessage),
      refreshDeps: [firstFileId],
    },
  )

  if (loading || !showMessage || !data || data.graph_status !== 'updatable') {
    return null
  }
  return (
    <div
      className={cn(
        'p-2 mt-4 mx-8 flex items-center gap-1',
        'rounded-[6px] bg-[#9194971A] ',
      )}
    >
      提示：检测到文档更新，图谱可能失效。为了确保使用最新的知识结构，建议更新图谱。
      <Link
        to={`/graph/edit?graphId=${data?.graph_id}`}
        target='_blank'
        className='flex items-center gap-1 text-[#0C99FF]'
      >
        立即更新
        <ArrowRight />
      </Link>
      <Close onClick={toggle} className='ml-auto cursor-pointer' />
    </div>
  )
}
