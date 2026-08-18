import { FC, useMemo } from 'react'
import { Checkbox, Skeleton, Table, TableProps, Tooltip } from 'antd'
import { useRequest } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import { ForestPermItem, getForestPermSet } from '@/api/perm'
import { getUpdatedPerm } from '../utils'

type Item = ForestPermItem & { forest: { name: string } }
export type Forest = {
  uin?: number
  /** 父组件记录的变更权限，key 为 forest.ID */
  value?: Record<number, ForestPermItem>
  /** 通知父组件权限变更 */
  onChange: (id: number, perm: ForestPermItem) => void
  className?: string
}
export const Forest: FC<Forest> = (props) => {
  const { t } = useTranslation('pages')
  const { value: changedPerms = {}, onChange, className, uin } = props

  // 获取全量数据
  const { data: allData, loading } = useRequest(
    async () => {
      // 当没有 uin 时，获取所有 forest 列表（用于新用户权限设置）
      const { perm_set } = await getForestPermSet({
        ...(uin ? { uin } : {}),
      })
      const list = (perm_set ?? []).map((item) => {
        if (item.forest.public_scope !== 'company') return item
        return { ...item, use_perm: true }
      })
      return list
    },
    {
      refreshDeps: [uin],
    },
  )

  // 应用父组件传入的变更权限
  const mergedList = useMemo(() => {
    if (!allData) return []

    // 应用变更权限
    return allData.map((item) => {
      const changed = changedPerms[item.forest.ID]
      if (changed) {
        return {
          ...item,
          manage_perm: changed.manage_perm,
          use_perm: changed.use_perm,
        }
      }
      return item
    })
  }, [allData, changedPerms])

  const handlePermChange = (id: number, key: 'manage_perm' | 'use_perm') => {
    // 找到当前项（从全量数据中查找）
    const currentItem = allData?.find((item) => item.forest.ID === id)
    if (!currentItem) return

    // 应用变更权限后的值
    const changed = changedPerms[id]
    const baseItem = changed
      ? {
          ...currentItem,
          manage_perm: changed.manage_perm,
          use_perm: changed.use_perm,
        }
      : currentItem

    // 计算新的权限值
    const updatedItem = getUpdatedPerm(baseItem, key)

    // 传递完整的权限项信息给父组件（用于构建增量数据）
    onChange(id, {
      forest: { ID: id, public_scope: currentItem.forest.public_scope },
      manage_perm: updatedItem.manage_perm,
      use_perm: updatedItem.use_perm,
    })
  }

  const columns: TableProps<Item>['columns'] = [
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
              onChange={() => handlePermChange(record.forest.ID, 'manage_perm')}
            >
              {t('settings.management')}
            </Checkbox>
            <Tooltip title={tooltip}>
              <Checkbox
                checked={record.use_perm}
                onChange={() => handlePermChange(record.forest.ID, 'use_perm')}
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

  if (loading && !allData) return <Skeleton active className={className} />

  return (
    <Table<Item>
      className={cn('overflow-auto break-all', className)}
      dataSource={mergedList.map((item) => ({
        ...item,
        key: item.forest.ID,
      }))}
      columns={columns}
      loading={loading}
    />
  )
}
