import { Badge, Button, Image, Input, message } from 'antd'
import {
  CloseCircleOutlined,
  FileImageOutlined,
  LoadingOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { cn, sleep } from '@/utils'
import { uploadImage } from '@/api/common'
import KnowledgeCreate from '@/assets/icons/knowledge-create.svg?react'
import KnowledgeDelete2 from '@/assets/icons/knowledge-delete2.svg?react'
import KnowledgeUpload from '@/assets/icons/knowledge-upload.svg?react'
import { ForestUploadBtn } from '../../../components/ForestUploadBtn'
import type { FileItem } from '../../types'
import { SearchInput } from './SearchInput'
import { UploadButton } from './UploadButton'

export type { FileItem } from '../../types'

interface ActionButtonsProps {
  forest_id: number
  parent_id: number
  refreshTable: () => void
  disabled?: boolean
  uploadLoading: boolean
  onUpload: () => void
  onCreateFolder: () => void
  onBatchDelete: () => void
  onSearch: (values: { value: string; image_url?: string }) => void
  searchKeyword: string
  setSearchKeyword: (value: string) => void
  uploadRef: React.RefObject<HTMLInputElement>
  allowedFileTypes: string
  onFileSelect: (e: React.ChangeEvent<HTMLInputElement>) => void
  onWordCloud?: () => void
  onKnowledgeGraph?: () => void
}

export default function ActionButtons({
  forest_id,
  parent_id,
  refreshTable: refresh,
  disabled,
  uploadLoading,
  onUpload,
  onCreateFolder,
  onBatchDelete,
  onSearch,
  searchKeyword,
  setSearchKeyword,
  uploadRef,
  allowedFileTypes,
  onFileSelect,
  onWordCloud,
  onKnowledgeGraph,
}: ActionButtonsProps) {
  const { t: tM } = useTranslation('messages')
  const { t: tC } = useTranslation('common')
  return (
    <div className={cn('mb-6')}>
      {/* 隐藏的文件上传输入框 */}
      <input
        type='file'
        ref={uploadRef}
        className='hidden'
        accept={allowedFileTypes}
        onChange={onFileSelect}
      />

      {/* 按钮区域 */}
      <div className={cn('flex items-center justify-between mb-4 h-8')}>
        <div className={cn('flex items-center w-full')}>
          <div
            onClickCapture={(e) => {
              if (disabled) {
                message.error(tM('noPermissionContactKbAdmin'))
                e.stopPropagation()
              }
            }}
            className={cn(
              'flex items-center gap-[10px] bg-[#E8F3FF] rounded px-[10px] py-[5px] text-sm border border-[#D6E7FF] h-8',
              {
                grayscale: disabled,
              },
            )}
          >
            <UploadButton
              disabled={disabled}
              forest_id={forest_id}
              parent_id={parent_id}
            ></UploadButton>
            <div className='h-5 w-[1px] bg-[#CCE1FF]'></div>
            <Button
              className={cn(
                'flex items-center !gap-1 !px-0 !h-[22px] !bg-[#E8F3FF] hover:!bg-[#E8F3FF] !border-none !shadow-none !rounded !text-[#4080FF] !font-medium !text-sm !leading-[22px]',
                {
                  '!opacity-40 cursor-not-allowed grayscale': disabled,
                },
              )}
              onClick={onCreateFolder}
            >
              <span>{tC('button.delete', { target: tC('file.folder') })}</span>
              <KnowledgeCreate className='w-4 h-4' />
            </Button>
            <div className='h-5 w-[1px] bg-[#CCE1FF]'></div>
            <Button
              className='flex items-center !gap-1 !px-0 !h-[22px] !bg-[#E8F3FF] !bg-transparent hover:!bg-[#E8F3FF] !border-none !shadow-none !rounded-none !text-[#3D7FFF] !font-medium !text-sm !leading-[22px]'
              onClick={onBatchDelete}
            >
              <span>{tC('button.delete')}</span>
              <KnowledgeDelete2 className='w-4 h-4' />
            </Button>
          </div>

          {/* 搜索框 - 位于按钮右侧 */}
          <div className='ml-12 relative !rounded-sm'>
            <SearchInput
              forest_id={forest_id}
              onSearch={onSearch}
              onChange={setSearchKeyword}
            />
          </div>
          <ForestUploadBtn
            forest_id={forest_id}
            onUploadOne={refresh}
            className='ml-auto'
          />
        </div>
      </div>
    </div>
  )
}

type ImageUploaderProps = {
  onDel: () => void
}
type ImageUploaderRef = {
  getImageUrl: () => string | undefined
}
const ImageUploader = forwardRef<ImageUploaderRef, ImageUploaderProps>(
  (props, ref) => {
    const [src, setSrc] = useState('')
    const [loading, setLoading] = useState(false)
    useImperativeHandle(ref, () => ({ getImageUrl: () => src }))
    if (loading) {
      return <LoadingOutlined />
    }
    if (src) {
      return (
        <Badge
          count={
            <CloseCircleOutlined
              onClick={() => {
                setSrc('')
                props.onDel()
              }}
            />
          }
        >
          <Image src={src} className='w-4 h-4'></Image>
        </Badge>
      )
    }
    const upload = async (files: FileList | null) => {
      const file = files?.[0]
      if (!file) return
      setLoading(true)
      try {
        const { url } = await uploadImage(
          { file, purpose: 'yg-chat' },
          {
            timeout: 0,
            headers: { 'Content-Type': 'multipart/form-data' },
          },
        )
        setSrc(url!)
      } catch {
        setSrc('')
      } finally {
        setLoading(false)
      }
    }
    return (
      <label>
        <FileImageOutlined className=' cursor-pointer' />
        <input
          accept='.png,.jpg,.jepg'
          type='file'
          className='hidden'
          onChange={(e) => upload(e.target.files)}
        ></input>
      </label>
    )
  },
)
