import { FC } from 'react'
import { App, Button, Tooltip } from 'antd'
import { InfoCircleOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { spliteFileName } from '@/utils'
import { useForestUploadStore } from '@/stores/ForestUploadStore'
import { uploadDir } from '@/utils/Forest'
import { convertFileListToDirectory, loadFile } from '@/utils/loadFile'
import { useDeployConfig } from '@/utils/useDeployConfig'

const legalFileExt = ['csv', 'xls', 'xlsx']
export type UploaderProps = {
  forest_id: number
  parent_id: number
  afterUpload?: () => void
}

export const FilesUploader: FC<UploaderProps> = (props) => {
  const { version } = useDeployConfig()
  const { t: tM } = useTranslation('messages')
  const { t: tC } = useTranslation('common')
  const { t } = useTranslation('pages')
  const { forest_id, parent_id, afterUpload } = props
  const upload = useForestUploadStore((state) => state.upload)
  const { message } = App.useApp()
  const onLoadFile = async (originFileList: FileList) => {
    const fileList = Array.from(originFileList)
    // 预检查大小，若有任一超出限制，提示一次
    if (version === 'saas') {
      const hasOversize = fileList.some((f) => isOversize(f))
      if (hasOversize) {
        message.error(tM('uploadOversizeFixedTip'))
      }
    }
    // 非法excel文件过滤不提示，同时过滤超限文件
    const filteredFileList = fileList.filter((file) => {
      if (!isLegalExcel(file)) return false
      // 在 saas 版本下，同时检查文件大小
      if (version === 'saas' && isOversize(file)) {
        return false // 超限文件不通过过滤
      }
      return true
    })
    const files = renameFiles(filteredFileList)
    const uploadCount = files.length
    let successCount = 0
    files.forEach((file) => {
      upload(
        {
          file,
          forest_id,
          parent_id,
          forestPrefix: 'excel',
          onSuccess: () => {
            successCount++
            if (successCount === uploadCount) {
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
  }
  return (
    <Button
      type='text'
      onClick={() =>
        loadFile(onLoadFile, {
          multiple: true,
          accept: legalFileExt.map((s) => `.${s}`).join(','),
        })
      }
      className='justify-between'
      iconPosition='end'
      icon={
        <Tooltip
          title={t('app.docs.supportUploadParse', {
            fileType: legalFileExt.join('、'),
            size: '100M',
          })}
          placement='bottom'
        >
          <InfoCircleOutlined className='m-2 mr-4 self-start' />
        </Tooltip>
      }
    >
      {tC('button.upload', { target: tC('file.excel') })}
    </Button>
  )
}

export const DirUploader: FC<UploaderProps> = (props) => {
  const { version } = useDeployConfig()
  const { t: tM } = useTranslation('messages')
  const { t: tC } = useTranslation('common')
  const { forest_id, parent_id, afterUpload } = props
  const upload = useForestUploadStore((state) => state.upload)
  const { message } = App.useApp()
  const onLoadFile = async (originFileList: FileList) => {
    const fileList = Array.from(originFileList)
    // 非法excel文件过滤不提示
    // 预检查大小，若有任一超出限制，提示一次
    if (version === 'saas') {
      const hasOversize = fileList.some((f) => isOversize(f))
      if (hasOversize) {
        message.error(tM('uploadOversizeFixedTip'))
      }
    }
    const dir = convertFileListToDirectory(fileList.filter((file) => {
      if (!isLegalExcel(file)) return false
      // 在 saas 版本下，同时检查文件大小
      if (version === 'saas' && isOversize(file)) {
        return false // 超限文件不通过过滤
      }
      return true
    }))
    const totalFiles = fileList.length
    let successCount = 0
    uploadDir('excel', forest_id, parent_id, dir, (info) =>
      upload(
        {
          ...info,
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
  }
  return (
    <Button
      type='text'
      className='justify-between'
      onClick={() => loadFile(onLoadFile, { directory: true })}
    >
      {tC('button.upload', { target: tC('file.folder') })}
    </Button>
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
  return file.size >= 100 * 1024 * 1024
}

const isLegalExcel = (f: File) => {
  const { ext } = spliteFileName(f.name)
  return ext ? legalFileExt.includes(ext.toLowerCase()) : false
}
