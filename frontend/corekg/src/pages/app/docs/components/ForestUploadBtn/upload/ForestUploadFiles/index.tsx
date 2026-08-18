import { createContext, FC } from 'react'
import { Button, Empty, Progress } from 'antd'
import { useShallow } from 'zustand/react/shallow'
import { cn, formatFileSize } from '@/utils'
import { FileItem, useForestUploadStore, type UploadBaseInfo } from '@/stores/ForestUploadStore'
import DirIcon from './dir.svg?react'
import StopIcon from './stop.svg?react'
import RefreshIcon from '../ForestAnalyzeFiles/refresh.svg?react'

export type ForestUploadFiles = Style & {
  forest_id: number
  onUploadOne: () => void

  closeModal?: () => void
}
export const ForestUploadFiles: FC<ForestUploadFiles> = (props) => {
  const { forest_id, onUploadOne, closeModal, className, style } = props
  const files = useForestUploadStore(
    useShallow((state) => {
      return state.getFilesByOptions({ forest_id })
    }),
  )
  const uploadingNum = useMemo(() => {
    return files.filter((item) => item.status === 'uploading').length
  }, [files])
  // 在一个文件上传完毕后 立刻加载解析列表
  const prevUploadingNum = useRef(uploadingNum)
  if (prevUploadingNum.current > uploadingNum) {
    onUploadOne()
  }
  prevUploadingNum.current = uploadingNum

  if (files.length === 0) {
    return <Empty className={cn('mt-10', className)} style={style} />
  }

  return (
    <div
      className={cn('h-full flex flex-col overflow-hidden', className)}
      style={style}
    >
      <div className='flex-1 flex flex-col overflow-auto'>
        {files.map((item) => {
          return (
            <AnalyzeUploadContext.Provider
              value={{ item, closeModal }}
              key={item.key}
            >
              <UploadFileItem />
            </AnalyzeUploadContext.Provider>
          )
        })}
      </div>
    </div>
  )
}
const AnalyzeUploadContext = createContext<{
  item: FileItem
  closeModal?: () => void
} | null>(null)

const UploadFileItem: FC = () => {
  const { item } = useContext(AnalyzeUploadContext)!
  const {
    status,
    file,
    percent = 0,
    speed = 0,
    restTime,
    illeagalReason,
  } = item
  const strokeColor = useMemo(() => {
    switch (status) {
      case 'uploading':
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
      case 'uploading':
        return `上传中（${formatFileSize(speed)}/s） · 预计剩余${formatTime(restTime)}`
      case 'waiting':
        return '排队中'
      case 'error':
        return '已中断'
      case 'finished':
        return '已完成'
      case 'illegal':
        return illeagalReason
    }
  }, [illeagalReason, restTime, speed, status])
  return (
    <div className='px-6 py-2.5 flex flex-col gap-1.5 hover:bg-[#F8F9FD]'>
      <div className='flex items-center justify-between pr-[25px]'>
        <span
          className='text-base font-medium text-[#6e757f] truncate flex-1 mr-4'
          title={file.name}
        >
          {file.name}
        </span>
        <span className='text-xs font-medium text-[#abafb2] leading-[16px] flex-shrink-0'>
          {formatFileSize(file.size)}
        </span>
      </div>
      <div className='flex items-center gap-4'>
        <div className='flex-1 flex flex-col gap-1.5'>
          <div className='relative'>
            <div className='bg-[#dfe0eb] h-[6px] rounded-xl w-full' />
            <div
              className='absolute top-0 left-0 h-[6px] rounded-lg'
              style={{
                width: `${percent * 100}%`,
                backgroundColor: strokeColor
              }}
            />
          </div>
        </div>
        <div className='w-4 h-4 flex items-center justify-center'>
          <FileOperator />
        </div>
      </div>
      <div className='flex items-center justify-between pr-6'>
        <span className='text-xs font-medium text-[#abafb2] leading-[16px]'>{Math.floor(percent * 100)}%</span>
        <span className='text-xs font-medium text-[#abafb2] leading-[16px]'>{textStatus}</span>
      </div>
    </div>
  )
}

/**
 * 格式化时间显示\
 */
const formatTime = (seconds?: number): string => {
  if (!seconds || seconds < 0) {
    return '-'
  }
  if (seconds < 59) return `${Math.ceil(seconds)}秒`

  const min = seconds / 60
  if (min < 60) return `${Math.ceil(min)}分钟`
  const hours = min / 60
  if (hours < 24) return `${Math.ceil(hours)}小时`
  const days = hours / 24
  return `${Math.ceil(days)}天`
}

const FileOperator: FC = () => {
  const { item, closeModal } = useContext(AnalyzeUploadContext)!
  const { status, cancel, key, forest_id, parent_id, forestPrefix } = item
  const retryUpload = useForestUploadStore(state => state.retryUpload)

  switch (status) {
    case 'waiting':
    case 'uploading':
      return <StopIcon className=' cursor-pointer' onClick={cancel} />
    case 'finished': {
      // let url = `/docs/${forestPrefix}/${forest_id}`
      // if (parent_id) {
      //   url += `/folder/${parent_id}`
      // }
      // 跳转到单文件详情页
      const fileId = item.file_id
      if (!fileId) {
        return <DirIcon />
      }
      const url = `/docs/detail/${forest_id}/file/${fileId}`
      return (
        <Link to={url} onClick={closeModal}>
          <DirIcon />
        </Link>
      )
    }
    case 'error':
      return (
        <RefreshIcon
          className='cursor-pointer'
          onClick={() => {
            retryUpload(key)
          }}
        />
      )
  }
}
