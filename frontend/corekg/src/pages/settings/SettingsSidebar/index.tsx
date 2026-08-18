import type { ReactNode } from 'react'
import { useMemo, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import globalConfig from '@/config'
import IndustryTermIcon from '@/assets/icons/settings/industry-term.svg?react'
import SynonymIcon from '@/assets/icons/settings/synonym.svg?react'
import {
  EExpandableStatus,
  SidebarWrapper,
  Item,
} from '@/components/Layout/SidebarWrapper'
import TagIcon from '@/components/Layout/SidebarWrapper/images/tag.svg?react'
import { useMatchRoute } from '@/hooks/useMatchRoute'
import useLocalStore from '@/stores/local'
import { useDeployConfig } from '@/utils/useDeployConfig'
import CollectIcon from './images/collect.svg?react'
import ExternalDataIcon from './images/external-data.svg?react'
import LikeIcon from './images/like.svg?react'
import MemberSettingIcon from './images/member-setting.svg?react'
import ModelSettingIcon from './images/model-setting.svg?react'
import Order from './images/order.svg?react'
import PhoneIcon from './images/phone.svg?react'
import UserIcon from './images/user.svg?react'
import styles from './index.module.scss'

export type SettingsItem = {
  name: string
  path: string
  icon: ReactNode
}
/** setting item 的组别信息 */
type GroupItem = {
  key: string
  item: SettingsItem
  adminOnly?: boolean
  /** 有值则筛选 */
  versions?: DeployConfig['version'][]
  hidden?: boolean
}

export default function SettingsSidebar() {
  const { t } = useTranslation('pages')
  const { version, mode } = useDeployConfig()
  const { equalPathActive } = useMatchRoute()
  const { uinList, userInfo } = useLocalStore()
  const location = useLocation()
  const currentUin = useMemo(() => {
    return uinList.find((item) => String(item.id) === String(userInfo.uinId))
  }, [uinList, userInfo.uinId])

  const [expandableStatus, setExpandableStatus] = useState<EExpandableStatus>(
    EExpandableStatus.EXPAND,
  )

  const groups = useMemo(() => {
    // 环境判断：统一使用 import.meta.env.MODE
    const isDevEnv = import.meta.env.MODE === 'development'
    const isTestEnv = import.meta.env.MODE === 'test'
    const isProdEnv = import.meta.env.MODE === 'production'
    // custom 私有化环境下 cimc 和 h3c 模式均支持这些功能
    const isCimcOrH3cMode =
      version === 'custom' && (mode === 'cimc' || mode === 'h3c')

    // 标签管理、标签分类管理、我的点赞、我的收藏：本地环境、测试环境、生产环境、cimc/h3c环境
    const shouldShowTagManagement =
      isDevEnv || isTestEnv || isProdEnv || isCimcOrH3cMode

    // 同义词管理、行业名词管理：本地环境、测试环境、cimc/h3c环境（不包括生产环境）
    const shouldShowSynonymManagement = isDevEnv || isTestEnv || isCimcOrH3cMode

    // 外部数据源：仅本地环境、测试环境、custom环境且mode为cimc时展示
    const shouldShowExternalDataSource =
      isDevEnv || isTestEnv || (version === 'custom' && mode === 'cimc')

    // 模型管理：本地环境、测试环境均可见；生产环境仅 custom 版本可见
    const shouldShowModelManagement =
      isDevEnv || isTestEnv || version === 'custom'

    const _groups: GroupItem[][] = [
      [
        {
          key: 'profile',
          item: {
            name: t('app.sidebar.personalCenter'),
            path: '/settings/profile',
            icon: <UserIcon />,
          },
        },
        {
          key: 'order-management',
          adminOnly: true,
          versions: ['saas'],
          item: {
            name: '订单管理',
            path: '/settings/order-management',
            icon: <Order />,
          },
        },
        {
          key: 'my-likes',
          item: {
            name: '我的点赞',
            path: '/settings/my-likes',
            icon: <LikeIcon />,
          },
          hidden: !shouldShowTagManagement,
        },
        {
          key: 'my-collections',
          item: {
            name: '我的收藏',
            path: '/settings/my-collections',
            icon: <CollectIcon />,
          },
          hidden: !shouldShowTagManagement,
        },
      ],
      [
        {
          key: 'organization',
          adminOnly: true,
          item: {
            name: t('organization.title'),
            path: '/settings/organization',
            icon: <MemberSettingIcon />,
          },
        },
        {
          key: 'personnel',
          adminOnly: true,
          item: {
            name: t('app.sidebar.contacts'),
            path: '/settings/personnel',
            icon: <PhoneIcon />,
          },
        },
        {
          key: 'account-bindings',
          adminOnly: true,
          item: {
            name: t('settings.externalDataSource'),
            path: '/settings/account-bindings',
            icon: <ExternalDataIcon />,
          },
          hidden: !shouldShowExternalDataSource,
        },
        {
          key: 'model',
          adminOnly: true,
          item: {
            name: t('app.sidebar.modelManagement'),
            path: '/settings/model',
            icon: <ModelSettingIcon />,
          },
          hidden: !shouldShowModelManagement,
        },
      ],
      // 标签管理和标签分类管理组（加上生产环境）
      [
        {
          key: 'tag',
          item: {
            name: '标签分类管理',
            path: '/settings/tag-group',
            icon: <TagIcon />,
          },
          hidden: !shouldShowTagManagement,
        },
        {
          key: 'tag',
          item: {
            name: '标签管理',
            path: '/settings/tag',
            icon: <TagIcon />,
          },
          hidden: !shouldShowTagManagement,
        },
      ],
      // 同义词管理和行业名词管理组（本地环境、测试环境、cimc/h3c环境，不包括生产环境）
      [
        {
          key: 'synonym',
          item: {
            name: '同义词管理',
            path: '/settings/synonym',
            icon: <SynonymIcon />,
          },
          hidden: !shouldShowSynonymManagement,
        },
        {
          key: 'industry-term',
          item: {
            name: '行业名词管理',
            path: '/settings/industry-term',
            icon: <IndustryTermIcon />,
          },
          hidden: !shouldShowSynonymManagement,
        },
      ],
    ]

    return _groups
  }, [t, version, mode])

  const items: SettingsItem[] = useMemo(() => {
    // 基础配置相关的路径：标签、同义词、行业名词
    const basicConfigKeys = ['tag', 'synonym', 'industry-term']
    const isBasicConfigPath = basicConfigKeys.some((key) =>
      location.pathname.includes(key),
    )

    // 如果当前路径是基础配置相关路径，需要合并多个组
    let targetGroups: GroupItem[][] = []
    if (isBasicConfigPath) {
      // 找到所有基础配置相关的组（标签组 + 同义词/行业名词组）
      targetGroups = groups.filter((g) =>
        g.some((item) => basicConfigKeys.includes(item.key)),
      )
    } else {
      // 其他路径，只显示匹配的组
      const matchedGroup = groups.find((g) =>
        g.some((item) => location.pathname.includes(item.key)),
      )
      if (matchedGroup) {
        targetGroups = [matchedGroup]
      }
    }

    // 合并所有匹配组中的菜单项
    const allItems = targetGroups.flat().filter((item) => {
      const { adminOnly, versions, hidden } = item
      return (
        !hidden &&
        (!adminOnly || currentUin?.role === 'sys_admin') &&
        (!versions || versions.includes(version))
      )
    })

    return allItems.map((v) => v.item)
  }, [groups, location.pathname, currentUin?.role, version])
  const renderItem = (item: SettingsItem) => {
    return (
      <Link key={item.path} to={item.path}>
        <Item
          active={equalPathActive(item.path)}
          text={item.name}
          icon={item.icon}
        />
      </Link>
    )
  }

  return (
    <SidebarWrapper
      expandableStatus={expandableStatus}
      updateExpandableStatus={setExpandableStatus}
    >
      <div className={styles.settingsSidebar}>{items.map(renderItem)}</div>
    </SidebarWrapper>
  )
}
