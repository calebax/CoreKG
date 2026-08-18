import { FC } from 'react'
import { App } from 'antd'
import { useMount } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { AppIcon } from 'tdesign-icons-react'
import { cn } from '@/utils'
import { Layout } from '@/components/Layout/Layout'
import { MenuWrapper } from '@/components/Layout/MenuWrapper'
import { useMatchRoute } from '@/hooks/useMatchRoute'
import useLocalStore from '@/stores/local'
import SettingsSidebar from './SettingsSidebar'
// 知识精灵
import Elf from './images/elf.svg?react'
import MemberManagement from './images/member-management.svg?react'
import User from './images/user.svg?react'

const Menu: FC = () => {
  const { t } = useTranslation('pages')
  const { pathname } = useLocation()
  const items = [
    {
      name: t('settings.memberManagement'),
      path: '/settings',
      icon: <MemberManagement />,
    },
    {
      name: '模型设置',
      path: '/settings/model',
      icon: <AppIcon />,
    },
    // { name: '知识精灵', path: '/settings/elf', icon: <Elf /> },
  ]
  return (
    <MenuWrapper hiddenCollapseIcon>
      <div className='flex flex-col px-4 gap-3'>
        {items.map((item) => {
          const { name, path, icon } = item
          const active = (() => {
            if (path === '/settings') return path === pathname
            return pathname.startsWith(path)
          })()
          return (
            <Link
              key={path}
              to={path}
              className={cn(
                'w-full rounded px-3 py-[9px]',
                'flex items-center gap-2',
                'bg-transparent text-base text-[#1E1F28] font-medium leading-[26px]',
                {
                  'bg-[#E6E8F0]': active,
                  'hover:bg-[#fcfcfe] hover:shadow-[0_1px_3px_rgba(29,33,41,0.1)]':
                    !active,
                },
              )}
            >
              {icon}
              {name}
            </Link>
          )
        })}
      </div>
    </MenuWrapper>
  )
}

const SettingLayout: FC = () => {
  const { pathname } = useLocation()
  const { t: tM } = useTranslation('messages')
  const { message } = App.useApp()
  const navigate = useNavigate()
  const uinList = useLocalStore((state) => state.uinList)
  const userInfo = useLocalStore((state) => state.userInfo)
  const currentUin = useMemo(() => {
    return uinList.find((item) => '' + item.id === '' + userInfo.uinId)
  }, [uinList, userInfo.uinId])
  // 使用useEffect代替useMount，并添加pathname作为依赖
  useEffect(() => {
    if (pathname === '/settings') {
      navigate('/settings/profile', { replace: true })
    }
    if (
      currentUin?.role !== 'sys_admin' &&
      ['/settings/organization', '/settings/personnel'].includes(pathname)
    ) {
      navigate('/', { replace: true })
      message.warning(tM('adminOnlySystemSettingsAccess'))
    }
  }, [pathname, currentUin?.role, navigate, message, tM])

  if (
    currentUin?.role !== 'sys_admin' &&
    ['/settings/organization', '/settings/personnel'].includes(pathname)
  ) {
    return null
  }
  return (
    <Layout>
      <SettingsSidebar />
    </Layout>
  )
}
export default SettingLayout
