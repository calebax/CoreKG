import { FC } from 'react'
import { Button, Popover } from 'antd'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import KnowledgeUpload from '@/assets/icons/knowledge-upload.svg?react'
import { useQuotaLimitModal } from '@/hooks/useQuotaLimitModal'
import { useVersion } from '@/utils/useVersion'
import { DirUploader, FilesUploader, UploaderProps } from './Uploader'

export type { FileItem } from '..'

type UploadButtonProps = {
  disabled?: boolean
  forest_id: number
  parent_id: number
}
export const UploadButton: FC<UploadButtonProps> = (props) => {
  const { t } = useTranslation('pages')
  const { t: tC } = useTranslation('common')

  const { disabled, forest_id, parent_id } = props
  const uploaderProps: UploaderProps = {
    forest_id,
    parent_id,
  }
  const content = (
    <div className='flex flex-col p-2'>
      <FilesUploader {...uploaderProps}></FilesUploader>
      <DirUploader {...uploaderProps}></DirUploader>
    </div>
  )
  const btnClass =
    'flex items-center px-0 gap-1 h-[22px] bg-transparent hover:bg-[#E6F2FF] border-none rounded-[6px] text-[#3D7FFF] font-normal text-sm leading-[22px]'
  const { version } = useVersion()
  const { show: showQuotaLimitModal } = useQuotaLimitModal()
  const isQuotaLimited = version && version.disk.use_ratio >= 1

  if (isQuotaLimited) {
    return (
      <Button
        className={cn(btnClass, {
          'opacity-40 cursor-not-allowed grayscale': isQuotaLimited || disabled,
        })}
        onClick={() => {
          showQuotaLimitModal({ type: 'knowledge' })
        }}
      >
        {tC('button.upload')}
        <KnowledgeUpload className='w-4 h-4' />
      </Button>
    )
  }

  return (
    <Popover
      content={disabled ? null : content}
      placement='bottomLeft'
      arrow={false}
    >
      <Button
        className={cn(btnClass, {
          'opacity-40': disabled,
        })}
        disabled={disabled}
      >
        {tC('button.upload')}
        <KnowledgeUpload className='w-4 h-4' />
      </Button>
    </Popover>
  )
}
