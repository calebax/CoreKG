import { FC, Fragment } from 'react'
import { Link } from 'react-router-dom'
import { getFileURL } from '@/utils/Forest'
import { PicTitle } from '../../components/SearchMode/Title'
import { FileInSearchResult } from '../../components/SearchMode/searchType'
import EmptyState from '../components/EmptyState'

type Content = {
  value?: FileInSearchResult[]
  maxItems?: number // 新增：控制最大显示数量
}
export const ImageContent: FC<Content> = (props) => {
  const { value, maxItems } = props
  if (!value || value.length === 0)
    return <EmptyState message='暂未查询到相关内容～' />

  // 将所有highlights扁平化为一个数组，然后限制数量
  const allImages: Array<{
    image_url: string
    location: any
    forest_id: string
    id: string
    highlighted_file_name: string
  }> = []

  for (const file of value) {
    const { forest_id, highlighted_file_name, highlights, id } = file
    for (const item of highlights!) {
      allImages.push({
        image_url: item.image_url!,
        location: item.location,
        forest_id,
        id,
        highlighted_file_name: highlighted_file_name!,
      })
    }
  }

  // 根据maxItems限制显示数量
  const displayImages = maxItems ? allImages.slice(0, maxItems) : allImages

  return (
    <div className='grid grid-cols-2 md:grid-cols-4 2xl:grid-cols-6 gap-4'>
      {displayImages.map((imageData, index) => {
        const { image_url, location, forest_id, id, highlighted_file_name } =
          imageData
        const urlWithFileLocation = getFileURL(forest_id, id, location)
        return (
          <Link
            key={`${id}-${index}`}
            to={urlWithFileLocation}
            target='_blank'
            className='text-[unset] hover:opacity-80 transition-opacity'
          >
            <PicTitle image={image_url} name={highlighted_file_name} />
          </Link>
        )
      })}
    </div>
  )
}
