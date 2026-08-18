import { FC } from 'react'
import { getLikeList } from '@/api/knowledge'
import LikeCollectList from '../components/LikeCollectList'

const MyLikes: FC = () => {
  return (
    <LikeCollectList type='like' fetchList={getLikeList} title='我的点赞' />
  )
}

export default MyLikes

