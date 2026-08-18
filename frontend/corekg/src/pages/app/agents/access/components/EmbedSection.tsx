import { FC } from 'react'
import { Button } from 'antd'
import IframeIcon from '@/assets/svgs/iframe.svg?react'

interface EmbedSectionProps {
  onViewCode: () => void
}

const EmbedSection: FC<EmbedSectionProps> = ({ onViewCode }) => {
  return (
    <div>
      <div className='flex items-center justify-start mb-4'>
        <span className='text-gray-600'>
          支持嵌入第三方页面，如下图样式（父级组件全屏）。
        </span>
        <button
          onClick={onViewCode}
          className='text-[#0C99FF] hover:text-blue-600 text-sm cursor-pointer border-none bg-transparent p-0 font-bold'
        >
          查看嵌入代码
        </button>
      </div>

      {/* 预览区域 */}
      <div className='relative mb-4'>
        <IframeIcon />

        {/* 访问限制按钮 */}
        {/* <div className='absolute top-4 right-4'>
          <Button size='small' className='bg-white border border-gray-300'>
            访问限制
          </Button>
        </div> */}
      </div>
    </div>
  )
}

export default EmbedSection
