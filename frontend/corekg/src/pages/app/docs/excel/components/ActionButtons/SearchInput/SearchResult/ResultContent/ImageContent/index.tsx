import { FC, Fragment } from 'react'
import { useTranslation } from 'react-i18next'
import { getFileURL } from '@/utils/Forest'
import { FileInSearchResult } from '../../..'

type Content = {
  value?: FileInSearchResult[]
}
export const ImageContent: FC<Content> = (props) => {
  const { t: tC } = useTranslation('common')
  const { value } = props
  if (!value || value.length === 0) return tC('empty.noData')
  return (
    <div className='flex flex-wrap gap-3'>
      {value.map((file) => {
        const { forest_id, highlighted_file_name, highlights, id } = file
        return (
          <Fragment key={id}>
            {highlights.map((item) => {
              const { image_url, location } = item
              const urlWithFileLocation = getFileURL(forest_id, id, location)
              return (
                <Link
                  key={image_url}
                  to={urlWithFileLocation}
                  target='_blank'
                  className='w-40 flex flex-col text-[unset]'
                >
                  <img
                    src={image_url}
                    className='w-40 h-40 object-scale-down'
                  />
                  <span
                    dangerouslySetInnerHTML={{ __html: highlighted_file_name }}
                    className=' whitespace-pre-wrap'
                  ></span>
                </Link>
              )
            })}
          </Fragment>
        )
      })}
    </div>
  )
}
