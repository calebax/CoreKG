import { FC, PropsWithChildren } from 'react'
import { Skeleton } from 'antd'
import useLocalStore from '@/stores/local'

const AuthRouter: FC<PropsWithChildren> = (props) => {
  const { children } = props
  const token = useLocalStore((state) => state.token)
  const nav = useNavigate()
  useEffect(() => {
    if (!token) {
      nav('/login')
    }
  }, [nav, token])
  if (token) {
    return children
  }
  return <Skeleton active className='p-4' />
}
export default AuthRouter
