import { useEffect } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { message } from 'antd'
import { useTranslation } from 'react-i18next'
import { changeWx } from '@/api/account'

const ChangeWx = () => {
  const { t: tM } = useTranslation('messages')
  const { t } = useTranslation('pages')
  const navigate = useNavigate()
  const location = useLocation()

  useEffect(() => {
    const handleLogin = async () => {
      try {
        const query = new URLSearchParams(location.search)
        const code = query.get('code')
        const state = query.get('state')

        if (code && state) {
          await changeWx({ code })
          message.success(
            tM('changeBindingSuccess', { target: t('profile.wechat') }),
          )
          navigate('/profile/my-info')
        } else {
          message.error(
            tM('targetBindingChangeFailedMissingRequiredParameters', {
              target: t('profile.wechat'),
            }),
          )
          navigate('/profile/my-info')
        }
      } catch (error) {
        console.error('微信换绑失败', error)
        message.error(
          tM('changeBindingFailed', { target: t('profile.wechat') }),
        )
        navigate('/profile/my-info')
      }
    }

    handleLogin()
  }, [location, navigate, t, tM])

  return (
    <div className='w-full h-screen flex items-center justify-center'>
      <div className='text-xl'>
        {t('profile.targetBindingChangingPleaseWait', {
          target: t('profile.wechat'),
        })}
      </div>
    </div>
  )
}

export default ChangeWx
