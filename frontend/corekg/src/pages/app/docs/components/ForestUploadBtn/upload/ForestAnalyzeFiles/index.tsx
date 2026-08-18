import { FC, createContext } from 'react'
import { App, Empty, Progress, Skeleton, Tabs } from 'antd'
import { cn, formatFileSize } from '@/utils'
import { retryParse } from '@/api/knowledge'
import DirIcon from './dir.svg?react'
import RefreshIcon from './refresh.svg?react'

export type AnalyzeFile = {
  name: string
  size: number
  percent: number
  status: 'waiting' | 'analyzing' | 'finished' | 'error'
  forest_id: number
  parent_id: number
  file_id: number
  forestPrefix: string
}

export type ForestAnalyzeFiles = Style & {
  files?: AnalyzeFile[]
  loading?: boolean
  reload?: () => void
  closeModal?: () => void
}
export const ForestAnalyzeFiles: FC<ForestAnalyzeFiles> = (props) => {
  const { files, reload, closeModal, className, style } = props
  type StatusFilter = AnalyzeFile['status'] | 'finished'
  const [status, setStatus] = useState<StatusFilter>('finished')
  const filteredFiles = useMemo(() => {
    if (!files || files.length === 0) return files
    return files.filter((file) => file.status === status)
  }, [files, status])
  const tabItems = useMemo(() => {
    const statusCounts = {
      finished: files?.filter((file) => file.status === 'finished').length || 0,
      waiting: files?.filter((file) => file.status === 'waiting').length || 0,
      analyzing:
        files?.filter((file) => file.status === 'analyzing').length || 0,
      error: files?.filter((file) => file.status === 'error').length || 0,
    }

    const items: {
      key: StatusFilter
      label: string
    }[] = [
      // { label: '全部', key: 'all' },
      { label: `已完成（${statusCounts.finished}）`, key: 'finished' },
      { label: `待开始（${statusCounts.waiting}）`, key: 'waiting' },
      { label: `进行中（${statusCounts.analyzing}）`, key: 'analyzing' },
      { label: `解析失败（${statusCounts.error}）`, key: 'error' },
    ]
    return items
  }, [files])
  if (!files) {
    return <Skeleton active className={cn('mt-10', className)} style={style} />
  }
  if (files.length === 0) {
    return <Empty className={cn('mt-10', className)} style={style} />
  }
  const analyzingNum = files.filter(
    (item) => item.status === 'analyzing',
  ).length

  return (
    <div
      className={cn(
        'max-h-full p-2.5 flex flex-col overflow-hidden',
        className,
      )}
      style={style}
    >
      <div className='flex py-2.5 text-[#616373]'>
        {/* <span className='font-semibold'>
          正在解析（{analyzingNum}/{files.length}）
        </span> */}
        系统仅展示近五天的任务
      </div>
      <Tabs
        activeKey={status}
        onChange={(k) => setStatus(k as any)}
        items={tabItems}
      />
      <div className='flex-1 flex flex-col overflow-auto'>
        {!filteredFiles || filteredFiles.length === 0 ? (
          <Empty />
        ) : (
          filteredFiles.map((file) => {
            return (
              <AnalyzeFileContext.Provider
                value={{ file, reload, closeModal }}
                key={file.file_id}
              >
                <AnalyzeFileItem />
              </AnalyzeFileContext.Provider>
            )
          })
        )}
      </div>
    </div>
  )
}

const AnalyzeFileContext = createContext<{
  file: AnalyzeFile
  reload?: () => void
  closeModal?: () => void
} | null>(null)

const AnalyzeFileItem: FC = () => {
  const { file } = useContext(AnalyzeFileContext)!
  const { status, percent, size, name } = file
  const strokeColor = useMemo(() => {
    switch (status) {
      case 'analyzing':
      case 'waiting':
        return '#3D7FFF'
      case 'error':
        return '#FF3B33'
      case 'finished':
        return '#30CD28'
    }
  }, [status])
  const textStatus = useMemo(() => {
    switch (status) {
      case 'waiting':
        return '待开始'
      case 'analyzing':
        return '进行中'
      case 'finished':
        return '已完成'
      case 'error':
        return '解析失败'
    }
  }, [status])
  return (
    <div className='p-2.5 flex flex-col gap-1.5 hover:bg-[#F8F9FD]'>
      <div className='flex mr-16 items-center'>
        <span className='text-[#1E1F28] text-base'>{name}</span>
        <span className='ml-auto text-[#616373]'>{formatFileSize(size)}</span>
      </div>
      <div className='flex items-center'>
        <Progress
          percent={percent * 100}
          type='line'
          strokeColor={strokeColor}
          showInfo={false}
        />
        <div className='w-16 flex-none flex items-center justify-center'>
          <FileOperator />
        </div>
      </div>
      <div className='flex mr-16'>
        <span className='text-[#616373]'>{Math.floor(percent * 100)}%</span>
        <span className='text-[#616373] ml-auto'>{textStatus}</span>
      </div>
    </div>
  )
}

const FileOperator: FC = () => {
  const { file, reload, closeModal } = useContext(AnalyzeFileContext)!
  const { message } = App.useApp()
  const { status, forest_id, parent_id, file_id, forestPrefix } = file
  switch (status) {
    case 'error':
      return (
        <RefreshIcon
          className='cursor-pointer'
          onClick={async () => {
            await retryParse(file_id)
            message.success('重试成功')
            reload?.()
          }}
        />
      )
    case 'finished': {
      let url = `/docs/${forestPrefix}/${forest_id}`
      if (parent_id) {
        url += `/folder/${parent_id}`
      }
      return (
        <Link to={url} onClick={closeModal}>
          <DirIcon />
        </Link>
      )
    }
    default:
      return null
  }
}
