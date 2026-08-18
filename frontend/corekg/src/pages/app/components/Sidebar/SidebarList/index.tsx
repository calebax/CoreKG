import { ReactNode, useState, useMemo, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate, useLocation } from 'react-router-dom'
import { Skeleton, message, Tooltip } from 'antd'
import EditFileIcon from '@/assets/icons/docs/edit-file.svg?react'
import { Item } from '@/components/Layout/SidebarWrapper'
import { useMatchRoute } from '@/hooks/useMatchRoute'
import DeleteIcon from '../images/delete.svg?react'
import ProjectIcon from '../images/project.svg?react'
import HomeGroupIcon from '@/assets/icons/home/home-group.svg?react'
import UnGroupedIcon from '@/assets/icons/home/home-up.svg?react'
import DownGroupedIcon from '@/assets/icons/home/home-down.svg?react'
import AddIcon from '@/assets/icons/home/home-add.svg?react'
import MoveIcon from '@/assets/icons/home/home-move-into.svg?react'
import { EExpandableStatus } from '../types'
import useSidebarList from './hook'
import CreateGroupModal from './CreateGroupModal'
import DeleteGroupModal from './DeleteGroupModal'
import AgentListPopover from './AgentListPopover'
import ProjectListPopover from './ProjectListPopover'
import styles from './index.module.scss'

enum EMenuOperator {
  DELETE,
  RENAME,
  MOVE_TO_GROUP,
}

type MenuItem = {
  text: string
  key: EMenuOperator
  icon: ReactNode
  popoverContent?: ReactNode
}
interface ISidebarListProps {
  expandableStatus: EExpandableStatus
}

