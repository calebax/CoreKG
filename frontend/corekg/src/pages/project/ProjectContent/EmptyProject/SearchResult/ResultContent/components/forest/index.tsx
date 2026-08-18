import { FC } from 'react'
import { CommonResultItem } from '../../CommonResultItem'

const ForestItem: FC<{ value?: any }> = (props) => {
  const { value } = props
  const {
    id,
    forest_type,
    data_source_type,
    data_source_subtype,
    user_name,
    avatar_url,
    created_at,
    highlighted_forest_name,
    highlighted_description,
  } = value
  const url_gap = useMemo(() => {
    if (forest_type === 'file') return 'detail'
    if (forest_type !== 'data') return forest_type
    return data_source_type
  }, [data_source_type, forest_type])
  return (
    <CommonResultItem
      className='w-full'
      type='forest'
      creator={user_name}
      creatorAvatar={avatar_url}
      time={created_at}
      title_html={highlighted_forest_name}
      to={`/docs/${url_gap}/${id}`}
    >
      <div
        className='text-[#3C4149]'
        dangerouslySetInnerHTML={{ __html: highlighted_description }}
      ></div>
    </CommonResultItem>
  )
}
export default ForestItem
