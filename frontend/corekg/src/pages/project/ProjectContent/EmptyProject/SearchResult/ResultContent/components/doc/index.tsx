import { FC, useState, useMemo } from 'react'
import { message, Tooltip } from 'antd'
import { toggleDocLike, toggleDocCollect } from '@/api/knowledge'
import { useDeployConfig } from '@/utils/useDeployConfig'
import CollectIcon from '../../../../images/collect.svg?react'
import CollectActiveIcon from '../../../../images/collectActive.svg?react'
import LikeIcon from '../../../../images/like.svg?react'
import LikeActiveIcon from '../../../../images/likeActive.svg?react'
import { CommonResultItem } from '../../CommonResultItem'

const DocItem: FC<{ value?: any }> = (props) => {
  const { value } = props
  const {
    id,
    forest_id,
    user_name,
    avatar_url,
    created_at,
    highlighted_file_name,
    highlights,
    is_like,
    is_collect,
  } = value
  const { highlighted_description, location } = highlights[0]

  // 环境判断：只在本地环境、测试环境、生产环境、或 custom 版本且 mode 为 cimc/h3c 时显示点赞收藏功能
  const isDevEnv = import.meta.env.MODE === 'development'
  const isTestEnv = import.meta.env.MODE === 'test'
  const isProdEnv = import.meta.env.MODE === 'production'
  const { version, mode } = useDeployConfig()
  const shouldShowLikeCollect =
    isDevEnv ||
    isTestEnv ||
    isProdEnv ||
    (version === 'custom' && (mode === 'cimc' || mode === 'h3c'))

  // 点赞和收藏状态
  const [liked, setLiked] = useState(is_like)
  const [collected, setCollected] = useState(is_collect)
  const [likeLoading, setLikeLoading] = useState(false)
  const [collectLoading, setCollectLoading] = useState(false)

  const url = useMemo(() => {
    const searchParams = new URLSearchParams()
    searchParams.append(
      'location',
      encodeURIComponent(JSON.stringify(location)),
    )
    return `/docs/detail/${forest_id}/file/${id}?${searchParams.toString()}`
  }, [forest_id, id, location])

  // 处理点赞
  const handleLike = async (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (likeLoading) return

    setLikeLoading(true)
    try {
      const enable = !liked
      await toggleDocLike({
        resource_id: id,
        resource_type: 'forest_file',
        enable,
      })
      setLiked(enable)
      message.success(enable ? '点赞成功' : '取消点赞成功')
    } catch (error) {
      console.log('点赞操作失败:', error)
    } finally {
      setLikeLoading(false)
    }
  }

  // 处理收藏
  const handleCollect = async (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (collectLoading) return

    setCollectLoading(true)
    try {
      const enable = !collected
      await toggleDocCollect({
        resource_id: id,
        resource_type: 'forest_file',
        enable,
      })
      setCollected(enable)
      message.success(enable ? '收藏成功' : '取消收藏成功')
    } catch (error) {
      console.log('收藏操作失败:', error)
    } finally {
      setCollectLoading(false)
    }
  }

  const titleExtra = (
    <div className='flex items-center gap-2 flex-shrink-0'>
      {/* 点赞在前 */}
      <Tooltip title={liked ? '取消点赞' : '点赞'}>
        <button
          type='button'
          onClick={handleLike}
          disabled={likeLoading}
          className='flex items-center justify-center w-6 h-6 hover:bg-gray-100 rounded transition-colors cursor-pointer'
        >
          {liked ? (
            <LikeActiveIcon className='w-4 h-4' />
          ) : (
            <LikeIcon className='w-4 h-4' />
          )}
        </button>
      </Tooltip>
      {/* 收藏在后 */}
      <Tooltip title={collected ? '取消收藏' : '收藏'}>
        <button
          type='button'
          onClick={handleCollect}
          disabled={collectLoading}
          className='flex items-center justify-center w-6 h-6 hover:bg-gray-100 rounded transition-colors cursor-pointer'
        >
          {collected ? (
            <CollectActiveIcon className='w-4 h-4' />
          ) : (
            <CollectIcon className='w-4 h-4' />
          )}
        </button>
      </Tooltip>
    </div>
  )

  return (
    <CommonResultItem
      className='w-full'
      type='doc'
      creator={user_name}
      creatorAvatar={avatar_url}
      time={created_at}
      title_html={highlighted_file_name}
      titleExtra={shouldShowLikeCollect ? titleExtra : undefined}
      to={url}
    >
      <div
        className='text-[#3C4149]'
        dangerouslySetInnerHTML={{ __html: highlighted_description }}
      ></div>
    </CommonResultItem>
  )
}
export default DocItem
