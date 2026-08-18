import { useState, useEffect } from 'react'
import { Modal, Checkbox } from 'antd'
import { useTranslation } from 'react-i18next'
import styles from './index.module.scss'

interface DeleteGroupModalProps {
  open: boolean
  onCancel: () => void
  onConfirm: (moveToFree: boolean) => Promise<void>
}

export default function DeleteGroupModal({
  open,
  onCancel,
  onConfirm,
}: DeleteGroupModalProps) {
  const { t } = useTranslation('pages')
  const [moveToFree, setMoveToFree] = useState(true)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (open) {
      setMoveToFree(true)
      setLoading(false)
    }
  }, [open])

  const handleConfirm = async () => {
    setLoading(true)
    try {
      await onConfirm(moveToFree)
    } catch (error) {
      console.error('删除会话分组失败:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleCancel = () => {
    onCancel()
  }

  return (
    <Modal
      title={t('app.sidebar.deleteGroup')}
      open={open}
      onCancel={handleCancel}
      onOk={handleConfirm}
      okText={t('app.sidebar.confirmDelete')}
      cancelText={t('app.sidebar.cancel')}
      okButtonProps={{ loading }}
      closable={!loading}
      keyboard={!loading}
      maskClosable={!loading}
      destroyOnHidden // antd 5 替代已弃用的 destroyOnClose
      centered
      width={520}
      className={styles.deleteGroupModal}
    >
      <div className={styles.warning}>
        {t('app.sidebar.deleteGroupWarning')}
      </div>
      <Checkbox
        checked={moveToFree}
        onChange={(e) => setMoveToFree(e.target.checked)}
        className={styles.checkbox}
        disabled={loading}
      >
        {t('app.sidebar.moveToUncategorized')}
      </Checkbox>
    </Modal>
  )
}

