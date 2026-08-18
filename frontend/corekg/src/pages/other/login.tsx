import { useSearchParams, useNavigate } from 'react-router-dom'
import { message } from 'antd'
import { useMount } from 'ahooks'
import { CrossStorageClient } from 'cross-storage'
import { useTranslation } from 'react-i18next'
import config from '@/config'
import useLocalStore from '@/stores/local'

const storage = new CrossStorageClient(config.loginHub)

export default function Login() {
  const { t } = useTranslation('pages')
  const { t: tM } = useTranslation('messages')
  const localStore = useLocalStore()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const key = searchParams.get('key')

  const init = async () => {
    try {
      await storage.onConnect()
      let data = await storage.get(key)
      await storage.del(key)
      if (!data || !key) {
        message.error(tM('loginFailedPleaseRelogin'))
        localStore.setLogout()
      } else {
        data = JSON.parse(data)
        localStore.setLogin({
          jwt_token: data.jwt_token,
          user_info: {
            id: data.user_id,
            username: data.user_info.name,
            avatar: data.user_info.avatar_url,
          },
        })
        message.success(tM('loginSuccess'))
        navigate('/')
      }
    } catch (error) {
      console.log('error', error)
      console.error(tM('loginFailedPleaseRelogin'))
      localStore.setLogout()
    }
  }

  useMount(() => {
    init()
  })

  return (
    <div className='w-screen h-screen flex items-center justify-center'>
      {t('other.login', { target: '...' })}
    </div>
  )
}
