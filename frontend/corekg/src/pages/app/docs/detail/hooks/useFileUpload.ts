import { useState, useRef } from 'react'
import { message } from 'antd'
import { uploadFile } from '@/api/knowledge'
import { useQuotaLimitModal } from '@/hooks/useQuotaLimitModal'
import { useVersion } from '@/utils/useVersion'

interface UseFileUploadProps {
  knowledgeBaseId: number
  parentId: number
  onSuccess?: () => void
}

export const useFileUpload = ({
  knowledgeBaseId,
  parentId,
  onSuccess,
}: UseFileUploadProps) => {
  const [uploadLoading, setUploadLoading] = useState<boolean>(false)
  const uploadRef = useRef<HTMLInputElement>(null)
  const { version, refresh: refreshVersion } = useVersion()
  const { show: showQuotaLimitModal } = useQuotaLimitModal()

  // 允许上传的文件类型
  // const allowedFileTypes = '.doc,.docx,.xls,.xlsx,.ppt,.pptx,.csv,.pdf,.png,.jpg,.jpeg,.ofd,.svg,.txt,.md'

  // 暂时先只支持这几种格式的文件，等后续算法后端那里支持更多格式了，再调整
  const allowedFileTypes = '.doc,.docx,.ppt,.pptx,.pdf,.mp4,.jpg,.jpeg,.png'

  // 处理文件选择
  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const uploadedFiles = e.target.files
    if (!uploadedFiles || uploadedFiles.length === 0) return

    const file = uploadedFiles[0]

    // 检查文件类型
    const fileExtension = file.name.split('.').pop()?.toLowerCase() || ''
    const validExtensions = [
      'doc',
      'docx',
      'xls',
      'xlsx',
      'csv',
      'ppt',
      'pptx',
      'pdf',
      'png',
      'jpg',
      'jpeg',
      'ofd',
      'svg',
      'txt',
      'md',
      'mp4',
    ]

    if (!validExtensions.includes(fileExtension)) {
      message.error('不支持的文件类型，请上传指定格式的文件')
      // 重置文件输入框
      if (uploadRef.current) uploadRef.current.value = ''
      return
    }

    // 检查版本额度（私有化版本不受限制）
    if (version && version.disk.use_ratio >= 1) {
      showQuotaLimitModal({ type: 'knowledge' })
      // 重置文件输入框
      if (uploadRef.current) uploadRef.current.value = ''
      return
    }

    // 执行上传逻辑
    try {
      setUploadLoading(true)

      // 调用上传文件接口
      const uploadData = {
        forest_id: knowledgeBaseId,
        parent_id: parentId,
        file: file,
      }

      await uploadFile(uploadData, {
        timeout: 0,
        headers: { 'Content-Type': 'multipart/form-data' },
      })

      message.success('上传文件成功')

      // 刷新版本信息
      refreshVersion()

      // 执行成功回调
      onSuccess?.()
    } catch (error) {
      console.error('上传文件失败:', error)
      // message.error('上传文件失败，请重试')
    } finally {
      setUploadLoading(false)
      // 重置文件输入框
      if (uploadRef.current) uploadRef.current.value = ''
    }
  }

  // 处理上传按钮点击
  const handleUpload = () => {
    uploadRef.current?.click()
  }

  return {
    uploadLoading,
    uploadRef,
    allowedFileTypes,
    handleFileSelect,
    handleUpload,
  }
}
