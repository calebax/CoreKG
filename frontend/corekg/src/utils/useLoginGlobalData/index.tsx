import { FC, PropsWithChildren, useMemo } from 'react'
import { Skeleton } from 'antd'
import { useRequest } from 'ahooks'
import { formatFileSize } from '@/utils'
import {
  getCompanyAdmins,
  getAllUin,
  DetailPersonalCenter,
} from '@/api/account'
import { getLicenseInfo } from '@/api/auth'
import { getCommonInfo, getMessageCount } from '@/api/common'
import useLocalStore from '@/stores/local'
import type { UinInfo } from '@/stores/local'
import { type AdminContextValue } from '../useAdmin'
import { useDeployConfig } from '../useDeployConfig'
import { type ContextValue as VersionContextValue } from '../useVersion'
import { LoginGlobalContext } from './context'

type CommonInfo =
  ReturnType<typeof getCommonInfo> extends Promise<infer T> ? T : never

// eslint-disable-next-line react-refresh/only-export-components
export { useLoginGlobalData } from './context'

// 比较两个 UinInfo 数组是否相同（按 id 排序后比较，避免顺序不同导致的误判）
const isUinListEqual = (oldList: UinInfo[], newList: UinInfo[]): boolean => {
  if (oldList.length !== newList.length) return false
  const sortById = (a: UinInfo, b: UinInfo) =>
    String(a.id).localeCompare(String(b.id))
  return (
    JSON.stringify([...oldList].sort(sortById)) ===
    JSON.stringify([...newList].sort(sortById))
  )
}

export type GlobalContextValue = {
  admin: AdminContextValue
  version: VersionContextValue
  license?: { modules: CustomModule[] }
  commonInfo: {
    data: CommonInfo
    refresh: () => void
  }
  messageNotificationCount: {
    count: number
    refresh: () => void
  }
}

/**
 * 提供登录后可用的全局数据
 */
