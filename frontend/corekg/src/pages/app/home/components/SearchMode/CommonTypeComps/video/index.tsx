import { FC, Fragment } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import { getFileURL } from '@/utils/Forest'
import EmptyState from '../../../../search/components/EmptyState'
import { PicTitle } from '../../Title'
import { FileInSearchResult } from '../../searchType'
import VideoIcon from './video.svg?react'

type Content = {
  value?: FileInSearchResult[]
}
const VideoContent: FC<Content> = (props) => {
  const { t: tC } = useTranslation('common')
  const { value } = props
  if (!value || value.length === 0)
    return <EmptyState message={tC('empty.noFind')} />
  return (
    <div className='grid grid-cols-4 gap-3'>
      {value.map((file) => {
        const { forest_id, highlighted_file_name, highlights, id } = file
        return (
          <Fragment key={id}>
            {highlights!.map((item) => {
              const { image_url, location } = item
              const time = formatSecond(location[1])
              const urlWithFileLocation = getFileURL(forest_id, id, location)
              return (
                <Link
                  key={image_url}
                  to={urlWithFileLocation}
                  target='_blank'
                  className='text-[unset] hover:opacity-80 transition-opacity'
                >
                  <PicTitle image={image_url!} name={highlighted_file_name!}>
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
          </Fragment>
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

export default VideoContent
