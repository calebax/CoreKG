import { useMemo, type ReactNode, type MouseEvent } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { message } from 'antd'
import { useTranslation } from 'react-i18next'
import { getCoze } from '@/api'
import { hasModulePermission } from '@/utils'
import { Item } from '@/components/Layout/SidebarWrapper'
import { useMatchRoute } from '@/hooks/useMatchRoute'
import useLocalStore from '@/stores/local'
import { resolveDeployUrl, useDeployConfig } from '@/utils/useDeployConfig'
import { useLoginGlobalData } from '@/utils/useLoginGlobalData'
import ApplicationIcon from '../images/application.svg?react'
import AgentIcon from '../images/agent.svg?react'
import CozeIcon from '../images/coze.svg?react'
import GraphIcon from '../images/graph.svg?react'
import KnowledgeBaseIcon from '../images/knowledgeBase.svg?react'
import { EExpandableStatus } from '../types'
import styles from './index.module.scss'

const AGENT_HOME_STATUS_KEY = 'agent_resources_is_home'
const AGENT_NAVIGATE_HOME_EVENT = 'agent-resources-navigate-home'

interface IMainSidebarProps {
  expandableStatus: EExpandableStatus
  onExpandableStatusChange: () => void
}

type MenuItem = {
  name: string
  path?: string
  icon: ReactNode
  external?: boolean
}

export default function MainSidebar(props: IMainSidebarProps) {
  const { t: tM } = useTranslation('messages')
  const { t } = useTranslation('pages')
  const { isPathActive } = useMatchRoute()
  const location = useLocation()
  const { coze_url, version } = useDeployConfig()
  const { license } = useLoginGlobalData()
  const baseUrl = useMemo(() => {
    const { protocol, hostname, port } = window.location
    return `${protocol}//${hostname}${port ? `:${port}` : ''}`
  }, [])

  // 原有逻辑：从配置中获取 coze_url
  // const cozeHref = useMemo(() => {
  //   return resolveDeployUrl(baseUrl, coze_url, `${baseUrl}:30080/`)
  // }, [baseUrl, coze_url])

  // 新逻辑：使用原主机名 + /coze 路径
  const cozeHref = useMemo(() => {
    const { protocol, hostname, port } = window.location
    return `${protocol}//${hostname}${port ? `:${port}` : ''}/coze`
  }, [])

  const menuList = useMemo(() => {
    const items: MenuItem[] = [
      {
        name: t('app.sidebar.applications'),
        path: '/apps',
        icon: <ApplicationIcon />,
      },
      {
        name: t('app.sidebar.knowledgeBase'),
        path: '/docs',
        icon: <KnowledgeBaseIcon />,
      },
    ]
    if (version !== 'international') {
      const baseMenu = [
        {
          name: t('app.sidebar.lightApplication'),
          path: '/agents',
          icon: <AgentIcon />,
        },
        // {
        //   name: t('app.sidebar.coze'),
        //   icon: <CozeIcon />,
        //   path: '/coze',
        //   external: true,
        // },
      ]
      // graph 模块权限检查
      if (hasModulePermission(license, 'graph')) {
        baseMenu.unshift({
          name: t('app.sidebar.knowledgeGraph'),
          path: '/graph',
          icon: <GraphIcon />,
        })
      }

      items.push(...baseMenu)
    }

    return items
  }, [t, version, license])

  const handleCozeClick = async (e: MouseEvent) => {
    e.preventDefault()
    const response = await getCoze({})
    console.log(response)
    if (response.code === 0) {
      message.success(
        tM('accessTargetSuccess', { target: t('app.sidebar.coze') }),
      )
      const token = useLocalStore.getState().token
      let targetUrl = cozeHref
      try {
        const url = new URL(cozeHref)
        if (token) {
          url.searchParams.set('auth_token', token)
        }
        targetUrl = url.toString()
      } catch (err) {
        // 如果 URL 构造失败，直接使用原始链接
      }

      const newWindow = window.open(targetUrl, '_blank')
      if (newWindow && token) {
        try {
          const targetOrigin = new URL(targetUrl).origin
          // 通过 postMessage 传递 token，作为 URL 携带的备选方案
          newWindow.postMessage(
            { type: 'set_token', payload: { token } },
            targetOrigin,
          )
          // 轻微延迟重发一次，提升跨域页面监听尚未就绪时的可靠性
          setTimeout(() => {
            newWindow.postMessage(
              { type: 'set_token', payload: { token } },
              targetOrigin,
            )
          }, 500)
        } catch (err) {
          // 忽略跨域或目标窗口不可用异常
        }
      }
    } else {
      message.error(
        tM('accessTargetFailedPleaseTryAgainLater', {
          target: t('app.sidebar.coze'),
        }),
      )
    }
  }

  const renderItem = (item: MenuItem) => {
    if (item.external) {
      return (
        <Link to={item.path!} target='_blank'>
          <Item
            key={item.path || item.name}
            status={props.expandableStatus}
            text={item.name}
            path={item.path}
            icon={item.icon}
            active={false}
          />
        </Link>
      )
    }

    if (item.path === '/agents') {
      return (
        <Link
          to={item.path}
          key={item.path}
          onClick={(event) => {
            if (location.pathname !== '/agents') return

            event.preventDefault()
            const isAgentHome = sessionStorage.getItem(AGENT_HOME_STATUS_KEY)

            // coze 已上报当前就在首页时，点击左侧菜单不做任何刷新动作
            if (isAgentHome === 'true') {
              return
            }

            // 已经在智能体页但不在首页时，让资源页转发一条“回首页”消息给 coze iframe
            window.dispatchEvent(new CustomEvent(AGENT_NAVIGATE_HOME_EVENT))
          }}
        >
          <Item
            status={props.expandableStatus}
            text={item.name}
            path={item.path}
            icon={item.icon}
            active={isPathActive(item.path!)}
          />
        </Link>
      )
    }

    return (
      <Link to={item.path!} key={item.path}>
        <Item
          status={props.expandableStatus}
          text={item.name}
          path={item.path}
          icon={item.icon}
          active={isPathActive(item.path!)}
        />
      </Link>
    )
  }

  return (
    <div className={styles.mainSidebar}>
      <div className={styles.mainSidebarContent}>
        {menuList.map(renderItem)}
      </div>
    </div>
  )
}
