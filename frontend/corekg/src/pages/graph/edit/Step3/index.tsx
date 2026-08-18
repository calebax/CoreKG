import { App, Button, Empty, Input, Skeleton } from 'antd'
import { useMount, useRequest } from 'ahooks'
import { createForestGraph } from '@/api/graph'
import type { StepComponent } from '..'
import { useGraphId, useGraphInfo } from '../../GraphProvider'
import { useKnowledge } from '../KnowledgeContext'
import { ForestSelect } from './ForestSelect'

export const Step3: StepComponent = (props) => {
  const { modal } = App.useApp()
  const { increase } = props
  const { data: graphInfo, updateBaseInfo } = useGraphInfo()
  const graphId = useGraphId()
  const { loading, data, loadData } = useKnowledge()
  useMount(() => {
    if (!data) loadData()
  })

  const [currentForestId, setForestId] = useState(graphInfo?.forest_id)
  const navigate = useNavigate()
  const { run: next, loading: submitting } = useRequest(
    async () => {
      if (!graphId) {
        if (!currentForestId) {
          modal.warning({
            content: '请勾选您希望进行分析的文档，系统将据此构建知识图谱',
          })
          return
        } else {
          const { data } = await createForestGraph({
            forest_id: currentForestId,
            avatar_url: `default${Math.ceil(Math.random() * 6)}`,
          })
          const { ID: id } = data
          navigate(`/graph/edit?graphId=${id}`)
          updateBaseInfo({ id, forest_id: currentForestId })
        }
      }
      increase()
    },
    { manual: true },
  )

  if (loading) {
    return <Skeleton active />
  }
  if (!data?.length) return <Empty />

  return (
    <div className='h-full overflow-hidden flex flex-col gap-2'>
      <ForestSelect
        disabled={Boolean(graphId)}
        value={currentForestId}
        onChange={setForestId}
        data={data}
        className='flex-1 overflow-hidden'
      />
      <span className='flex items-center self-end gap-4'>
        <Button type='primary' loading={submitting} onClick={next}>
          下一步
        </Button>
      </span>
    </div>
  )
}
