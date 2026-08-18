import { useMemo } from 'react'
import { useSearchParams } from 'react-router-dom'

/** 从查询参数获取location信息 以定位文件位置 */
export const useFileLocation = () => {
  const [searchParams] = useSearchParams()
  const locationString = searchParams.get('location')
  const location = useMemo(() => {
    if (!locationString) return null
    try {
      const location = JSON.parse(decodeURIComponent(locationString))
      if (!Array.isArray(location)) return null
      return location
    } catch {
      return null
    }
  }, [locationString])
  return location
}
