import React from 'react'
import { Modal } from 'antd'
import ApiKeyWarning2 from '@/assets/icons/apiKey-warning2.svg?react'
import { useVersion } from '@/utils/useVersion'

interface DeleteConfirmModalProps {
  visible: boolean
  isFolder?: boolean
  isMultiple?: boolean
  customText?: string
  customTitle?: string
  customOkText?: string
  onCancel: () => void
  onConfirm: () => Promise<void>
}

const DeleteConfirmModal: React.FC<DeleteConfirmModalProps> = ({
  visible,
  isFolder = false,
  isMultiple = false,
  customText,
  customTitle,
  customOkText,
  onCancel,
  onConfirm,
}) => {
  const { refresh } = useVersion()
  // 根据选中的内容确定提示文字
  let confirmText = customText || '删除后，所选项目将无法恢复，请谨慎操作。'

  if (!customText) {
    if (isMultiple) {
      confirmText = '删除后，所选项目将无法恢复，请谨慎操作。'
    } else if (isFolder) {
      confirmText = '删除后，该文件夹将无法恢复，请谨慎操作。'
    } else {
      confirmText = '删除后，该文件将无法恢复，请谨慎操作。'
    }
  }

  const title = customTitle || '确认删除'
  const okText = customOkText || '确认'

  return (
    <Modal
      open={visible}
      onCancel={onCancel}
      onOk={async () => {
        await onConfirm()
        refresh()
      }}
      centered
      closable={false}
      className='delete-api-key-modal !w-[30%]'
      okText={okText}
      okButtonProps={{
        className:
          'bg-[#0C99FF] hover:!bg-[#0C99FF] !w-[77px] !h-[32px] !rounded-md !text-sm !px-5 !py-2 !font-medium text-[#ffffff]',
        danger: false,
      }}
      cancelButtonProps={{
        className:
          '!bg-[#F5F5F5] text-[#0C1F17] !w-[77px] !h-[32px] !rounded-md !text-sm !border-none !px-4 !py-2 !font-medium',
      }}
      cancelText='取消'
    >
      <div className='relative'>
        <div className='flex justify-between'>
          <div className='text-[22px] font-[500] mb-2 text-[#000000E5]'>
            {title}
          </div>
          {/* <ApiKeyWarning2 className='w-[26px] h-[26px]' /> */}
        </div>
        <div className='h-[0.5px] w-[calc(100%+48px)] bg-[#C9CDD4] mt-4 -mx-6'></div>
        <div className='mt-6 text-base text-[#616373] mb-22 font-medium fontFamily-pingFangSC'>
          {confirmText}
        </div>
      </div>
    </Modal>
  )
}

export default DeleteConfirmModal
