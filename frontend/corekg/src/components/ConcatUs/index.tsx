import { FC, PropsWithChildren } from 'react'
import { Modal, Button, Image } from 'antd'
import { useBoolean } from 'ahooks'
import { cn } from '@/utils'
import ConcatIcon from './images/concat.svg?react'
import EmailIcon from './images/email.svg?react'
import PhoneIcon from './images/phone.svg?react'
import QRCode from './images/qrcode.png'
import WechatIcon from './images/wechat.svg?react'

export const ConcatModal: FC<{ open: boolean; onClose: () => void }> = (
  props,
) => {
  const { open, onClose } = props
  return (
    <Modal
      open={open}
      title='联系我们'
      onCancel={onClose}
      width={500}
      footer={
        <Button type='primary' onClick={onClose}>
          我知道了
        </Button>
      }
    >
      <div className={cn('flex flex-col gap-5', 'font-medium text-base')}>
        <div className='flex items-center gap-1'>
          <WechatIcon />
          官方公众号
          <Image
            src={QRCode}
            className='w-25 h-25'
            wrapperClassName='ml-auto'
            preview
          />
        </div>
        <div className='flex items-center gap-1'>
          <PhoneIcon />
          电话联系：
          <span className='ml-auto'>17791089086</span>
        </div>
        <div className='flex items-center gap-1'>
          <EmailIcon />
          邮箱联系：
          <span className='ml-auto'>support@yygu.cn</span>
        </div>
      </div>
    </Modal>
  )
}

export const ConcatUs: FC<Style & PropsWithChildren> = (props) => {
  const { children, className, style } = props
  const [open, { toggle }] = useBoolean()
  return (
    <>
      <div
        className={cn('cursor-pointer', className)}
        style={style}
        onClick={toggle}
      >
        {children ?? (
          <span className='flex items-center text-[#0C99FF] gap-1'>
            <ConcatIcon />
            联系人工客服
          </span>
        )}
      </div>
      <ConcatModal open={open} onClose={toggle} />
    </>
  )
}
