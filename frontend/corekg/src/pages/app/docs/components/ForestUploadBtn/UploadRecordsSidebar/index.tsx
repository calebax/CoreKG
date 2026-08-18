import { FC, useMemo } from 'react'
import { Button, Drawer } from 'antd'
import { useShallow } from 'zustand/react/shallow'
import { cn } from '@/utils'
import { ForestUploadFiles } from '../upload/ForestUploadFiles'
import { useForestUploadStore } from '@/stores/ForestUploadStore'

export interface UploadRecordsSidebarProps {
  forest_id: number
  open: boolean
  onClose: () => void
  reloadAnalyzeFiles: () => void
}

export const UploadRecordsSidebar: FC<UploadRecordsSidebarProps> = ({
  forest_id,
  open,
  onClose,
  reloadAnalyzeFiles,
}) => {
  const files = useForestUploadStore(
    useShallow((state) => {
      return state.getFilesByOptions({ forest_id })
    }),
  )

  const uploadingNum = useMemo(() => {
    return files.filter((item) => item.status === 'uploading').length
  }, [files])

  const handleCancelAll = () => {
    files.forEach((item) => {
      // 只取消排队中和上传中的文件
      if (item.status === 'waiting' || item.status === 'uploading') {
        item.cancel?.()
      }
    })
  }

  return (
    <Drawer
      title={null}
      placement='right'
      open={open}
      onClose={onClose}
      width={480}
      closable={false}
      styles={{
        body: { padding: 0 },
        header: { padding: 0, border: 'none' },
      }}
    >
      <div className='h-full flex flex-col bg-white relative'>
        {/* 标题区域*/}
        <div className='bg-white border-b border-[#e3e6ed] h-[44px] flex items-center justify-between px-6 py-2.5'>
          <span className='text-base font-medium text-[#0c1f17]'>上传记录</span>
          <div
            className='w-6 h-6 flex items-center justify-center rounded cursor-pointer'
            onClick={onClose}
          >
            <span className='text-sm'>✕</span>
          </div>
        </div>

        {/* 标题下方装饰线*/}
        <div className='absolute top-[42px] left-6 w-[58px] h-[2px] bg-[#0c1f17]' />

        {/* 状态区域 - 只在有文件时显示*/}
        {files.length > 0 && (
          <div className='h-[44px] flex items-center justify-between px-6 py-1.5'>
            <span className='text-sm font-medium text-[#abafb2]'>
              正在上传（{uploadingNum}/{files.length}）
            </span>
            <Button
              type='default'
              size='small'
              onClick={handleCancelAll}
              className='bg-neutral-100 text-[#0c1f17] text-sm font-medium px-[15px] py-[9px] h-auto border-0 cursor-pointer rounded'
            >
              取消全部
            </Button>
          </div>
        )}

        {/* 文件列表区域 */}
        <div className='flex-1'>
          <ForestUploadFiles
            forest_id={forest_id}
            onUploadOne={reloadAnalyzeFiles}
            closeModal={onClose}
            className='h-full'
          />
        </div>
      </div>
    </Drawer>
  )
}