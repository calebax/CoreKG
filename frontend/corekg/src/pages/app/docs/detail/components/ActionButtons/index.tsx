import { Button } from 'antd'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import FolderIcon from '@/assets/icons/docs/folder.svg?react'
import type { FileItem } from '../../types'
import { UploadButton } from './UploadButton'
import type { KnowledgeBaseType } from './UploadButton/Uploader'

export type { FileItem } from '../../types'

interface ActionButtonsProps {
  forest_id: number
  parent_id: number
  refreshTable: () => void
  disabled?: boolean
  onCreateFolder: () => void
  knowledgeBaseType?: KnowledgeBaseType
}

export default function ActionButtons({
  forest_id,
  parent_id,
  refreshTable: refresh,
  disabled,
  onCreateFolder,
  knowledgeBaseType = 'file',
}: ActionButtonsProps) {
  const { t } = useTranslation('pages')

  return (
    <div>
      {/* 按钮区域 - 移到右侧 */}
      <div className='flex items-center justify-end gap-3'>
        <UploadButton
          disabled={disabled}
          forest_id={forest_id}
          parent_id={parent_id}
          afterUpload={refresh}
          knowledgeBaseType={knowledgeBaseType}
        />

        <Button
          type='default'
          className={cn(
            'flex items-center leading-none rounded-md gap-1 bg-[#FAE8FF] text-[#CC5DE8] text-sm font-medium border-none hover:border-none px-2.5 py-2.5',
            {
              'opacity-40 cursor-not-allowed grayscale': disabled,
            },
          )}
          icon={<FolderIcon className='w-[14px] h-[14px]' />}
          onClick={onCreateFolder}
          disabled={disabled}
        >
          {t('app.docs.detail.newFolder')}
        </Button>
      </div>
    </div>
  )
}
