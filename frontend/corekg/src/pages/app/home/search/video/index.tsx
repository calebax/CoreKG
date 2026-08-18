import { FC, Fragment } from 'react'
import { Link } from 'react-router-dom'
import { cn } from '@/utils'
import { getFileURL } from '@/utils/Forest'
import { PicTitle } from '../../components/SearchMode/Title'
import { FileInSearchResult } from '../../components/SearchMode/searchType'
import EmptyState from '../components/EmptyState'
import VideoIcon from './video.svg?react'

type Content = {
  value?: FileInSearchResult[]
  maxItems?: number // 新增：控制最大显示数量
}
export const VideoContent: FC<Content> = (props) => {
  const { value, maxItems } = props
  if (!value || value.length === 0)
    return <EmptyState message='暂未查询到相关内容～' />

  // 将所有highlights扁平化为一个数组，然后限制数量
  const allVideos: Array<{
    image_url: string
    location: any
    forest_id: string
    id: string
    highlighted_file_name: string
  }> = []

  for (const file of value) {
    const { forest_id, highlighted_file_name, highlights, id } = file
    for (const item of highlights!) {
      allVideos.push({
        image_url: item.image_url!,
        location: item.location,
        forest_id,
        id,
        highlighted_file_name: highlighted_file_name!,
      })
    }
  }

  // 根据maxItems限制显示数量
  const displayVideos = maxItems ? allVideos.slice(0, maxItems) : allVideos

  return (
    <div className='grid grid-cols-2 md:grid-cols-4 2xl:grid-cols-6 gap-4'>
      {displayVideos.map((videoData, index) => {
        const { image_url, location, forest_id, id, highlighted_file_name } =
          videoData
        const time = formatSecond(location[1])
        const urlWithFileLocation = getFileURL(forest_id, id, location)
        return (
          <Link
            key={`${id}-${index}`}
            to={urlWithFileLocation}
            target='_blank'
            className='text-[unset]'
          >
            <PicTitle image={image_url} name={highlighted_file_name}>
              <div
                className={cn(
                  'px-2.5 rounded-full bg-[#00000033] text-white',
                  'flex items-center gap-1',
                )}
              >
                <VideoIcon />
                {time}
              </div>
            </PicTitle>
          </Link>
        )
      })}
    </div>
  )
}

const formatSecond = (second: number) => {
  const h = Math.floor(second / 3600)
  const m = Math.floor((second % 3600) / 60)
  const s = second % 60

  const parts: (number | string)[] = []
  if (h > 9) {
    parts.push(h)
  } else if (h > 0) {
    parts.push(`0${h}`)
  }

  if (m > 9) {
    parts.push(m)
  } else {
    parts.push(`0${m}`)
  }

  if (s > 9) {
    parts.push(s)
  } else {
    parts.push(`0${s}`)
  }

  return parts.join(':')
}
