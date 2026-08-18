import { FC, useState } from 'react'
import { Button, Modal, Typography } from 'antd'
import { useBoolean } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import { getBindCompanyKeyWithPermSet } from '@/api/perm'
import { useQuotaLimitModal } from '@/hooks/useQuotaLimitModal'
import { useVersion } from '@/utils/useVersion'
import { PermModal } from '../PermModal'

export type InviteBtn = {
  className?: string
}
export const InviteBtn: FC<InviteBtn> = (props) => {
  const { t } = useTranslation('pages')
  const { t: tC } = useTranslation('common')
  const { className } = props
  const [permOpen, { toggle: togglePerm }] = useBoolean(false)
  const [url, setUrl] = useState('')
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
          togglePerm()
        }}
      >
        {t('settings.inviteMember')}
      </Button>
      <PermModal
        open={permOpen}
        close={togglePerm}
        okText={t('settings.generateLink')}
        onSubmit={async (val) => {
          const { key } = await getBindCompanyKeyWithPermSet({
            perm_set: {
              chatPs: val.chatPs.map((item) => {
                const {
                  use_perm,
                  manage_perm,
                  agent: { ID },
                } = item
                return {
                  use_perm,
                  manage_perm,
                  agent: { ID },
                  act_option: 'update',
                }
              }),
              forestPs: val.forestPs.map((item) => {
                const {
                  use_perm,
                  manage_perm,
                  forest: { ID },
                } = item
                return {
                  use_perm,
                  manage_perm,
                  forest: { ID },
                  act_option: 'update',
                }
              }),
            },
            count: 1,
            invitation_role: 'sys_employee',
            issuer: 'yygu',
          })
          // 刷新版本信息
          refreshVersion()
          const url = `${location.origin}/invite?key=${encodeURIComponent(key)}`
          setUrl(url)
        }}
      ></PermModal>
      <Modal
        open={Boolean(url)}
        onCancel={() => setUrl('')}
        title={t('settings.copyLink')}
        maskClosable={false}
        keyboard={false}
        width={'50vw'}
        footer={
          <div className='flex gap-2 justify-end'>
            <Button onClick={() => setUrl('')}>{tC('button.cancel')}</Button>
            <Typography.Text
              copyable={{
                icon: [
                  <Button type='primary'>{t('settings.copyLink')}</Button>,
                  <Button type='primary'>{t('settings.copyLink')}</Button>,
                ],
                text: url,
              }}
            ></Typography.Text>
          </div>
        }
      >
        <div className='flex flex-col'>
          <span className='text-[#4E5969] text-base'>
            {t('settings.invitationLinkGeneratedCopySend')}
          </span>
          <div className='bg-[#F0F2F7] text-[#0E42D2] text-base rounded p-6'>
            {url}
          </div>
        </div>
      </Modal>
    </>
  )
}
