import { useParams, useNavigate } from 'react-router-dom'
import FileExplorer from './components/FileExplorer'

export default function KnowledgeBaseDetail() {
  const { id, folderId } = useParams<{ id: string; folderId: string }>()
  const navigate = useNavigate()
  // console.log(id, folderId)

  // 根据路径参数决定显示知识库根目录还是文件夹内容
  if (folderId) {
    return <FileExplorer isRootLevel={false} />
  }

  return <FileExplorer isRootLevel={true} />
}
