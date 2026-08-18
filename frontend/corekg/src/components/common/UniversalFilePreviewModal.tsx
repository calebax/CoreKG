import { FC } from 'react'
import { Modal } from 'antd'
import { UniversalFilePreview } from './UniversalFilePreview'

interface UniversalFilePreviewModalProps {
  id?: string
  visible: boolean
  onClose: () => void
  url: string
  fileName: string
  fileType?: string
}

export const UniversalFilePreviewModal: FC<UniversalFilePreviewModalProps> = ({
  id,
  visible,
  onClose,
  url,
  fileName,
  fileType,
}) => {
  return (
    <Modal
      open={visible}
      onCancel={onClose}
      footer={null}
      width='90vw'
      centered
      title={fileName}
      styles={{
        body: {
          height: '80vh',
          padding: '0',
          overflow: 'hidden',
          backgroundColor: '#f5f5f5',
        },
      }}
      destroyOnClose
    >
      <UniversalFilePreview
        id={id}
        url={url}
        fileName={fileName}
        fileType={fileType}
      />
    </Modal>
  )
}

