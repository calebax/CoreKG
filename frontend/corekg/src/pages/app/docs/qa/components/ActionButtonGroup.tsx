import { App, Button, Divider, Tooltip } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useBoolean } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { CloudUploadIcon } from 'tdesign-icons-react'
import { cn } from '@/utils'
import AddIcon from '../images/add.svg?react'
import UploadIcon from '../images/upload.svg?react'
import styles from '../styles.module.scss'
import { ImportModal } from './ImportModal'

interface ActionButtonGroupProps {
  forest_id: number
  onAdd: () => void
  isAdmin?: boolean
  reloadData: () => void
}

export default function ActionButtonGroup({
  forest_id,
  onAdd,
  reloadData,
  isAdmin = false,
}: ActionButtonGroupProps) {
  const { t: tM } = useTranslation('messages')
  const { t: tC } = useTranslation('common')
  const { t } = useTranslation('pages')
  const { message } = App.useApp()
  const handleAdd = () => {
    if (!isAdmin) {
      message.error(tM('noPermissionContactKbAdmin'))
      return
    }
    onAdd()
  }
  const [open, { toggle }] = useBoolean()
  return (
    <>
      <div className={cn('  flex justify-end items-center gap-[12px]')}>
        <Tooltip
          placement='top'
          title={
            isAdmin
              ? t('app.docs.supportUploadParse', {
                  file: 'xlsx、xls、csv',
                  size: '100M',
                })
              : null
          }
        >
          <Button
            className={`h-[30px] text-sm font-medium rounded-[6px] border-[#0C99FF] hover:border-[#0C99FF] text-[#0C99FF] active:text-[#0C99FF] shadow-none ${styles.createBtn} ${!isAdmin ? styles.createBtnDisabled : ''}`}
            onClick={() => {
              if (!isAdmin) {
                message.error(tM('noPermissionContactKbAdmin'))
                return
              }
              toggle()
            }}
          >
            <UploadIcon className={`${styles.createBtnIcon}`} /> 上传文件
          </Button>
        </Tooltip>
        <Button
          className={`h-[30px] text-sm font-medium rounded-[6px] border-[#0C99FF] hover:border-[#0C99FF] text-[#0C99FF] active:text-[#0C99FF] shadow-none ${styles.createBtn} ${!isAdmin ? styles.createBtnDisabled : ''}`}
          onClick={handleAdd}
        >
          <AddIcon className={`${styles.createBtnIcon}`} /> 新建问答对
        </Button>
      </div>
      {open ? (
        <ImportModal
          forest_id={forest_id}
          open={open}
          onCancel={toggle}
          onSuccess={() => {
            message.success(tM('importSuccess'))
            toggle()
            reloadData()
          }}
        />
      ) : null}
    </>
  )
}
