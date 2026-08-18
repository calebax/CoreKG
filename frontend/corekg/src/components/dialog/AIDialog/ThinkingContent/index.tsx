import { FC } from 'react'
import { DownOutlined, UpOutlined, CheckCircleFilled } from '@ant-design/icons'
import { useBoolean } from 'ahooks'
import DownIcon from '@/assets/icons/arrow-down2.svg?react'
import UpIcon from '@/assets/icons/arrow-up.svg?react'

export type ThinkingContent = {
  loading: boolean
  content: string
}
export const ThinkingContent: FC<ThinkingContent> = (props) => {
  const { content, loading } = props
  const [hidden, { toggle }] = useBoolean(true)
  if (!content) {
    return null
  }

  if (loading) {
    return (
      <div className='bg-[#F8F9FD] rounded-md px-3 py-2'>
        <div className='text-sm text-[#1E1F28] font-medium leading-[22px] mb-3'>
          思考中...
        </div>
        <div className='text-sm text-[#616373] leading-[22px] font-normal'>
          {content}
        </div>
      </div>
    )
  }

  return (
    <div className='bg-[#FCFCFE] rounded-md'>
      <div
        className='flex items-center gap-2 px-3 py-3 cursor-pointer'
        onClick={toggle}
      >
        <div className='flex items-center gap-2'>
          <CheckCircleFilled className='text-[#3B82F6] text-base' />
          <span className='text-sm text-[#3B82F6] font-medium leading-[22px]'>
            思考完成
          </span>
        </div>
        {hidden ? (
          <DownIcon className='w-4 h-4 text-[#3473EC]' />
        ) : (
          <UpIcon className='w-4 h-4 text-[#3473EC]' />
        )}
      </div>
      {!hidden && (
        <div className='relative px-3 pb-3'>
          {/* 与思考完成图标对齐的竖线 */}
          <div className='absolute left-[20px] top-0 bottom-0 w-[1px] bg-[#E6E8F0]'></div>
          {/* 内容，左边留出竖线空间 */}
          <div className='ml-[24px] text-sm text-[#616373] leading-[22px] font-normal'>
            {content}
          </div>
        </div>
      )}
    </div>
  )
}
