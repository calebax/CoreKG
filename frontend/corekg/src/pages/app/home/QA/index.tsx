import { FC } from 'react'
import { QAContent } from './QAContent'

/**
 * 接受路径参数
 * search创建新会话 或者 session_id使用已有会话
 */
const QA: FC = () => {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()

  const session_id = useMemo(() => {
    return parseInt(searchParams.get('session_id')!)
  }, [searchParams])
  const question_id = useMemo(() => {
    const _id = searchParams.get('question_id')!
    return _id
    // return undefined
  }, [searchParams])
  if (!Number.isInteger(session_id)) {
    setTimeout(() => {
      navigate('/')
    }, 0)
    return null
  }
  return (
    <QAContent
      session_id={session_id}
      question_id={question_id}
      key={session_id}
    />
  )
}

export default QA
