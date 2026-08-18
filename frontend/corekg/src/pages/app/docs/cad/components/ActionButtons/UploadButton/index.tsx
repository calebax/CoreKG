import { FC } from 'react'
import { Button, Popover } from 'antd'
import KnowledgeUpload from '@/assets/icons/knowledge-upload.svg?react'
import {
  DirUploader,
  FilesUploader,
  UploaderProps,
  FileTypeConfig,
} from './Uploader'

export type { FileItem } from '..'

type UploadButtonProps = {
  disabled?: boolean
  forest_id: number
  parent_id: number
}

export const UploadButton: FC<UploadButtonProps> = (props) => {
  const { disabled, forest_id, parent_id } = props

  // 定义允许的文件格式
  const acceptedFileTypes: FileTypeConfig = {
    extensions: ['.pdf'],
    mimeTypes: ['application/pdf'],
    description: 'PDF格式图纸',
  }

  const uploaderProps: UploaderProps = {
    forest_id,
    parent_id,
    acceptedFileTypes,
  }

  const content = (
    <div className='flex flex-col p-2'>
      <FilesUploader {...uploaderProps}></FilesUploader>
      <DirUploader {...uploaderProps}></DirUploader>
    </div>
  )

  const btnClass =
    'flex items-center !gap-1 !py-2.5 !h-10 !bg-[#E8F3FF] hover:!bg-[#E8F3FF] !border-none !rounded-full !text-[#4080FF] !font-medium'

  return (
    <Popover
      content={disabled ? null : content}
      placement='bottomLeft'
      arrow={false}
    >
      <Button className={btnClass}>
        上传
        <KnowledgeUpload className='w-4 h-4' />
      </Button>
    </Popover>
  )
}
