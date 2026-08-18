import { useState } from 'react'
import { Modal, Input, message } from 'antd'
import { useTranslation } from 'react-i18next'
import styles from './index.module.scss'

interface CreateGroupModalProps {
  open: boolean
  onCancel: () => void
  onSuccess: (name: string) => Promise<void>
}

export default function CreateGroupModal({
  open,
  onCancel,
  onSuccess,
}: CreateGroupModalProps) {
  const { t } = useTranslation('pages')
  const [groupName, setGroupName] = useState('')
  const [loading, setLoading] = useState(false)

  const handleOk = async () => {
    if (!groupName.trim()) {
      message.warning(t('app.sidebar.groupNameRequired'))
      return
    }

    setLoading(true)
    try {
      await onSuccess(groupName.trim())
      setGroupName('')
      onCancel()
    } catch (error) {
      console.error('创建会话分组失败:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleCancel = () => {
    setGroupName('')
    onCancel()
  }

  const isDisabled = !groupName.trim() || loading

  return (
    <Modal
      title={t('app.sidebar.createGroup')}
      open={open}
      onCancel={handleCancel}
      onOk={handleOk}
      okText={t('app.sidebar.create')}
      cancelText={t('app.sidebar.cancel')}
      okButtonProps={{ loading, disabled: isDisabled }}
      closable={!loading}
      keyboard={!loading}
      maskClosable={!loading}
      // antd 5：destroyOnClose 已弃用，关闭后卸载子节点请用 destroyOnHidden
      destroyOnHidden
      className={styles.createGroupModal}
    >
      <div className={styles.description}>
        {t('app.sidebar.groupDescription')}
      </div>
      {/* <div className={styles.label}>
        {t('app.sidebar.groupName')}
      </div> */}
      <Input
        value={groupName}
        onChange={(e) => setGroupName(e.target.value)}
        placeholder={t('app.sidebar.groupNamePlaceholder')}
        maxLength={10}
        onPressEnter={isDisabled ? undefined : handleOk}
        disabled={loading}
        className={styles.input}
      />
    </Modal>
  )
}

