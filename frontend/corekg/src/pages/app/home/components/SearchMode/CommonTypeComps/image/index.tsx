import { FC, Fragment } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { getFileURL } from '@/utils/Forest'
import EmptyState from '../../../../search/components/EmptyState'
import { PicTitle } from '../../Title'
import { FileInSearchResult } from '../../searchType'

type Content = {
  value?: FileInSearchResult[]
}
const ImageContent: FC<Content> = (props) => {
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
              const urlWithFileLocation = getFileURL(forest_id, id, location)
              return (
                <Link
                  key={image_url}
                  to={urlWithFileLocation}
                  target='_blank'
                  className='text-[unset] hover:opacity-80 transition-opacity'
                >
                  <PicTitle image={image_url!} name={highlighted_file_name!} />
                </Link>
              )
            })}
          </Fragment>
        )
      })}
    </div>
  )
}

export default ImageContent
