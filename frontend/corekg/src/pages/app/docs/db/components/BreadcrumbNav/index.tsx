import { useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { Breadcrumb, BreadcrumbProps } from 'antd'
import { useRequest } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import { getNameByModuleIDs } from '@/api/knowledge'
import NavigationIcon from '@/assets/icons/docs/navigation.svg?react'
import { IDInfo } from '../../hooks/useIds'

export default function BreadcrumbNav(props: Style & { id_info: IDInfo }) {
  const { id_info, className, style } = props
  const { forest_id, forest_db_name, forest_table_name } = id_info
  const { t } = useTranslation('pages')
  const navigate = useNavigate()
  const { data: forestInfo } = useRequest(
    async () => {
      const res = await getNameByModuleIDs({
        module_id_list: [
          {
            ids: [forest_id],
            module: 'forest',
          },
        ],
      })
      return res.name_list[0] as { id: any; name: string }
    },
    {
      refreshDeps: [forest_id],
    },
  )
  const items = useMemo(() => {
    const itemClassName =
      'text-base  hover:text-[#0C99FF] font-medium bg-transparent'
    const _items: BreadcrumbProps['items'] = [
      {
        title: (
          <span
            onClick={() => {
              navigate(`/docs`)
            }}
            className='text-sm cursor-pointer hover:text-[#0C99FF] font-medium text-[#3C4149]'
          >
            {t('app.docs.knowledgeBase')}
          </span>
        ),
      },
    ]
    if (!forestInfo) return
    const name_id_list = [forestInfo]
    if (forest_db_name) {
      name_id_list.push({ id: forest_db_name, name: forest_db_name })
      if (forest_table_name) {
        name_id_list.push({ id: forest_table_name, name: forest_table_name })
      }
    }
    const hrefs: string[] = []

    name_id_list.forEach((item, i) => {
      const { id, name } = item
      if (i === 0) {
        hrefs.push(`/docs/db/${id}`)
      } else {
        hrefs.push(`${hrefs[i - 1]}/${id}`)
      }
      // 下一级的href从上一级的href拼接而来
      const href = hrefs[i]
      _items.push({
        title: (
          <span
            onClick={() => {
              navigate(href)
            }}
            className={
              i === name_id_list.length - 1
                ? 'text-sm font-medium text-[#3C4149]'
                : 'text-sm cursor-pointer hover:text-[#0C99FF] font-medium text-[#3C4149]'
            }
          >
            {name}
          </span>
        ),
      })
    })

    return _items
  }, [forest_db_name, forestInfo, forest_table_name])

  const breadcrumbClassName = cn(
    '[&_span.ant-breadcrumb-separator]:inline-flex [&_span.ant-breadcrumb-separator]:items-center [&_span.ant-breadcrumb-separator]:align-middle',
    className, // 外部传入的className拼接到后面
  )

  return (
    <Breadcrumb
      className={breadcrumbClassName}
      separator={<NavigationIcon className='inline-block' />}
      items={items}
      style={style}
    />
  )
}
