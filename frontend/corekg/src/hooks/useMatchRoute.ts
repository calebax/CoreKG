import { useLocation, matchPath } from 'react-router-dom'

export function useMatchRoute() {
  const { pathname } = useLocation()
  const isPathActive = (targetPath: string, pageId?: string) => {
    const exec = pageId ? true : false
    const matchRoutePath = exec ? `${targetPath}/${pageId}` : targetPath
    // 使用matchPath进行路由匹配
    const match = matchPath({ path: matchRoutePath, end: false }, pathname)

    return !!match
  }

  const equalPathActive = (targetPath: string) => {
    return targetPath === pathname
  }

  return {
    isPathActive,
    equalPathActive,
  }
}
