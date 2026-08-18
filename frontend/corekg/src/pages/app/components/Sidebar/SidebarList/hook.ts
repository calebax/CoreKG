import { useState, useCallback, useEffect, useRef } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { message } from 'antd'
import { useTranslation } from 'react-i18next'
import {
  getSessionHistory,
  removeChatSession,
  renameChatSession,
  moveChatSession,
} from '@/api'
import { getFileQaProject } from '@/api/knowledge'
import useProjectStore from '@/stores/project'

type UngroupedSession = {
  id: number
  name: string
  nameLoading?: boolean
  pendingActive?: boolean
}

export default function useSidebarList() {
  const navigate = useNavigate()
  const { pathname } = useLocation()
  const { t } = useTranslation('pages')
  const { load, add, del, rename, projectList } = useProjectStore()
  const [expend, setExpend] = useState<boolean>(true)
  const [ungroupedSessions, setUngroupedSessions] = useState<
    UngroupedSession[]
  >([])
  const pendingSessionIdsRef = useRef(new Set<number>())
  const [loadingUngrouped, setLoadingUngrouped] = useState(false)

  const [disabledIds, setDisabledIds] = useState<string[]>([])
  const [agentProjectId, setAgentProjectId] = useState<number | null>(null)

  const pushProjectPage = useCallback(
    (id: number) => {
      navigate(`/project/${id}`)
    },
    [navigate],
  )

  const loadUngroupedSessions = useCallback(async () => {
    setLoadingUngrouped(true)
    try {
      const res = await getSessionHistory({
        limit: 9999,
        offset: 0,
        project_id: -1,
      })
      const list: UngroupedSession[] = (res.Data || []).map((item: any) => {
        const hasName = Boolean(item.name?.trim())
        if (hasName) pendingSessionIdsRef.current.delete(item.ID)
        return {
          id: item.ID,
          name: item.name,
          nameLoading: pendingSessionIdsRef.current.has(item.ID) && !hasName,
        }
      })
      const loadedIds = new Set(list.map((item) => item.id))
      pendingSessionIdsRef.current.forEach((id) => {
        if (!loadedIds.has(id)) {
          list.unshift({ id, name: '', nameLoading: true })
        }
      })
      setUngroupedSessions(list)
    } catch (error) {
      console.log('加载未分组会话失败：', error)
    } finally {
      setLoadingUngrouped(false)
    }
  }, [])

  const addPendingUngroupedSession = useCallback((sessionId: number) => {
    pendingSessionIdsRef.current.add(sessionId)
    setUngroupedSessions((prev) => [
      { id: sessionId, name: '', nameLoading: true, pendingActive: true },
      ...prev
        .filter((item) => item.id !== sessionId)
        .map((item) => ({ ...item, pendingActive: false })),
    ])
  }, [])

  const clearPendingUngroupedActive = useCallback(() => {
    setUngroupedSessions((prev) =>
      prev.map((item) => ({ ...item, pendingActive: false })),
    )
  }, [])

  const loadList = useCallback(
    async function () {
      const { file, agent } = await getFileQaProject()
      setDisabledIds([file.project_id, agent.project_id])
      setAgentProjectId(agent.project_id)
      load()
      loadUngroupedSessions()
    },
    [load, loadUngroupedSessions],
  )

  useEffect(() => {
    loadList()
  }, [loadList])

  useEffect(() => {
    if (pathname === '/') {
      navigate('global')
    }
  }, [pathname, navigate])

  const handleAddProject = async (name?: string) => {
    try {
      const id = await add(name)
      message.success(t('app.sidebar.createSuccess'))
      pushProjectPage(id)
    } catch (error) {
      console.log('创建项目失败：', error)
    }
  }
  const handleDeleteProject = async (id: number, moveToFree: boolean) => {
    await del(id, moveToFree)
    message.success(t('app.sidebar.deleteSuccess'))
    if (pathname.includes(`/project/${id}`)) {
      navigate('/')
    }
  }

  const handleDeleteSession = async (id: number) => {
    try {
      await removeChatSession(id)
      message.success(t('app.sidebar.deleteSuccess'))
      // 如果当前正处于该未分组会话下，则返回首页
      if (window.location.pathname.endsWith(`/project/0//${id}`)) {
        navigate('/')
      }
      await loadUngroupedSessions()
    } catch (error) {
      // 保持静默失败处理与其他逻辑一致
      // eslint-disable-next-line no-console
      console.log('删除未分组会话失败：', error)
    }
  }

  const handleRenameSession = async (id: number, name: string) => {
    try {
      await renameChatSession({ id, name })
      message.success(t('app.docs.detail.fileEdit.renameSuccess'))
      await loadUngroupedSessions()
    } catch (error) {
      // eslint-disable-next-line no-console
      console.log('重命名未分组会话失败：', error)
    }
  }

  const handleMoveSession = async (sessionId: number, projectId: number) => {
    try {
      await moveChatSession({ id: sessionId, subject_id: projectId })
      await loadUngroupedSessions()
      await load()
      message.success('移动成功')
      navigate(`/project/${projectId}/${sessionId}`)
    } catch (error) {
      console.log('移动会话失败：', error)
    }
  }

  return {
    expend,
    setExpend,
    projectList,
    handleCreateProject: handleAddProject,
    handleDeleteProject,
    handleRenameProject: rename,
    handleChangeExpend: () => setExpend(!expend),
    pushProjectPage,
    disabledIds,
    agentProjectId,
    ungroupedSessions,
    loadingUngrouped,
    reloadUngroupedSessions: loadUngroupedSessions,
    addPendingUngroupedSession,
    clearPendingUngroupedActive,
    handleDeleteSession,
    handleRenameSession,
    handleMoveSession,
    reloadProjectList: load,
  }
}
