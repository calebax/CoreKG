import { FC, PropsWithChildren, useEffect, useMemo, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { App, Button, Checkbox, ConfigProvider, Spin, Tooltip } from 'antd'
import { CheckCircleOutlined, SyncOutlined } from '@ant-design/icons'
import { useMemoizedFn, useRequest } from 'ahooks'
import dayjs from 'dayjs'
import { ArrowRightIcon } from 'tdesign-icons-react'
import { match, P } from 'ts-pattern'
import { cn } from '@/utils'
import { createForestGraph, restockGraph } from '@/api/graph'
import { getKnowledgeBaseDetail } from '@/api/knowledge'
import Icon from './icon.svg'
import SuccessNewIcon from './success-new.svg'
import UpdateRocketIcon from './update-rocket.svg'
import UpdateIcon from './update.svg?react'

type GraphStatus =
  | 'none'
  | 'draft'
  | 'pending'
  | 'running'
  | 'success'
  | 'failed'
  | 'updatable'

/** 多模态知识库及其图谱 */
export const MutimodalGraph: FC<
  Style & {
    is_admin?: boolean
    /** 图谱状态 */
    graph_info?: any & {
      /** 任务总数 */
      task_count?: number
      /** 成功任务数 */
      success_task_count?: number
      /** 创建时间 */
      created_at?: string
      /** 更新时间 */
      updated_at?: string
    }
    /** 知识库id */
    knowledgebase_id: number
    graph_updatable?: boolean
    knowledge_status?: string
  }
> = (props) => {
  const {
    knowledgebase_id,
    graph_updatable,
    graph_info,
    is_admin,
    knowledge_status,
    className,
    style,
  } = props
  const [graphInfo, setGraphInfo] = useState(graph_info)
  useEffect(() => {
    setGraphInfo(graph_info)
  }, [graph_info])
  const { ID: graph_id } = graphInfo ?? {}
  const [graph_status, setStatus] = useState<GraphStatus>(() => {
    // 未创建图谱
    if (!graph_id) return 'none'
    // 图谱需要更新
    if (graph_updatable) return 'updatable'
    // 其他
    return graph_info.status as GraphStatus
  })
  const { message } = App.useApp()
  const ops = useGraphOperator(
    knowledgebase_id,
    setStatus,
    graphInfo,
    setGraphInfo,
  )

  // 计算构建进度百分比
  const buildProgress = useMemo(() => {
    if (!graphInfo) return 0
    const { task_count = 0, success_task_count = 0 } = graphInfo
    if (task_count === 0) return 0
    return Math.floor((success_task_count / task_count) * 100)
  }, [graphInfo])

  // 判断是否显示新图标（更新时间不超过7天）
  const shouldShowNewIcon = useMemo(() => {
    if (!graph_info?.updated_at) return false
    // 使用精确的时间差计算（精确到秒），判断是否在7天内
    // dayjs转换为本地时间
    const updatedDate = dayjs(graph_info.updated_at)
    const now = dayjs()
    const diffSeconds = now.diff(updatedDate, 'second')
    // 7天 = 604800秒，不超过7天（可以等于7天）
    const isWithin7Days = diffSeconds <= 604800
    return isWithin7Days
  }, [graph_info])

  return (
    <div
      className={cn(
        className,
        'border-[1.5px] border-[#45CF7533] rounded-md text-[#5AB479] whitespace-nowrap',
        'h-17 min-w-48 p-1.5 flex gap-4 items-center',
        'shadow-[4px_4px_10px_rgba(69,207,117,0.2)]',
      )}
      style={style}
    >
      <div className='flex flex-col items-center'>
        <img src={Icon} className='h-10 w-10' />
        <span className='text-xs font-medium'>知识图谱洞察</span>
      </div>
      {/* 按钮样式 */}
      <ButtonStyleProvider>
        <div className='flex-1 flex flex-col items-end justify-center'>
          {match({ graph_status, knowledge_status })
            .with(
              { graph_status: 'none', knowledge_status: P.not('success') },
              () => (
                <Button
                  type='text'
                  icon={<ArrowRightIcon />}
                  iconPosition='end'
                  className='text-[#5AB479] text-sm font-medium cursor-not-allowed'
                  onClick={() => {
                    message.warning('资源准备中，请稍候~')
                  }}
                >
                  立即体验
                </Button>
              ),
            )
            .with({ graph_status: 'none' }, () => (
              <Button
                type='text'
                icon={<ArrowRightIcon />}
                iconPosition='end'
                onClick={ops.create}
                disabled={!is_admin}
                className='text-[#5AB479] text-sm font-medium'
              >
                立即体验
              </Button>
            ))
            .with({ graph_status: 'draft' }, () => (
              <Link
                to={is_admin ? `/graph/edit?graphId=${graph_id}` : ''}
                className='text-[#5AB479] text-sm font-medium'
              >
                <Button
                  type='text'
                  icon={<ArrowRightIcon />}
                  iconPosition='end'
                  disabled={!is_admin}
                  className='text-[#5AB479] text-sm font-medium'
                >
                  继续构建
                </Button>
              </Link>
            ))
            .with({ graph_status: 'success' }, () => (
              <div className='relative inline-block'>
                <Link
                  to={`/graph/detail?graphId=${graph_id}`}
                  className='flex items-center gap-1 text-[#5AB479] text-sm font-medium'
                >
                  图谱构建成功
                  <Tooltip title='图谱更新成功后系统会自动采用最新版本进行问答，您可在图谱洞察模式中查看溯源路径。'>
                    <CheckCircleOutlined className='pointer-events-auto' />
                  </Tooltip>
                </Link>
                {shouldShowNewIcon && (
                  <img
                    src={SuccessNewIcon}
                    alt='new'
                    className='absolute bottom-6 -right-4 w-11 h-11 pointer-events-none'
                  />
                )}
              </div>
            ))
            .with({ graph_status: 'updatable' }, () => (
              <div className='relative inline-block'>
                <Button
                  type='text'
                  icon={<UpdateIcon />}
                  iconPosition='end'
                  onClick={ops.update}
                  disabled={!is_admin}
                  className='text-[#5AB479] text-sm font-medium'
                >
                  立即更新
                </Button>
                <img
                  src={UpdateRocketIcon}
                  alt='update rocket'
                  className='absolute bottom-6 -right-4 w-11 h-11 pointer-events-none'
                />
              </div>
            ))
            .with({ graph_status: P.union('pending', 'running') }, () => (
              <div className='flex flex-col items-end'>
                <Button
                  type='text'
                  icon={<Spin />}
                  iconPosition='end'
                  onClick={ops.showRunningInfo}
                  disabled={!is_admin}
                  className='text-[#5AB479] text-sm font-medium -mb-1'
                >
                  图谱构建中
                </Button>
                <div className='text-[#919497] text-xs font-normal self-start mt-0.5'>
                  当前进度：{buildProgress}%
                </div>
              </div>
            ))

            .with({ graph_status: 'failed' }, () => (
              <Button
                type='text'
                icon={
                  <Tooltip title='点击重试'>
                    <RefreshIcon onClick={is_admin ? ops.retry : undefined} />
                  </Tooltip>
                }
                iconPosition='end'
                disabled={!is_admin}
                className='text-[#5AB479] text-sm font-medium'
              >
                图谱构建失败
              </Button>
            ))
            .exhaustive()}
        </div>
      </ButtonStyleProvider>
    </div>
  )
}

const RefreshIcon: FC<{ onClick?: () => Promise<void> }> = (props) => {
  const { onClick } = props
  const { loading, run } = useRequest(
    async () => {
      return onClick?.()
    },
    { manual: true },
  )
  return <SyncOutlined onClick={run} spin={loading} className=' rotate-90' />
}

const useGraphOperator = (
  knowledgebase_id: number,
  setStatus: (val: GraphStatus) => void,
  graph_info?: any,
  setGraphInfo?: (info: any) => void,
) => {
  const { modal, message } = App.useApp()
  const navigate = useNavigate()
  const create = useMemoizedFn(() => {
    modal.confirm({
      title: '创建图谱',
      content: (
        <div className='text-[#6E757F] flex flex-col'>
          尚未生成关联图谱。确认创建后系统将自动：
          <ul className='list-disc pl-5'>
            <li>提取核心实体</li>
            <li>建立关系链</li>
            <li>生成可视化图谱</li>
          </ul>
          构建完成后，可在问答的"图谱洞察"模式查看来源与覆盖范围。
        </div>
      ),
      cancelText: '稍后再说',
      okText: '立刻创建',
      onOk: async () => {
        const { data } = await createForestGraph({
          forest_id: knowledgebase_id,
          avatar_url: `default${Math.ceil(Math.random() * 6)}`,
        })
        const { ID } = data
        navigate(`/graph/edit?graphId=${ID}`)
      },
    })
  })
  const update = useMemoizedFn(() => {
    // 使用对象引用来存储选择框状态，避免闭包问题
    const stateRef = { isFullUpdate: false }
    modal.confirm({
      // 使用与删除确认弹窗一致的布局和按钮样式（统一交互体验）
      title: null,
      icon: null,
      centered: true,
      closable: false,
      className: 'delete-api-key-modal !w-[30%] px-5 py-6',
      okText: '立刻更新',
      cancelText: '取消',
      okButtonProps: {
        className:
          'bg-[#0C99FF] hover:!bg-[#0C99FF] !w-[77px] !h-[32px] !rounded-md !text-sm !px-[10.5] !py-[9px] !font-medium text-[#ffffff]',
        danger: false,
      },
      cancelButtonProps: {
        className:
          '!bg-[#F5F5F5] text-[#0C1F17] !w-[77px] !h-[32px] !rounded-md !text-sm !border-none !px-[24.5] !py-[9px] !font-medium',
      },
      content: (
        <div className='relative'>
          <div className='flex justify-between'>
            <div className='text-[18px] font-[500] mb-2.5 text-[#0C1F17]'>
              更新图谱
            </div>
          </div>
          <div className='h-[0.5px] w-[calc(100%+60px)] bg-[#E3E6ED] -mx-6' />
          <div className='mt-3 text-base text-[#6E757F] mb-2 font-normal fontFamily-pingFangSC px-1 leading-[22px]'>
            <div>检测到存在文档更新，图谱可能失效。更新后将：</div>
            <ul className='list-disc pl-5 mt-1.5'>
              <li>同步实体变更</li>
              <li>修复失效关系</li>
            </ul>
            <div className='mt-1.5'>
              在更新过程中您仍可继续使用当前图谱进行问答。更新完成后系统会自动切换到最新图谱版本
            </div>
            <div className='mt-4 flex items-start justify-center gap-1'>
              <Checkbox
                onChange={(e) => {
                  stateRef.isFullUpdate = e.target.checked
                }}
              />
              <span className='text-sm font-normal text-[#0C1F17] leading-[20px]'>
                勾选后将执行全量图谱更新，并可重新选择图谱模板及调整实体规则，整体耗时较长。
              </span>
            </div>
          </div>
        </div>
      ),
      onOk: async () => {
        if (!stateRef.isFullUpdate) {
          // 不勾选：调用全量更新接口，使用图谱ID
          const graphId = graph_info?.ID
          await restockGraph({ graph_id: graphId })
          message.success('更新成功')
          // 刷新图谱状态
          try {
            const res = await getKnowledgeBaseDetail({ id: knowledgebase_id })
            if (res && res.data) {
              setGraphInfo?.(res.graph_info)
              const newGraphStatus =
                res.data.graph_status === 'updatable'
                  ? 'updatable'
                  : (res.graph_info?.status as GraphStatus) || 'success'
              setStatus(newGraphStatus)
            }
          } catch (error) {
            console.log('刷新图谱状态失败:', error)
          }
        } else {
          // 勾选：执行原有逻辑（创建新图谱）
          const { data } = await createForestGraph({
            forest_id: knowledgebase_id,
          })
          const { ID } = data
          navigate(`/graph/edit?graphId=${ID}`)
        }
      },
    })
  })
  const showRunningInfo = useMemoizedFn(() => {
    modal.info({
      title: '图谱构建中',
      content: (
        <div className='flex justify-start text-[#6E757F] whitespace-pre-wrap'>
          <Spin className='mr-2' />
          在更新过程中您仍可继续使用当前图谱进行问答。更新完成后系统会自动切换到最新图谱版本。
        </div>
      ),
    })
  })
  const retry = useMemoizedFn(async () => {
    // retry
    message.success('重试成功')
    setStatus('pending')
  })
  return {
    create,
    update,
    showRunningInfo,
    retry,
  }
}

const ButtonStyleProvider: FC<PropsWithChildren> = (props) => {
  return (
    <ConfigProvider
      theme={{
        components: {
          Button: {
            textHoverBg: 'rgba(0,0,0,0)',
            colorBgTextActive: 'rgba(0,0,0,0)',
            contentFontSize: 12,
            paddingInline: 0,
          },
        },
      }}
    >
      {props.children}
    </ConfigProvider>
  )
}
