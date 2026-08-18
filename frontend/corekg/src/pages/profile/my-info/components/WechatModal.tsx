import { useEffect, useState } from 'react'
import { Modal } from 'antd'
import { useTranslation } from 'react-i18next'
import { getWxLoginCode } from '@/utils/wx'
import styles from '../styles.module.scss'

interface WechatModalProps {
  visible: boolean
  onCancel: () => void
  appId?: string
}

const WechatModal = ({ visible, onCancel, appId }: WechatModalProps) => {
  const { t } = useTranslation('pages')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (visible && appId) {
      // 确保在DOM准备好后再加载脚本，避免找不到DOM元素
      const script = document.createElement('script')
      script.src =
        'https://res.wx.qq.com/connect/zh_CN/htmledition/js/wxLogin.js'
      script.async = true
      script.onload = () => {
        try {
          getWxLoginCode(appId, '/settings/profile')

          // 获取iframe并监听其加载完成事件
          const wxQrcodeElem = document.querySelector('#wxQrcode')
          if (wxQrcodeElem) {
            const iframeElem = wxQrcodeElem.getElementsByTagName('iframe')[0]
            if (iframeElem) {
              iframeElem.onload = () => {
                setLoading(false)
              }
            }
          }
        } catch (error) {
          console.error(t('profile.loadWechatQrCodeFailed'), error)
          setLoading(false)
        }
      }
      document.body.appendChild(script)

      return () => {
        // 清理脚本
        document.body.removeChild(script)
        setLoading(true)
      }
    }
  }, [visible, appId, t])
  console.log(styles.weChatModal)

  return (
    <Modal
      title={
        <div className='text-lg font-[500] text-[#1D2129] mb-6'>
          {t('profile.changeWechatBinding')}
        </div>
      }
      open={visible}
      onCancel={onCancel}
      footer={null}
      destroyOnHidden
      width={460}
      className={`weChat-modal ${styles.weChatModal}`}
      centered
    >
      <div className='flex flex-col items-center py-6'>
        <div className='text-2xl text-[#1D2129] font-500 mb-8'>
          {t('profile.changeWechatBinding')}
        </div>
        <div className='w-[300px] h-[300px] overflow-hidden'>
          <div className={`flex items-center justify-center relative`}>
            {loading && (
              <div className='text-gray-400 absolute z-10 mt-90'>
                {t('profile.qrCodeLoading')}
              </div>
            )}
            <div id='wxQrcode' className='mx-auto'></div>
          </div>
        </div>
        <div className='text-[#1D2129] font-400 text-sm mb-2'>
          {t('profile.useWechatScanToChangeWechatBinding')}
        </div>
      </div>
    </Modal>
  )
}

export default WechatModal
