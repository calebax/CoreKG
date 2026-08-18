import { FC } from 'react'
import { Checkbox, Skeleton, Table, TableProps, Tooltip } from 'antd'
import { produce } from 'immer'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import { AgentPermItem } from '@/api/perm'
import { getUpdatedPerm } from '../utils'

type Item = AgentPermItem & { agent: { show_name: string } }
export type Agent = {
  value?: Item[]
  onChange: (updater: (val?: Item[]) => Item[] | undefined) => void
  className?: string
}
export const Agent: FC<Agent> = (props) => {
  const { t } = useTranslation('pages')
  const { value, onChange, className } = props
  const updatePerm = (id: number, key: 'manage_perm' | 'use_perm') => {
    onChange((val) => {
      if (!val) return
      return produce(val, (draft) => {
        const index = draft.findIndex((item) => item.agent.ID === id)!
        if (index === -1) return
        draft[index] = getUpdatedPerm(draft[index], key)
      })
    })
  }
  const columns: TableProps['columns'] = [
    {
      title: t('settings.name', { target: t('settings.app') }),
      render: (record: Item) => record.agent.show_name,
    },
    {
      title: t('settings.permission'),
      width: 300,
      render: (record: Item) => {
        const manage_perm = record.manage_perm
        const company = record.agent.public_scope === 'company'
        const tooltip = (() => {
          if (company)
            return t('settings.contentVisibleToAll', {
              target: t('settings.app'),
            })
          if (manage_perm) return t('settings.adminDefaultVisibility')
          return null
        })()
        return (
          <span className='flex gap-2'>
            <Checkbox
              checked={record.manage_perm}
              onChange={() => updatePerm(record.agent.ID, 'manage_perm')}
            >
              {t('settings.management')}
            </Checkbox>
            <Tooltip title={tooltip}>
              <Checkbox
                checked={record.use_perm}
                onChange={() => updatePerm(record.agent.ID, 'use_perm')}
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
      dataSource={value.map((item) => ({ ...item, key: item.agent.ID }))}
      columns={columns}
    ></Table>
  )
}
