import { FC, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  Button,
  Empty,
  Input,
  InputRef,
  Popover,
  Typography,
  message,
} from 'antd'
import { LoadingOutlined } from '@ant-design/icons'
import {
  useBoolean,
  useClickAway,
  useMemoizedFn,
  useMount,
  useRequest,
} from 'ahooks'
import { useTranslation } from 'react-i18next'
import { EnterIcon } from 'tdesign-icons-react'
import { moveChatSession } from '@/api'
import { cn } from '@/utils'
import { deleteSession, updateSessionName } from '@/api/agent'
import Rename from '@/assets/icons/docs/edit-file.svg?react'
import MoveIcon from '@/assets/icons/home/home-move-into.svg?react'
import MoveOutIcon from '@/assets/icons/home/home-move-out.svg?react'
import ProjectListPopover from '@/pages/app/components/Sidebar/SidebarList/ProjectListPopover'
import { useProject } from '@/pages/project'
import useProjectStore from '@/stores/project'
import { useSessionInfo } from '../..'
import Del from './images/del.svg?react'
import Icon from './images/icon.svg?react'
import Op from './images/op.svg?react'
import styles from './styles.module.scss'

export const SessionHistory: FC = () => {
  const {
    data: { sessions },
    session_id,
  } = useProject()
  const { t } = useTranslation('common')
  return (
    <div
      className={cn('flex-1 overflow-auto', 'py-5 px-3 flex flex-col gap-1')}
    >
      {sessions.length === 0 ? (
        <Empty description={t('empty.noData')} />
      ) : (
        sessions.map((item) => {
          return (
            <HistoryItem
              {...item}
              key={item.session_id}
              active={item.session_id === session_id}
            />
          )
        })
      )}
    </div>
  )
}
const HistoryItem: FC<{
  name: string
  session_id: number
  active: boolean
  nameLoading?: boolean
}> = (props) => {
  const {
    project_id,
    session_id: paramSessionId,
    setSessionId,
    setData,
    reloadSessions,
  } = useProject()
  const { setSessionStatus } = useSessionInfo()
  const { load: reloadProjectList } = useProjectStore()
  const { name, session_id } = props
  const { t } = useTranslation('common')
  const { t: tPages } = useTranslation('pages')
  const navigate = useNavigate()
  const [editing, { toggle }] = useBoolean()
  const [movePopoverOpen, setMovePopoverOpen] = useState(false)
  const [menuPopoverOpen, setMenuPopoverOpen] = useState(false)

  const onEdit = useMemoizedFn(async (val: string) => {
    try {
      await updateSessionName(session_id, val)
      message.success(tPages('app.docs.detail.fileEdit.renameSuccess'))
      // 刷新历史记录列表
      await reloadSessions()
      toggle()
    } catch (error) {
      console.log('重命名会话失败:', error)
    }
  })
  const onDel = useMemoizedFn(async () => {
    await deleteSession(session_id)
    setData((draft) => {
      const index = draft.sessions.findIndex(
        (item) => item.session_id === session_id,
      )
      if (index !== -1) draft.sessions.splice(index, 1)
    })
    if (paramSessionId === session_id) {
      // 如果删除的是当前会话
      setSessionStatus('none')
      setSessionId(undefined)
    }
  })

  const handleMoveToProject = useMemoizedFn(async (targetProjectId: number) => {
    try {
      await moveChatSession({ id: session_id, subject_id: targetProjectId })
      message.success('移动成功')
      // 刷新历史记录列表
      await reloadSessions()
      // 刷新会话分组列表
      await reloadProjectList()
      navigate(`/project/${targetProjectId}/${session_id}`)
      setMovePopoverOpen(false)
      setMenuPopoverOpen(false) // 同时关闭三点弹窗
    } catch (error) {
      console.log('移动会话失败:', error)
    }
  })

  // 处理三点弹窗的打开/关闭
  const handleMenuOpenChange = useMemoizedFn((open: boolean) => {
    setMenuPopoverOpen(open)
    if (!open) {
      // 关闭三点弹窗时，也关闭会话分组弹窗
      setMovePopoverOpen(false)
    }
  })

  // 处理会话分组弹窗的打开/关闭
  const handleMovePopoverOpenChange = useMemoizedFn((open: boolean) => {
    setMovePopoverOpen(open)
  })

  // 从分组中移出
  const handleMoveOutFromGroup = useMemoizedFn(async () => {
    try {
      await moveChatSession({ id: session_id, subject_id: 0 })
      message.success('移出成功')
      // 刷新历史记录列表
      await reloadSessions()
      // 刷新未分组列表（通过触发自定义事件通知 SidebarList 刷新）
      window.dispatchEvent(new CustomEvent('refreshUngroupedSessions'))
      setMenuPopoverOpen(false)
    } catch (error) {
      console.log('移出分组失败:', error)
    }
  })
  return (
    <Link
      to={`/project/${project_id}/${session_id}`}
      className={cn(
        styles.item,
        'rounded text-black',
        'p-1 pl-2.5 flex items-center',
        {
          [styles.itemActive]: props.active,
        },
      )}
      onClick={() => {
        setSessionStatus('created')
        setSessionId(session_id)
      }}
    >
      <Icon />
      <div className='flex-1 min-w-0 overflow-hidden ml-2.5 mr-2 font-medium'>
        {editing ? (
          <EditingText text={name} onComplete={onEdit} onCancel={toggle} />
        ) : props.nameLoading ? (
          <span className={styles.loadingName} />
        ) : (
          <Typography.Paragraph
            className='m-0'
            ellipsis={{ rows: 1, tooltip: name }}
          >
            {name}
          </Typography.Paragraph>
        )}
      </div>
      <div className='flex-shrink-0'>
        <Popover
          placement='bottomLeft'
          arrow={false}
          trigger={['click']}
          open={menuPopoverOpen}
          onOpenChange={handleMenuOpenChange}
          content={
            <div
              className='p-2.5 flex flex-col items-center gap-1 min-w-[190px] max-w-[230px]'
              onClick={(e) => {
                e.preventDefault()
                e.stopPropagation()
              }}
            >
              <Button
                type='text'
                icon={
                  <span className='anticon align-middle'>
                    <Rename />
                  </span>
                }
                className=' justify-start items-center w-50'
                onClick={() => {
                  toggle()
                  setMenuPopoverOpen(false)
                }}
              >
                {t('button.rename')}
              </Button>
              <ProjectListPopover
                open={movePopoverOpen}
                onOpenChange={handleMovePopoverOpenChange}
                onSelect={handleMoveToProject}
                excludeProjectId={project_id}
              >
                <Button
                  type='text'
                  icon={
                    <span className='anticon align-middle'>
                      <MoveIcon />
                    </span>
                  }
                  className='justify-start items-center w-50'
                  onClick={(e) => {
                    e.preventDefault()
                    e.stopPropagation()
                    setMovePopoverOpen(true)
                  }}
                >
                  移至会话分组
                </Button>
              </ProjectListPopover>
              {project_id != null && project_id !== 0 && (
                <Button
                  type='text'
                  icon={
                    <span className='anticon align-middle'>
                      <MoveOutIcon />
                    </span>
                  }
                  className='justify-start items-center w-50'
                  onClick={() => {
                    handleMoveOutFromGroup()
                  }}
                >
                  从分组中移出
                </Button>
              )}
              <Button
                type='text'
                icon={
                  <span className='anticon align-middle'>
                    <Del />
                  </span>
                }
                className='justify-start items-center w-50'
                onClick={() => {
                  onDel()
                  setMenuPopoverOpen(false)
                }}
              >
                {t('button.delete')}
              </Button>
            </div>
          }
          destroyTooltipOnHide
        >
          <Op
            className={cn(styles.op)}
            onClick={(e) => {
              e.preventDefault()
              e.stopPropagation()
            }}
          />
        </Popover>
      </div>
    </Link>
  )
}

const EditingText: FC<{
  text: string
  onComplete: (val: string) => any
  onCancel: () => void
}> = (props) => {
  const { text, onComplete: _onComplete, onCancel } = props
  const [cache, setCache] = useState(text)
  const { run: onComplete, loading: submitting } = useRequest(
    async () => {
      await _onComplete(cache)
    },
    { manual: true },
  )
  const inputRef = useRef<InputRef>(null)
  useMount(() => {
    inputRef.current?.focus()
  })
  useClickAway(
    () => {
      if (submitting) return
      onCancel()
    },
    () => inputRef.current?.nativeElement,
  )
  return (
    <Input
      ref={inputRef}
      onClick={(e) => {
        e.preventDefault()
        e.stopPropagation()
      }}
      onKeyDown={(e) => {
        if (e.key === 'Escape') {
          onCancel()
        }
      }}
      disabled={submitting}
      onPressEnter={onComplete}
      value={cache}
      onChange={(e) => setCache(e.target.value)}
      suffix={
        submitting ? <LoadingOutlined /> : <EnterIcon onClick={onComplete} />
      }
      minLength={1}
      maxLength={50}
      showCount
    />
  )
}
