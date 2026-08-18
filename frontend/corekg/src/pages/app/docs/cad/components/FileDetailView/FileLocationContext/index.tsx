import { createContext, FC, PropsWithChildren } from 'react'

type FileLocation = any[]
const FileLocationContext = createContext<FileLocation | null>(null)
export const FileLocationProvider: FC<PropsWithChildren> = (props) => {
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
  return (
    <FileLocationContext.Provider value={location}>
      {props.children}
    </FileLocationContext.Provider>
  )
}

/** fileLocation用于定位文件中的位置 例如pdf页数 视频时间等 */
// eslint-disable-next-line react-refresh/only-export-components
export const useFileLocation = () => useContext(FileLocationContext)
