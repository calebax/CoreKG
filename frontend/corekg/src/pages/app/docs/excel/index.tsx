import { useParams } from 'react-router-dom'
import FileExplorer from './components/FileExplorer'

export default function KnowledgeBaseDetail() {
  const { id, folderId } = useParams<{ id: string; folderId: string }>()
  return <FileExplorer isRootLevel={!folderId} />
}
