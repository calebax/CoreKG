import { Spin } from 'antd'
import { LoadingOutlined, HighlightOutlined } from '@ant-design/icons'
import { cn } from '@/utils'

interface AIPolishButtonProps {
  disabled: boolean
  loading: boolean
  onClick: () => void
  className?: string
}

export const AIPolishButton = ({
  disabled,
  loading,
  onClick,
  className,
}: AIPolishButtonProps) => {
  const isDisabledOrLoading = disabled || loading

  return (
    <div
      className={cn(
        'cursor-pointer border rounded-full',
        'text-[13px]',
        'py-1 px-3 flex items-center gap-1 font-[500]',
        'transition-colors',
        // 激活状态（紫色）
        !isDisabledOrLoading &&
          'bg-[#FBE9FF] text-[#CC5DE8] border-[#CC5DE833] hover:bg-[#f5d8ff] hover:text-[#B34DD8]',
        // 禁用状态（灰色）
        isDisabledOrLoading &&
          'bg-[#f7f7f7] text-[#6e757f] border-[#00000033] cursor-not-allowed opacity-60',
        className,
      )}
      onClick={() => {
        if (!isDisabledOrLoading) {
          onClick()
        }
      }}
    >
      {loading ? (
        <Spin
          indicator={<LoadingOutlined spin />}
          size='small'
          className='w-[15.368px] h-[15.368px]'
        />
      ) : (
        <HighlightOutlined
          className='w-[15.368px] h-[15.368px]'
          style={{ color: 'currentColor' }}
        />
      )}
      <span>AI润色</span>
    </div>
  )
}
