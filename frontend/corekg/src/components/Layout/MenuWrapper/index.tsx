import { FC, PropsWithChildren, ReactNode, useMemo } from 'react'
import { Link, NavLink, useNavigate } from 'react-router-dom'
import { Dropdown, Button, Modal, App, Popover } from 'antd'
import type { MenuProps } from 'antd'
import { ExclamationCircleOutlined, CheckOutlined } from '@ant-design/icons'
import config from '@/config'
import { cn } from '@/utils'
import { getAllUin, switchLogin } from '@/api/account'
import useLocalStore from '@/stores/local'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { useVersion } from '@/utils/useVersion'
import { Help } from '../Header/Help'
import { VersionBtn } from '../Header/Version'
import Logout from '../Header/images/logout.svg?react'
import Settings from '../Header/images/setting.svg?react'
import Team from '../Header/images/team.svg?react'
import User from '../Header/images/user.svg?react'
import styles from '../Header/styles.module.scss'
import Arrow2Left from './images/arrow2-left.svg?react'
import Arrow2Right from './images/arrow2-right.svg?react'

export type MenuWrapper = PropsWithChildren & {
  className?: string
  style?: React.CSSProperties
  collapsed?: boolean
  setCollapsed?: (val: boolean) => void
  /** 菜单顶部 传入undefined展示默认顶部 */
  header?: ReactNode
  /** 是否隐藏菜单折叠控制按钮 */
  hiddenCollapseIcon?: boolean
}
export const MenuWrapper: FC<MenuWrapper> = (props) => {
  const {
    className,
    style,
    children,
    collapsed,
    setCollapsed,
    header,
    hiddenCollapseIcon,
  } = props

  const { message } = App.useApp()
  const { uinList, userInfo, setLogin, setLogout, token } = useLocalStore()
  const navigate = useNavigate()

  // 使用部署配置Hook
  const deployConfig = useDeployConfig()

  const currentUin = useMemo(() => {
    return uinList.find((item) => item.id === userInfo.uinId)
  }, [uinList, userInfo.uinId])

  const items: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <User />,
      label: <NavLink to='/profile'>个人中心</NavLink>,
    },
    {
      key: 'logout',
      label: '退出登录',
      danger: true,
      icon: <Logout />,
      onClick: () => {
        Modal.confirm({
          title: '退出登录',
          icon: <ExclamationCircleOutlined />,
          content: '确定退出登录吗？',
          okText: '确认',
          cancelText: '取消',
          onOk: () => {
            setLogout()
          },
        })
      },
    },
  ]

  if (currentUin?.role === 'sys_admin') {
    items.splice(1, 0, {
      key: 'settings',
      icon: <Settings />,
      label: <NavLink to={'/settings'}>系统设置</NavLink>,
    })
  }

  const menuHeader = useMemo(() => {
    if (header !== undefined) {
      return header
    }
    if (collapsed) {
      return (
        <span
          className={cn('h-16 relative', 'flex items-center justify-center')}
        >
          {hiddenCollapseIcon ? null : (
            <Arrow2Right
              className='cursor-pointer'
              onClick={() => setCollapsed?.(false)}
            />
          )}
        </span>
      )
    }

    return (
      <span className={cn('pl-8 h-16', 'flex gap-4 items-center')}>
        <Link to={'/'} className='flex gap-1.5 items-center'>
          <img src={deployConfig.logo} alt='logo' className='w-6 h-6' />
          <img src={deployConfig.title} alt='title' />
        </Link>
        {hiddenCollapseIcon ? null : (
          <Arrow2Left
            className='cursor-pointer ml-auto mr-5'
            onClick={() => setCollapsed?.(true)}
          />
        )}
      </span>
    )
  }, [collapsed, header, hiddenCollapseIcon, setCollapsed])

  const userArea = useMemo(() => {
    if (collapsed) {
      return (
        <div className='flex flex-col items-center py-2 gap-2'>
          <Dropdown
            menu={{ items }}
            rootClassName={styles.menu}
            placement='topRight'
          >
            <img
              src={userInfo.avatar}
              alt='avatar'
              className='w-8 h-8 rounded-full cursor-pointer'
            />
          </Dropdown>
        </div>
      )
    }

    return (
      <div className='px-4 py-3'>
        <div className='flex flex-col gap-6 mb-4'>
          <div className='flex flex-col items-center py-1 w-full max-w-[198px] mx-auto'>
            <div className='flex flex-col gap-4 items-center w-full'>
              <VersionBtn
                className='bg-[#EAF2FF] border-[#E6E8F0] text-[#1E1F28] hover:bg-[#E6E8F0] hover:border-[#E6E8F0] hover:text-[#1E1F28] h-10 w-full rounded font-medium text-[14px] leading-[19.759px]'
                style={{
                  fontFamily: 'Inter ',
                  background: '#EAF2FF',
                  border: '1px solid #EAF2FF',
                  color: '#1E1F28',
                  fontWeight: 500,
                }}
              />
              {/* API市场按钮暂时注释掉 */}
              {/* <Button
                className={cn(
                  'w-full h-10',
                  'border-2 border-[#165DFF] text-[#165DFF] font-medium bg-transparent hover:bg-transparent hover:border-[#165DFF] hover:text-[#165DFF]',
                  'rounded text-[14px]',
                )}
                style={{ fontFamily: 'Inter ' }}
                onClick={() => {
                  message.info('API市场正在开发中，敬请期待')
                }}
              >
                API市场
              </Button> */}
              <Button
                className={cn(
                  'w-full h-10',
                  'border-2 border-[#165DFF] text-[#165DFF] font-medium bg-transparent hover:bg-transparent hover:border-[#165DFF] hover:text-[#165DFF]',
                  'rounded text-[14px]',
                )}
                style={{ fontFamily: 'Inter ' }}
                onClick={() => {
                  const { protocol, hostname, port } = window.location
                  window.open(
                    `${protocol}//${hostname}${port ? `:${port}` : ''}/usage_help/?mode=${deployConfig.mode}`,
                    '_blank',
                  )
                }}
              >
                帮助文档
              </Button>
            </div>
          </div>
        </div>

        <div className=''>
          <Dropdown
            menu={{ items }}
            rootClassName={styles.menu}
            placement='topRight'
          >
            <div className='flex items-center justify-between p-2 hover:bg-[#F0F2F7] rounded cursor-pointer'>
              <div className='flex items-center gap-1.5'>
                <img
                  src={userInfo.avatar}
                  alt='avatar'
                  className='w-6 h-6 rounded-full'
                />
                <div className='flex flex-col justify-center'>
                  <div
                    className='text-[10.286px] font-medium text-[#1E1F28] leading-normal'
                    style={{ fontFamily: 'Inter ' }}
                  >
                    {currentUin?.uinName}
                  </div>
                </div>
              </div>
              <div className='flex items-center justify-center w-4 h-4 transform scale-y-[-1]'>
                <svg width='16' height='16' viewBox='0 0 16 16' fill='none'>
                  <path
                    d='M4 6l4 4 4-4'
                    stroke='currentColor'
                    strokeWidth='1.5'
                    strokeLinecap='round'
                    strokeLinejoin='round'
                  />
                </svg>
              </div>
            </div>
          </Dropdown>
        </div>

        {config.env !== 'production' ? (
          <div className='text-xs text-center py-1'>
            {config.env === 'development' ? 'dev' : 'test'}: {config.version}
          </div>
        ) : null}
      </div>
    )
  }, [collapsed, items, userInfo.avatar, currentUin, message])

  return (
    <div
      className={cn(
        'flex flex-col',
        'bg-[#F8F9FD]',
        'transition-all duration-200',
        collapsed ? 'w-16' : 'w-57.5',
        className,
      )}
      style={style}
    >
      {menuHeader}
      <div className='flex-1 overflow-hidden'>{children}</div>
      {userArea}
    </div>
  )
}
