import { FC, Fragment } from 'react'
import { Button, Divider } from 'antd'
import { ArrowRightOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { getFileURL } from '@/utils/Forest'
import { FileInSearchResult, ResultType as OriginResultType } from '../..'
import word from '../../images/word.png'

type ResultType = Exclude<OriginResultType, 'all'>

const ResultType_TitleMap: Record<ResultType, any> = {
  doc: 'file.doc',
  image: 'file.image',
  video: 'file.video',
}

type LinkItem = {
  /** 判断value的类型 */
  type: ResultType
  value?: FileInSearchResult[]
  onMoreItem?: () => void
}
export const LinkItem: FC<LinkItem> = (props) => {
  const { t: tC } = useTranslation('common')
  const { type, value, onMoreItem } = props
  if (!value || !value.length) return null

  return (
    <div className='flex flex-col p-2'>
      <span className='text-[#165DFF]'>{tC(ResultType_TitleMap[type])}</span>
      {value.map((file) => {
        const { forest_id, id, highlighted_file_name, highlights } = file
        const { highlighted_description, image_url = word } = highlights[0]
        const url = getFileURL(forest_id, id)
        return (
          <Fragment key={url}>
            <Link to={url} target='_blank' className=' text-[unset]'>
              <span className='flex gap-2 items-center'>
                <img loading='lazy' src={image_url} className=' w-12 h-12' />
                <div className='flex flex-col'>
                  <span
                    className='text-base'
                    dangerouslySetInnerHTML={{ __html: highlighted_file_name }}
                  ></span>
                  <span
                    dangerouslySetInnerHTML={{
                      __html: highlighted_description,
                    }}
                    className='text-sm line-clamp-2 text-[#758BBE]'
                  ></span>
                </div>
              </span>
            </Link>
            <Divider className='m-2' />
          </Fragment>
        )
      })}
      <Button
        type='text'
        size='small'
        icon={<ArrowRightOutlined />}
        className='opacity-50 self-start'
        onClick={onMoreItem}
      >
        {tC('button.viewMore')}
      </Button>
    </div>
  )
}
