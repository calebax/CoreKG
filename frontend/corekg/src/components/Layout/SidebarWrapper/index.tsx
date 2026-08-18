import { FC, useCallback, useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { Button, Dropdown, Modal, Tooltip, Badge, type MenuProps } from 'antd'
import { Input, Image } from 'antd'
import { useRequest } from 'ahooks'
import { useTranslation } from 'react-i18next'
import globalConfig from '@/config'
import { cn } from '@/utils'
import { getAllUin, switchLogin } from '@/api/account'
import { fetchOrganizationProfile } from '@/api/organization'
import HomeQAIcon from '@/assets/icons/home/home-qa.svg?react'
import SearchIcon from '@/assets/icons/home/home-search.svg?react'
import SwitchIcon from '@/assets/icons/home/home-switch-organization.svg?react'
import UserIconDefault from '@/assets/icons/userIcon-default.svg'
import { ConcatUs } from '@/components/ConcatUs'
import GroupIcon from '@/pages/app/components/Sidebar/images/group.svg?react'
import LogoutIcon from '@/pages/app/components/Sidebar/images/logout.svg?react'
import UserIcon from '@/pages/app/components/Sidebar/images/user.svg?react'
import useLocalStore, { type UinInfo } from '@/stores/local'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { useLoginGlobalData } from '@/utils/useLoginGlobalData'
import Item from './Item'
import { NotificationSidebar } from './NotificationSidebar'
import { SearchOverlay } from './SearchOverlay'
import SwitchOrganizationDialog from './SwitchOrganizationDialog'
import ConcatIcon from './images/concat.svg?react'
import DefaultOrgLogo from './images/corekg.svg'
import ExpandIcon from './images/expand.svg?react'
import FoldIcon from './images/fold.svg?react'
import HelpIcon from './images/help.svg?react'
import MessageIcon from './images/messge.svg?react'
import ReleaseAnnouncementIcon from './images/release-announcement.svg?react'
import TagIcon from './images/tag.svg?react'
import styles from './index.module.scss'
import { EExpandableStatus, ISidebarWrapperProps } from './types'

export { EExpandableStatus }
export type { ISidebarWrapperProps }

export default function SidebarWrapper({
  children,
  expandableStatus,
  updateExpandableStatus,
}: ISidebarWrapperProps) {
  const { version, mode } = useDeployConfig()
  const title = 'CoreKg'
  const { t } = useTranslation('pages')
  const { t: tCommon } = useTranslation('common')
  const { uinList, userInfo, setLogin, setLogout, token } = useLocalStore()
  const { messageNotificationCount } = useLoginGlobalData()
  const [searchOverlayVisible, setSearchOverlayVisible] = useState(false)
  const [switchOrgDialogVisible, setSwitchOrgDialogVisible] = useState(false)
  const [notificationVisible, setNotificationVisible] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const [currentPathname, setCurrentPathname] = useState(location.pathname)
  // 延迟子内容显示，让宽度动画先完成
  const [showContent, setShowContent] = useState(true)

  useEffect(() => {
    setCurrentPathname(location.pathname)
  }, [location.pathname])

  useEffect(() => {
    const handleGlobalRouteChange = (event: Event) => {
      const pathname = (event as CustomEvent<{ pathname?: string }>).detail
        ?.pathname
      if (pathname) {
        setCurrentPathname(pathname)
      }
    }
    window.addEventListener('globalSessionRouteChange', handleGlobalRouteChange)
    return () => {
      window.removeEventListener(
        'globalSessionRouteChange',
        handleGlobalRouteChange,
      )
    }
  }, [])

  const handleExpandableStatusChange = () => {
    switch (expandableStatus) {
      case EExpandableStatus.FOLD:
        // 展开时，先隐藏子内容，让宽度动画先完成
        setShowContent(false)
        // 使用 requestAnimationFrame 确保在下一帧更新，让浏览器先处理宽度变化
        requestAnimationFrame(() => {
          updateExpandableStatus(EExpandableStatus.EXPAND)
          // 在下一帧显示子内容，确保宽度变化已完成
          requestAnimationFrame(() => {
            setShowContent(true)
          })
        })
        break
      case EExpandableStatus.EXPAND:
        updateExpandableStatus(EExpandableStatus.FOLD)
        setShowContent(true) // 收起时立即显示，因为收起很快
        break
    }
  }

  const isFold = expandableStatus === EExpandableStatus.FOLD

  const currentOrg = useMemo(() => {
    try {
      const current = uinList.find(
        (x: any) => String(x.id) === String(userInfo.uinId),
      )
      if (current) {
        return {
          name: current.companyName || title,
          logo: current.logo || DefaultOrgLogo,
        }
      }
    } catch (e) {
      // ignore and fallback
    }
    return {
      name: title,
      logo: DefaultOrgLogo,
    }
  }, [uinList, userInfo.uinId, title])

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
            logo: x.company_logo, // 组织logo
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

  const handleLogout = useCallback(() => {
    Modal.confirm({
      title: t('app.sidebar.loginOut'),
      content: t('app.sidebar.confirmToLogOut'),
      okText: tCommon('button.confirm'),
      cancelText: tCommon('button.cancel'),
      cancelButtonProps: {
        className: 'hover:!text-[#0C99FF] hover:!border-[#0C99FF]',
      },
      okButtonProps: {
        className: '!bg-[#0C99FF] !text-[#ffffff]',
      },
      onOk: () => {
        setLogout()
      },
    })
  }, [setLogout, t, tCommon])

  const renderItem = (item: UinInfo, active: boolean) => {
    const roleIsAdmin = item.role === 'sys_admin'
    const roleText = roleIsAdmin ? t('app.sidebar.admin') : ''
    return (
      <div
        className={cn(styles.settingsDropDownItem, {
          [styles.settingsDropDownItemActive]: active,
        })}
      >
        <div className={styles.settingsDropDownItemName}>
          <div>{item.companyName}</div>
        </div>
        {roleIsAdmin && (
          <div className={styles.settingsDropDownItemRole}>{roleText}</div>
        )}
      </div>
    )
  }

  const renderSwitcher = () => {
    if (version !== 'saas') return null

    return (
      <Dropdown
        rootClassName={styles.settingsDropDown}
        menu={{
          items: uinList.map((item) => {
            const active = String(item.id) === String(userInfo.uinId)
            return {
              key: String(item.id),
              label: renderItem(item, active),
              onClick: async ({ domEvent }) => {
                domEvent.stopPropagation()
                if (active) return
                await handleSwitchOrganization(item.id)
              },
            }
          }),
        }}
        placement='bottomLeft'
        align={{ offset: [-50, 8] }}
        trigger={['click']}
        mouseLeaveDelay={0.3}
        overlayStyle={{ marginTop: 8, marginLeft: -20 }}
      ></Dropdown>
    )
  }

  const currentUin = useMemo(() => {
    return uinList.find((item) => String(item.id) === String(userInfo.uinId))
  }, [uinList, userInfo.uinId])

  const displayUserName = useMemo(() => {
    return userInfo.name
  }, [userInfo.name])
  const buildMenuItem = (
    key: string,
    label: string,
    icon: React.ReactNode,
    onClick: () => void,
  ): Exclude<MenuProps['items'], undefined>[number] => {
    return {
      key,
      label: (
        <div
          className={cn(styles.userMenuItem, {
            [styles.userMenuItemActive]:
              (key === 'profile' &&
                window.location.pathname.includes('/settings/profile')) ||
              (key === 'organization' &&
                window.location.pathname.includes('/settings/organization')) ||
              (key === 'personnel' &&
                window.location.pathname.includes('/settings/personnel')) ||
              (key === 'external-data' &&
                window.location.pathname.includes(
                  '/settings/account-bindings',
                )) ||
              (key === 'model-management' &&
                window.location.pathname.includes('/settings/model')) ||
              (key === 'announcement' &&
                window.location.pathname.includes('/announcement')),
          })}
          onClick={onClick}
        >
          <div className={styles.userMenuItemIcon}>{icon}</div>
          <span>{label}</span>
        </div>
      ),
    }
  }

  const renderUserMenuItems = () => {
    const menuItems: Exclude<MenuProps['items'], undefined> = [
      buildMenuItem(
        'profile',
        t('app.sidebar.personalCenter'),
        <UserIcon />,
        () => {
          navigate('/settings/profile')
        },
      ),
    ]

    if (currentUin?.role === 'sys_admin') {
      menuItems.push(
        buildMenuItem(
          'organization',
          t('app.sidebar.organizationManagement'),
          <GroupIcon />,
          () => {
            navigate('/settings/organization')
          },
        ),
      )
    }

    // custom 版本环境下不显示切换组织选项
    if (version !== 'custom') {
      menuItems.push(
        buildMenuItem(
          'switch-org',
          t('app.sidebar.switchTeam'),
          <SwitchIcon />,
          () => {
            setSwitchOrgDialogVisible(true)
          },
        ),
      )
      // 没有联系我们
      menuItems.push({
        key: 'concat',
        label: (
          <ConcatUs className={styles.userMenuItem}>
            <div className={styles.userMenuItemIcon}>
              <ConcatIcon />
            </div>
            联系我们
          </ConcatUs>
        ),
      })
    }

    // custom 版本环境下不显示发版公告
    if (version !== 'custom') {
      menuItems.push(
        buildMenuItem(
          'announcement',
          '发版公告',
          <ReleaseAnnouncementIcon />,
          () => {
            navigate('/announcement')
          },
        ),
      )
    }

    // 环境判断：只在本地环境、测试环境、生产环境、或 custom 版本且 mode 为 cimc/h3c 时显示基础配置
    const isDevEnv = import.meta.env.MODE === 'development'
    const isTestEnv = import.meta.env.MODE === 'test'
    const isProdEnv = import.meta.env.MODE === 'production'
    const shouldShowBasicConfig =
      isDevEnv ||
      isTestEnv ||
      isProdEnv ||
      (version === 'custom' && (mode === 'cimc' || mode === 'h3c'))
    if (shouldShowBasicConfig) {
      menuItems.push(
        buildMenuItem('tag-management', '基础配置', <TagIcon />, () => {
          navigate('/settings/tag-group')
        }),
      )
    }

    menuItems.push(
      buildMenuItem(
        'logout',
        t('app.sidebar.loginOut'),
        <LogoutIcon />,
        handleLogout,
      ),
    )

    return menuItems
  }

  const renderUserInfo = () => {
    if (expandableStatus === EExpandableStatus.FOLD) {
      return (
        <Dropdown
          rootClassName={styles.userInfoDropDown}
          menu={{
            items: [
              {
                key: 'user-header',
                label: (
                  <div className={styles.userInfoHeader}>
                    <img
                      src={userInfo.avatar || UserIconDefault}
                      alt={userInfo.name}
                      className={styles.userInfoAvatar}
                    />
                    <div className={styles.userInfoText}>
                      <div
                        className={styles.userInfoName}
                        title={displayUserName}
                      >
                        {displayUserName}
                      </div>
                      <div className={styles.userInfoOrgRow}>
                        <span className={styles.userInfoOrgLabel}>
                          当前组织
                        </span>
                        <span
                          className={styles.userInfoOrgName}
                          title={currentOrg.name}
                        >
                          {currentOrg.name}
                        </span>
                      </div>
                    </div>
                  </div>
                ),
                disabled: true,
              },
              ...renderUserMenuItems(),
            ],
          }}
          placement='topLeft'
          trigger={['click']}
        >
          <div className={styles.userInfoTrigger}>
            <img
              src={userInfo.avatar || UserIconDefault}
              alt={userInfo.name}
              className={styles.userInfoAvatarSmall}
            />
          </div>
        </Dropdown>
      )
    }

    return (
      <Dropdown
        rootClassName={styles.userInfoDropDown}
        menu={{
          items: [
            {
              key: 'user-header',
              label: (
                <div className={styles.userInfoHeader}>
                  <img
                    src={userInfo.avatar || UserIconDefault}
                    alt={userInfo.name}
                    className={styles.userInfoAvatar}
                  />
                  <div className={styles.userInfoText}>
                    <div
                      className={styles.userInfoName}
                      title={displayUserName}
                    >
                      {displayUserName}
                    </div>
                    <div className={styles.userInfoOrgRow}>
                      <span className={styles.userInfoOrgLabel}>当前组织</span>
                      <span
                        className={styles.userInfoOrgName}
                        title={currentOrg.name}
                      >
                        {currentOrg.name}
                      </span>
                    </div>
                  </div>
                </div>
              ),
              disabled: true,
            },
            ...renderUserMenuItems(),
          ],
        }}
        placement='topLeft'
        trigger={['click']}
      >
        <div className={styles.userInfoArea}>
          <img
            src={userInfo.avatar || UserIconDefault}
            alt={userInfo.name}
            className={styles.userInfoAvatarSmall}
          />
          <div className={styles.userInfoContent}>
            <div className={styles.userInfoName} title={displayUserName}>
              {displayUserName || 'admin'}
            </div>
            <div className={styles.userInfoOrg} title={currentOrg.name}>
              @{currentOrg.name || 'CoreKg'}
            </div>
          </div>
        </div>
      </Dropdown>
    )
  }

  const renderSidebarFooter = () => {
    return (
      <>
        {/* 消息与通知 - 非custom环境以及custom环境的mode==='cimc'版本要显示，其他环境不显示 */}
        {version !== 'custom' || (version === 'custom' && mode === 'cimc') ? (
          <Item
            icon={
              <Badge
                dot={messageNotificationCount.count > 0}
                className={styles.messageBadge}
                offset={[-3, 3]}
              >
                <MessageIcon />
              </Badge>
            }
            status={expandableStatus}
            text='消息与通知'
            className={styles.messageNotificationItem}
            onClick={() => setNotificationVisible(true)}
          />
        ) : null}
        {version !== 'international' && mode !== 'cimc' ? (
          <Link
            to={
              version === 'saas'
                ? 'https://docs.corekg.com/docs/corekg/'
                : `/usage_help/?mode=${mode}`
            }
            target='_blank'
          >
            <Item
              icon={<HelpIcon />}
              status={expandableStatus}
              text={t('app.sidebar.helpAndSupport')}
            />
          </Link>
        ) : null}
        {renderUserInfo()}
      </>
    )
  }

  const renderSidebarHeader = () => {
    return (
      <div className={styles.sidebarHeader}>
        {/* 当展开时显示 logo 标题；收起时在左上显示展开图标 */}
        {expandableStatus === EExpandableStatus.EXPAND ? (
          <div className={styles.sidebarHeaderTitle}>
            <Link
              to='/global'
              className={
                styles.sidebarHeaderTitle + ' ' + styles.sidebarHeaderTitleLink
              }
            >
              <img
                src={currentOrg.logo || DefaultOrgLogo}
                className='w-7 h-7 rounded-full'
              />
              {/* 如果有组织名则显示组织名，否则回退到部署 title */}
              <span
                className={styles.sidebarHeaderOrgName}
                title={currentOrg.name || title}
              >
                {currentOrg.name || title}
              </span>
            </Link>
            {/* 将 SwitchIcon 放在 logo 旁边，使其与左侧一体 */}
            {renderSwitcher()}
          </div>
        ) : (
          <div className={styles.sidebarHeaderTitle}>
            <div
              className={cn(styles.sidebarHeaderExpandable, {
                [styles.sidebarHeaderExpandableFold]: isFold,
              })}
            >
              <ExpandIcon onClick={handleExpandableStatusChange} />
            </div>
          </div>
        )}

        <div
          className={styles.sidebarHeaderActions}
          style={{ marginTop: '2px' }}
        >
          {/* 在右侧仅显示收起图标（展开时），Switch 已移至左侧 */}
          {expandableStatus === EExpandableStatus.EXPAND ? (
            <div
              className={cn(styles.sidebarHeaderExpandable, {
                [styles.sidebarHeaderExpandableFold]: isFold,
              })}
            >
              <FoldIcon onClick={handleExpandableStatusChange} />
            </div>
          ) : null}
        </div>
      </div>
    )
  }

  const handleSearchInputClick = useCallback(() => {
    setSearchOverlayVisible(true)
  }, [])

  const handleCloseSearchOverlay = useCallback(() => {
    setSearchOverlayVisible(false)
  }, [])

  const shouldHideSearchAndQA = useMemo(() => {
    const pathname = location.pathname
    // 在任何 settings 二级页面均隐藏（组织管理后台与个人中心均属于 settings 下的页面）
    return pathname.startsWith('/settings')
  }, [location.pathname])

  const renderSearchInput = useMemo(() => {
    if (shouldHideSearchAndQA) return null
    if (expandableStatus === EExpandableStatus.FOLD) {
      return (
        <div
          className={styles.sidebarSearchInput}
          onClick={handleSearchInputClick}
        >
          <Tooltip title='搜索' placement='right'>
            <div className='flex items-center justify-center w-full h-8 cursor-pointer pr-2'>
              <SearchIcon />
            </div>
          </Tooltip>
        </div>
      )
    }
    return (
      <div
        className={styles.sidebarSearchInput}
        onClick={handleSearchInputClick}
      >
        <Input placeholder='搜索' prefix={<SearchIcon />} readOnly />
      </div>
    )
  }, [shouldHideSearchAndQA, expandableStatus, handleSearchInputClick])

  const renderQAMenu = useMemo(() => {
    if (shouldHideSearchAndQA) return null
    return (
      <div className={styles.sidebarQAMenu}>
        <Link
          to='/global'
          onClick={() => window.dispatchEvent(new Event('startGlobalSession'))}
        >
          <Item
            icon={<HomeQAIcon />}
            status={expandableStatus}
            text='问AI'
            className={styles.qaMenuItem}
            active={currentPathname === '/global'}
          />
        </Link>
      </div>
    )
  }, [shouldHideSearchAndQA, expandableStatus, currentPathname])

  return (
    <>
      <div
        className={cn(styles.sidebarWrapper, {
          [styles.sidebarWrapperFold]: isFold,
        })}
      >
        {renderSidebarHeader()}
        {renderSearchInput}
        {showContent ? (
          <div className={styles.sidebarNavigationContainer}>
            {renderQAMenu}
            {children}
          </div>
        ) : (
          <div className={styles.sidebarNavigationContainer} />
        )}
        {renderSidebarFooter()}
      </div>
      <SearchOverlay
        visible={searchOverlayVisible}
        onClose={handleCloseSearchOverlay}
      />
      <SwitchOrganizationDialog
        open={switchOrgDialogVisible}
        onClose={() => setSwitchOrgDialogVisible(false)}
      />
      <NotificationSidebar
        visible={notificationVisible}
        onClose={() => setNotificationVisible(false)}
      />
    </>
  )
}

export { SidebarWrapper, Item }