export const LoginGlobalProvider: FC<PropsWithChildren> = (props) => {
  const token = useLocalStore((st) => st.token)
  const { version: deployVersion } = useDeployConfig()
  // Admin
  const { data: adminData, refresh: refreshAdmin } = useRequest(
    async () => {
      const { Data } = await getCompanyAdmins()
      const admin = (Data ?? []) as AdminContextValue['admin']
      const adminIds = admin.map((item) => item.uin)
      return {
        admin,
        adminIds,
      }
    },
    {
      refreshDeps: [token, deployVersion],
    },
  )

  // CommonInfo - 获取通用信息
  const { data: commonInfoData, refresh: refreshCommonInfo } = useRequest(
    async () => {
      // custom 环境下不调用该接口，返回默认配额
      if (deployVersion === 'custom') {
        return {
          company_quota: {
            agent_quota: 0,
            agent_quota_used: 0,
            article_quota: 0,
            article_quota_used: 0,
            disk_quota: 0,
            disk_quota_used: 0,
            employee_quota: 0,
            employee_quota_used: 0,
            qa_quota: 0,
            qa_quota_used: 0,
            is_purchased: true,
            company_quota: 0,
          },
        }
      }
      try {
        return await getCommonInfo({ timeout: 2000 })
      } catch {
        return {
          company_quota: {
            agent_quota: 5,
            agent_quota_used: 5,
            article_quota: 0,
            article_quota_used: 0,
            disk_quota: 10737418240,
            disk_quota_used: 10737418240,
            employee_quota: 5,
            employee_quota_used: 5,
            qa_quota: 100,
            qa_quota_used: 100,
            is_purchased: false,
            company_quota: 2,
          },
        }
      }
    },
    {
      pollingInterval: deployVersion === 'custom' ? undefined : 60 * 1000,
      refreshDeps: [token, deployVersion],
    },
  )

  // MessageNotificationCount - 获取消息通知数量
  const {
    data: messageNotificationCountData,
    refresh: refreshMessageNotificationCount,
  } = useRequest(
    async () => {
      const result = await getMessageCount([
        { field: 'read_status', value: ['unread'] },
      ])
      return { count: result.count ?? 0 }
    },
    {
      pollingInterval: 60 * 1000,
      refreshDeps: [token, deployVersion],
    },
  )

  // Version - 从 commonInfo 构建版本数据
  const version = useMemo(() => {
    if (deployVersion === 'custom' || !commonInfoData) {
      return undefined
    }
    const { company_quota } = commonInfoData
    const diskUseRatio =
      company_quota.disk_quota > 0
        ? Number(
            (company_quota.disk_quota_used / company_quota.disk_quota).toFixed(
              4,
            ),
          )
        : 0
    const _version: VersionContextValue['version'] = {
      is_purchased: company_quota.is_purchased,
      name: company_quota.is_purchased ? '专业版' : '社区版',
      qa: {
        used: company_quota.qa_quota_used,
        quota: company_quota.qa_quota,
      },
      agent: {
        used: company_quota.agent_quota_used,
        quota: company_quota.agent_quota,
      },
      disk: {
        used: formatFileSize(company_quota.disk_quota_used),
        quota: formatFileSize(company_quota.disk_quota),
        use_ratio: diskUseRatio,
      },
      employee: {
        used: company_quota.employee_quota_used,
        quota: company_quota.employee_quota,
      },
    }
    return _version
  }, [commonInfoData, deployVersion])

  // license
  const { data: license } = useRequest(
    async () => {
      if (deployVersion !== 'custom') return
      const res = await getLicenseInfo({})
      return { modules: res.modules ?? [] }
    },
    { refreshDeps: [token, deployVersion] },
  )

  // 进入页面时同步uin 昵称等信息
  const setLogin = useLocalStore((state) => state.setLogin)
  useRequest(
    async () => {
      const { uin } = await getAllUin()
      const latestUinList: UinInfo[] = (uin as any[]).map((x) => ({
        id: x.uin.ID,
        role: x.role,
        uinName: x.uin.Name,
        companyName: x.company_name,
        companyStatus: x.company_status,
        uinStatus: x.uin.UinStatus,
        subjectType: x.uin.SubjectType,
        subjectId: x.uin.SubjectID,
        logo: x.company_logo,
        companyUserId: x.company_user_id,
      }))
      const currentUinList = useLocalStore.getState().uinList
      if (!isUinListEqual(currentUinList, latestUinList)) {
        setLogin({ uinList: latestUinList })
      }
    },
    {
      refreshDeps: [token],
    },
  )

  // 进入页面时同步个人信息
  const setUserInfo = useLocalStore((state) => state.setUserInfo)
  useRequest(
    async () => {
      const data = await DetailPersonalCenter()
      const currentUserInfo = useLocalStore.getState().userInfo
      const latestUserInfo = {
        id: data?.user_info?.id ?? currentUserInfo.id,
        name: data?.user_info?.name ?? currentUserInfo.name,
        avatar: data?.user_info?.avatar_url ?? currentUserInfo.avatar,
        uinId: data?.user_info?.uin ?? currentUserInfo.uinId,
        loginWay: data?.user_info?.loginWay ?? currentUserInfo.loginWay,
      }
      setUserInfo(latestUserInfo)
    },
    {
      refreshDeps: [token],
    },
  )

  // 必须有admin commonInfo
  // saas版本必须有version信息
  // 私有化版本必须有license信息
  if (
    !adminData ||
    !commonInfoData ||
    (!version && deployVersion === 'saas') ||
    (!license && deployVersion === 'custom')
  ) {
    return <Skeleton active className='p-4' />
  }

  return (
    <LoginGlobalContext.Provider
      value={{
        admin: {
          ...adminData,
          refresh: refreshAdmin,
        },
        version: {
          version,
          refresh: refreshCommonInfo,
        },
        commonInfo: {
          data: commonInfoData,
          refresh: refreshCommonInfo,
        },
        license,
        messageNotificationCount: {
          count: messageNotificationCountData?.count ?? 0,
          refresh: refreshMessageNotificationCount,
        },
      }}
    >
      {props.children}
    </LoginGlobalContext.Provider>
  )
}
