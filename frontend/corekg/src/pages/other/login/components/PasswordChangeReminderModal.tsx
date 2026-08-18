import React, { useState } from 'react'
import { Modal, Checkbox, Button, message } from 'antd'
import { useTranslation } from 'react-i18next'

interface PasswordChangeReminderModalProps {
  visible: boolean
  onCancel: () => void
  onSkip: (dontShowAgain?: boolean) => void
  onModifyPassword: () => void
  loading?: boolean
}

const PasswordChangeReminderModal: React.FC<PasswordChangeReminderModalProps> = ({
  visible,
  onCancel,
  onSkip,
  onModifyPassword,
  loading = false,
}) => {
  const { t } = useTranslation('pages')
  const [dontShowAgain, setDontShowAgain] = useState(false)

  const handleSkip = () => {
    onSkip(dontShowAgain)
  }

  const handleModifyPassword = () => {
    onModifyPassword()
  }

  return (
    <Modal
      open={visible}
      onCancel={onCancel}
      footer={null}
      closable={false}
      centered
      width={440}
      className="password-change-reminder-modal"
      styles={{
        content: {
          borderRadius: '4px',
          overflow: 'hidden'
        }
      }}
    >
      {/* Header */}
      <div className="border-b border-[#e3e6ed] px-0 py-[10px]">
        <h3 className="text-[#0c1f17] text-[22px] font-medium leading-none m-0">
          {t('other.passwordChangeReminder.title')}
        </h3>
      </div>

      {/* Content */}
      <div className="flex flex-col h-[151px] justify-between py-3">
        <p className="text-[#6e757f] text-[16px] font-normal leading-[24px] m-0">
          {t('other.passwordChangeReminder.description')}
        </p>

        {/* Checkbox */}
        <div className="flex justify-end items-center">
          <Checkbox
            checked={dontShowAgain}
            onChange={(e) => setDontShowAgain(e.target.checked)}
            className="text-sm"
            style={{
              borderRadius: '3.2px',
            }}
          >
            <span className="text-[#abafb2] text-sm font-normal leading-[22px]">
              {t('other.passwordChangeReminder.dontShowAgain')}
            </span>
          </Checkbox>
        </div>
      </div>

      {/* Footer */}
      <div className="flex gap-[6px] h-[44px] items-center justify-end py-[6px]">
        <Button
          className="w-[77px] h-8 rounded-md bg-[#F5F5F5] border-[#F5F5F5] text-[#0c1f17] hover:bg-[#F5F5F5] hover:text-[#0C99FF] hover:border-[#F5F5F5]"
          onClick={handleSkip}
          disabled={loading}
          style={{
            height: '32px',
            borderRadius: '6px',
            fontSize: '14px',
            fontWeight: 500,
            backgroundColor: '#f5f5f5',
            borderColor: '#f5f5f5',
            padding: '9px 24.5px'
          }}
        >
          {t('other.passwordChangeReminder.skip')}
        </Button>
        <Button
          type="primary"
          className="w-[77px] h-8 rounded-md"
          onClick={handleModifyPassword}
          disabled={loading}
          style={{
            height: '32px',
            borderRadius: '6px',
            fontSize: '14px',
            fontWeight: 500,
            backgroundColor: '#0C99FF',
            borderColor: '#0C99FF',
            color: '#ffffff',
            padding: '9px 24.5px'
          }}
        >
          {t('other.passwordChangeReminder.modify')}
        </Button>
      </div>
    </Modal>
  )
}

export default PasswordChangeReminderModal