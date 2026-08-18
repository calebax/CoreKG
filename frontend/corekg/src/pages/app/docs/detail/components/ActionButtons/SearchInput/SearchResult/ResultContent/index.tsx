import { FC } from 'react'
import { ResultType } from '../..'
import { AllTypeContent } from './AllTypeContent'
import { DocContent } from './DocContent'
import { ImageContent } from './ImageContent'
import { VideoContent } from './VideoContent'

export type ResultContent = {
  type: ResultType
  setType: (type: ResultType) => void
  value: any
}
export const ResultContent: FC<ResultContent> = (props) => {
  const { type } = props
  const Comp = getCompByType(type)
  return (
    <div className='max-h-[60vh] overflow-auto pr-2'>
      <Comp {...props} />
    </div>
  )
}

function getCompByType(type: ResultType): FC<ResultContent> {
  if (type === 'all') {
    return AllTypeContent
  }
  if (type === 'doc') {
    return DocContent
  }
  if (type === 'image') {
    return ImageContent
  }
  if (type === 'video') {
    return VideoContent
  }
  throw new Error(`'${type}'是不支持的类型`)
}
