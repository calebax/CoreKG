import { FC } from 'react'
import { Badge, Button, Modal } from 'antd'
import { useBoolean } from 'ahooks'
import { cn } from '@/utils'
import { useForestUploadStore } from '@/stores/ForestUploadStore'
import Icon from './icon.svg?react'
import styles from './styles.module.scss'
import { ForestUpload } from './upload'
import { useAnalyzeFiles } from './upload/utils'

export type ForestUploadBtn = Style & {
  forest_id: number
  onUploadOne?: () => void
}
export const ForestUploadBtn: FC<ForestUploadBtn> = (props) => {
  const { forest_id, onUploadOne, className, style } = props
  const { analyzeFiles, reload } = useAnalyzeFiles()
  const uploadingLength = useForestUploadStore(
    (state) =>
      state.getFilesByOptions({ forest_id, status: 'uploading' }).length,
  )

  const count = useMemo(() => {
    if (!analyzeFiles) return uploadingLength
    return (
      uploadingLength +
      analyzeFiles.filter((item) => item.status === 'analyzing').length
    )
  }, [analyzeFiles, uploadingLength])

  const prevLength = useRef(uploadingLength)
  if (uploadingLength < prevLength.current) {
    onUploadOne?.()
  }
  prevLength.current = uploadingLength

  const [open, { toggle }] = useBoolean()
  return (
    <>
      <Badge count={count} color='#2E7CF7' className={className} style={style}>
        <Button
          className={cn('outline-none shadow-none border-none bg-[#E8F3FF]!')}
          style={style}
          onClick={toggle}
        >
          <Icon />
        </Button>
      </Badge>
      <Modal
        open={open}
        onCancel={toggle}
        footer={null}
        width='60vw'
        className={styles.modal}
      >
        <ForestUpload
          forest_id={forest_id}
          closeModal={toggle}
          analyzeFiles={analyzeFiles}
          reloadAnalyzeFiles={reload}
        />
      </Modal>
    </>
  )
}
