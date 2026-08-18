import { FC, useEffect, useMemo } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { message } from 'antd'
import { getCoze } from '@/api'
import { cn } from '@/utils'
import { MenuWrapper } from '@/components/Layout/MenuWrapper'
import useLocalStore from '@/stores/local'
import { resolveDeployUrl, useDeployConfig } from '@/utils/useDeployConfig'
import { Item } from './Item'
import { SessionHistory } from './SessionHistory'
import AgentActive from './images/agent-active.svg?react'
import Agent from './images/agent.svg?react'
import Coze from './images/coze.svg?react'
import ForestActive from './images/forest-active.svg?react'
import Forest from './images/forest.svg?react'
import Graph from './images/graph.svg?react'
import MainActive from './images/main-active.svg?react'
import Main from './images/main.svg?react'

export const Menu: FC = () => {
  const { coze_url } = useDeployConfig()
  const { pathname } = useLocation()
  const { sidebarCollapsed, setSidebarCollapsed } = useLocalStore()
  const baseUrl = useMemo(() => {
    const { protocol, hostname, port } = window.location
    return `${protocol}//${hostname}${port ? `:${port}` : ''}`
  }, [])
  const cozeHref = useMemo(() => {
    return resolveDeployUrl(baseUrl, coze_url, `${baseUrl}:30080/`)
  }, [baseUrl, coze_url])

  useEffect(() => {
    // 如果当前页面不是这些pages或主页 则自动收起
    const pages = ['/docs', '/agents', '/QA']
    if (pathname === '/' || pages.includes(pathname)) return
    setSidebarCollapsed(true)
  }, [pathname, setSidebarCollapsed])

  const menuList = [
    { name: '主页', path: '/', icon: <Main />, iconActive: <Main /> },
    {
      name: '知识库',
      path: '/docs',
      icon: <Forest />,
      iconActive: <Forest />,
    },
    {
      name: '智能体',
      path: '/agents',
      icon: <Agent />,
      iconActive: <Agent />,
    },
    {
      name: '知识图谱',
      path: '/graph',
      icon: <Graph />,
      iconActive: <Graph />,
    },
    {
      name: 'Coze',
      path: coze_url || '/coze',
      icon: <Coze />,
      iconActive: <Coze />,
      isSpecial: true, // 标记为特殊项，需要特殊处理
    },
  ]

  const handleCozeClick = async (e: React.MouseEvent) => {
    e.preventDefault()
    const response = await getCoze({})
    console.log(response)
    if (response.code === 0) {
      message.success('访问Coze成功')
      // 构造目标URL，附带当前登录token
      const token = useLocalStore.getState().token
      try {
        const urlStr = cozeHref || coze_url || '/coze'
        const targetUrl = new URL(urlStr, baseUrl)
        if (token) {
          targetUrl.searchParams.set('auth_token', token)
        }
        const newWin = window.open(targetUrl.toString(), '_blank')
        if (newWin && token) {
          const targetOrigin = targetUrl.origin
          try {
            newWin.postMessage(
              { type: 'set_token', payload: { token } },
              targetOrigin,
            )
            setTimeout(() => {
              newWin.postMessage(
                { type: 'set_token', payload: { token } },
                targetOrigin,
              )
            }, 500)
          } catch (err) {
            // noop
          }
        }
      } catch (err) {
        window.open(`${coze_url}`, '_blank')
      }
    } else {
      message.error('访问Coze失败，请稍后重试')
    }
  }

  return (
    <MenuWrapper
      collapsed={sidebarCollapsed}
      setCollapsed={setSidebarCollapsed}
    >
      <div className={cn('h-full overflow-hidden', 'flex flex-col gap-3')}>
        {menuList.map((item) => {
          const { path, icon, iconActive, name, isSpecial } = item
          const active = (() => {
            if (path === '/') {
              return path === pathname
            } else {
              return pathname.startsWith(path)
            }
          })()

          // Coze特殊处理
          if (isSpecial && name === 'Coze') {
            return (
              <div
                key={path}
                className={cn(
                  'text-base cursor-pointer',
                  sidebarCollapsed ? 'mx-3' : 'mx-4',
                )}
                onClick={handleCozeClick}
              >
                <Item
                  collapsed={sidebarCollapsed}
                  title={name}
                  active={false} // Coze不需要激活状态
                  icon={icon}
                  iconActive={iconActive}
                />
              </div>
            )
          }

          return (
            <Link
              to={path}
              key={path}
              className={cn('text-base', sidebarCollapsed ? 'mx-3' : 'mx-4')}
            >
              <Item
                collapsed={sidebarCollapsed}
                title={name}
                active={active}
                icon={icon}
                iconActive={iconActive}
              />
            </Link>
          )
        })}
        <div
          className={cn(
            'border-b border-[#E5E6EB]',
            sidebarCollapsed ? 'mx-3' : 'mx-4',
          )}
        />
        <SessionHistory className='flex-1' collapsed={sidebarCollapsed} />
      </div>
    </MenuWrapper>
  )
}
