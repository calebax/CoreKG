import { FC } from 'react'
import { getCollectionList } from '@/api/knowledge'
import LikeCollectList from '../components/LikeCollectList'

const MyCollections: FC = () => {
  return (
    <LikeCollectList
      type='collect'
      fetchList={getCollectionList}
      title='我的收藏'
    />
  )
}

export default MyCollections

