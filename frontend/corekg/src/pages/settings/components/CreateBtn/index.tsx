import { FC } from 'react'
import { App, Button } from 'antd'
import { useBoolean } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import { createEmployee } from '@/api/apiKey'
import { useQuotaLimitModal } from '@/hooks/useQuotaLimitModal'
import { encryptPassword } from '@/utils/crypto'
import { useVersion } from '@/utils/useVersion'
import { EmployeeInfoModal } from '../EmployeeInfoModal'

export type CreateBtn = {
  className?: string
  refresh: () => void
}
export const CreateBtn: FC<CreateBtn> = (props) => {
  const { t } = useTranslation('pages')
  const { t: tM } = useTranslation('messages')
  const { className, refresh } = props
  const { message } = App.useApp()
  const [open, { toggle }] = useBoolean(false)
  const { version, refresh: refreshVersion } = useVersion()
  const { show: showQuotaLimitModal } = useQuotaLimitModal()
  const isQuotaLimited =
    version && version.employee.used >= version.employee.quota

  return (
    <>
      <Button
        type='primary'
        className={cn(className, {
          'opacity-50': isQuotaLimited,
        })}
        onClick={() => {
          if (isQuotaLimited) {
            showQuotaLimitModal({ type: 'member' })
            return
          }
          toggle()
        }}
      >
        {t('settings.createMember')}
      </Button>
      <EmployeeInfoModal
        open={open}
        onClose={toggle}
        requirePassword
        onOk={async (val) => {
          if (val.password) {
            // 对密码进行加密
            const encryptedVal = {
              ...val,
              password: encryptPassword(val.password),
            }
            await createEmployee(encryptedVal as any)
            message.success(tM('createSuccess'))
            refresh()
            refreshVersion()
          }
        }}
        title={t('settings.createMember')}
      />
    </>
  )
}
