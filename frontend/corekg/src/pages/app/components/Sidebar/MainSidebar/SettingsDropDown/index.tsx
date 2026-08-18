import { useMemo, useCallback, type ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { Dropdown, Modal, type MenuProps } from 'antd'
import { useTranslation } from 'react-i18next'
import globalConfig from '@/config'
import { cn } from '@/utils'
import IndustryTermIcon from '@/assets/icons/settings/industry-term.svg?react'
import SynonymIcon from '@/assets/icons/settings/synonym.svg?react'
import { Item } from '@/components/Layout/SidebarWrapper'
import { useMatchRoute } from '@/hooks/useMatchRoute'
import ModelSettingIcon from '@/pages/settings/SettingsSidebar/images/model-setting.svg?react'
import useLocalStore from '@/stores/local'
import { useDeployConfig } from '@/utils/useDeployConfig'
import GroupIcon from '../../images/group.svg?react'
import LogoutIcon from '../../images/logout.svg?react'
import PhoneIcon from '../../images/phone.svg?react'
import SettingsIcon from '../../images/settings.svg?react'
import UserIcon from '../../images/user.svg?react'
import { EExpandableStatus } from '../../types'
import styles from './index.module.scss'

interface ISettingsDropDown {
  status: EExpandableStatus
}

export default function SettingsDropDown(props: ISettingsDropDown) {
  const { version, mode } = useDeployConfig()
  const { t } = useTranslation('pages')
  const { t: tCommon } = useTranslation('common')
  const navigate = useNavigate()
  const { isPathActive } = useMatchRoute()
  const { uinList, userInfo, setLogout } = useLocalStore()
  const currentUin = useMemo(() => {
    return uinList.find((item) => String(item.id) === String(userInfo.uinId))
  }, [uinList, userInfo.uinId])

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

  const buildMenuItem = (
    key: string,
    label: string,
    icon: ReactNode,
    onClick: () => void,
    customClassName?: string,
  ): Exclude<MenuProps['items'], undefined>[number] => {
    return {
      key,
      label: (
        <Item
          className={cn(customClassName && styles[customClassName])}
          text={label}
          icon={icon}
          onClick={onClick}
        />
      ),
    }
  }

  const menuItems = [
    buildMenuItem(
      'profile',
      t('app.sidebar.personalCenter'),
      <UserIcon />,
      () => {
        navigate('settings/profile')
      },
      'profile-item',
    ),
  ] as Exclude<MenuProps['items'], undefined>

  if (currentUin?.role === 'sys_admin') {
    menuItems.push(
      buildMenuItem(
        'organization',
        t('app.sidebar.organizationManagement'),
        <GroupIcon />,
        () => {
          navigate('settings/organization')
        },
      ),
      buildMenuItem(
        'personnel',
        t('app.sidebar.contacts'),
        <PhoneIcon />,
        () => {
          navigate('settings/personnel')
        },
      ),
    )

    if (version !== 'saas') {
      menuItems.push(
        buildMenuItem(
          'model-management',
          t('settings.modelSettings'),
          <ModelSettingIcon />,
          () => {
            navigate('/settings/model')
          },
        ),
      )
    }

    if (
      globalConfig.apiEnv === 'test' ||
      (version === 'custom' && (mode === 'cimc' || mode === 'h3c')) ||
      import.meta.env.DEV
    ) {
      menuItems.push(
        buildMenuItem('synonym', '同义词管理', <SynonymIcon />, () => {
          navigate('/settings/synonym')
        }),
        buildMenuItem(
          'industry-term',
          '行业名词管理',
          <IndustryTermIcon />,
          () => {
            navigate('/settings/industry-term')
          },
        ),
      )
    }
  }

  // 添加退出登录选项作为最后一项
  menuItems.push(
    buildMenuItem(
      'logout',
      t('app.sidebar.loginOut'),
      <LogoutIcon />,
      handleLogout,
      'logout-item',
    ),
  )

  const isActive = useMemo(() => {
    return ['/settings', '/organization', '/personnel'].some((path) =>
      isPathActive(path),
    )
  }, [isPathActive])

  return (
    <Dropdown
      rootClassName={styles.settingsDropDown}
      menu={{ items: menuItems }}
      placement='bottomLeft'
    >
      <div>
        <Item
          status={props.status}
          text={t('app.sidebar.settings')}
          icon={<SettingsIcon />}
          active={isActive}
        />
      </div>
    </Dropdown>
  )
}
