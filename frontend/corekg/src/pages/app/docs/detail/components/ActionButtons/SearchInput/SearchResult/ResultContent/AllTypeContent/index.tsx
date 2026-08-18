import { FC, Fragment, PropsWithChildren } from 'react'
import { Button, Divider } from 'antd'
import { ArrowRightOutlined } from '@ant-design/icons'
import { getFileURL } from '@/utils/Forest'
import { ResultContent } from '..'
import { FileInSearchResult, ResultType } from '../../..'
import word from '../../../images/word.png'

export const AllTypeContent: FC<ResultContent> = (props) => {
  const { setType } = props
  const value: Record<string, FileInSearchResult[] | undefined> =
    props.value ?? {}
  const { doc_search_result, image_search_result, video_search_result } = value
  if (!doc_search_result && !image_search_result && !video_search_result) {
    return '暂无数据'
  }
  const items: { value: any; title: string; type: ResultType }[] = [
    { value: doc_search_result, title: '文档', type: 'doc' },
    { value: image_search_result, title: '图片', type: 'image' },
    { value: video_search_result, title: '视频', type: 'video' },
  ]
  return (
    <div className='flex flex-col'>
      {items.map((item) => {
        const { value, title, type } = item
        const Comp = getCompByType(type)
        if (!Comp || !Array.isArray(value) || !value.length) return null
        return (
          <Fragment key={type}>
            <span className='text-[#165DFF]'>{title}</span>
            <Comp value={value}>
              <Button
                type='text'
                size='small'
                icon={<ArrowRightOutlined />}
                className='opacity-50 self-start'
                onClick={() => setType(type)}
              >
                查看更多
              </Button>
            </Comp>
          </Fragment>
        )
      })}
    </div>
  )
}

function getCompByType(type: ResultType) {
  if (type === 'doc') {
    return DocContentInAny
  }
  if (type === 'image') {
    return ImageContentInAny
  }
  if (type === 'video') {
    return VideoContentInAny
  }
}

type Title = {
  src?: string
  title: string
}
const Title: FC<Title> = (props) => {
  const { src, title } = props
  return (
    <span className='flex gap-2 items-center'>
      {src ? <img loading='lazy' src={src} className=' w-12 h-12' /> : null}
      <span
        dangerouslySetInnerHTML={{ __html: title }}
        className='text-base'
      ></span>
    </span>
  )
}

type Content = PropsWithChildren & {
  value: FileInSearchResult[]
}
const DocContentInAny: FC<Content> = (props) => {
  const { value, children } = props
  return (
    <div className='flex flex-col gap-2'>
      {value.map((file) => {
        const { forest_id, highlighted_file_name, highlights, id } = file
        const url = getFileURL(forest_id, id)
        return (
          <div className=' p-2 bg-[#F7F7F7] flex flex-col' key={id}>
            <Link to={url} target='_blank' className='text-[unset]'>
              <Title src={word} title={highlighted_file_name}></Title>
            </Link>
            <Divider className='m-2' />
            {highlights.map((item) => {
              const { highlighted_description, location } = item
              const urlWithFileLocation = getFileURL(forest_id, id, location)
              return (
                <Fragment key={highlighted_description}>
                  <Link
                    to={urlWithFileLocation}
                    target='_blank'
                    className='text-[unset]'
                  >
                    <span
                      dangerouslySetInnerHTML={{
                        __html: highlighted_description,
                      }}
                      className='text-sm text-[#758BBE]'
                    ></span>
                  </Link>

                  <Divider className='m-2' />
                </Fragment>
              )
            })}

            {children}
          </div>
        )
      })}
    </div>
  )
}

const ImageContentInAny: FC<Content> = (props) => {
  const { value, children } = props

  return (
    <div className=' p-2 bg-[#F7F7F7] flex flex-col'>
      <div className='flex flex-col'>
        {value.map((file) => {
          const { forest_id, highlighted_file_name, highlights, id } = file
          const img = highlights[0].image_url
          const url = getFileURL(forest_id, id)
          return (
            <Fragment key={url}>
              <Link to={url} target='_blank' className=' text-[unset]'>
                <Title src={img} title={highlighted_file_name} />
              </Link>
              <Divider className='m-2' />
            </Fragment>
          )
        })}
      </div>
      {children}
    </div>
  )
}

const VideoContentInAny: FC<Content> = ImageContentInAny
