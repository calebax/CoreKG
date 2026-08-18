import { createContext } from 'react'
import { useRequest } from 'ahooks'
// 模型列表和外部数据源
import { listCustomModel } from '@/api'
import { loadAccountBindingList } from '@/api/accountBindings'
import { getFileQaProject, getFileSession } from '@/api/knowledge'
import Project from '@/pages/project'
import { useFileDetailViewProject } from '../FileDetailView'

export type Knowledge = {
  forest_id: number
  name: string
  node_type: string
  key: string
  knowledgeType: 'file' | 'excel_sheet' | 'mysql_table' | 'qa' | 'other'
  parentKey?: string
  children?: Knowledge[]
  [key: string]: any
}

export type ExternalDataSourceInfo = {
  bindings: Array<{
    account: string
    boundAt: string
    id: number
    provider: string
    valid: boolean
  }>
  supported: Array<{
    logo: string
    provider: string
  }>
}

export type ExternalDataSourceItem = {
  id?: number
  provider: string
  logo: string
  label: string
}

interface ProjectValue {
  fileId: number
  fileInfo: {
    id: number
    name: string
    knowledgeBaseName: string
  }
}

interface QaContextValue {
  fileId: number
  getStorageKey: (fileId: number) => string
}

// eslint-disable-next-line react-refresh/only-export-components
export const QaContext = createContext<QaContextValue | null>(null)

// eslint-disable-next-line react-refresh/only-export-components
export const useQaContext = () => {
  const context = useContext(QaContext)
  if (!context) {
    throw new Error('useQaContext must be used within QaContext.Provider')
  }
  return context
}

export default function Qa() {
  const { fileId, fileInfo } = useFileDetailViewProject<ProjectValue>()!
  const [projectId, setProjectId] = useState<number>(0)
  const [sessionId, setSessionId] = useState<number>(0)
  const [loading, setLoading] = useState<boolean>(true)

  // 生成当前文件的 sessionStorage key
  const getStorageKey = (fileId: number) => `file_session_${fileId}`

  const fetchFileSessionId = async () => {
    try {
      setLoading(true)
      const {
        file: { project_id },
      } = await getFileQaProject(0)

      // setProjectId(project_id)

      // 检查是否有保存的 session_id
      const storageKey = getStorageKey(fileId)
      const savedSessionId = sessionStorage.getItem(storageKey)

      if (savedSessionId) {
        // 如果有保存的 session_id，使用它来恢复历史记录
        setSessionId(Number(savedSessionId))
      } else {
        // 首次进入，不加载历史记录
        setSessionId(0)
      }

      setLoading(false)
    } catch (error) {
      console.log(error)
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchFileSessionId()
  }, [fileId])

  return loading ? null : (
    <QaContext.Provider value={{ fileId, getStorageKey }}>
      <Project
        project_id={projectId}
        session_id={sessionId}
        defaultKnowBase={fileId}
        type='single-file'
      />
    </QaContext.Provider>
  )
}
