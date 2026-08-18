import React, { useState } from 'react'
import { App, Button } from 'antd'
import { useMount } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import { chooseUin } from '@/api/account'
import CreateActiveIcon from '@/assets/icons/login/login-create-active.svg?react'
import CreateIcon from '@/assets/icons/login/login-create.svg?react'
import LoadingCover from '@/components/common/LoadingCover'
import useLocalStore from '@/stores/local'
import { useDeployConfig } from '@/utils/useDeployConfig'

interface ChooseUinInfo {
  refreshToken: string
  userInfo: any
  uinList: any[]
  companyQuota?: number
  companyUserId?: number
}
interface ChooseUinProps {
  info: ChooseUinInfo
  onLogin?: () => void
  onCreateClick?: () => void
}

const sysRoleList = {
  sys_admin: 'other.administrator',
  sys_employee: 'other.employee',
}

const ChooseUin: React.FC<ChooseUinProps> = ({
  info,
  onLogin,
  onCreateClick,
}) => {
  const { version } = useDeployConfig()
  const { t } = useTranslation('pages')
  const { t: tM } = useTranslation('messages')
  const localStore = useLocalStore()
  const { message } = App.useApp()
  const [submitLoading, setSubmitLoading] = useState(false)

  // 判断是否达到组织配额上限
  const filteredUinList = info.uinList.filter(
    (x: any) => x.companyUserId === info.userInfo.id,
  )
  const isCompanyQuotaReached =
    info.companyQuota !== undefined &&
    filteredUinList.length >= info.companyQuota
  // 处理创建组织点击事件
  const handleCreateClick = () => {
    if (isCompanyQuotaReached) {
      message.warning('您创建的组织已达上限，不可创建新组织')
      return
    }
    onCreateClick?.()
  }
  const handleChooseUin = async (x: any) => {
    setSubmitLoading(true)
    const body = {
      login_way: info.userInfo.loginWay,
      refresh_token: info.refreshToken,
      uin_id: x.id,
      user_id: info.userInfo.id,
    }
    try {
      const res = await chooseUin(body)
      localStore.setLogin({
        token: res.jwt_token,
        userInfo: {
          ...info.userInfo,
          uinId: x.id,
        },
        uinList: info.uinList,
      })
      message.success(tM('loginSuccess'))
      onLogin?.()
    } catch (error) {
      setSubmitLoading(false)
      console.log('login error', error)
    }
  }

  useMount(() => {
    if (version === 'custom') {
      handleChooseUin(info.uinList[0])
    }
  })
  return (
    <>
      <div
        className={cn(
          'w-full h-full flex flex-col pt-6 relative',
          submitLoading && 'pointer-events-none opacity-70',
        )}
      >
        {/* custom 版本环境下不显示新建组织/企业选项 */}
        {version !== 'custom' && (
          <div
            className={cn(
              'w-full mb-3 flex items-center justify-end gap-2 pr-3',
              isCompanyQuotaReached
                ? 'cursor-not-allowed opacity-50'
                : 'cursor-pointer group',
            )}
            onClick={handleCreateClick}
            role='button'
            tabIndex={0}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault()
                handleCreateClick()
              }
            }}
          >
            <CreateIcon
              className={cn(
                'w-4 h-4 transition-opacity',
                isCompanyQuotaReached ? '' : 'group-hover:hidden',
              )}
            />
            <CreateActiveIcon
              className={cn(
                'w-4 h-4 transition-opacity',
                isCompanyQuotaReached ? 'hidden' : 'hidden group-hover:block',
              )}
            />
            <span
              className={cn(
                'text-sm font-normal transition-colors',
                isCompanyQuotaReached
                  ? 'text-[#6E757F]'
                  : 'text-[#6E757F] group-hover:text-[#0C99FF]',
              )}
            >
              {t('other.createNewOrganization')}
            </span>
          </div>
        )}

        <div className='flex-1 overflow-y-auto flex flex-col gap-3 min-h-0 pr-2'>
          {info.uinList.map((x) => (
            <Button
              key={x.id}
              className='w-full p-x-0 py-2! h-auto!'
              onClick={() => handleChooseUin(x)}
              disabled={submitLoading}
            >
              <div className='w-full'>
                <div className='flex items-center gap-2'>
                  <span className=''>{x.uinName}</span>
                  <span className='text-xs text-gray-400'>
                    {t(sysRoleList[x.role as keyof typeof sysRoleList] as any)}
                  </span>
                </div>
                <div className='text-gray-400 text-xs text-left'>
                  {x.subjectType === 'company'
                    ? x.companyName
                    : t('other.personalUser')}
                </div>
              </div>
            </Button>
          ))}
        </div>

        <LoadingCover loading={submitLoading} />
      </div>
    </>
  )
}

export default ChooseUin
