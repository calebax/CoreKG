import { FC, PropsWithChildren } from 'react'
import { Typography } from 'antd'

type Title = {
  image: string
  name: string
  desc?: string
  isDocument?: boolean
}
/** 用于文档、应用等 头像+名称 */
export const Title: FC<Title> = (props) => {
  const { image, name, desc, isDocument } = props
  return (
    <div
      className='flex gap-3'
      style={{ alignItems: isDocument ? 'center' : 'center' }}
    >
      <img
        loading='lazy'
        src={image}
        className='w-14 h-14 rounded-md flex-shrink-0'
      />
      <div
        className='flex flex-col min-w-0 justify-center'
        style={isDocument ? { minHeight: '52px' } : { minHeight: '52px' }}
      >
        <span
          className='text-lg font-medium text-[#1E1F28] line-clamp-1 leading-[26px] py-2'
          dangerouslySetInnerHTML={{ __html: name }}
        ></span>
        {desc ? (
          <div
            dangerouslySetInnerHTML={{
              __html: desc,
            }}
            className='text-base line-clamp-1 text-[#616373] font-normal leading-[30px] h-[34px] [&_span]:!leading-[30px] [&_span]:!box-border [&_span]:!align-top'
            style={{
              display: '-webkit-box',
              WebkitLineClamp: 1,
              WebkitBoxOrient: 'vertical' as const,
              lineHeight: '30px',
              overflow: 'hidden',
            }}
          ></div>
        ) : null}
      </div>
    </div>
  )
}

type PicTitle = {
  image: string
  name: string
}
/** 图片和视频 大图片+下方名称 */
export const PicTitle: FC<PropsWithChildren<PicTitle>> = (props) => {
  const { image, name, children } = props
  return (
    <div className='flex flex-col gap-3 min-w-0'>
      <div className='aspect-[4/3] relative overflow-hidden bg-gray-100 rounded border border-[#D7D9E5] hover:shadow-lg transition-all duration-300'>
        <img
          src={image}
          className='object-cover w-full h-full rounded hover:scale-120 transition-transform duration-500'
        />
        <div className='z-10 absolute bottom-[7px] right-0'>{children}</div>
      </div>
      <Typography.Paragraph
        className='m-0 text-lg text-[#1e1f28] font-medium break-words !fontFamily-pingFangSC'
        style={{
          lineHeight: '24px',
          paddingBottom: '4px',
        }}
        ellipsis={{ rows: 2, tooltip: name.replace(/<[^>]*>/g, '') }}
      >
        <span dangerouslySetInnerHTML={{ __html: name }} />
      </Typography.Paragraph>
    </div>
  )
}
