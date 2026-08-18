import { FC, Fragment } from 'react'
import { Link } from 'react-router-dom'
import { Divider } from 'antd'
import CADIcon from '@/assets/icons/docs/CAD-search.svg'
import DataBaseIcon from '@/assets/icons/docs/DataBase-search.svg'
import ODSIcon from '@/assets/icons/docs/ODS-search.svg'
import QAIcon from '@/assets/icons/docs/QA-search.svg'
import EmptyState from '../../../../search/components/EmptyState'
import { Title } from '../../Title'
import { FileInSearchResult } from '../../searchType'

type Content = {
  value?: FileInSearchResult[]
}

// 根据forest_type映射对应的图片
const getForestTypeIcon = (forestType: string) => {
  switch (forestType) {
    case 'file':
      return ODSIcon // 多模态
    case 'qa':
      return QAIcon // 问答对
    case 'cad':
      return CADIcon // CAD
    case 'db':
      return DataBaseIcon // 数据库
    default:
      return ODSIcon // 默认使用多模态图标
  }
}

const ForestContent: FC<Content> = (props) => {
  const { value } = props
  if (!value || value.length === 0)
    return <EmptyState message='暂未查询到相关内容～' />

  return (
    <>
      {value.map((file, index) => {
        const {
          highlighted_forest_name,
          highlighted_description,
          id,
          forest_type,
        } = file as any // 临时使用any类型，因为类型定义可能还没更新
        const url = `/docs/detail/${id}`
        const iconUrl = getForestTypeIcon(forest_type || 'file')

        return (
          <div key={url}>
            <Link
              to={url}
              className='text-[unset] block hover:bg-[#EFF0F6] rounded-[10px] p-2.5 transition-colors'
              target='_blank'
            >
              <Title
                image={iconUrl}
                name={highlighted_forest_name!}
                desc={highlighted_description}
              ></Title>
            </Link>
            {/* {index < value.length - 1 && (
              <Divider className='my-3 border-gray-100' />
            )} */}
          </div>
        )
      })}
    </>
  )
}
export default ForestContent
