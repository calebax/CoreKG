import { useNavigate } from 'react-router-dom'
import { Breadcrumb } from 'antd'

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
}

export default function BreadcrumbNav({
  isRootLevel,
  knowledgeBaseId,
  knowledgeBaseName,
  folderInfo,
}: BreadcrumbNavProps) {
  const navigate = useNavigate()

  // 构建面包屑导航项
  const getBreadcrumbItems = () => {
    const items = [
      {
        title: (
          <span
            className='text-base cursor-pointer hover:text-[#0C99FF] font-medium'
            onClick={() => navigate('/docs')}
          >
            知识库
          </span>
        ),
      },
    ]

    if (isRootLevel) {
      // 根级别只添加知识库名称（有名称时才显示）
      if (folderInfo.name) {
        items.push({
          title: (
            <span className='text-base font-semibold'>{folderInfo.name}</span>
          ),
        })
      }
    } else {
      // 添加知识库名称作为第二级（有名称时才显示）
      if (knowledgeBaseName) {
        items.push({
          title: (
            <span
              className='text-base cursor-pointer hover:text-[#0C99FF] font-medium'
              onClick={() => navigate(`/docs/cad/${knowledgeBaseId}`)}
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
              className='text-base cursor-pointer hover:text-[#0C99FF] font-medium'
              onClick={() =>
                navigate(
                  `/docs/cad/${knowledgeBaseId}/folder/${item.id}?path=${pathParam}`,
                )
              }
            >
              {item.name}
            </span>
          ),
        })
      })

      // 添加当前文件夹作为最后一级
      items.push({
        title: (
          <span className='text-base font-semibold'>{folderInfo.name}</span>
        ),
      })
    }

    return items
  }

  return (
    <div className='mb-6'>
      <Breadcrumb
        separator={<span style={{ color: '#616373' }}>/</span>}
        items={getBreadcrumbItems()}
      />
    </div>
  )
}
