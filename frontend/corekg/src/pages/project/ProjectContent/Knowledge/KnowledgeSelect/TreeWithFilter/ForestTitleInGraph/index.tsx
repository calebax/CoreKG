import { FC, ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { App, Checkbox } from 'antd'
import { InfoCircleOutlined } from '@ant-design/icons'
import { useRequest } from 'ahooks'
import { cn } from '@/utils'
import { createForestGraph, restockGraph } from '@/api/graph'
import { getKnowledgeBaseDetail } from '@/api/knowledge'

export type ForestTitleInGraph = {
  id: number
  title: ReactNode
}
/** 查看该图谱是否有更新并给出提示 */
export const ForestTitleInGraph: FC<ForestTitleInGraph> = (props) => {
  const { id, title } = props
  const { modal, message } = App.useApp()
  const navigate = useNavigate()
  const { data: graphData, refresh: refreshGraphStatus } = useRequest(
    async () => {
      const res = await getKnowledgeBaseDetail({ id })
      return {
        graph_status: res.data.graph_status,
        graph_info: res.graph_info,
        is_admin: res.data.is_admin ?? res.graph_info?.is_admin,
      }
    },
  )
  const isAdmin = Boolean(
    graphData?.is_admin ?? graphData?.graph_info?.is_admin ?? false,
  )
  if (graphData?.graph_status !== 'updatable') return title
  const updateGraph = () => {
    // 使用对象引用来存储选择框状态，避免闭包问题
    const stateRef = { isFullUpdate: false }
    const graphId = graphData?.graph_info?.ID
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
          if (!graphId) {
            message.error('获取图谱ID失败')
            return
          }
          await restockGraph({ graph_id: graphId })
          message.success('更新成功')
          // 刷新图谱状态
          refreshGraphStatus()
        } else {
          // 勾选：执行原有逻辑（创建新图谱）
          const { data } = await createForestGraph({
            forest_id: id,
          })
          const { ID } = data
          navigate(`/graph/edit?graphId=${ID}`)
        }
      },
    })
  }
  return (
    <div className='flex items-center gap-1'>
      {title}
      <InfoCircleOutlined
        className={cn('text-[#0C99FF]', {
          'cursor-pointer': isAdmin,
          'cursor-not-allowed': !isAdmin,
        })}
        onClick={(e) => {
          if (!isAdmin) {
            e.stopPropagation()
            return
          }
          updateGraph()
          e.stopPropagation()
        }}
      />
    </div>
  )
}
