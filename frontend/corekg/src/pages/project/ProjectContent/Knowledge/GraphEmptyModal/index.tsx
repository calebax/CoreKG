import { FC } from 'react'
import { Modal } from 'antd'
import { useNavigate } from 'react-router-dom'
import styles from './index.module.scss'

export type GraphEmptyModal = {
  open: boolean
  onCancel: () => void
}

/** 图谱模式下知识库为空时的提示弹窗 */
export const GraphEmptyModal: FC<GraphEmptyModal> = ({ open, onCancel }) => {
  const navigate = useNavigate()

  const handleCreate = () => {
    // 跳转到知识图谱创建页面
    navigate('/graph/edit')
    onCancel()
  }

  return (
    <Modal
      title='温馨提示'
      open={open}
      onCancel={onCancel}
      onOk={handleCreate}
      width={440}
      okText='去创建'
      cancelText='我知道了'
      closable={true}
      maskClosable={true}
      keyboard={false}
      centered
      destroyOnClose
      className={styles.graphEmptyModal}
    >
      <div className={styles.content}>
        暂无已构建的知识图谱资源，请先选择知识库并创建知识图谱。
      </div>
    </Modal>
  )
}

