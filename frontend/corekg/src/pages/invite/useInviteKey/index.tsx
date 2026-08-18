import { useMemo } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { App } from 'antd'
import { useMount, useRequest } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { bindCompanyWithPermSet, getInviteInfo } from '@/api/perm'
import useLocalStore from '@/stores/local'
import { encryptPassword } from '@/utils/crypto'

export const useInviteCode = () => {
  const { t: tM } = useTranslation('messages')
  const { message } = App.useApp()

  const navigate = useNavigate()
  const setLogin = useLocalStore((state) => state.setLogin)

  const [searchParams] = useSearchParams()
  const keyParam = searchParams.get('key')
  // 将key存放在sessionStorage里 扫码登录后页面跳转可以取出
  const key = useMemo(() => {
    const storageKey = 'inviteKey'
    if (keyParam) {
      const _key = decodeURIComponent(keyParam)
      sessionStorage.setItem(storageKey, decodeURIComponent(keyParam))
      return _key
    }
    const _key = sessionStorage.getItem(storageKey)
    sessionStorage.removeItem(storageKey)
    return _key
  }, [keyParam])

  const wxInfo = useMemo(() => {
    return {
      code: searchParams.get('code'),
      state: searchParams.get('state'),
    }
  }, [searchParams])

  const type: 'wx' | 'key' | 'illegal' = useMemo(() => {
    if (wxInfo.code && wxInfo.state && key) {
      // 已扫码
      return 'wx'
    } else if (key) {
      // 未扫码
      return 'key'
    }
    // 不合法
    return 'illegal'
  }, [key, wxInfo.code, wxInfo.state])

  useMount(async () => {
    if (type === 'illegal') {
      navigate('/', { replace: true })
    } else if (type === 'wx') {
      const body = {
        key: key!,
        code: wxInfo.code!,
        way: 'wechat_web',
        domain_name: location.origin,
      }
      try {
        const res = await bindCompanyWithPermSet(body)
        if (res.login_status === 'failed') {
          throw new Error()
        }
        const userInfo = {
          id: res.user_id,
          avatar: res.user_info.avatar_url,
          name: res.user_info.name,
          uinId: res.user_info.uin,
          loginWay: res.login_way,
        }
        const uinList = res.uin.map((x: any) => {
          return {
            id: x.uin.ID,
            role: x.role,
            uinName: x.uin.Name,
            companyName: x.company_name,
            companyStatus: x.company_status, // passed
            uinStatus: x.uin.UinStatus, // normal
            subjectType: x.uin.SubjectType, // company, individual (企业，个人)
            subjectId: x.uin.SubjectID, // 企业id，个人为0
            logo: x.company_logo,
            companyUserId: x.company_user_id,
          }
        })
        setLogin({
          token: res.jwt_token,
          userInfo: userInfo,
          uinList: uinList,
        })
        message.success(tM('joinOrganizationSuccess'))
        navigate('/', { replace: true })
      } catch {
        navigate(`?key=${key}`, { replace: true })
      }
    }
  })

  const { data } = useRequest(async () => {
    if (!key || type !== 'key') {
      return undefined
    }
    const { inviter_name: invitor, company_name: companyName } =
      await getInviteInfo({ key })
    return { invitor, companyName }
  })
  const { invitor, companyName } = data ?? {}

  // 处理账号密码登录
  const handlePasswordLogin = async (username: string, password: string) => {
    if (!key) {
      message.error('邀请链接无效')
      return
    }
    try {
      const encryptedPassword = encryptPassword(password)
      const body = {
        key: key,
        way: 'password_web',
        domain_name: location.origin,
        // domain_name: 'https://example.com', //为了本地开发登录页面，临时设置的，开发完之后需要注释掉这行代码，恢复上一行的代码
        username: username,
        password: encryptedPassword,
      }
      const res = await bindCompanyWithPermSet(body)
      if (res.login_status === 'failed') {
        throw new Error()
      }
      const userInfo = {
        id: res.user_id,
        avatar: res.user_info.avatar_url,
        name: res.user_info.name,
        uinId: res.user_info.uin,
        loginWay: res.login_way,
      }
      const uinList = res.uin.map((x: any) => {
        return {
          id: x.uin.ID,
          role: x.role,
          uinName: x.uin.Name,
          companyName: x.company_name,
          companyStatus: x.company_status, // passed
          uinStatus: x.uin.UinStatus, // normal
          subjectType: x.uin.SubjectType, // company, individual (企业，个人)
          subjectId: x.uin.SubjectID, // 企业id，个人为0
          logo: x.company_logo,
          companyUserId: x.company_user_id,
        }
      })
      setLogin({
        token: res.jwt_token,
        userInfo: userInfo,
        uinList: uinList,
      })
      message.success(tM('joinOrganizationSuccess'))
      navigate('/', { replace: true })
    } catch (error) {
      console.log('账号密码登录失败:', error)
    }
  }

  return {
    key,
    invitor,
    companyName,
    type,
    handlePasswordLogin,
  }
}
