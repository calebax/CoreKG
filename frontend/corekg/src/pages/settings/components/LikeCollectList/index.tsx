import { FC, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import {
  Table,
  Button,
  Space,
  message,
  Pagination,
  Breadcrumb,
  Tooltip,
} from 'antd'
import { useRequest } from 'ahooks'
import dayjs from 'dayjs'
import { toggleDocLike, toggleDocCollect } from '@/api/knowledge'
import SeparatorIcon from '@/assets/separator.svg?react'
import HelpIcon from '@/pages/app/docs/detail/images/help-tip.svg?react'
import { getFileIcon } from '@/pages/app/docs/detail/utils/fileUtils'
import styles from './styles.module.scss'

interface LikeCollectListProps {
  type: 'like' | 'collect'
  fetchList: (data: { limit: number; offset: number }) => Promise<any>
  title: string
}

const LikeCollectList: FC<LikeCollectListProps> = ({
  type,
  fetchList,
  title,
}) => {
  const navigate = useNavigate()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)

  // 获取列表数据
  const {
    data,
    loading,
    run: refresh,
  } = useRequest(
    () =>
      fetchList({
        limit: pageSize,
        offset: (page - 1) * pageSize,
      }),
    {
      refreshDeps: [page, pageSize],
    },
  )

  const handleToggleStatus = async (record: any) => {
    try {
      if (type === 'like') {
        await toggleDocLike({
          resource_id: record.resource_id,
          resource_type: record.resource_type || 'forest_file',
          enable: false,
        })
        message.success('取消点赞成功')
      } else {
        await toggleDocCollect({
          resource_id: record.resource_id,
          resource_type: record.resource_type || 'forest_file',
          enable: false,
        })
        message.success('取消收藏成功')
      }
      refresh()
    } catch (error) {
      console.log(`${title}操作失败:`, error)
    }
  }

  const columns = [
    {
      title: '名称',
      dataIndex: 'resource_name',
      key: 'resource_name',
      width: 300,
      render: (text: string, record: any) => {
        const ext = text?.split('.').pop()?.toLowerCase() || ''
        const icon = getFileIcon(ext, false)
        // 跳转地址：添加 kbName 和 fileName 参数
        const params = new URLSearchParams()
        params.append('kbName', '知识库')
        params.append('fileName', text || '')
        // 使用 resource_id 作为跳转的文件 ID
        const previewUrl = `/docs/detail/${record.forest_id}/file/${record.resource_id}?${params.toString()}`

        return (
          <div className='flex items-center gap-2 w-full overflow-hidden'>
            <div className='flex-shrink-0 flex items-center justify-center w-6 h-6'>
              {icon}
            </div>
            <Tooltip title={text}>
              <Link
                to={previewUrl}
                target='_blank'
                className='flex-1 min-w-0 truncate text-[#0C99FF] hover:underline font-medium'
              >
                {text}
              </Link>
            </Tooltip>
          </div>
        )
      },
    },
    {
      title: (
        <div className='flex items-center gap-[4px] text-[#000000] font-semibold'>
          分段规则
          <Tooltip
            title={
              <div className='text-xs'>
                1、系统将使用默认分段规则解析导入文件：
                <br />
                2、您可在文件详情页手动调整分段规则。
              </div>
            }
          >
            <HelpIcon className='cursor-pointer w-3.5 h-3.5' />
          </Tooltip>
        </div>
      ),
      dataIndex: 'file_config',
      key: 'file_config',
      width: 150,
      render: (config: any) => {
        let typeText = '用户自定义'
        if (
          !config?.split_config ||
          config?.split_config?.split_mode === 'auto'
        ) {
          typeText = '系统默认'
        } else if (config?.split_config?.split_mode === 'rule') {
          typeText = '自定义分段'
        }
        return <div className='text-[#3C4149]'>{typeText}</div>
      },
    },
    {
      title: '标签',
      dataIndex: 'tag_list',
      key: 'tag_list',
      width: 200,
      render: (tagList: any[]) => {
        if (!tagList || tagList.length === 0) {
          return <span className='text-[#919497]'>-</span>
        }
        const tagNames = tagList
          .map((tag: any) => tag.tag_name || tag.name || '')
          .filter(Boolean)
        const tagText = tagNames.join('、')
        return (
          <Tooltip title={tagText}>
            <div className='max-w-[180px] truncate text-[#3C4149]'>
              {tagText}
            </div>
          </Tooltip>
        )
      },
    },
    {
      title: '所属知识库',
      dataIndex: 'forest_name',
      key: 'forest_name',
      width: 180,
      render: (text: string) => (
        <span className='text-[#3C4149] truncate block max-w-[160px]'>
          {text || '-'}
        </span>
      ),
    },
    {
      title: type === 'like' ? '点赞时间' : '收藏时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (time: number | string) => {
        if (!time) return '-'
        const date =
          typeof time === 'number' && time < 10000000000 ? time * 1000 : time
        return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
      },
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      fixed: 'right' as const,
      render: (_: any, record: any) => (
        <Button
          type='link'
          className='p-0 h-auto text-[#0C99FF] hover:text-[#38b2ff]'
          onClick={() => handleToggleStatus(record)}
        >
          {type === 'like' ? '取消点赞' : '取消收藏'}
        </Button>
      ),
    },
  ]

  return (
    <div className='w-full h-full flex flex-col bg-[#FAFAFA]'>
      <div className='w-full h-12 bg-white flex items-center px-5 pt-[2px] border-b border-[#EFF1F4]'>
        <Breadcrumb
          className={styles.layoutHeader}
          separator={<SeparatorIcon />}
          items={[
            {
              title: (
                <span
                  className='text-sm cursor-pointer hover:text-[#0C99FF] font-medium text-[#3C4149]'
                  onClick={() => navigate('/')}
                >
                  问答
                </span>
              ),
            },
            {
              title: (
                <span className='text-sm font-medium text-[#3C4149]'>
                  {title}
                </span>
              ),
            },
          ]}
        />
      </div>
      <div className='flex-1 overflow-auto p-6'>
        <div className='flex flex-col h-full bg-white rounded-lg shadow-sm overflow-hidden'>
          <Table
            columns={columns}
            dataSource={data?.data || []}
            loading={loading}
            pagination={false}
            rowKey='id'
            tableLayout='fixed'
            className='flex-1 overflow-auto'
          />
          <div className='p-4 border-t border-[#EFF1F4] flex justify-end'>
            <Pagination
              current={page}
              pageSize={pageSize}
              total={data?.total || 0}
              onChange={(p, ps) => {
                setPage(p)
                setPageSize(ps)
              }}
              showSizeChanger
              showQuickJumper
              showTotal={(total) => `共计 ${total} 条数据`}
            />
          </div>
        </div>
      </div>
    </div>
  )
}

export default LikeCollectList
