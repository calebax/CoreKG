import { FC } from 'react'
import { Button, Input, List, Popover, Typography } from 'antd'
import { useRequest } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { listCustomModel } from '@/api'
import { cn } from '@/utils'
import ArrowDown from '@/assets/icons/arrow-down.svg?react'
import SearchIcon from '@/assets/icons/search.svg?react'
import { scroll } from '@/styles/scroll'
import styles from './styles.module.scss'

export type ModelSelect = {
  className?: string
  /** 如果传string则展示这个名称 且不再允许选择 */
  value?: number | string
  onChange?: (val: number) => void
}
export const ModelSelect: FC<ModelSelect> = (props) => {
  const { t: tC } = useTranslation('common')
  const { t } = useTranslation('pages')
  const { value, onChange, className } = props
  const disabled = typeof value === 'string'
  const [search, setSearch] = useState('')
  const modelList = useRequest(
    async () => {
      const { Data: chat_models = [] } = (await listCustomModel()) as any as {
        Data?: any[]
      }
      if (chat_models.length > 0) {
        onChange?.(chat_models[0].ID)
      }
      return chat_models.map((item) => {
        const { ID, show_name, description } = item
        return {
          id: ID,
          name: show_name,
          desc: description,
        }
      })
    },
    { ready: !disabled },
  )
  const getModelBtn = (text: string) => (
    <Button
      // icon={disabled ? null : <ArrowDown className='ml-auto' />}
      iconPosition='end'
      className={cn(
        'w-auto rounded-[28px] px-2.5',
        className,
        styles.modelSelect,
        disabled && styles.modelSelectDisabled,
      )}
    >
      {text}
    </Button>
  )
  if (disabled) {
    return getModelBtn(value)
  }
  if (!value || !modelList.data) return null
  const selectedModel = modelList.data.find((item) => item.id === value)

  return (
    <Popover
      trigger={['hover']}
      placement='bottom'
      arrow={false}
      content={
        <div className='bg-white rounded-[10px] shadow-[0px_2px_12px_0px_rgba(0,0,0,0.1)] flex flex-col gap-2.5 p-[10px] min-w-44 max-w-[40vw] mt-4'>
          <Input
            prefix={<SearchIcon className='w-3.5 h-3.5' />}
            allowClear
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={tC('button.search')}
            className={cn(
              'h-8 rounded-[19px] px-2 text-sm text-[#1e1f28]',
              className,
              styles.searchInput,
            )}
            style={{
              borderColor: search ? '#a895fc' : '#d7d9e5',
            }}
          />
          <div
            className={cn(
              'flex flex-col gap-1 max-h-[25vh] overflow-auto',
              scroll,
            )}
          >
            {modelList.data
              .filter((item) => item.name.includes(search))
              .map((item) => (
                <div
                  key={item.id}
                  className={cn(
                    'h-[48px] flex flex-col justify-center px-2 pb-1 cursor-pointer hover:bg-[#f8f9fd] rounded',
                    item.id === value ? 'bg-[#f8f9fd] rounded' : '',
                  )}
                  onClick={() => onChange?.(item.id)}
                >
                  <div className='flex flex-col justify-center mb-[4px] w-full h-[48px]'>
                    <p
                      className={cn(
                        'block leading-6 text-sm text-[#1e1f28]',
                        item.id === value ? 'font-medium' : 'font-normal',
                      )}
                    >
                      {item.name}
                    </p>
                  </div>
                  <div className='flex flex-col justify-center w-full font-normal'>
                    <p className='block leading-6 text-xs text-[#616373]'>
                      {item.desc}
                    </p>
                  </div>
                </div>
              ))}
          </div>
        </div>
      }
    >
      {getModelBtn(
        selectedModel ? selectedModel.name : t('app.home.selectLargeModel'),
      )}
    </Popover>
  )
}
