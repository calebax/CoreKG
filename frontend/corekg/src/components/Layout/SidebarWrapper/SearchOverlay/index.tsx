import { FC, useState, useRef, useEffect, useMemo } from 'react'
import { Input, Skeleton, Popover, Button, Tabs } from 'antd'
import { useRequest } from 'ahooks'
import { cn } from '@/utils'
import { globalSearch, getHotWords, getFileList } from '@/api/knowledge'
import TableEmptyIcon from '@/assets/icons/docs/table-empty.svg?react'
import CloseIcon from '@/assets/icons/home/home-search-close.svg?react'
import SearchIcon from '@/assets/icons/home/home-search.svg?react'
import type { SessionConfig, SessionInfo } from '@/pages/project/ProjectContent'
import { SessionContext, useSessionInfo } from '@/pages/project/ProjectContent'
import type { ResultType } from '@/pages/project/ProjectContent/EmptyProject/SearchResult'
import { ResultContent } from '@/pages/project/ProjectContent/EmptyProject/SearchResult/ResultContent'
import {
  KnowledgeSelect,
  KnowledgeStatus,
  withKnowledgeDataProvider,
  useKnowledgeData,
  type Knowledge,
} from '@/pages/project/ProjectContent/Knowledge'
import { useDeployConfig } from '@/utils/useDeployConfig'
import styles from './index.module.scss'

export type SearchOverlayProps = {
  visible: boolean
  onClose: () => void
}

// 带确定按钮的知识库选择器包装组件
const KnowledgeSelectWithConfirm: FC<{
  onConfirm: (config: Partial<SessionConfig>) => void
  onCancel?: () => void
  onClose?: () => void
  initialConfig: Partial<SessionConfig>
}> = ({ onConfirm, onCancel, onClose, initialConfig }) => {
  const { knowledgeList } = useKnowledgeData()
  const [tempConfig, setTempConfig] = useState<Partial<SessionConfig>>({
    knowledge: initialConfig?.knowledge || [],
    tag_ids: initialConfig?.tag_ids || [],
  })

  // 获取所有原子节点（用于判断全选）
  const getAllAtomNodes = (nodes: Knowledge[]): Knowledge[] => {
    const result: Knowledge[] = []
    const traverse = (node: Knowledge) => {
      if (node.knowledgeType !== 'other') {
        result.push(node)
      } else if (node.children) {
        node.children.forEach(traverse)
      }
    }
    nodes.forEach(traverse)
    return result
  }

  // 文档模式下的知识库列表（去除表格文件节点）
  const knowledgeListForDoc = useMemo(() => {
    if (!knowledgeList) return []
    return knowledgeList.filter(
      (item) =>
        !['excel_sheet', 'mysql_table'].includes(item.knowledgeType) &&
        !(
          item.forest_type === 'data' &&
          item.forest_data_source_type === 'excel'
        ),
    )
  }, [knowledgeList])

  // 获取所有原子节点
  const allAtomNodes = useMemo(() => {
    return getAllAtomNodes(knowledgeListForDoc)
  }, [knowledgeListForDoc])

  // 判断是否全选
  const isSelectAll = useMemo(() => {
    const selected = tempConfig.knowledge || []
    if (!selected.length || !allAtomNodes.length) return false
    const selectedKeys = new Set(selected.map((item) => item.key))
    return allAtomNodes.every((node) => selectedKeys.has(node.key))
  }, [tempConfig.knowledge, allAtomNodes])

  // 临时设置配置
  const setTempSessionConfig = (
    val: Partial<SessionConfig> | ((draft: Partial<SessionConfig>) => void),
  ) => {
    setTempConfig((prev) => {
      const newVal = typeof val === 'function' ? { ...prev } : val
      if (typeof val === 'function') {
        val(newVal)
      }
      return { ...prev, ...newVal }
    })
  }

  // 创建临时 SessionContext
  const tempSessionContextValue: any = {
    sessionStatus: 'new',
    sessionConfig: tempConfig,
    setSessionConfig: setTempSessionConfig,
  }

  const handleConfirm = () => {
    onConfirm(tempConfig)
    onClose?.()
  }

  const handleCancel = () => {
    setTempConfig({
      knowledge: initialConfig?.knowledge || [],
      tag_ids: initialConfig?.tag_ids || [],
    })
    onCancel?.()
    onClose?.()
  }

  // 获取选中标签的文件总数
  const [tempTagFileCount, setTempTagFileCount] = useState(0)
  useRequest(
    async () => {
      const tag_ids = tempConfig.tag_ids
      const actualTagIds = tag_ids?.filter((id) => id >= 0) || []
      if (actualTagIds.length === 0) {
        setTempTagFileCount(0)
        return
      }
      const res = await getFileList({
        forest_id: 0,
        filters: [{ field: 'tag_ids', value: actualTagIds.map(String) }],
        limit: 1,
      })
      setTempTagFileCount(res.total ?? 0)
    },
    {
      refreshDeps: [tempConfig.tag_ids],
    },
  )

  // 临时选择知识展示的文字
  const tempSelectedKnowledgeText = useMemo(() => {
    const { knowledge, tag_ids } = tempConfig
    const actualTagIds = tag_ids?.filter((id) => id >= 0) || []

    if (actualTagIds.length > 0) {
      return `已选资源(${tempTagFileCount})`
    }

    if (isSelectAll) {
      return '全部'
    }

    if (knowledge && knowledge.length > 0) {
      return `已选资源(${knowledge.length})`
    }

    return '选择资源'
  }, [tempConfig, tempTagFileCount, isSelectAll])

  return (
    <SessionContext.Provider value={tempSessionContextValue}>
      <div className='flex flex-col min-w-80'>
        <KnowledgeSelect
          className='min-w-100 max-h-[40vh]'
          globalSearch
          tabBarExtraContent={
            <div className='flex items-center gap-2'>
              <Button
                size='small'
                onClick={handleCancel}
                className='bg-[#ffffff] border-[#d9d9d9]'
              >
                取消
              </Button>
              <Button
                size='small'
                type='primary'
                onClick={handleConfirm}
                className='bg-[#0C99FF]'
              >
                确定
              </Button>
            </div>
          }
        />
      </div>
    </SessionContext.Provider>
  )
}