export default function SidebarList(props: ISidebarListProps) {
  const { t } = useTranslation('pages')
  const { t: tC } = useTranslation('common')

  const {
    expend,
    handleChangeExpend,
    handleCreateProject,
    handleDeleteProject,
    handleRenameProject,
    pushProjectPage,
    projectList,
    disabledIds,
    agentProjectId,
    ungroupedSessions,
    loadingUngrouped,
    reloadUngroupedSessions,
    addPendingUngroupedSession,
    clearPendingUngroupedActive,
    handleDeleteSession,
    handleRenameSession,
    handleMoveSession,
    reloadProjectList,
  } = useSidebarList()
  const { isPathActive } = useMatchRoute()
  const navigate = useNavigate()
  const { pathname } = useLocation()
  const [currentPathname, setCurrentPathname] = useState(pathname)

  useEffect(() => {
    setCurrentPathname(pathname)
  }, [pathname])

  useEffect(() => {
    const handleGlobalRouteChange = (event: Event) => {
      const nextPathname = (event as CustomEvent<{ pathname?: string }>).detail
        ?.pathname
      if (nextPathname) {
        setCurrentPathname(nextPathname)
      }
    }
    window.addEventListener('globalSessionRouteChange', handleGlobalRouteChange)
    return () => {
      window.removeEventListener(
        'globalSessionRouteChange',
        handleGlobalRouteChange,
      )
    }
  }, [])

  // 获取当前激活的会话ID（用于未分组会话的激活状态判断）
  // 未分组会话的路由格式：/project/0/:session_id
  const currentSessionId = useMemo(() => {
    if (currentPathname.startsWith('/project/0/')) {
      // 从路径中提取 session_id（路径格式：/project/0/:session_id）
      const match = currentPathname.match(/^\/project\/0\/(\d+)/)
      if (match && match[1]) {
        const sessionId = parseInt(match[1])
        return Number.isInteger(sessionId) ? sessionId : null
      }
    }
    return null
  }, [currentPathname])

  // 监听刷新未分组列表的事件
  useEffect(() => {
    const handleRefresh = () => {
      reloadUngroupedSessions()
    }
    const handlePending = (event: Event) => {
      const sessionId = (event as CustomEvent<{ id?: number }>).detail?.id
      if (typeof sessionId === 'number') {
        setUngroupedExpanded(true)
        addPendingUngroupedSession(sessionId)
      }
    }
    window.addEventListener('refreshUngroupedSessions', handleRefresh)
    window.addEventListener('pendingUngroupedSession', handlePending)
    return () => {
      window.removeEventListener('refreshUngroupedSessions', handleRefresh)
      window.removeEventListener('pendingUngroupedSession', handlePending)
    }
  }, [addPendingUngroupedSession, reloadUngroupedSessions])

  // 未分组部分的展开状态
  const [ungroupedExpanded, setUngroupedExpanded] = useState(true)
  // 新建分组弹窗状态
  const [createModalOpen, setCreateModalOpen] = useState(false)
  // 删除分组弹窗状态
  const [deleteModalOpen, setDeleteModalOpen] = useState(false)
  const [deleteProjectId, setDeleteProjectId] = useState<number | null>(null)
  // 智能体列表弹窗状态
  const [agentListPopoverOpen, setAgentListPopoverOpen] = useState(false)
  // 会话分组弹窗状态（key为sessionId）
  const [projectPopoverOpenMap, setProjectPopoverOpenMap] = useState<
    Record<number, boolean>
  >({})

  // 渲染头部（通用函数）
  const renderHeader = (
    title: string,
    expanded: boolean,
    onToggle: () => void,
    showAddButton = false,
  ) => {
    const isExpanded = props.expandableStatus === EExpandableStatus.EXPAND
    return (
      <div
        className={`${styles.sidebarListHeader} ${
          !isExpanded ? styles.sidebarListHeaderFold : ''
        }`}
      >
        {isExpanded && (
          <div
            onClick={onToggle}
            className={styles.sidebarListHeaderTitle}
          >
            <span>{title}</span>
            <span className={styles.sidebarListHeaderArrow}>
              {expanded ? <DownGroupedIcon /> : <UnGroupedIcon />}
            </span>
          </div>
        )}
        {showAddButton && (
          <div
            className={`${styles.sidebarListHeaderAdd} ${
              !isExpanded ? styles.sidebarListHeaderAddFold : ''
            }`}
          >
            <Tooltip title='新建会话分组' placement='right'>
              {/* 包一层原生元素，避免 Tooltip 对 SVG 子组件走 findDOMNode（React 严格模式告警） */}
              <span
                role='button'
                tabIndex={0}
                className={styles.sidebarListHeaderAddTrigger}
                onClick={() => setCreateModalOpen(true)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault()
                    setCreateModalOpen(true)
                  }
                }}
              >
                <AddIcon />
              </span>
            </Tooltip>
          </div>
        )}
      </div>
    )
  }

  // 会话分组列表菜单
  const menuList: MenuItem[] = [
    {
      text: tC('button.rename'),
      key: EMenuOperator.RENAME,
      icon: <EditFileIcon />,
    },
    {
      text: tC('button.delete'),
      key: EMenuOperator.DELETE,
      icon: <DeleteIcon />,
    },
  ]

  // 未分组会话的菜单列表（包含移至会话分组）
  const getUngroupedSessionMenuList = (sessionId: number): MenuItem[] => {
    const setProjectPopoverOpen = (open: boolean) => {
      setProjectPopoverOpenMap((prev) => ({
        ...prev,
        [sessionId]: open,
      }))
    }

    return [
      {
        text: tC('button.rename'),
        key: EMenuOperator.RENAME,
        icon: <EditFileIcon />,
      },
      {
        text: '移至会话分组',
        key: EMenuOperator.MOVE_TO_GROUP,
        icon: <MoveIcon />,
        popoverContent: (
          <ProjectListPopover
            open={projectPopoverOpenMap[sessionId] || false}
            onOpenChange={setProjectPopoverOpen}
            onSelect={(projectId) => handleMoveSession(sessionId, projectId)}
            excludeProjectId={agentProjectId}
          >
            <span />
          </ProjectListPopover>
        ),
      },
      {
        text: tC('button.delete'),
        key: EMenuOperator.DELETE,
        icon: <DeleteIcon />,
      },
    ]
  }

  const [editId, setEditId] = useState<number>(-1)
  const [editSessionId, setEditSessionId] = useState<number>(-1)

  const handleMenuClick = async (item: MenuItem, id: number) => {
    switch (item.key) {
      case EMenuOperator.DELETE:
        setDeleteProjectId(id)
        setDeleteModalOpen(true)
        break
      case EMenuOperator.RENAME:
        setEditId(id)
        break
    }
  }

  const handleEditClose = async (id: number, isChange: boolean, newText: string) => {
    if (isChange) {
      try {
        await handleRenameProject({
          id,
          name: newText,
        })
        message.success(t('app.docs.detail.fileEdit.renameSuccess'))
        // 重新加载项目列表，确保UI更新
        await reloadProjectList()
      } catch (error) {
        console.log('重命名会话分组失败:', error)
      }
    }
    setEditId(-1)
  }

  const handleSessionEditClose = (
    id: number,
    isChange: boolean,
    newText: string,
  ) => {
    if (isChange) {
      handleRenameSession(id, newText)
    }
    setEditSessionId(-1)
  }

  const renderProjectItem = (item: any) => {
    const isAgentProject = agentProjectId !== null && item.id === agentProjectId

    const handleAgentProjectClick = () => {
      // 直接打开弹窗，由弹窗内部处理数据加载和空状态显示
      setAgentListPopoverOpen(true)
    }

    const itemContent = (
      <Item<MenuItem>
        style={{
          marginBottom: '2px',
        }}
        status={props.expandableStatus}
        text={item.name}
        icon={<HomeGroupIcon />}
        key={item.id}
        isEdit={editId === item.id}
        active={isPathActive('project', item.id)}
        onMenuClick={(menuItem) => handleMenuClick(menuItem, item.id)}
        menuList={!disabledIds.includes(item.id) ? menuList : []}
        onClick={() => {
          if (isAgentProject) {
            handleAgentProjectClick()
          } else {
            pushProjectPage(item.id)
          }
        }}
        onEditClose={(isChange: boolean, newText: string) =>
          handleEditClose(item.id, isChange, newText)
        }
      />
    )

    if (isAgentProject) {
      return (
        <AgentListPopover
          open={agentListPopoverOpen}
          onOpenChange={setAgentListPopoverOpen}
        >
          {itemContent}
        </AgentListPopover>
      )
    }

    return itemContent
  }

  const handleUngroupedSessionClick = (sessionId: number) => {
    // 未分组会话使用特殊的项目ID 0 来进入项目问答页面
    clearPendingUngroupedActive()
    navigate(`/project/0/${sessionId}`)
  }

  const renderUngroupedSessionItem = (item: {
    id: number
    name: string
    nameLoading?: boolean
    pendingActive?: boolean
  }) => {
    // 检查是否在未分组会话的项目页面中（project/0/:session_id）
    const isActive =
      item.pendingActive ||
      (currentPathname.startsWith('/project/0/') && currentSessionId === item.id)
    if (item.nameLoading) {
      return (
        <div
          key={item.id}
          className={`${styles.pendingSessionItem} ${
            isActive ? styles.pendingSessionItemActive : ''
          }`}
          onClick={() => handleUngroupedSessionClick(item.id)}
        >
          <span className={styles.pendingSessionName} />
        </div>
      )
    }
    return (
      <Item
        style={{
          marginBottom: '2px',
        }}
        status={props.expandableStatus}
        text={item.name}
        icon={null}
        key={item.id}
        active={isActive}
        isEdit={editSessionId === item.id}
        onMenuClick={(menuItem: MenuItem) => {
          switch (menuItem.key) {
            case EMenuOperator.DELETE:
              handleDeleteSession(item.id)
              break
            case EMenuOperator.RENAME:
              setEditSessionId(item.id)
              break
            case EMenuOperator.MOVE_TO_GROUP:
              // 不需要处理，由popoverContent处理
              break
          }
        }}
        menuList={getUngroupedSessionMenuList(item.id)}
        onClick={() => handleUngroupedSessionClick(item.id)}
        onEditClose={(isChange: boolean, newText: string) =>
          handleSessionEditClose(item.id, isChange, newText)
        }
      />
    )
  }

  return (
    <>
      <div className={styles.sidebarList}>
        {/* 会话分组部分 */}
        <div className={styles.sidebarListSection}>
          {renderHeader(
            t('app.sidebar.sessionGroup'),
            expend,
            handleChangeExpend,
            true,
          )}
          <div className={styles.sidebarListContent}>
            <div
              className={`${styles.sidebarListContentScroll} ${
                !expend ? styles.sidebarListContentScrollHidden : ''
              }`}
            >
              {projectList.map(renderProjectItem)}
            </div>
          </div>
        </div>

        {/* 未分组部分 */}
        <div className={styles.sidebarListSection}>
          {renderHeader(
            t('app.sidebar.ungrouped'),
            ungroupedExpanded,
            () => setUngroupedExpanded(!ungroupedExpanded),
          )}
          <div className={styles.sidebarListContent}>
            <div
              className={`${styles.sidebarListContentScroll} ${
                !ungroupedExpanded ? styles.sidebarListContentScrollHidden : ''
              }`}
            >
              {loadingUngrouped ? (
                <Skeleton active paragraph={{ rows: 3 }} />
              ) : (
                ungroupedSessions.map(renderUngroupedSessionItem)
              )}
            </div>
          </div>
        </div>
      </div>

      <CreateGroupModal
        open={createModalOpen}
        onCancel={() => setCreateModalOpen(false)}
        onSuccess={async (name) => {
          await handleCreateProject(name)
        }}
      />

      <DeleteGroupModal
        open={deleteModalOpen}
        onCancel={() => {
          setDeleteModalOpen(false)
          setDeleteProjectId(null)
        }}
        onConfirm={async (moveToFree) => {
          if (deleteProjectId !== null) {
            await handleDeleteProject(deleteProjectId, moveToFree)
            setDeleteModalOpen(false)
            setDeleteProjectId(null)
            // 如果移动到未分组，刷新未分组列表
            if (moveToFree) {
              reloadUngroupedSessions()
            }
          }
        }}
      />
    </>
  )
}
