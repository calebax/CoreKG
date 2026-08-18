import { FC, useState } from 'react'
import { Image } from 'antd'
import { EyeOutlined, CloseCircleFilled, LoadingOutlined } from '@ant-design/icons'
import { cn } from '@/utils'
import { getFileIcon as getKnowledgeFileIcon } from '@/components/common/ReferencePreviewModal/utils'
import { UniversalFilePreviewModal } from '@/components/common/UniversalFilePreviewModal'
import type { Attachment } from '../UserDialog'

interface AttachmentListProps {
  attachments: Attachment[]
  isCompact?: boolean
  className?: string
  onRemove?: (index: number) => void
  canRemove?: boolean
}

export const AttachmentList: FC<AttachmentListProps> = ({
  attachments,
  isCompact,
  className,
  onRemove,
  canRemove = false,
}) => {
  const [previewFile, setPreviewFile] = useState<Attachment | null>(null)
  const [previewImage, setPreviewImage] = useState<string | undefined>()

  if (!attachments?.length) return null

  const handleFileClick = (file: Attachment) => {
    if (file.loading || !file.url) return
    
    // 图片类型使用 Image 组件的预览功能
    if (file.type === 'image') {
      setPreviewImage(file.url)
    } else {
      // 非图片类型使用通用预览弹窗
      setPreviewFile(file)
    }
  }

  return (
    <div
      className={cn(
        'flex flex-wrap gap-2 mb-1',
        {
          'justify-end': !isCompact,
        },
        className,
      )}
    >
      {attachments.map((file, index) => (
        <div
          key={index}
          className={cn(
            'flex items-center gap-2 bg-[#F7F7F7] border border-[#EFF1F4] rounded-lg px-2 py-1.5',
            'hover:bg-[#F0F0F0] transition-colors cursor-pointer',
            'max-w-[200px] group relative',
          )}
          onClick={() => handleFileClick(file)}
        >
          <div className='flex-shrink-0 w-8 h-8 rounded overflow-hidden bg-white flex items-center justify-center relative'>
            {file.type === 'image' && file.url ? (
              <Image
                src={file.url}
                alt={file.name}
                className='w-full h-full object-cover'
                preview={{
                  visible: previewImage === file.url,
                  onVisibleChange: (visible) => {
                    if (!visible) {
                      setPreviewImage(undefined)
                    }
                  },
                  mask: (
                    <div className='flex flex-col items-center justify-center text-white'>
                      <EyeOutlined className='text-xs mb-0.5' />
                      <span className='text-[10px] scale-90'>预览</span>
                    </div>
                  ),
                }}
              />
            ) : (
              <div className='w-full h-full flex items-center justify-center'>
                {getKnowledgeFileIcon(file.name) || (
                  <span className='text-lg'>📎</span>
                )}
              </div>
            )}
          </div>
          <span className='text-xs text-[#0C1F17] truncate flex-1'>
            {file.name}
          </span>

          {(file.loading || canRemove) && (
            <div className='relative shrink-0 size-4 ml-1 flex items-center justify-center'>
              {file.loading && (
                <LoadingOutlined className='text-blue-500 text-sm group-hover:invisible' />
              )}
              {canRemove && (
                <CloseCircleFilled
                  className={cn(
                    'absolute text-[#919497] hover:text-[#616373] cursor-pointer',
                    file.loading
                      ? 'invisible group-hover:visible'
                      : 'opacity-0 group-hover:opacity-100',
                  )}
                  onClick={(e) => {
                    e.stopPropagation()
                    onRemove?.(index)
                  }}
                />
              )}
            </div>
          )}
        </div>
      ))}

      {/* 非图片预览 Modal */}
      {previewFile && (
        <UniversalFilePreviewModal
          visible={!!previewFile}
          onClose={() => setPreviewFile(null)}
          id={previewFile.id}
          url={previewFile.url!}
          fileName={previewFile.name}
          fileType={previewFile.type}
        />
      )}
    </div>
  )
}
