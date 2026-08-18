import { useState } from 'react'
import { Modal, Input, message } from 'antd'
import { useTranslation } from 'react-i18next'
import styles from './CreateOrganizationModal.module.scss'

interface CreateOrganizationModalProps {
  open: boolean
  onCancel: () => void
  onSuccess: (companyName: string, userDisplayName: string) => Promise<void>
}

export default function CreateOrganizationModal({
  open,
  onCancel,
  onSuccess,
}: CreateOrganizationModalProps) {
  const { t } = useTranslation('pages')
  const [companyName, setCompanyName] = useState('')
  const [userDisplayName, setUserDisplayName] = useState('')
  const [loading, setLoading] = useState(false)

  const handleOk = async () => {
    if (!companyName.trim()) {
      message.warning(t('other.pleaseEnterOrganizationName'))
      return
    }
    if (!userDisplayName.trim()) {
      message.warning(t('other.pleaseEnterUserNickname'))
      return
    }
    if (companyName.trim().length > 50) {
      message.warning(t('other.organizationNameMaxLength'))
      return
    }
    if (userDisplayName.trim().length > 20) {
      message.warning(t('other.userNicknameMaxLength'))
      return
    }

    setLoading(true)
    try {
      await onSuccess(companyName.trim(), userDisplayName.trim())
      setCompanyName('')
      setUserDisplayName('')
      onCancel()
    } catch (error) {
      console.error('创建组织失败:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleCancel = () => {
    setCompanyName('')
    setUserDisplayName('')
    onCancel()
  }

  const isDisabled = !companyName.trim() || !userDisplayName.trim() || loading

  return (
    <Modal
      title={t('other.createOrganization')}
      open={open}
      onCancel={handleCancel}
      onOk={handleOk}
      okText={t('app.sidebar.create')}
      cancelText={t('app.sidebar.cancel')}
      okButtonProps={{ loading, disabled: isDisabled }}
      closable={!loading}
      keyboard={!loading}
      maskClosable={!loading}
      destroyOnClose
      className={styles.createOrganizationModal}
    >
      <Input
        value={companyName}
        onChange={(e) => setCompanyName(e.target.value)}
        placeholder={t('other.pleaseEnterOrganizationName')}
        maxLength={50}
        onPressEnter={isDisabled ? undefined : handleOk}
        disabled={loading}
        className={styles.input}
      />
      <Input
        value={userDisplayName}
        onChange={(e) => setUserDisplayName(e.target.value)}
        placeholder={t('other.pleaseEnterUserNickname')}
        maxLength={20}
        onPressEnter={isDisabled ? undefined : handleOk}
        disabled={loading}
        className={styles.input}
      />
    </Modal>
  )
}

