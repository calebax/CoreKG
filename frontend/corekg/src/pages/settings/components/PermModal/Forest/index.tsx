import { FC } from 'react'
import { Checkbox, Skeleton, Table, TableProps, Tooltip } from 'antd'
import { produce } from 'immer'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import { ForestPermItem } from '@/api/perm'
import { getUpdatedPerm } from '../utils'

type Item = ForestPermItem & { forest: { name: string } }
export type Forest = {
  value?: Item[]
  onChange: (updater: (val?: Item[]) => Item[] | undefined) => void
  className?: string
}
export const Forest: FC<Forest> = (props) => {
  const { t } = useTranslation('pages')
  const { value, onChange, className } = props
  const updatePerm = (id: number, key: 'manage_perm' | 'use_perm') => {
    onChange((val) => {
      if (!val) return
      return produce(val, (draft) => {
        const index = draft.findIndex((item) => item.forest.ID === id)!
        if (index === -1) return
        draft[index] = getUpdatedPerm(draft[index], key)
      })
    })
  }
  const columns: TableProps['columns'] = [
    {
      title: t('settings.name', { target: t('settings.knowledgeBase') }),
      render: (record: Item) => record.forest.name,
    },
    {
      title: t('settings.permission'),
      width: 300,
      render: (record: Item) => {
        const manage_perm = record.manage_perm
        const company = record.forest.public_scope === 'company'
        const tooltip = (() => {
          if (company)
            return t('settings.contentVisibleToAll', {
              target: t('settings.knowledgeBase'),
            })
          if (manage_perm) return t('settings.adminDefaultVisibility')
          return null
        })()
        return (
          <span className='flex gap-2'>
            <Checkbox
              checked={record.manage_perm}
              onChange={() => updatePerm(record.forest.ID, 'manage_perm')}
            >
              {t('settings.management')}
            </Checkbox>
            <Tooltip title={tooltip}>
              <Checkbox
                checked={record.use_perm}
                onChange={() => updatePerm(record.forest.ID, 'use_perm')}
                disabled={manage_perm || company}
              >
                {t('settings.use')}
              </Checkbox>
            </Tooltip>
          </span>
        )
      },
    },
  ]
  if (!value) return <Skeleton active className={className} />
  return (
    <Table
      className={cn('overflow-auto break-all', className)}
      dataSource={value.map((item) => ({ ...item, key: item.forest.ID }))}
      columns={columns}
    ></Table>
  )
}
