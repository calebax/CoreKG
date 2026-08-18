import { FC, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { App, Button, Dropdown, Popover, Tooltip } from 'antd'
import { GraphBaseInfo } from 'Graph'
import { useBoolean, useRequest } from 'ahooks'
import dayjs from 'dayjs'
import { match } from 'ts-pattern'
import { cn, uniqueArray } from '@/utils'
import { deleteGraph, getGraphInfo, updateGraph } from '@/api/graph'
import { ItemCard } from '@/components/ItemCard'
import { getAvatar } from '../../getAvatar'
import { GraphModal } from '../GraphModal'
import { GraphMetaValues } from '../GraphModal/types'
import StatusDraft from './images/draft.svg'
import StatusRunning from './images/running.svg'
import StatusSuccess from './images/success.svg'

export type GraphCard = Style & {
  value: GraphBaseInfo
  reload: () => void
}
export const GraphCard: FC<GraphCard> = (props) => {
  const { modal, message } = App.useApp()
  const navigate = useNavigate()
  const { value, reload, className, style } = props
  const [showMenu, setShowMenu] = useState(false)
  const { status } = value

  // 计算构建进度百分比
  const buildProgress = useMemo(() => {
    const { task_count = 0, success_task_count = 0 } = value
    if (task_count === 0) return 0
    return Math.floor((success_task_count / task_count) * 100)
  }, [value])

  // 计算实际显示的状态：如果 status 是 updatable 且进度是 100%，则视为成功状态
  const displayStatus = useMemo(() => {
    if ((status as string) === 'updatable' && buildProgress === 100) {
      return 'success' as const
    }
    return status
  }, [status, buildProgress])
  const to = useMemo(() => {
    switch (value.status) {
      case 'draft':
        return `/graph/edit?graphId=${value.id}`
      case 'success':
      case 'updatable':
        return `/graph/detail?graphId=${value.id}`
      case 'pending':
      case 'running':
        return `/graph/detail?graphId=${value.id}`
      case 'failed':
        return ''
    }
  }, [value.id, value.status])

  const [open, { toggle }] = useBoolean()
  const { data, run } = useRequest(
    async () => {
      const { manager_ids, public_scope, scope_ids } = await getGraphInfo({
        graph_id: value.id,
      })
      const res: Pick<
        GraphMetaValues,
        'manager_ids' | 'public_scope' | 'scope_ids'
      > = {
        manager_ids,
        public_scope,
        scope_ids: uniqueArray(manager_ids, scope_ids),
      }
      return res
    },
    { manual: true },
  )

  const handleClick = (value: GraphBaseInfo) => {
    if (!to) {
      message.info('该图谱当前无法访问')
      return
    }
    // 如果节点数和边数都为 0，添加 empty 参数标识空状态
    const isEmpty = value.totalNodes === 0 && value.totalRelationships === 0
    const url = isEmpty ? `${to}&empty=true` : to
    navigate(url)
  }

  // 菜单项配置（无权限也展示，但禁用操作项）
  const isReadOnly = !value.is_admin
  const sharedItemClass =
    'text-[#2D2D2D] hover:!bg-[#F5F5F5] px-[10px] py-[7px] rounded'
  const disabledItemClass =
    'cursor-not-allowed !text-[#C0C4CC] hover:!bg-transparent'

  const handleEdit = () => {
    run()
    toggle()
  }

  const handleDelete = () => {
    modal.confirm({
      title: '删除图谱',
      content: '确定删除图谱吗？',
      onOk: async () => {
        await deleteGraph({ graph_id: value.id })
        reload()
        message.success('操作成功')
      },
    })
  }

  return (
    <>
      <ItemCard
        className={cn(className, 'relative')}
        style={style}
        onClick={() => handleClick(value)}
        avatar={getAvatar(value.avatar_url)}
        title={value.name}
        desc={value.description}
        extra={
          <div className='flex flex-col '>
            <span>
              {`${value.totalNodes} 节点数 | ${value.totalRelationships} 关系数`}
            </span>
            <span>
              {dayjs(value.updateAt).format('YYYY.MM.DD')}
              {' 更新'}
            </span>
          </div>
        }
        operators={{
          items: [
            {
              key: 'edit',
              label: '编辑图谱',
              onClick: isReadOnly ? undefined : handleEdit,
              className: `${sharedItemClass} ${isReadOnly ? disabledItemClass : ''}`,
              disabled: isReadOnly,
            },
            {
              key: 'delete',
              label: '删除图谱',
              onClick: isReadOnly ? undefined : handleDelete,
              className: `${sharedItemClass} ${isReadOnly ? disabledItemClass : ''}`,
              disabled: isReadOnly,
            },
          ],
          onClick: (e) => e.domEvent.stopPropagation(),
        }}
      >
        {match(displayStatus)
          .with('success', () => (
            <img
              src={StatusSuccess}
              className=' absolute left-0 top-0 -translate-y-1/2 -translate-x-3 z-10 w-[70px] h-[48px]'
              width={70}
              height={48}
            />
          ))
          .with('draft', () => (
            <img
              src={StatusDraft}
              className=' absolute left-0 top-0 -translate-y-1/2 -translate-x-3 z-10 w-[70px] h-[48px]'
              width={70}
              height={48}
            />
          ))
          .otherwise(() => (
            <div className='absolute left-0 top-0 -translate-y-1/2 -translate-x-3 z-10 w-[100px] h-[47px] overflow-hidden'>
              <img
                src={StatusRunning}
                className='absolute inset-0 w-full h-full z-0'
                width={100}
                height={47}
                alt='构建中'
              />
              <span className='absolute left-1/2 top-1/2 -translate-x-[47%] -translate-y-1/2 z-10 text-[#FFFFFF] text-[12px] font-semibold leading-none whitespace-nowrap'>
                构建中-{buildProgress}%
              </span>
            </div>
          ))}
      </ItemCard>
      {open ? (
        <GraphModal
          onCancel={toggle}
          okText='提交'
          title='编辑图谱'
          loading={!data}
          initialValues={data ? { ...value, ...data } : undefined}
          onOk={async (val) => {
            await updateGraph({ graph_id: value.id, ...val })
            reload()
          }}
        />
      ) : null}
    </>
  )
}
