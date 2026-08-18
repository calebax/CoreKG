import { useNavigate } from 'react-router-dom'
import { Breadcrumb } from 'antd'
import { useTranslation } from 'react-i18next'
import KnowledgeBaseIcon from '@/assets/icons/docs/knowledge-base-icon.svg?react'
import NavigationIcon from '@/assets/icons/docs/navigation.svg?react'

interface PathItem {
  id: number
  name: string
  level: number
}

interface BreadcrumbNavProps {
  isRootLevel: boolean
  knowledgeBaseId: string
  knowledgeBaseName: string
  folderInfo: {
    name: string
    level: number
    path: PathItem[]
  }
  onNavigateAway?: () => void
}

export default function BreadcrumbNav({
  isRootLevel,
  knowledgeBaseId,
  knowledgeBaseName,
  folderInfo,
  onNavigateAway,
}: BreadcrumbNavProps) {
  const navigate = useNavigate()
  const { t } = useTranslation('pages')

  // 构建面包屑导航项
  const getBreadcrumbItems = () => {
    const items = [
      {
        title: (
          <span
            className='flex items-center gap-2 text-sm cursor-pointer hover:text-[#0C99FF] font-medium text-[#3C4149]'
            onClick={() => {
              onNavigateAway?.()
              navigate('/docs')
            }}
          >
            <KnowledgeBaseIcon className='w-4 h-4' />
            <span>{t('app.docs.knowledgeBase')}</span>
          </span>
        ),
      },
    ]

    if (isRootLevel) {
      // 根级别只添加知识库名称（有名称时才显示）
      if (folderInfo.name) {
        items.push({
          title: (
            <span className='text-sm font-medium text-[#3C4149]'>
              {folderInfo.name}
            </span>
          ),
        })
      }
    } else {
      // 添加知识库名称作为第二级（有名称时才显示）
      if (knowledgeBaseName) {
        items.push({
          title: (
            <span
              className='text-sm cursor-pointer hover:text-[#0C99FF] font-medium text-[#3C4149]'
              onClick={() => {
                onNavigateAway?.()
                navigate(`/docs/detail/${knowledgeBaseId}`)
              }}
            >
              {knowledgeBaseName}
            </span>
          ),
        })
      }

      // 添加路径中的文件夹作为后续级别
      folderInfo.path.forEach((item, index) => {
        // 获取到该层级的路径
        const prevPath = folderInfo.path.slice(0, index)
        const pathParam = encodeURIComponent(JSON.stringify(prevPath))

        items.push({
          title: (
            <span
              className='text-sm cursor-pointer hover:text-[#0C99FF] font-medium text-[#3C4149]'
              onClick={() => {
                onNavigateAway?.()
                navigate(
                  `/docs/detail/${knowledgeBaseId}/folder/${item.id}?path=${pathParam}`,
                )
              }}
            >
              {item.name}
            </span>
          ),
        })
      })

      // 添加当前文件夹作为最后一级
      items.push({
        title: (
          <span className='text-sm font-medium text-[#3C4149]'>
            {folderInfo.name}
          </span>
        ),
      })
    }

    return items
  }

  return (
    <div className='w-full h-12 bg-white flex items-center px-5 pt-[2px] font-medium'>
      <Breadcrumb
        className='[&_span.ant-breadcrumb-separator]:inline-flex [&_span.ant-breadcrumb-separator]:items-center [&_span.ant-breadcrumb-separator]:align-middle'
        separator={<NavigationIcon className='inline-block' />}
        items={getBreadcrumbItems()}
      />
    </div>
  )
}