const SearchOverlayInner: FC<SearchOverlayProps> = ({ visible, onClose }) => {
  const [search, setSearch] = useState<string>('')
  const inputRef = useRef<any>(null)
  const [isDebouncing, setIsDebouncing] = useState(false)
  const [popoverOpen, setPopoverOpen] = useState(false)
  const { version, mode } = useDeployConfig()
  const { knowledgeList } = useKnowledgeData()

  // 知识库选择相关状态（已确认的选择）
  const [sessionConfig, setSessionConfigState] = useState<
    Partial<SessionConfig>
  >({
    mode: 'knowledge',
    knowledge: [],
    tag_ids: [],
  })

  const setSessionConfig = (
    val: Partial<SessionConfig> | ((draft: Partial<SessionConfig>) => void),
  ) => {
    setSessionConfigState((prev) => {
      const newVal = typeof val === 'function' ? { ...prev } : val
      if (typeof val === 'function') {
        val(newVal)
      }
      return { ...prev, ...newVal }
    })
  }

  // 获取所有原子节点（用于判断全选）
  const getAllAtomNodes = (nodes: Knowledge[]): Knowledge[] => {
    const result: Knowledge[] = []
    const traverse = (node: Knowledge) => {
      if (node.knowledgeType !== 'other') {
        result.push(node)
      } else if (node.children) {
        node.children.forEach(traverse)
      }
    }
    nodes.forEach(traverse)
    return result
  }

  // 文档模式下的知识库列表（去除表格文件节点）
  const knowledgeListForDoc = useMemo(() => {
    if (!knowledgeList) return []
    return knowledgeList.filter(
      (item) =>
        !['excel_sheet', 'mysql_table'].includes(item.knowledgeType) &&
        !(
          item.forest_type === 'data' &&
          item.forest_data_source_type === 'excel'
        ),
    )
  }, [knowledgeList])

  // 获取所有原子节点
  const allAtomNodes = useMemo(() => {
    return getAllAtomNodes(knowledgeListForDoc)
  }, [knowledgeListForDoc])

  // 判断是否全选
  const isKnowledgeSelectAll = useMemo(() => {
    const selected = sessionConfig.knowledge || []
    if (!selected.length || !allAtomNodes.length) return false
    const selectedKeys = new Set(selected.map((item) => item.key))
    return allAtomNodes.every((node) => selectedKeys.has(node.key))
  }, [sessionConfig.knowledge, allAtomNodes])

  // 获取选中标签的文件总数
  const [tagFileCount, setTagFileCount] = useState(0)
  useRequest(
    async () => {
      const tag_ids = sessionConfig.tag_ids
      const actualTagIds = tag_ids?.filter((id) => id >= 0) || []
      if (actualTagIds.length === 0) {
        setTagFileCount(0)
        return
      }
      const res = await getFileList({
        forest_id: 0,
        filters: [{ field: 'tag_ids', value: actualTagIds.map(String) }],
        limit: 1,
      })
      setTagFileCount(res.total ?? 0)
    },
    {
      refreshDeps: [sessionConfig.tag_ids],
    },
  )

  // 提取选中的 forest_ids
  const [resolvedTagFileIds, setResolvedTagFileIds] = useState<number[]>([])

  useRequest(
    async () => {
      const tag_ids = sessionConfig.tag_ids
      const actualTagIds = tag_ids?.filter((id) => id >= 0) || []
      if (actualTagIds.length === 0) {
        setResolvedTagFileIds([])
        return
      }
      try {
        const fileListRes = await getFileList({
          forest_id: 0,
          filters: [{ field: 'tag_ids', value: actualTagIds.map(String) }],
          limit: 9999,
        })
        const ids = (fileListRes.data || []).map((file: any) => Number(file.ID))
        setResolvedTagFileIds(ids)
      } catch (error) {
        console.log('获取标签文件失败', error)
        setResolvedTagFileIds([])
      }
    },
    {
      refreshDeps: [sessionConfig.tag_ids],
    },
  )

  const forestIds = useMemo(() => {
    const ids: number[] = []
    if (sessionConfig.knowledge?.length) {
      // 全局搜索场景下，只提取知识库ID（forest_id），不包含文件ID
      // 只处理知识库节点（node_type === 'forest'）
      sessionConfig.knowledge.forEach((item) => {
        // 只提取知识库节点的forest_id，不提取文件ID
        if (
          item.node_type === 'forest' &&
          item.forest_id !== undefined &&
          item.forest_id !== null
        ) {
          const forestId = Number(item.forest_id)
          if (!ids.includes(forestId)) {
            ids.push(forestId)
          }
        }
      })
    }
    // 全局搜索场景下不使用标签，注释掉标签相关的逻辑
    // // 合并标签解析出的文件 ID
    // if (resolvedTagFileIds.length > 0) {
    //   resolvedTagFileIds.forEach((id) => {
    //     if (!ids.includes(id)) ids.push(id)
    //   })
    // }
    return ids
  }, [sessionConfig.knowledge])

  /** 选择知识展示的文字 */
  const selectedKnowledgeText = useMemo(() => {
    const { knowledge, tag_ids } = sessionConfig
    const actualTagIds = tag_ids?.filter((id) => id >= 0) || []

    // 标签模式优先展示
    if (actualTagIds.length > 0) {
      return `已选资源(${tagFileCount})`
    }

    // 全选时显示"全部"
    if (isKnowledgeSelectAll) {
      return '全部'
    }

    if (knowledge && knowledge.length > 0) {
      return `已选资源(${knowledge.length})`
    }

    return '选择资源'
  }, [sessionConfig, tagFileCount, isKnowledgeSelectAll])

  // 环境判断：只在本地环境、测试环境、生产环境、或 custom 版本且 mode 为 cimc/h3c 时显示搜索热词
  const isDevEnv = import.meta.env.MODE === 'development'
  const isTestEnv = import.meta.env.MODE === 'test'
  const isProdEnv = import.meta.env.MODE === 'production'
  const shouldShowHotWords =
    isDevEnv ||
    isTestEnv ||
    isProdEnv ||
    (version === 'custom' && (mode === 'cimc' || mode === 'h3c'))

  // 获取热词数据
  const { data: hotWordsData } = useRequest(
    async () => {
      if (!visible) return null
      const res = await getHotWords()
      return (res as { words?: string[] })?.words || []
    },
    {
      refreshDeps: [visible],
      ready: shouldShowHotWords,
    },
  )

  useEffect(() => {
    if (visible) {
      // 弹窗打开时聚焦输入框
      if (inputRef.current) {
        setTimeout(() => {
          inputRef.current?.focus()
        }, 100)
      }
    } else {
      // 弹窗关闭时，重置搜索状态，确保下次打开时显示热词
      setSearch('')
      // 重置知识库选择
      setSessionConfigState({
        mode: 'knowledge',
        knowledge: [],
        tag_ids: [],
      })
      // 重置选择资源弹窗状态
      setPopoverOpen(false)
    }
  }, [visible])

  const { data, loading, run } = useRequest(
    async () => {
      if (!search?.trim()) return null

      // 准备搜索参数
      const params: any = { text: search! }
      params.forest_ids = forestIds // 默认传数组，空数组也可以

      const {
        forest_search_result,
        doc_search_result,
        agent_search_result,
        video_search_result,
        image_search_result,
        external_search_result,
      } = (await globalSearch(params)) as Record<string, any | undefined>
      const _result: { type: ResultType; values: any[] }[] = []
      if (forest_search_result) {
        _result.push({ type: 'forest', values: forest_search_result })
      }
      if (doc_search_result) {
        _result.push({ type: 'doc', values: doc_search_result })
      }
      // 不同类型的外部数据源
      const {
        gmail_search,
        gmail_drive_search,
        confluence_search,
        slack_search,
      } = external_search_result ?? {}
      const externalData: { external_type: string; value: any }[] = []
      if (Array.isArray(gmail_search?.items)) {
        gmail_search.items.forEach((value: any) => {
          externalData.push({ external_type: 'gmail', value })
        })
      }
      if (Array.isArray(gmail_drive_search?.files)) {
        gmail_drive_search.files.forEach((value: any) => {
          externalData.push({ external_type: 'google_drive', value })
        })
      }
      if (Array.isArray(confluence_search?.results)) {
        confluence_search.results.forEach((value: any) => {
          externalData.push({ external_type: 'confluence', value })
        })
      }
      if (Array.isArray(slack_search?.files?.files)) {
        slack_search.files.files.forEach((value: any) => {
          externalData.push({ external_type: 'slack', value })
        })
      }

      if (externalData.length > 0) {
        _result.push({ type: 'connect_app', values: externalData })
      }

      if (agent_search_result) {
        _result.push({ type: 'agent', values: agent_search_result })
      }
      if (video_search_result) {
        _result.push({ type: 'video', values: video_search_result })
      }
      if (image_search_result) {
        _result.push({ type: 'pic', values: image_search_result })
      }
      return _result
    },
    {
      manual: true,
      debounceWait: 1000,
      onFinally: () => setIsDebouncing(false),
    },
  )

  // 手动触发搜索（当确认选择知识库后）
  const handleConfirmKnowledge = (config: Partial<SessionConfig>) => {
    setSessionConfigState(config)
    // 如果当前有搜索文本，触发搜索
    if (search?.trim()) {
      setIsDebouncing(true)
      run()
    }
  }

  useEffect(() => {
    if (search?.trim()) {
      setIsDebouncing(true)
      run()
      return
    }
    setIsDebouncing(false)
  }, [search, run]) // 移除 forestIds 依赖，只有确认选择后才触发搜索

  const hasSearchText = Boolean(search?.trim())
  const hasResults = data && data.length > 0
  const isShowLoading = loading || isDebouncing

  const handleOverlayClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose()
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      onClose()
    }
  }

  const handleClearInput = (e: React.MouseEvent) => {
    e.stopPropagation()
    setSearch('')
    inputRef.current?.focus()
  }

  // 点击热词处理函数
  const handleHotWordClick = (word: string) => {
    setSearch(word)
    // 触发搜索（通过设置 search state，useRequest 会自动触发）
    inputRef.current?.focus()
  }

  const handleHotWordKeyDown = (e: React.KeyboardEvent, word: string) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault()
      handleHotWordClick(word)
    }
  }

  if (!visible) return null

  // 模拟 SessionContext 内容
  const sessionContextValue: any = {
    sessionStatus: 'new',
    sessionConfig,
    setSessionConfig,
  }

  return (
    <SessionContext.Provider value={sessionContextValue}>
      <div
        className={styles.searchOverlay}
        onClick={handleOverlayClick}
        onKeyDown={handleKeyDown}
        tabIndex={-1}
      >
        <div
          className={cn(styles.searchModal, {
            [styles.searchModalWithContent]: hasSearchText && hasResults,
          })}
        >
          <div className={styles.searchHeader}>
            <Input
              ref={inputRef}
              className={styles.searchInput}
              placeholder='请输入'
              prefix={<SearchIcon className={styles.searchIcon} />}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              suffix={
                <div className='flex items-center gap-2'>
                  {hasSearchText && (
                    <CloseIcon
                      className={styles.closeIcon}
                      onClick={handleClearInput}
                      tabIndex={0}
                      role='button'
                      aria-label='清空输入框'
                      onKeyDown={(e) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                          e.preventDefault()
                          handleClearInput(e as any)
                        }
                      }}
                    />
                  )}
                  <Popover
                    arrow={false}
                    placement='bottomRight'
                    trigger='click'
                    open={popoverOpen}
                    onOpenChange={(open) => {
                      // 只允许通过取消按钮关闭，点击外部不关闭
                      if (!open) {
                        // 尝试关闭时，不执行任何操作（保持打开状态）
                        return
                      }
                      setPopoverOpen(open)
                    }}
                    content={
                      <KnowledgeSelectWithConfirm
                        onConfirm={handleConfirmKnowledge}
                        onClose={() => setPopoverOpen(false)}
                        initialConfig={sessionConfig}
                      />
                    }
                    overlayClassName='resource-select-popover'
                  >
                    <KnowledgeStatus
                      active={Boolean(
                        sessionConfig.knowledge?.length ||
                          sessionConfig.tag_ids?.length,
                      )}
                      title={selectedKnowledgeText}
                      defaultTitle='选择资源'
                    />
                  </Popover>
                </div>
              }
            />
          </div>

          {hasSearchText ? (
            <div className={styles.searchContent}>
              {isShowLoading ? (
                <div className={styles.loadingWrapper}>
                  <Skeleton className='p-4' active />
                </div>
              ) : !hasResults ? (
                <div className={styles.emptyWrapper}>
                  <TableEmptyIcon className={styles.emptyIcon} />
                  <div className={styles.emptyText}>暂无搜索结果</div>
                </div>
              ) : (
                <div className={styles.resultsWrapper}>
                  <ResultContent result={data} />
                </div>
              )}
            </div>
          ) : (
            shouldShowHotWords &&
            hotWordsData &&
            hotWordsData.length > 0 && (
              <div className={styles.hotWordsWrapper}>
                {hotWordsData.map((word, index) => (
                  <div
                    key={index}
                    className={styles.hotWordItem}
                    onClick={() => handleHotWordClick(word)}
                    onKeyDown={(e) => handleHotWordKeyDown(e, word)}
                    tabIndex={0}
                    role='button'
                    aria-label={`搜索热词：${word}`}
                  >
                    {word}
                  </div>
                ))}
              </div>
            )
          )}
        </div>
      </div>
    </SessionContext.Provider>
  )
}

export const SearchOverlay = withKnowledgeDataProvider(SearchOverlayInner)
