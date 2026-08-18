import { Badge, Button, Image, Input, message } from 'antd'
import {
  CloseCircleOutlined,
  FileImageOutlined,
  LoadingOutlined,
} from '@ant-design/icons'
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
      <div className={cn('flex items-center justify-between mb-4')}>
        <div className={cn('flex items-center')}>
          <div
            onClickCapture={(e) => {
              if (disabled) {
                message.error('无权限，请联系知识库管理员')
                e.stopPropagation()
              }
            }}
            className={cn('flex items-center bg-[#E8F3FF] rounded text-sm', {
              grayscale: disabled,
            })}
          >
            <UploadButton
              disabled={disabled}
              forest_id={forest_id}
              parent_id={parent_id}
            ></UploadButton>
            <div className='h-5 w-[1px] bg-[#4080FF]'></div>
            <Button
              className='flex items-center !gap-1 !py-2.5 !h-10 !bg-[#E8F3FF] hover:!bg-[#E8F3FF] !border-none !rounded-full !text-[#4080FF] !font-medium'
              onClick={onCreateFolder}
            >
              <span>新建文件夹</span>
              <KnowledgeCreate className='w-4 h-4' />
            </Button>
            <div className='h-5 w-[1px] bg-[#4080FF]'></div>
            <Button
              className='flex items-center !gap-1 !py-2.5 !h-10 !bg-[#E8F3FF] hover:!bg-[#E8F3FF] !border-none !rounded-full  !text-[#4080FF] !font-medium'
              onClick={onBatchDelete}
            >
              <span>删除</span>
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
        </div>
        <ForestUploadBtn
          forest_id={forest_id}
          onUploadOne={refresh}
          className='ml-auto mr-4'
        />
        {/* 词云和知识图谱按钮 - 位于右侧，改为类似knownow项目的Tab样式 */}
        <div className='flex items-center overflow-hidden rounded-md border border-black/10 text-[rgba(0,0,0,0.88)]'>
          <button
            className='cursor-pointer px-4 py-2 text-sm transition-colors duration-300 text-[#165DFF] font-medium hover:text-white hover:bg-[#165DFF] border-none bg-black/[0.02] font-medium'
            onClick={onWordCloud}
          >
            词云
          </button>
          <div className='h-[22px] w-[2px] bg-black/10'></div>
          <button
            className='cursor-pointer px-4 py-2 text-sm transition-colors duration-300 text-[#165DFF] font-medium hover:text-white hover:bg-[#165DFF] border-none bg-black/[0.02] font-medium'
            onClick={onKnowledgeGraph}
          >
            知识图谱
          </button>
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
        const { url } = (await uploadImage(
          { file, purpose: 'yg-chat' },
          {
            timeout: 0,
            headers: { 'Content-Type': 'multipart/form-data' },
          },
        )) as unknown as { url: string }
        setSrc(url)
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
