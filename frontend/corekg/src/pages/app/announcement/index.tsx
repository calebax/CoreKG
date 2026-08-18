import { FC, useEffect, useMemo } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Breadcrumb, Spin } from 'antd'
import { useRequest } from 'ahooks'
import { cn } from '@/utils'
import { listAnnouncement } from '@/api/announcement'
import SeparatorIcon from '@/assets/separator.svg?react'
import MarkdownPreview from '@/components/common/MarkdownPreview'
import styles from './index.module.scss'

interface AnnouncementItem {
  id: number
  created_at: string
  uin: number
  company_id: number
  creator: string
  tag: string
  content: string
}

const Announcement: FC = () => {
  const navigate = useNavigate()
  const { id } = useParams<{ id?: string }>()

  // 获取公告列表
  const { data, loading } = useRequest(
    async () => {
      const result = await listAnnouncement({
        Limit: 9999,
        Offset: 0,
        OrderBy: ['created_at desc'],
      })
      return (result?.data || []) as AnnouncementItem[]
    },
    {
      refreshDeps: [],
    },
  )

  // 根据路由参数或默认选中第一个
  useEffect(() => {
    if (!data || data.length === 0) return

    const routeId = id ? Number(id) : null
    const isValidId = routeId && data.some((item) => item.id === routeId)

    // 如果路由中没有 id，或者 id 无效，则导航到第一个
    if (!id || !isValidId) {
      navigate(`/announcement/${data[0].id}`, { replace: true })
    }
  }, [data, id, navigate])

  // 从路由参数获取选中的 id
  const selectedId = useMemo(() => {
    if (!id) return null
    const numId = Number(id)
    return isNaN(numId) ? null : numId
  }, [id])

  const selectedAnnouncement = useMemo(() => {
    if (!data || !selectedId) return undefined
    return data.find((item) => item.id === selectedId)
  }, [data, selectedId])

  // 解码 Unicode 转义序列并处理内容
  const processedContent = useMemo(() => {
    if (!selectedAnnouncement?.content) return ''
    // 解码 Unicode 转义序列（如 \u003c -> <）
    return selectedAnnouncement.content.replace(
      /\\u([0-9a-fA-F]{4})/g,
      (_, hex) => String.fromCharCode(parseInt(hex, 16)),
    )
  }, [selectedAnnouncement?.content])

  // 判断是否为纯 HTML（包含 HTML 标签且没有 Markdown 语法）
  const isPureHtml = useMemo(() => {
    if (!processedContent) return false
    // 检查是否包含 HTML 标签（如 <ul>, <li>, <h1>, <p> 等）
    const hasHtmlTags = /<[a-z][a-z0-9]*(\s[^>]*)?>/i.test(processedContent)
    // 检查是否包含 Markdown 语法（如果包含，则不是纯 HTML）
    const hasMarkdownSyntax =
      /^#{1,6}\s|^\*\s|^-\s|^\d+\.\s|```|\[.*\]\(.*\)/m.test(processedContent)
    return hasHtmlTags && !hasMarkdownSyntax
  }, [processedContent])

  if (loading) {
    return (
      <div className='w-full h-full flex items-center justify-center'>
        <Spin size='large' />
      </div>
    )
  }

  return (
    <div className='w-full h-full overflow-hidden flex flex-col'>
      {/* 面包屑导航 */}
      <div className='border-b border-[#EFF1F4] py-3 pl-4'>
        <Breadcrumb
          className={styles.layoutHeader}
          separator={<SeparatorIcon />}
          items={[
            {
              title: (
                <span
                  className='text-sm cursor-pointer hover:text-[#0C99FF] font-medium text-[#3C4149]'
                  onClick={() => {
                    navigate('/global')
                  }}
                >
                  问答
                </span>
              ),
            },
            {
              title: (
                <span className='cursor-pointer text-sm font-medium text-[#3C4149]'>
                  发版公告
                </span>
              ),
            },
          ]}
        />
      </div>

      {/* 主体内容区域 */}
      <div className='flex-1 overflow-hidden flex'>
        {/* 左侧版本列表 */}
        <div className='w-50 border-r border-[#EFF1F4] overflow-y-auto overflow-x-hidden scrollbar-hide'>
          <div className='p-4'>
            {data?.map((item) => {
              const isSelected = item.id === selectedId
              return (
                <div
                  key={item.id}
                  className={cn(
                    'px-3 py-2 mb-2 rounded cursor-pointer transition-colors',
                    {
                      'bg-[#E6F7FF] text-[#0C99FF]': isSelected,
                      'hover:bg-[#F5F5F5] text-[#3C4149]': !isSelected,
                    },
                  )}
                  onClick={() => {
                    navigate(`/announcement/${item.id}`)
                  }}
                >
                  <div className='text-sm font-medium'>
                    {item.tag ? `${item.tag}版本` : '未命名版本'}
                  </div>
                </div>
              )
            })}
          </div>
        </div>

        {/* 右侧内容区域 */}
        <div className='flex-1 overflow-hidden'>
          {selectedAnnouncement ? (
            <div className='h-full overflow-y-auto overflow-x-hidden'>
              {isPureHtml ? (
                <div
                  className={cn(
                    'markdown-body h-full',
                    styles.announcementContent,
                  )}
                  style={{
                    backgroundColor: '#ffffff',
                    padding: '24px',
                    fontSize: '16px',
                    lineHeight: '1.6',
                    color: '#0C1F17',
                  }}
                  dangerouslySetInnerHTML={{
                    __html: processedContent,
                  }}
                />
              ) : (
                <MarkdownPreview
                  content={processedContent}
                  className='h-full'
                  style={{
                    backgroundColor: '#ffffff',
                    padding: '24px',
                  }}
                  disableReference
                />
              )}
            </div>
          ) : (
            <div className='h-full flex items-center justify-center text-[#999999]'>
              请选择版本查看内容
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default Announcement
