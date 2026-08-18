import { FC, PropsWithChildren, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { App, Breadcrumb, Skeleton } from 'antd'
import { useCounter } from 'ahooks'
import { createForestGraph, parseGraph } from '@/api/graph'
import NavigationIcon from '@/assets/icons/docs/navigation.svg?react'
import { useGraphId, useGraphInfo, withGraphProvider } from '../GraphProvider'
import GraphIcon from '../images/graph.svg?react'
import { GraphSteps } from './GraphSteps'
import { KnowledgeProvider } from './KnowledgeContext'
import { Step1 } from './Step1'
import { Step2 } from './Step2'
import { Step3 } from './Step3'
import { Step4 } from './Step4'

export type StepComponent<T = object> = FC<
  {
    decrease: () => void
    increase: () => void
  } & T
>

// 注意顺序
const StepComps: StepComponent[] = [Step3, Step1, Step2, Step4]

/** 从查询参数中获取graphId edit可能没有graphId */
const Edit: FC = withGraphProvider(() => {
  const { loading, data } = useGraphInfo()
  const { modal } = App.useApp()
  const navigate = useNavigate()
  const graphId = useGraphId()
  // 有id必定有知识库 从第二个步骤开始
  const [step, { inc, dec }] = useCounter(graphId ? 1 : 0)
  const StepContent = StepComps[step]

  // 构建面包屑导航项
  const breadcrumbItems = useMemo(() => {
    const items = [
      {
        title: (
          <span
            className='flex items-center gap-2 text-sm cursor-pointer hover:text-[#0C99FF] font-medium text-[#3C4149]'
            onClick={() => navigate('/graph')}
          >
            <GraphIcon className='w-4 h-4' />
            <span>知识图谱</span>
          </span>
        ),
      },
      {
        title: (
          <span
            className='text-sm font-medium text-[#3C4149]'
            onClick={() => navigate(`/graph/detail?graphId=${data!.id}`)}
          >
            {data?.name}
          </span>
        ),
      },
      {
        title: (
          <span className='text-sm font-medium text-[#3C4149]'>编辑规则</span>
        ),
      },
    ]

    return items
  }, [data, navigate])

  if (loading) return <Skeleton active className='p-4' />
  if (data?.isEditingRules) {
    // 正在编辑规则
    return (
      <ManagePerm>
        <div className='overflow-hidden h-full flex flex-col gap-4'>
          <div className='w-full h-12 bg-white flex items-center px-5 pt-[2px] font-medium'>
            <Breadcrumb
              className='[&_span.ant-breadcrumb-separator]:inline-flex [&_span.ant-breadcrumb-separator]:items-center [&_span.ant-breadcrumb-separator]:align-middle'
              separator={<NavigationIcon className='inline-block' />}
              items={breadcrumbItems}
            />
          </div>
          <Step2
            decrease={() => navigate(`/graph/detail?graphId=${data.id}`)}
            increase={() => {
              modal.confirm({
                title:
                  '修改规则后图谱将进行全量更新，更新过程中不可使用图谱进行问答，请确认',
                onOk: async () => {
                  await createForestGraph({ forest_id: data.forest_id })
                  await parseGraph({ graph_id: data.id })
                },
              })
            }}
            editRules
          />
        </div>
      </ManagePerm>
    )
  }
  return (
    <ManagePerm>
      <div className='overflow-hidden h-full flex flex-col gap-4'>
        <GraphSteps
          className='border-b border-[#D7D9E5]'
          value={step}
          steps={['选择知识', '选择模板', '编辑实体', '生成图谱']}
        />
        <div className='flex-1 overflow-hidden'>
          <KnowledgeProvider>
            <StepContent decrease={() => dec()} increase={() => inc()} />
          </KnowledgeProvider>
        </div>
      </div>
    </ManagePerm>
  )
}, false)
const ManagePerm: FC<PropsWithChildren> = (props) => {
  // const navigate = useNavigate()
  // const { message } = App.useApp()
  const { children } = props
  // const { data } = useGraphInfo()
  // useEffect(() => {
  //   if (data && !data.is_admin) {
  //     navigate('/graph')
  //     message.error('没有对此图谱的管理权限')
  //   }
  // }, [data, message, navigate])
  // if (!data?.is_admin) return <Skeleton active />
  return children
}

const EditWithWrapper: FC = () => {
  return (
    <div className='p-4 w-full h-full overflow-hidden bg-white'>
      <Edit />
    </div>
  )
}
export default EditWithWrapper
