import { useCallback, useMemo, useState, useEffect } from 'react'
import { Modal, App } from 'antd'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import { createCompany, getAllUin, switchLogin } from '@/api/account'
import { getCommonInfo } from '@/api/common'
import CloseIcon from '@/assets/icons/close.svg?react'
import HomeArrowRightHoverIcon from '@/assets/icons/home/home-arrow-hover.svg?react'
import HomeArrowRightIcon from '@/assets/icons/home/home-arrow.svg?react'
import CreateActiveIcon from '@/assets/icons/login/login-create-active.svg?react'
import CreateIcon from '@/assets/icons/login/login-create.svg?react'
import useLocalStore, { type UinInfo } from '@/stores/local'
import { useDeployConfig } from '@/utils/useDeployConfig'
import DefaultOrgLogo from '../images/defaultLogo.svg'
import CreateOrganizationModal from './CreateOrganizationModal'

interface SwitchOrganizationDialogProps {
  open: boolean
  onClose: () => void
}

const SwitchOrganizationDialog = ({
  open,
  onClose,
}: SwitchOrganizationDialogProps) => {
  const { version } = useDeployConfig()
  const { t } = useTranslation('pages')
  const { t: tM } = useTranslation('messages')
  const { uinList, userInfo, setLogin, token } = useLocalStore()
  const { message } = App.useApp()
  const [createModalOpen, setCreateModalOpen] = useState(false)
  // 临时保存新创建的组织ID，用于在列表中临时将其放到第一位
  const [newlyCreatedOrgId, setNewlyCreatedOrgId] = useState<number | null>(
    null,
  )
  // 保存组织配额
  const [companyQuota, setCompanyQuota] = useState<number | undefined>(
    undefined,
  )

  const handleSwitchOrganization = useCallback(
    async (targetUinId: number | string) => {
      if (targetUinId === userInfo.uinId) return
      try {
        const { jwt_token } = await switchLogin({
          login_way: userInfo.loginWay!,
          uin: Number(targetUinId) as any as number,
        })
        // 调整 uinList 顺序，将选中的组织移到第一位
        const reorderedUinList = [...uinList]
        const targetIndex = reorderedUinList.findIndex(
          (item) => String(item.id) === String(targetUinId),
        )
        if (targetIndex > 0) {
          const [targetOrg] = reorderedUinList.splice(targetIndex, 1)
          reorderedUinList.unshift(targetOrg)
        }
        setLogin({
          token: jwt_token,
          uinList: reorderedUinList,
          userInfo: {
            ...userInfo,
            uinId: Number(targetUinId),
          },
        })
      } catch {
        const { uin } = await getAllUin()
        const latestUinList = (uin as any[]).map((x) => {
          return {
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
          }
        })
        // 调整 latestUinList 顺序，将选中的组织移到第一位
        const reorderedLatestUinList = [...latestUinList]
        const targetIndex = reorderedLatestUinList.findIndex(
          (item) => String(item.id) === String(targetUinId),
        )
        if (targetIndex > 0) {
          const [targetOrg] = reorderedLatestUinList.splice(targetIndex, 1)
          reorderedLatestUinList.unshift(targetOrg)
        }
        setLogin({
          token,
          uinList: reorderedLatestUinList,
          userInfo: {
            ...userInfo,
            uinId: Number(targetUinId),
          },
        })
      }
      window.location.href = '/global'
    },
    [setLogin, token, uinList, userInfo],
  )

  const handleOrgClick = useCallback(
    async (org: UinInfo) => {
      if (String(org.id) === String(userInfo.uinId)) {
        onClose()
        return
      }
      await handleSwitchOrganization(org.id)
    },
    [handleSwitchOrganization, userInfo.uinId, onClose],
  )

  const getRoleText = (role: string) => {
    if (role === 'sys_admin') {
      return t('app.sidebar.admin')
    }
    return ''
  }

  // 处理创建组织
  const handleCreateOrganization = useCallback(
    async (companyName: string, userDisplayName: string) => {
      try {
        // 判断是否是本地开发环境
        const isLocalDev = import.meta.env.MODE === 'development'

        // 本地开发使用 https://example.com，其他环境使用 location.origin
        const domainName = isLocalDev
          ? 'https://example.com'
          : location.origin

        // 1. 创建组织（切换组织弹窗创建，不需要refresh_token参数，使用请求头token）
        const createResult = await createCompany({
          company_name: companyName,
          user_display_name: userDisplayName,
          domain_name: domainName,
          // 不传 refresh_token，使用请求头的 token
          user_id: Number(userInfo.id),
        })

        // 从创建接口返回值中获取新创建的组织ID
        const newOrgId = (createResult as any)?.uin?.ID

        // 2. 调用getAllUin获取最新的组织列表（包含新创建的组织）
        const { uin } = await getAllUin()
        const latestUinList = (uin as any[]).map((x) => {
          return {
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
          }
        })

        // 3. 更新localStorage中的uinList（保持正常顺序，不临时调整）
        setLogin({
          uinList: latestUinList,
        })

        // 4. 设置新创建的组织ID，用于临时将其放到列表第一位
        if (newOrgId) {
          setNewlyCreatedOrgId(newOrgId)
        }

        message.success('创建组织成功')
      } catch (error: any) {
        console.log('创建组织失败:', error)
      }
    },
    [userInfo, setLogin, message],
  )

  // 处理弹窗关闭，清除临时状态
  const handleClose = useCallback(() => {
    setNewlyCreatedOrgId(null)
    onClose()
  }, [onClose])

  // 临时调整组织列表顺序：如果存在新创建的组织，将其放到第一位（仅用于显示，不影响存储）
  const displayUinList = useMemo(() => {
    if (!newlyCreatedOrgId) {
      return uinList
    }
    const reorderedList = [...uinList]
    const targetIndex = reorderedList.findIndex(
      (item) => String(item.id) === String(newlyCreatedOrgId),
    )
    if (targetIndex > 0) {
      const [targetOrg] = reorderedList.splice(targetIndex, 1)
      reorderedList.unshift(targetOrg)
    }
    return reorderedList
  }, [uinList, newlyCreatedOrgId])

  // 判断是否达到组织配额上限
  // 对uinList进行筛选，提取出每一个companyUserId等于userInfo.id的uin
  // console.log('userInfo', userInfo)
  // console.log('uinList', uinList)
  const filteredUinList = uinList.filter(
    (x) => x.companyUserId === Number(userInfo.id),
  )
  // console.log('filteredUinList', filteredUinList)
  const isCompanyQuotaReached =
    companyQuota !== undefined && filteredUinList.length >= companyQuota
  // console.log('companyQuota', companyQuota)
  // console.log('filteredUinList.length', filteredUinList.length)
  // console.log('isCompanyQuotaReached', isCompanyQuotaReached)
  // 当弹窗打开时，获取组织配额
  useEffect(() => {
    // custom 环境下不调用获取配额接口
    if (open && version !== 'custom') {
      const fetchQuota = async () => {
        try {
          const commonInfo = await getCommonInfo()
          setCompanyQuota(commonInfo.company_quota.company_quota)
        } catch (error) {
          console.log('获取组织配额失败:', error)
        }
      }
      fetchQuota()
    }
  }, [open, version])

  // 处理创建组织点击事件
  const handleCreateClick = useCallback(() => {
    if (isCompanyQuotaReached) {
      message.warning('您创建的组织已达上限，不可创建新组织')
      return
    }
    setCreateModalOpen(true)
  }, [isCompanyQuotaReached, message])

  return (
    <Modal
      open={open}
      onCancel={handleClose}
      footer={null}
      closable={false}
      width={480}
      centered
      styles={{
        body: {
          padding: 0,
        },
        content: {
          padding: 0,
        },
      }}
    >
      <div className='flex flex-col gap-4 p-6'>
        {/* 弹窗头部 */}
        <div className='flex items-center justify-between pb-3 border-b border-[#EFF1F4]'>
          <div className='text-lg font-medium text-[#0C1F17]'>切换组织</div>
          <button
            type='button'
            onClick={handleClose}
            className='flex items-center justify-center w-6 h-6 rounded cursor-pointer hover:bg-[#F5F7FA] transition-colors'
            aria-label='关闭'
            tabIndex={0}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault()
                handleClose()
              }
            }}
          >
            <CloseIcon className='w-4 h-4' />
          </button>
        </div>

        {/* 创建组织按钮 - custom 版本环境下不显示 */}
        {version !== 'custom' && (
          <div
            className={cn(
              'w-full flex items-center justify-end gap-2',
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
            aria-label='创建新组织/企业'
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

        {/* 组织列表 */}
        <div className='flex flex-col gap-1 max-h-[380px] overflow-y-auto scrollbar-hide'>
          {displayUinList.map((org) => {
            const isCurrent = String(org.id) === String(userInfo.uinId)
            const roleText = getRoleText(org.role)
            const logoUrl = org.logo || DefaultOrgLogo

            return (
              <div
                key={org.id}
                className={cn(
                  'flex items-center rounded gap-4 px-2 py-3 cursor-pointer transition-colors group',
                  {
                    'bg-[#F7F7F7]': isCurrent,
                    'hover:bg-[#F7F7F7]': !isCurrent,
                  },
                )}
                onClick={() => handleOrgClick(org)}
                tabIndex={0}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault()
                    handleOrgClick(org)
                  }
                }}
                role='button'
                aria-label={`切换到组织：${org.companyName}`}
              >
                {/* 组织头像 */}
                <div className='flex-shrink-0'>
                  <img
                    src={logoUrl}
                    alt={org.companyName}
                    className='w-10 h-10 rounded-lg object-cover'
                    onError={(e) => {
                      const target = e.target as HTMLImageElement
                      target.src = DefaultOrgLogo
                    }}
                  />
                </div>

                {/* 组织信息 */}
                <div className='flex-1 min-w-0 flex flex-col gap-1'>
                  <div className='flex items-center gap-1 min-w-0'>
                    <span
                      className='text-sm font-medium text-[#0C1F17] truncate min-w-0'
                      title={org.companyName}
                    >
                      {org.companyName}
                    </span>
                    {roleText && (
                      <span className='flex-shrink-0 px-2 py-0.5 text-xs font-normal text-[#0C99FF] bg-[#0C99FF1A] rounded-xl'>
                        {roleText}
                      </span>
                    )}
                  </div>
                  <div className='flex items-center gap-[6px]'>
                    <span className='text-xs text-[#6E757F] font-normal truncate'>
                      {org.uinName}
                    </span>
                    {isCurrent && (
                      <span className='flex-shrink-0 text-xs font-normal text-[#0C99FF]'>
                        当前登录
                      </span>
                    )}
                  </div>
                </div>

                {/* 右侧操作区域 - 当前组织不显示 */}
                {!isCurrent && (
                  <div className='flex-shrink-0 flex items-center gap-1'>
                    <span className='text-sm font-normal text-[#0C99FF] opacity-0 group-hover:opacity-100 transition-opacity'>
                      立即进入
                    </span>
                    <div className='relative w-[14px] h-[14px] cursor-pointer'>
                      <HomeArrowRightIcon className='w-[14px] h-[14px] absolute group-hover:opacity-0 transition-opacity' />
                      <HomeArrowRightHoverIcon className='w-[14px] h-[14px] absolute opacity-0 group-hover:opacity-100 transition-opacity' />
                    </div>
                  </div>
                )}
              </div>
            )
          })}
        </div>
      </div>

      {/* 创建组织弹窗 */}
      <CreateOrganizationModal
        open={createModalOpen}
        onCancel={() => setCreateModalOpen(false)}
        onSuccess={handleCreateOrganization}
      />
    </Modal>
  )
}

export default SwitchOrganizationDialog
