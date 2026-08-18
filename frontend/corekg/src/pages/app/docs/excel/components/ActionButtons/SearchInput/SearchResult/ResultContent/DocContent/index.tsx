import { FC, Fragment } from 'react'
import { Button, Divider } from 'antd'
import { DownOutlined, UpOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { useImmer } from 'use-immer'
import { cn } from '@/utils'
import { getFileURL } from '@/utils/Forest'
import { FileInSearchResult } from '../../..'
import word from '../../../images/word.png'

type Content = {
  value?: FileInSearchResult[]
}
export const DocContent: FC<Content> = (props) => {
  const { t: tC } = useTranslation('common')
  const { value } = props
  const [hiddenList, setHiddenList] = useImmer<boolean[]>([])
  if (!value || value.length === 0) return tC('empty.noData')
  return (
    <div className='flex flex-col gap-2'>
      {value.map((file, i) => {
        const { forest_id, highlighted_file_name, highlights, id } = file
        const url = getFileURL(forest_id, id)
        const hidden = hiddenList[i]
        return (
          <div className=' p-2 flex flex-col' key={id}>
            <span className='flex gap-2 items-center'>
              <Link
                to={url}
                target='_blank'
                className={cn('flex gap-2 text-[unset]')}
              >
                <img loading='lazy' src={word} className=' w-12 h-12' />
                <span
                  className='text-base'
                  dangerouslySetInnerHTML={{ __html: highlighted_file_name }}
                ></span>
              </Link>
              <Button
                className='ml-auto'
                size='small'
                onClick={(e) => {
                  setHiddenList((draft) => {
                    draft[i] = !draft[i]
                  })
                  e.preventDefault()
                }}
                icon={hidden ? <UpOutlined /> : <DownOutlined />}
              >
                {tC(`button.${hidden ? 'expand' : 'collapse'}`)}
              </Button>
            </span>
            <Divider className='m-2' />
            {hidden
              ? null
              : highlights.map((item) => {
                  const { highlighted_description, location } = item
                  const urlWithFileLocation = getFileURL(
                    forest_id,
                    id,
                    location,
                  )
                  return (
                    <Fragment key={highlighted_description}>
                      <Link
                        to={urlWithFileLocation}
                        target='_blank'
                        className={cn('flex flex-col text-[unset]')}
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
          </div>
        )
      })}
    </div>
  )
}
