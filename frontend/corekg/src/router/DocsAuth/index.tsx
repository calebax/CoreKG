import { FC, PropsWithChildren } from 'react'
import { Skeleton } from 'antd'
import { useRequest } from 'ahooks'
import { getKnowledgeBaseDetail } from '@/api/knowledge'

export const DocsAuth: FC<PropsWithChildren> = (props) => {
  const navigate = useNavigate()
  const params = useParams()
  const id = useMemo(() => {
    const _id = parseInt(params.id!)
    if (Number.isInteger(_id)) {
      return _id
    }
    return null
  }, [params.id])
  const { error, loading } = useRequest(
    async () => {
      if (!id) throw new Error('bad id')
      return getKnowledgeBaseDetail({ id })
    },
    { refreshDeps: [id] },
  )
  useEffect(() => {
    if (error) {
      navigate('/docs')
    }
  }, [error, navigate])

  if (error || loading) {
    return <Skeleton active paragraph={{ rows: 10 }} />
  }
  return props.children
}
