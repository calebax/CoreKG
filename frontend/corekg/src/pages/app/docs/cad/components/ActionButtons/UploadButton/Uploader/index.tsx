import { FC } from 'react'
import { App, Button, Tooltip, Typography } from 'antd'
import { InfoCircleOutlined } from '@ant-design/icons'
import { spliteFileName } from '@/utils'
import { useForestUploadStore } from '@/stores/ForestUploadStore'
import { uploadDir } from '@/utils/Forest'
import { convertFileListToDirectory, loadFile } from '@/utils/loadFile'
import { useDeployConfig } from '@/utils/useDeployConfig'

export type FileTypeConfig = {
  extensions: string[] // 允许的文件扩展名
  mimeTypes: string[] // 允许的MIME类型
  description: string // 格式描述，用于错误提示
}

export type UploaderProps = {
  forest_id: number
  parent_id: number
  acceptedFileTypes?: FileTypeConfig // 新增：文件格式限制配置、
}

export const FilesUploader: FC<UploaderProps> = (props) => {
  const { version } = useDeployConfig()
  const { forest_id, parent_id, acceptedFileTypes } = props
  const upload = useForestUploadStore((state) => state.upload)
  const onLoadFile = async (originFileList: FileList) => {
    const fileList = Array.from(originFileList)
    const filteredFileList = fileList.filter((file) =>
      acceptedFileTypes ? isFileTypeAccepted(file, acceptedFileTypes) : true,
    )
    const files = renameFiles(filteredFileList)
    files.forEach(async (file) => {
      upload(
        { file, forest_id, parent_id, forestPrefix: 'cad' },
        {
          isOversize: (val) => {
            return version === 'saas' && isOversize(val)
          },
        },
      )
    })
  }

  return (
    <Button
      type='text'
      onClick={() =>
        loadFile(onLoadFile, {
          multiple: true,
          accept: 'application/pdf',
        })
      }
      className='justify-between'
      iconPosition='end'
      icon={
        <Tooltip
          title='目前仅支持上传PDF格式图纸，DXF格式即将支持'
          placement='bottom'
        >
          <InfoCircleOutlined className='m-2 mr-4 self-start' />
        </Tooltip>
      }
    >
      上传文件
    </Button>
  )
}

export const DirUploader: FC<UploaderProps> = (props) => {
  const { version } = useDeployConfig()
  const { forest_id, parent_id, acceptedFileTypes } = props
  const upload = useForestUploadStore((state) => state.upload)
  const onLoadFile = async (originFileList: FileList) => {
    const fileList = Array.from(originFileList)
    const filteredFileList = fileList.filter((file) =>
      acceptedFileTypes ? isFileTypeAccepted(file, acceptedFileTypes) : true,
    )
    const dir = convertFileListToDirectory(filteredFileList)
    uploadDir('cad', forest_id, parent_id, dir, (info) =>
      upload(info, {
        isOversize: (val) => {
          return version === 'saas' && isOversize(val)
        },
      }),
    )
  }

  return (
    <Button
      type='text'
      className='justify-between'
      onClick={() => loadFile(onLoadFile, { directory: true })}
    >
      上传文件夹
    </Button>
  )
}

/** 检查单个文件是否符合格式要求 */
const isFileTypeAccepted = (file: File, config: FileTypeConfig): boolean => {
  const fileName = file.name.toLowerCase()
  const fileType = file.type.toLowerCase()

  // 检查扩展名
  const hasValidExtension = config.extensions.some((ext) =>
    fileName.endsWith(ext.toLowerCase()),
  )

  // 检查MIME类型
  const hasValidMimeType =
    config.mimeTypes.length === 0 ||
    config.mimeTypes.some((mime) => fileType === mime.toLowerCase())

  return hasValidExtension && hasValidMimeType
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
  return file.size >= 50 * 1024 * 1024
}
