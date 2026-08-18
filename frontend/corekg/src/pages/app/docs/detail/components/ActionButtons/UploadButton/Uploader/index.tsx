import { FC, useCallback } from 'react'
import { App, Button, Upload, Popover, Tooltip } from 'antd'
import { InfoCircleOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { spliteFileName, cn } from '@/utils'
import FileIcon from '@/assets/icons/docs/file.svg?react'
import FolderIcon from '@/assets/icons/docs/fileList.svg?react'
import UploadIcon from '@/assets/icons/docs/upload-file.svg?react'
import { useForestUploadStore } from '@/stores/ForestUploadStore'
import { uploadDir } from '@/utils/Forest'
import { convertFileListToDirectory, loadFile } from '@/utils/loadFile'
import { useDeployConfig } from '@/utils/useDeployConfig'

export type KnowledgeBaseType = 'file' | 'excel' | 'qa' | 'data'

export type UploaderProps = {
  forest_id: number
  parent_id: number
  afterUpload?: () => void
  knowledgeBaseType?: KnowledgeBaseType
  disabled?: boolean
}

// 文件格式配置
const FILE_FORMATS: Record<
  KnowledgeBaseType,
  { accept: string; extensions: string[]; tooltip: string }
> = {
  file: {
    // 多模态知识库
    accept: '.pdf,.md,.txt,.doc,.docx,.ppt,.pptx,.jpg,.png,.jpeg,.mp4,.ofd',
    extensions: [
      'pdf',
      'md',
      'txt',
      'doc',
      'docx',
      'ppt',
      'pptx',
      'jpg',
      'png',
      'jpeg',
      'mp4',
      'ofd',
    ],
    tooltip: 'PDF、MD、TXT、DOC、DOCX、PPT、PPTX、JPG、PNG、JPEG、MP4、OFD',
  },
  excel: {
    // 表格知识库
    accept: '.csv,.xls,.xlsx',
    extensions: ['csv', 'xls', 'xlsx'],
    tooltip: 'CSV、XLS、XLSX',
  },
  qa: {
    // 问答对知识库
    accept: '.csv,.xlsx',
    extensions: ['csv', 'xlsx'],
    tooltip: 'CSV、XLSX',
  },
  data: {
    // 数据库类型
    accept: '.csv,.xlsx',
    extensions: ['csv', 'xlsx'],
    tooltip: 'CSV、XLSX',
  },
}

// 文件上传器
export const CombinedUploader: FC<UploaderProps> = (props) => {
  const {
    forest_id,
    parent_id,
    afterUpload,
    knowledgeBaseType = 'file',
    disabled,
  } = props
  const upload = useForestUploadStore((state) => state.upload)
  const { message } = App.useApp()
  const { t } = useTranslation('pages')
  const { t: tM } = useTranslation('messages')
  const { t: tC } = useTranslation('common')
  const { version } = useDeployConfig()
  // 获取文件格式配置
  const formatConfig = FILE_FORMATS[knowledgeBaseType] || FILE_FORMATS.file

  // 检查文件格式是否允许
  const isFileAllowed = (fileName: string): boolean => {
    const { name, ext } = spliteFileName(fileName)
    if (!ext) return false
    return formatConfig.extensions.includes(ext.toLowerCase())
  }

  // 处理文件上传
  const handleFilesUpload = useCallback(
    async (originFileList: FileList | File[]) => {
      // 使用默认的分段规则
      const defaultSegmentRule = {
        type: 'auto' as const,
      }

      const fileList = Array.from(originFileList)

      // 预检查大小，若有任一超出限制，提示一次
      if (version === 'saas') {
        const hasOversize = fileList.some((f) => isOversize(f))
        if (hasOversize) {
          message.error(tM('uploadOversizeFixedTip'))
        }
      }

      // 过滤不允许的文件
      const allowedFiles: File[] = []
      const rejectedFiles: string[] = []

      fileList.forEach((file) => {
        if (isFileAllowed(file.name)) {
          // 在 saas 版本下，同时检查文件大小
          if (version === 'saas' && isOversize(file)) {
            // 超限文件不加入 allowedFiles，这样就不会被上传
            return
          }
          allowedFiles.push(file)
        } else {
          rejectedFiles.push(file.name)
        }
      })

      // 提示被拒绝的文件
      if (rejectedFiles.length > 0) {
        message.warning(
          `以下文件格式不支持：${rejectedFiles.join(', ')}。当前知识库仅支持 ${formatConfig.tooltip} 格式`,
        )
      }

      if (allowedFiles.length === 0) {
        return
      }

      // 重命名超长文件名
      const files = renameFiles(allowedFiles)

      // 文件上传
      let uploadCount = files.length
      let successCount = 0

      files.forEach((file) => {
        upload(
          {
            file,
            forest_id,
            parent_id,
            forestPrefix: 'detail',
            segmentRule: defaultSegmentRule,
            onSuccess: () => {
              successCount++
              if (successCount === uploadCount) {
                message.success(
                  tM('uploadSuccessFileCount', { target: successCount }),
                )
                afterUpload?.()
              }
            },
            onError: (error) => {
              uploadCount--
              if (successCount === uploadCount) {
                if (successCount > 0) {
                  message.success(`成功上传 ${successCount} 个文件`)
                  afterUpload?.()
                }
              }
            },
          },
          {
            isOversize: (val) => {
              return version === 'saas' && isOversize(val)
            },
          },
        )
      })
    },
    [
      upload,
      forest_id,
      parent_id,
      afterUpload,
      message,
      formatConfig,
      isFileAllowed,
    ],
  )

  return (
    <Upload
      multiple
      showUploadList={false}
      beforeUpload={(file, fileList) => {
        // 处理所有文件
        if (fileList.length > 0 && fileList[0] === file) {
          // 只在第一个文件时处理整个列表，避免重复
          handleFilesUpload(fileList)
        }
        return false // 阻止默认上传行为
      }}
      accept={formatConfig.accept}
      disabled={disabled}
    >
      <Button
        type='default'
        className='flex items-center rounded-md gap-1 bg-[#FAE8FF] text-[#CC5DE8] text-sm font-medium border-none hover:border-none px-2.5 py-2.5'
        icon={<UploadIcon className='w-[14px] h-[14px]' />}
        disabled={disabled}
      >
        {t('app.docs.detail.uploadFile')}
      </Button>
    </Upload>
  )
}

// 带下拉选择（上传文件/上传文件夹）的上传按钮
export const CombinedUploaderWithMenu: FC<UploaderProps> = (props) => {
  const {
    forest_id,
    parent_id,
    afterUpload,
    knowledgeBaseType = 'file',
    disabled,
  } = props
  const upload = useForestUploadStore((state) => state.upload)
  const { message } = App.useApp()
  const { t } = useTranslation('pages')
  const { t: tM } = useTranslation('messages')
  const { t: tC } = useTranslation('common')
  const { version } = useDeployConfig()
  const formatConfig = FILE_FORMATS[knowledgeBaseType] || FILE_FORMATS.file

  const isFileAllowed = (fileName: string): boolean => {
    const { ext } = spliteFileName(fileName)
    if (!ext) return false
    return formatConfig.extensions.includes(ext.toLowerCase())
  }

  const handleFilesUpload = useCallback(
    async (originFileList: FileList) => {
      const defaultSegmentRule = { type: 'auto' as const }
      const fileList = Array.from(originFileList)

      // 预检查大小，若有任一超出限制，提示一次
      if (version === 'saas') {
        const hasOversize = fileList.some((f) => isOversize(f))
        if (hasOversize) {
          message.error(tM('uploadOversizeFixedTip'))
        }
      }

      const allowedFiles: File[] = []
      const rejectedFiles: string[] = []
      fileList.forEach((file) => {
        if (isFileAllowed(file.name)) {
          // 在 saas 版本下，同时检查文件大小
          if (version === 'saas' && isOversize(file)) {
            // 超限文件不加入 allowedFiles，这样就不会被上传
            return
          }
          allowedFiles.push(file)
        } else {
          rejectedFiles.push(file.name)
        }
      })
      if (rejectedFiles.length > 0) {
        message.warning(tM('uploadSupport', { target: formatConfig.tooltip }))
      }
      if (allowedFiles.length === 0) return
      const files = renameFiles(allowedFiles)
      let uploadCount = files.length
      let successCount = 0
      files.forEach((file) => {
        upload(
          {
            file,
            forest_id,
            parent_id,
            forestPrefix: 'detail',
            segmentRule: defaultSegmentRule,
            onSuccess: (response) => {
              successCount++
              if (successCount === uploadCount) {
                // 如果接口返回的code为20001，则不进行成功提示
                if (response?.code === 20001) {
                  return
                }
                message.success(
                  tM('uploadSuccessFileCount', { target: successCount }),
                )
                afterUpload?.()
              }
            },
            onError: () => {
              uploadCount--
              if (successCount === uploadCount && successCount > 0) {
                message.success(
                  tM('uploadSuccessFileCount', { target: successCount }),
                )
                afterUpload?.()
              }
            },
          },
          {
            isOversize: (val) => {
              return version === 'saas' && isOversize(val)
            },
          },
        )
      })
    },
    [upload, forest_id, parent_id, afterUpload, message, formatConfig],
  )

  const handleDirUpload = useCallback(
    async (originFileList: FileList) => {
      const fileList = Array.from(originFileList)
      const filtered = fileList.filter((f) => isFileAllowed(f.name))
      const dir = convertFileListToDirectory(filtered)
      const totalFiles = filtered.length
      if (totalFiles === 0) return
      let successCount = 0
      const defaultSegmentRule = { type: 'auto' as const }
      uploadDir('detail', forest_id, parent_id, dir, (info) =>
        upload(
          {
            ...info,
            forestPrefix: 'detail',
            segmentRule: defaultSegmentRule,
            onSuccess: () => {
              successCount++
              if (successCount === totalFiles) {
                message.success(
                  tM('folderUploadSuccessFileCount', { target: successCount }),
                )
                afterUpload?.()
              }
            },
          },
          {
            isOversize: (val) => {
              return version === 'saas' && isOversize(val)
            },
          },
        ),
      )
    },
    [upload, forest_id, parent_id, afterUpload, message],
  )

  const content = (
    <div className='p-2.5 min-w-[230px] rounded-lg shadow-lg border border-gray-200 bg-white mt-1'>
      <div
        className='px-[5px] py-[7px] flex items-center gap-2 rounded hover:bg-[#F7F7F7] cursor-pointer transition-colors'
        onClick={() =>
          loadFile((fl) => handleFilesUpload(fl), {
            multiple: true,
            accept: formatConfig.accept,
          })
        }
      >
        <FileIcon className='w-[18px] h-[18px] text-[#2D2D2D]' />
        <div className='flex-1 text-sm text-[#2D2D2D]'>
          {tC('button.upload', { target: tC('file.file') })}
        </div>
        <Tooltip
          title={t('app.docs.supportUploadParse', {
            fileType: formatConfig.tooltip,
            size: '100M',
          })}
          placement='bottom'
        >
          <InfoCircleOutlined className='text-[#9CA3AF]' />
        </Tooltip>
      </div>
      <div
        className='px-[5px] py-[7px] flex items-center gap-2 rounded hover:bg-[#F7F7F7] cursor-pointer transition-colors'
        onClick={() =>
          loadFile((fl) => handleDirUpload(fl), { directory: true })
        }
      >
        <FolderIcon className='w-[18px] h-[18px] text-[#2D2D2D]' />
        <div className='flex-1 text-sm text-[#2D2D2D]'>
          {tC('button.upload', { target: tC('file.folder') })}
        </div>
      </div>
    </div>
  )

  return (
    <Popover
      content={disabled ? null : content}
      placement='bottomLeft'
      arrow={false}
    >
      <Button
        type='default'
        className={cn(
          'flex items-center rounded-md h-[30px] gap-1 bg-white text-[#0C99FF] text-sm font-medium border-[#0C99FF] hover:border-[#0C99FF] px-2.5 py-2',
          {
            'opacity-40 cursor-not-allowed grayscale': disabled,
          },
        )}
        icon={<UploadIcon className='w-[14px] h-[14px]' />}
        disabled={disabled}
      >
        {t('app.docs.detail.uploadFile')}
      </Button>
    </Popover>
  )
}

/** 将超长名称截断 */
const renameFiles = (fileList: File[]): File[] => {
  const files = Array.from(fileList)
  const renamedFiles = files.map((file) => {
    const { name, ext } = spliteFileName(file.name)
    const newName = name.length > 50 ? name.slice(0, 50) : name
    const newFullName = ext ? newName + '.' + ext : newName
    return new File([file], newFullName, { type: file.type })
  })
  return renamedFiles
}

const isOversize = (file: File) => {
  // 视频不超过500MB 其他100
  const maxMB = file.type.startsWith('video/') ? 500 : 100
  return file.size >= maxMB * 1024 * 1024
}
