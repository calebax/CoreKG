import { FC } from 'react'
import { Button } from 'antd'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import UploadIcon from '@/assets/icons/docs/upload-file.svg?react'
import { useQuotaLimitModal } from '@/hooks/useQuotaLimitModal'
import { useVersion } from '@/utils/useVersion'
import {
  CombinedUploaderWithMenu as CombinedUploader,
  UploaderProps,
} from './Uploader'
import type { KnowledgeBaseType } from './Uploader'

export type { FileItem } from '..'

type UploadButtonProps = {
  disabled?: boolean
  forest_id: number
  parent_id: number
  afterUpload?: () => void
  knowledgeBaseType?: KnowledgeBaseType
}

export const UploadButton: FC<UploadButtonProps> = (props) => {
  const {
    disabled,
    forest_id,
    parent_id,
    afterUpload,
    knowledgeBaseType = 'file',
  } = props
  const { t } = useTranslation('pages')
  const uploaderProps: UploaderProps = {
    forest_id,
    parent_id,
    afterUpload,
    knowledgeBaseType,
  }

  const { version } = useVersion()
  const { show: showQuotaLimitModal } = useQuotaLimitModal()
  const isQuotaLimited = version && version.disk.use_ratio >= 1

  if (isQuotaLimited) {
    return (
      <Button
        type='default'
        className={cn(
          'flex gap-1 rounded-md items-center bg-white text-[#0C99FF] text-sm font-medium border-[#0C99FF] hover:border-[#0C99FF] px-2.5 py-2.5',
          {
            'opacity-40 cursor-not-allowed grayscale':
              isQuotaLimited || disabled,
          },
        )}
        icon={<UploadIcon className='w-[14px] h-[14px]' />}
        onClick={() => {
          showQuotaLimitModal({ type: 'knowledge' })
        }}
      >
        {t('app.docs.detail.uploadFile')}
      </Button>
    )
  }

  return <CombinedUploader {...uploaderProps} disabled={disabled} />
}
