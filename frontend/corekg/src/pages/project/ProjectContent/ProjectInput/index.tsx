import {
  FC,
  forwardRef,
  ReactNode,
  useImperativeHandle,
  useMemo,
  useRef,
  useEffect,
  useState,
} from 'react'
import { App, Popover } from 'antd'
import { useControllableValue, useMount, useRequest } from 'ahooks'
import axios from 'axios'
import { useTranslation } from 'react-i18next'
import { match } from 'ts-pattern'
import globalConfig from '@/config'
import { cn, hasModulePermission, spliteFileName } from '@/utils'
import { uploadImage, uploadAttachment } from '@/api/common'
import {
  getKnowledgeBaseList,
  getFileList,
  expansionQuestion,
} from '@/api/knowledge'
import NetworkingIcon from '@/assets/icons/home/home-networking.svg?react'
import UploadIcon from '@/assets/icons/home/home-upload.svg?react'
import { AIPolishButton } from '@/components/common/AIPolishButton'
import { loadFile } from '@/utils/loadFile'
import { useDeployConfig } from '@/utils/useDeployConfig'
import { useLoginGlobalData } from '@/utils/useLoginGlobalData'
import { useQaInputMaxLength } from '@/utils/useQaInputMaxLength'
import { SessionConfig, useSessionInfo } from '..'
import { useProject } from '../..'
import {
  DialogInput,
  DisabledBtn,
  ActiveBtn,
  StopBtn,
  Attachment,
} from '../../components/DialogInput'
import { ExternalDataSourceSelect } from '../../components/ExternalDataSourceSelect'
import { ModelSelect } from '../../components/ModelSelect'
import { PromptModeSelect } from '../../components/PromptModeSelect'
import { ExternalDataProviderSelect } from '../ExternalDataProviderSelect'
import {
  KnowledgeList,
  KnowledgeSelect,
  KnowledgeStatus,
  useKnowledgeData,
  GraphEmptyModal,
} from '../Knowledge'
import DatabaseIcon from './images/database.svg?react'
import ExternalDataIcon from './images/external-data.svg?react'
import GraphIcon from './images/graph.svg?react'
import H3C from './images/h3c.svg?react'
import KnowledgeIcon from './images/knowledge.svg?react'
import ModalIcon from './images/modal.svg?react'
import GraphSearchIcon from './images/search.svg?react'
import TableIcon from './images/table.svg?react'

export const ProjectInput = forwardRef<
  { startQA: () => void },
  Style &
    ValueController<string> & {
      onAsk?: () => void
      showRightMargin?: boolean
    }
>((props, ref) => {
  const { version, isH3CTest } = useDeployConfig()
  const qaInputMaxLength = useQaInputMaxLength()
  const { license } = useLoginGlobalData()
  const hasGraphPermission = hasModulePermission(license, 'graph')

  const { t } = useTranslation('pages')
  // 默认不显示 mr-3，只有在明确需要时才显示
  const { className, style, onAsk, showRightMargin = false } = props

  const {
    models,
    externalDataSourceList,
    data: { charts },
    isOtherPage,
    defaultKnowBase,
    type,
    isUngroupedSession,
    session_id: currentSessionId,
  } = useProject()
  const { message } = App.useApp()
  const [search, setSearch] = useControllableValue<string | undefined>(props)
  const { version: deployVersion, mode: deployMode } = useDeployConfig()

  // 上传文件和联网搜索相关状态
  const [enableWebSearch, setEnableWebSearch] = useState(false)
  const [attachments, setAttachments] = useState<Attachment[]>([])
  const cancelledUploadKeysRef = useRef(new Set<string>())
  const uploadControllersRef = useRef(new Map<string, AbortController>())
  const MAX_ATTACHMENTS = 5

  const cancelUploadByKey = (uploadKey: string) => {
    cancelledUploadKeysRef.current.add(uploadKey)
    uploadControllersRef.current.get(uploadKey)?.abort()
    uploadControllersRef.current.delete(uploadKey)
  }

  const cancelAllUploads = () => {
    uploadControllersRef.current.forEach((controller, uploadKey) => {
      cancelledUploadKeysRef.current.add(uploadKey)
      controller.abort()
    })
    uploadControllersRef.current.clear()
  }

  const isUploadAborted = (error: unknown) =>
    axios.isCancel(error) ||
    (error as { code?: string })?.code === 'ERR_CANCELED'
  // AI润色状态
  const [isPolishing, setIsPolishing] = useState(false)

  const {
    sessionStatus,
    dialogStatus,
    sessionConfig = {},
    setSessionConfig,
    startQA: _startQA,
    stopQA,
    dialog,
  } = useSessionInfo()

  const {
    model_id,
    knowledge,
    graphKnowledgeBase,
    tableKnowledgeBase,
    databaseKnowledgeBase,
    externalIds,
    externalDataProviders,
    mode,
    prompt_key,
    enable_web_search,
    tag_ids,
  } = sessionConfig

  // 获取选中标签的文件总数
  const [tagFileCount, setTagFileCount] = useState(0)
  useRequest(
    async () => {
      // 过滤出真正的标签 ID（正数或0）
      const actualTagIds = tag_ids?.filter((id) => id >= 0) || []
      if (mode !== 'knowledge' || actualTagIds.length === 0) {
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
      refreshDeps: [tag_ids, mode],
    },
  )

  // 判断是否第一次提问：没有对话记录就是第一次提问
  const isFirstQuestion = dialog.length === 0

  // 从 sessionConfig 同步 enable_web_search 状态
  useEffect(() => {
    // 如果 sessionConfig 中有 enable_web_search，使用该值；否则重置为 false
    setEnableWebSearch(enable_web_search ?? false)
  }, [enable_web_search])

  const dataSelected = match({ mode, type })
    .with({ mode: 'h3c-test' }, () => true)
    .with({ mode: 'model' }, () => true)
    .with({ mode: 'graph_search' }, () => true)
    .with({ mode: 'graph' }, () => graphKnowledgeBase?.length)
    .with({ mode: 'knowledge', type: 'single-file' }, () => true)
    .with({ mode: 'knowledge' }, () => {
      // 如果选择了标签，则需要标签下的文件数大于 0
      const actualTagIds = tag_ids?.filter((id) => id >= 0) || []
      if (actualTagIds.length > 0) {
        return tagFileCount > 0
      }
      // 否则检查是否选择了知识库或外部数据源
      return knowledge?.length || externalIds?.length
    })
    .with({ mode: 'table' }, () => tableKnowledgeBase?.length)
    .with({ mode: 'database' }, () => databaseKnowledgeBase?.length)
    .with({ mode: 'external_data' }, () => externalDataProviders?.length)
    .otherwise(() => false)

  const modelSelected = mode === 'h3c-test' || model_id
  // 检查是否有附件正在上传
  const isUploading = attachments.some((att) => att.loading)
  const allowAsk =
    modelSelected &&
    dataSelected &&
    search?.trim() &&
    dialogStatus === 'ready' &&
    !isUploading &&
    !isPolishing

  // AI润色按钮是否可用（需要有输入内容且选择了资源）
  const allowPolish =
    dataSelected &&
    search?.trim() &&
    dialogStatus === 'ready' &&
    !isUploading &&
    !isPolishing

  // 判断是否显示AI润色按钮（仅在本地、测试、生产或 custom+cimc/h3c 环境显示，且仅在文档模式和单文件问答中显示）
  const showPolishButton =
    (globalConfig.env === 'development' ||
      globalConfig.apiEnv === 'test' ||
      globalConfig.env === 'production' ||
      (deployVersion === 'custom' &&
        (deployMode === 'cimc' || deployMode === 'h3c'))) &&
    (mode === 'knowledge' || type === 'single-file')

  // 全局问答、未分组会话和单文件问答共用部署配置里的问答输入上限。
  const shouldUseQaInputMaxLength =
    type === 'single-file' || type === 'global' || isUngroupedSession

  const {
    loadData,
    hasCalledLoadFn,
    knowledgeList: allForestData,
  } = useKnowledgeData()
  useMount(async () => {
    if (sessionStatus === 'new' && !hasCalledLoadFn) {
      loadData()
    }
  })

  // 获取所有子节点（用于判断是否全选）
  const getAllAtomNodes = (
    nodes: typeof allForestData,
    isSimpleMode?: boolean,
  ): typeof allForestData => {
    if (!nodes) return []
    // 简化模式（数据库）：知识库节点本身就是原子节点
    if (isSimpleMode) {
      return nodes.filter((node) => node.node_type === 'forest')
    }
    // 完整模式（文档、表格）：返回文件级别的原子节点
    const result: typeof allForestData = []
    const traverse = (node: (typeof allForestData)[0]) => {
      if (node.knowledgeType !== 'other') {
        result.push(node)
      } else if (node.children) {
        node.children.forEach(traverse)
      }
    }
    nodes.forEach(traverse)
    return result
  }

  // 获取文档模式和表格模式的所有子节点
  const knowledgeAtomNodes = useMemo(() => {
    if ((mode !== 'knowledge' && mode !== 'table') || !allForestData) return []
    // 表格模式：筛选表格类型
    if (mode === 'table') {
      return getAllAtomNodes(
        allForestData.filter(
          (item) =>
            item.forest_type === 'data' &&
            item.forest_data_source_type === 'excel',
        ),
        false, // 传入 false 表示使用完整模式（文件级别）
      )
    }
    // 文档模式：去除表格文件节点
    return getAllAtomNodes(
      allForestData.filter(
        (item) =>
          !['excel_sheet', 'mysql_table'].includes(item.knowledgeType) &&
          !(
            item.forest_type === 'data' &&
            item.forest_data_source_type === 'excel'
          ),
      ),
      false, // 传入 false 表示使用完整模式（文件级别）
    )
  }, [mode, allForestData])

  // 记录上一个模式，用于检测模式切换
  const prevModeRef = useRef<SessionConfig['mode'] | undefined>(mode)

  // 模式切换时清除选择并重置为全选
  useEffect(() => {
    const prevMode = prevModeRef.current
    prevModeRef.current = mode

    // 检测模式切换
    if (prevMode && prevMode !== mode && sessionStatus === 'new') {
      // 切换模式时清空已上传的附件
      cancelAllUploads()
      setAttachments([])

      // 清除之前模式的选择
      if (prevMode === 'knowledge') {
        setSessionConfig({ knowledge: [], externalIds: undefined })
      } else if (prevMode === 'table') {
        setSessionConfig({ tableKnowledgeBase: [] })
      } else if (prevMode === 'database') {
        setSessionConfig({ databaseKnowledgeBase: [] })
      } else if (prevMode === 'external_data') {
        setSessionConfig({ externalDataProviders: [] })
      }

      // 重置初始化标志，以便触发下方的默认选择逻辑
      // 注意：这里依赖于下方 useEffect 在此之后执行
      initKeyRef.current = ''
    }
  }, [mode, sessionStatus, setSessionConfig])

  // 默认全选逻辑：确保初始化时显示"全部"（仅在首次加载时执行，模式切换由上面的逻辑处理）
  const initKeyRef = useRef<string>('')
  useEffect(() => {
    const currentKey = `${mode}-${sessionStatus}`
    // 只在首次加载时执行，避免与模式切换逻辑冲突
    if (
      initKeyRef.current === '' &&
      sessionStatus === 'new' &&
      allForestData?.length
    ) {
      if (
        (mode === 'knowledge' || mode === 'table') &&
        knowledgeAtomNodes.length > 0
      ) {
        if (mode === 'knowledge' && !knowledge?.length) {
          setSessionConfig({
            knowledge: knowledgeAtomNodes,
            externalIds: undefined,
            prompt_key: prompt_key || 'normal',
          })
        } else if (mode === 'table' && !tableKnowledgeBase?.length) {
          setSessionConfig({ tableKnowledgeBase: knowledgeAtomNodes })
        }
        initKeyRef.current = currentKey
      }
      // 数据库模式和图谱模式：默认不选择，让用户自己选
    }
  }, [
    sessionStatus,
    mode,
    knowledgeAtomNodes,
    knowledge,
    tableKnowledgeBase,
    setSessionConfig,
  ])

  // 判断是否全选
  const isKnowledgeSelectAll = useMemo(() => {
    const selected =
      mode === 'knowledge'
        ? knowledge
        : mode === 'table'
          ? tableKnowledgeBase
          : null
    if (!selected?.length || !knowledgeAtomNodes.length) return false
    const selectedKeys = new Set(selected.map((item) => item.key))
    return knowledgeAtomNodes.every((node) => selectedKeys.has(node.key))
  }, [mode, knowledge, tableKnowledgeBase, knowledgeAtomNodes])

  /** 选择corekg知识展示的文字 */
  const selectedKnowledgeText = (() => {
    const selectedKnowledge = mode === 'graph' ? graphKnowledgeBase : knowledge
    const knowledgeType = selectedKnowledge?.[0]?.knowledgeType

    // 标签模式优先展示（包括group节点，即使没有子节点也要显示）
    const actualTagIds = tag_ids?.filter((id) => id >= 0) || []
    if (mode === 'knowledge' && tag_ids && tag_ids.length > 0) {
      return `已选资源(${tagFileCount})`
    }

    // 单文件问答页面
    // 默认选中当前文件 如果只选中了当前文件 展示特定字符串
    if (type === 'single-file') {
      if (
        !knowledge ||
        knowledge.length === 0 ||
        (knowledge.length === 1 &&
          knowledge[0]?.key === `file_${defaultKnowBase}`)
      ) {
        return '已选择当前文件'
      }
    }
    // 文档模式：全选时显示"全部"
    if (mode === 'knowledge' && isKnowledgeSelectAll) {
      return '全部'
    }

    return isOtherPage &&
      selectedKnowledge?.length === 1 &&
      selectedKnowledge?.[0]?.key === `file_${defaultKnowBase}`
      ? '已选当前资源'
      : knowledgeType
        ? t(`project.selectedKnowledgeOfInput.${knowledgeType}` as any, {
            num: selectedKnowledge?.length,
          })
        : null
  })()

  // 表格模式的显示文字
  const selectedTableText = useMemo(() => {
    if (mode === 'table' && isKnowledgeSelectAll) {
      return '全部'
    }
    return `已选表格(${tableKnowledgeBase?.length || 0})`
  }, [mode, isKnowledgeSelectAll, tableKnowledgeBase?.length])
  const selectedDatabaseStatus =
    sessionStatus === 'new'
      ? `已选数据表(${databaseKnowledgeBase?.length || 0})`
      : databaseKnowledgeBase?.length
        ? '已选资源'
        : undefined

  // 外接数据模式的显示文字
  const selectedExternalDataStatus =
    sessionStatus === 'new'
      ? `已选数据源(${externalDataProviders?.length || 0})`
      : externalDataProviders?.length
        ? '已选数据源'
        : undefined

  // 处理文件上传（支持多选，追加到已有附件，最多保留 5 个）
  const handleUpload = () => {
    const allowedExtensions = [
      'pdf',
      'txt',
      'md',
      'png',
      'jpg',
      'jpeg',
      'mp4',
      'doc',
      'docx',
    ]

    const appendAttachments = (
      newOnes: (Attachment & { uploadKey?: string })[],
    ) => {
      setAttachments((prev) => {
        const next = [...prev, ...newOnes]
        if (next.length <= MAX_ATTACHMENTS) return next
        const removed = next.slice(0, next.length - MAX_ATTACHMENTS)
        removed.forEach((att) => {
          const key = (att as Attachment & { uploadKey?: string }).uploadKey
          if (key) cancelUploadByKey(key)
        })
        return next.slice(-MAX_ATTACHMENTS)
      })
    }

    loadFile(
      async (files) => {
        if (!files || files.length === 0) return

        const entries: {
          file: File
          uploadKey: string
          attachment: Attachment & { uploadKey: string }
        }[] = []

        for (const file of Array.from(files)) {
          const { ext } = spliteFileName(file.name)
          if (!ext || !allowedExtensions.includes(ext.toLowerCase())) {
            message.warning(`${file.name}: 不支持的文件格式`)
            continue
          }

          const isMP4 = ext.toLowerCase() === 'mp4'
          const maxMB = isMP4 ? 500 : 100
          if (file.size >= maxMB * 1024 * 1024) {
            message.warning(`${file.name}: 文件大小不能超过 ${maxMB}MB`)
            continue
          }

          const extLower = ext.toLowerCase()
          const uploadKey = `${Date.now()}-${Math.random()}`
          entries.push({
            file,
            uploadKey,
            attachment: {
              name: file.name,
              type: ['jpg', 'png', 'jpeg'].includes(extLower)
                ? 'image'
                : extLower === 'mp4'
                  ? 'video'
                  : extLower,
              mime_type: file.type,
              loading: true,
              uploadKey,
            },
          })
        }

        if (entries.length === 0) return

        const uploadCount = Math.min(entries.length, MAX_ATTACHMENTS)
        const entriesToUpload = entries.slice(0, uploadCount)

        if (entries.length > uploadCount) {
          message.warning(
            `最多上传 ${MAX_ATTACHMENTS} 个附件，已选取前 ${uploadCount} 个文件`,
          )
        }

        appendAttachments(entriesToUpload.map((entry) => entry.attachment))

        await Promise.all(
          entriesToUpload.map(async ({ file, uploadKey, attachment }) => {
            const controller = new AbortController()
            uploadControllersRef.current.set(uploadKey, controller)

            try {
              const res = await uploadAttachment(
                { file, purpose: 'yg-chat' },
                {
                  timeout: 0,
                  headers: { 'Content-Type': 'multipart/form-data' },
                  signal: controller.signal,
                },
              )
              if (cancelledUploadKeysRef.current.has(uploadKey)) {
                cancelledUploadKeysRef.current.delete(uploadKey)
                return
              }
              setAttachments((prev) =>
                prev.map((att) =>
                  (att as Attachment & { uploadKey?: string }).uploadKey ===
                  uploadKey
                    ? {
                        ...attachment,
                        url: res.url,
                        id: res.id,
                        md_url: res.md_url,
                        loading: false,
                        uploadKey: undefined,
                      }
                    : att,
                ),
              )
            } catch (e) {
              if (
                isUploadAborted(e) ||
                cancelledUploadKeysRef.current.has(uploadKey)
              ) {
                cancelledUploadKeysRef.current.delete(uploadKey)
                return
              }
              console.log('上传失败', e)
              setAttachments((prev) =>
                prev.filter(
                  (att) =>
                    (att as Attachment & { uploadKey?: string }).uploadKey !==
                    uploadKey,
                ),
              )
            } finally {
              uploadControllersRef.current.delete(uploadKey)
            }
          }),
        )
      },
      {
        multiple: true,
        accept: '.pdf,.txt,.md,.png,.jpg,.jpeg,.mp4,.doc,.docx',
      },
    )
  }

  // 移除附件（含上传中的附件，立即中断上传请求）
  const handleRemoveAttachment = (index: number) => {
    setAttachments((prev) => {
      const target = prev[index] as Attachment & { uploadKey?: string }
      if (target?.uploadKey) {
        cancelUploadByKey(target.uploadKey)
      }
      return prev.filter((_, i) => i !== index)
    })
  }

  // AI润色处理函数
  const handlePolish = async () => {
    if (!allowPolish || !search) return

    // 获取选中的文件ID
    let fileIds: string[] = []
    const actualTagIds = tag_ids?.filter((id) => id >= 0) || []

    // 从不同的知识库类型中提取文件ID
    if (type === 'single-file' && defaultKnowBase) {
      // 单文件问答：使用当前文件ID
      fileIds = [String(defaultKnowBase)]
    } else if (mode === 'knowledge') {
      if (actualTagIds.length > 0) {
        // 如果选择了标签，获取标签对应的文件
        try {
          const fileListRes = await getFileList({
            forest_id: 0,
            filters: [{ field: 'tag_ids', value: actualTagIds.map(String) }],
            limit: 9999,
          })
          fileIds = (fileListRes.data || []).map((file: any) => String(file.ID))
        } catch (error) {
          console.log('获取标签文件失败', error)
          return
        }
      } else if (knowledge?.length) {
        // 使用选中的知识库/文件
        fileIds = knowledge.map((item) => String(item.id))
      }
    }

    if (fileIds.length === 0) {
      message.warning('请先选择知识资源')
      return
    }

    const originalSearch = search
    setIsPolishing(true)

    try {
      const res = await expansionQuestion({
        file_ids: fileIds.map((item) => Number(item)),
        question: originalSearch,
        session_id: currentSessionId || 0,
      })
      if (res.expanded_question) {
        setSearch(res.expanded_question)
      } else {
        setSearch(originalSearch)
      }
    } catch (error) {
      console.log('AI润色失败', error)
      // 恢复原始内容
      setSearch(originalSearch)
    } finally {
      setIsPolishing(false)
    }
  }

  const startQA = () => {
    if (charts.length >= 100) {
      message.warning(t('project.graph.tooManyCharts'))
      return
    }
    if (!allowAsk) return
    onAsk?.()

    // 构造请求参数
    const qaData: any = { content: search! }

    // 添加联网搜索选项
    if (enableWebSearch) {
      qaData.options = {
        enable_web_search: true,
      }
    }

    // 添加附件
    if (attachments.length > 0) {
      qaData.input = {
        attachments: attachments.map((att) => ({
          id: att.id,
          url: att.url, // 用于前端展示
          ...(att.md_url ? { md_url: att.md_url } : {}), // 用于后端处理，仅在存在时传递
          type: att.type,
          name: att.name,
          ...(att.mime_type ? { mime_type: att.mime_type } : {}),
        })),
      }
    }

    _startQA(qaData)
    setSearch('')
    // 发送后清空附件
    setAttachments([])
  }

  useImperativeHandle(ref, () => {
    return {
      startQA,
    }
  })

  const sendBtn = (() => {
    if (dialogStatus === 'answering') {
      return <StopBtn onClick={stopQA} className='cursor-pointer' />
    }
    if (!allowAsk) {
      return (
        <DisabledBtn
          onClick={() => {
            if (!dataSelected) {
              const warningMessage = match(mode)
                .with('table', () => '请先选择关联表格')
                .with('database', () => '请先选择数据表')
                .with('external_data', () => '请先选择数据源')
                .otherwise(() => '请先选择知识资源')
              message.warning(warningMessage)
              return
            }
            if (!search?.trim()) {
              message.warning('请先输入您的问题')
              return
            }
          }}
        />
      )
    }
    return <ActiveBtn onClick={startQA} className='cursor-pointer' />
  })()

  // 渲染大模型模式下的左侧功能按钮（上传和联网搜索）
  const renderLeftActions = () => {
    if (mode !== 'model') return null

    const isCustomDeploy = deployVersion === 'custom'

    return (
      <div className='flex items-center gap-2'>
        <div
          className={cn(
            'cursor-pointer border border-[#eff1f4] rounded-full bg-[#f7f7f7]',
            'text-[13px] text-[#6e757f]',
            'py-1 px-3 flex items-center gap-1 font-[500]',
            'transition-colors hover:bg-[#FBE9FF] hover:text-[#CC5DE8] hover:border-[#CC5DE833]',
          )}
          onClick={handleUpload}
        >
          {UploadIcon && (
            <UploadIcon
              className='w-[15.368px] h-[15.368px]'
              style={{ color: 'currentColor' }}
            />
          )}
          <span>上传文件</span>
        </div>
        {!isCustomDeploy && (
          <div
            className={cn(
              'cursor-pointer border rounded-full',
              'text-[13px]',
              'py-1 px-3 flex items-center gap-1 font-[500]',
              'transition-colors',
              enableWebSearch
                ? 'bg-[#FBE9FF] text-[#CC5DE8] border-[#CC5DE833]'
                : 'bg-[#f7f7f7] text-[#6e757f] border-[#eff1f4] hover:bg-[#FBE9FF] hover:text-[#CC5DE8] hover:border-[#CC5DE833]',
            )}
            onClick={() => {
              const newValue = !enableWebSearch
              setEnableWebSearch(newValue)
              // 同步更新到 sessionConfig
              setSessionConfig({ enable_web_search: newValue })
            }}
          >
            {NetworkingIcon && (
              <NetworkingIcon
                className='w-[15.368px] h-[15.368px]'
                style={{ color: 'currentColor' }}
              />
            )}
            <span>联网搜索</span>
          </div>
        )}
      </div>
    )
  }

  useMount(async () => {
    if (sessionStatus === 'new' && !hasCalledLoadFn) {
      loadData()
    }
  })

  // 获取可以进行图谱问答的知识库（仅 graph_status 为 success 的）
  const { data: graphForests } = useRequest(async () => {
    const res = await getKnowledgeBaseList({ offset: 0, limit: 9999 })
    const data: any[] = res.Data ?? []
    return data.filter((item) => item.graph_status === 'success')
  })

  // 获取节点的知识库forest_id
  const getForestId = (item: any) => item.forest_id

  // 检查图谱模式下是否有可用的知识库
  const hasGraphKnowledge = useMemo(() => {
    if (mode !== 'graph') return true
    // 如果数据还在加载中，返回true（避免在加载时显示提示弹窗）
    if (!allForestData || !graphForests) return true
    const graphForestIds = graphForests.map((item) => item.ID)
    const availableKnowledge = allForestData?.filter((item) =>
      graphForestIds.includes(getForestId(item)),
    )
    return availableKnowledge && availableKnowledge.length > 0
  }, [mode, allForestData, graphForests])

  // 提示弹窗状态
  const [graphEmptyModalOpen, setGraphEmptyModalOpen] = useState(false)

  const modeItems = useMemo(() => {
    const items: {
      icon: ReactNode
      text: string
      mode: SessionConfig['mode']
    }[] = [
      { icon: <ModalIcon />, text: '大模型', mode: 'model' },
      { icon: <KnowledgeIcon />, text: '文档模式', mode: 'knowledge' },
      { icon: <TableIcon />, text: '表格模式', mode: 'table' },
      { icon: <DatabaseIcon />, text: '数据库模式', mode: 'database' },
    ]
    // 根据权限决定是否显示图谱洞察模式
    if (hasGraphPermission) {
      items.push({ icon: <GraphIcon />, text: '图谱洞察', mode: 'graph' })
      // 本地环境、测试环境或 custom+cimc 模式下显示图搜模式
      const showGraphSearch =
        globalConfig.env === 'development' ||
        globalConfig.apiEnv === 'test' ||
        (deployVersion === 'custom' && deployMode === 'cimc')
      // if (showGraphSearch) {
      //   items.push({
      //     icon: <GraphSearchIcon />,
      //     text: '图搜模式',
      //     mode: 'graph_search',
      //   })
      // }
    }
    if (isH3CTest) {
      items.push({ icon: <H3C />, text: '芯模Pilot', mode: 'h3c-test' })
    }
    // 外接数据模式：仅测试环境、本地环境、custom+cimc 环境显示（使用 import.meta.env.MODE 判定）
    const showExternalData =
      import.meta.env.MODE === 'development' ||
      import.meta.env.MODE === 'test' ||
      (deployVersion === 'custom' && deployMode === 'cimc')
    if (showExternalData) {
      items.push({
        icon: <ExternalDataIcon />,
        text: '外接数据',
        mode: 'external_data',
      })
    }

    return items
  }, [isH3CTest, hasGraphPermission, deployVersion, deployMode])
  const placeholder = useMemo(() => {
    switch (mode) {
      case 'model':
        return '输入内容并发送（或按 Enter），即可生成回答'
      case 'graph_search':
        return '输入内容并发送（或按 Enter），即可生成回答'
      case 'knowledge':
        return '输入内容并发送（或按 Enter），将自动检索相关知识资源并生成回答'
      case 'database':
        return '输入内容并发送（或按 Enter），将自动检索相关知识资源并生成回答'
      case 'graph':
        return '输入内容并发送（或按 Enter），将检索知识库并高亮图谱路径，提升问答准确性与可追溯性。'
      case 'table':
        return '输入内容并发送（或按 Enter），将自动检索相关表格资源并生成回答'
      case 'external_data':
        return '输入内容并发送（或按 Enter），将自动检索外部数据源并生成回答'
      case 'h3c-test':
        return '请输入您的问题，按Enter发送，按Ctrl+Enter'
    }
  }, [mode])

  return (
    <div className={cn(className, 'flex flex-col')} style={style}>
      <DialogInput
        className={cn(
          'border border-[rgb(230,178,243)] rounded-[20px] shadow-[0_0_10px_rgba(0,0,0,0.1)] focus-within:shadow-[0_0_20px_rgba(0,0,0,0.15)]',
          {
            'mr-3': showRightMargin,
          },
        )}
        value={search}
        onChange={setSearch}
        onEnter={startQA}
        // 只有问答类入口按环境放宽上限，普通项目内会话继续保持原有 500。
        maxLength={shouldUseQaInputMaxLength ? qaInputMaxLength : 500}
        placeholder={placeholder}
        attachments={attachments}
        onRemoveAttachment={handleRemoveAttachment}
        helperText={
          isPolishing ? '正在使用知识库润色问题，请稍后...' : undefined
        }
      >
        {/* 首页状态：模型选择在右边（与发送按钮一起） */}
        {sessionStatus === 'new' ? (
          <>
            {/* 大模型模式：显示上传文件和联网搜索按钮 */}
            {renderLeftActions()}
            {/* 知识库问答数据源 */}
            {/* 非新建页面 这两个数据源只展示一个 */}
            {/* {(version !== 'custom' && mode === 'knowledge') ||
            externalIds?.length ? (
              <ExternalDataSourceSelect
                list={externalDataSourceList}
                checkedList={externalIds}
                onChange={(externalIds) =>
                  setSessionConfig({ externalIds, knowledge: undefined })
                }
                allowSelect={sessionStatus === 'new'}
                disabled={mode === 'knowledge' && sessionStatus === 'new'}
                className='font-medium'
              />
            ) : null} */}
            {/* 单文件问答数据源 - 第一次提问之前可以修改 */}
            {mode === 'knowledge' &&
            type === 'single-file' &&
            isFirstQuestion ? (
              <>
                <Popover
                  arrow={false}
                  placement='topLeft'
                  trigger='click'
                  content={
                    <KnowledgeSelect className='max-h-[40vh] min-w-100' />
                  }
                >
                  <KnowledgeStatus
                    active={true}
                    title={selectedKnowledgeText || '已选择当前文件'}
                  />
                </Popover>
                <PromptModeSelect
                  value={prompt_key || 'normal'}
                  onChange={(val) => setSessionConfig({ prompt_key: val })}
                  className={cn(
                    'bg-[#FBE9FF] text-[#CC5DE8] border-[#CC5DE833]',
                    'border rounded-full py-1 px-3 text-[13px] font-[500]',
                  )}
                  showArrow={true}
                />
                {/* AI润色按钮 */}
                {showPolishButton && (
                  <AIPolishButton
                    disabled={!allowPolish}
                    loading={isPolishing}
                    onClick={handlePolish}
                  />
                )}
              </>
            ) : null}
            <Popover
              arrow={false}
              placement='topLeft'
              trigger='click'
              content={<KnowledgeSelect className='max-h-[40vh] min-w-100' />}
            >
              {mode === 'knowledge' && type !== 'single-file' ? (
                <KnowledgeStatus
                  active={Boolean(knowledge?.length || tag_ids?.length)}
                  title={selectedKnowledgeText}
                  defaultTitle={'选择资源'}
                />
              ) : null}
            </Popover>

            {/* 文档模式时显示问答模式选择器（排除单文件，单文件已在上面单独处理） */}
            {mode === 'knowledge' && type !== 'single-file' ? (
              <>
                <PromptModeSelect
                  value={prompt_key || 'normal'}
                  onChange={(val) => setSessionConfig({ prompt_key: val })}
                  className={cn(
                    'bg-[#FBE9FF] text-[#CC5DE8] border-[#CC5DE833]',
                    'border rounded-full py-1 px-3 text-[13px] font-[500]',
                  )}
                  showArrow={true}
                />
                {/* AI润色按钮 */}
                {showPolishButton && (
                  <AIPolishButton
                    disabled={!allowPolish}
                    loading={isPolishing}
                    onClick={handlePolish}
                  />
                )}
              </>
            ) : null}

            {/* 图谱问答数据源 */}
            <Popover
              arrow={false}
              placement='topLeft'
              trigger='click'
              open={hasGraphKnowledge ? undefined : false}
              onOpenChange={(open) => {
                // 如果知识库为空且尝试打开Popover，则显示提示弹窗
                if (open && !hasGraphKnowledge) {
                  setGraphEmptyModalOpen(true)
                }
              }}
              content={
                <KnowledgeSelect className='max-h-[40vh] min-w-100' graph />
              }
            >
              {mode === 'graph' && type !== 'single-file' ? (
                <KnowledgeStatus
                  active={Boolean(graphKnowledgeBase?.length)}
                  title={'已选资源'}
                  defaultTitle={'选择资源'}
                />
              ) : null}
            </Popover>

            {/* 表格问答数据源 */}
            {mode === 'table' && (
              <Popover
                arrow={false}
                placement='topLeft'
                trigger='click'
                content={
                  <KnowledgeSelect className='max-h-[40vh] min-w-100' table />
                }
              >
                <KnowledgeStatus
                  active={Boolean(tableKnowledgeBase?.length)}
                  title={selectedTableText}
                  defaultTitle={'选择表格'}
                />
              </Popover>
            )}
            {mode === 'database' && (
              <Popover
                arrow={false}
                placement='topLeft'
                trigger='click'
                content={
                  <KnowledgeSelect
                    className='max-h-[40vh] min-w-100'
                    database
                  />
                }
              >
                <KnowledgeStatus
                  active={Boolean(databaseKnowledgeBase?.length)}
                  title={selectedDatabaseStatus}
                  defaultTitle={'选择数据表'}
                />
              </Popover>
            )}
            {/* 外接数据模式数据源选择器 */}
            {mode === 'external_data' && (
              <Popover
                arrow={false}
                placement='topLeft'
                trigger='click'
                content={
                  <ExternalDataProviderSelect
                    className='max-h-[40vh]'
                    value={externalDataProviders}
                    onChange={(providers) =>
                      setSessionConfig({ externalDataProviders: providers })
                    }
                  />
                }
              >
                <KnowledgeStatus
                  active={Boolean(externalDataProviders?.length)}
                  title={selectedExternalDataStatus}
                  defaultTitle={'选择数据源'}
                />
              </Popover>
            )}
            {/* 右侧：模型选择 + 发送按钮 */}
            <div className='ml-auto flex items-center gap-3'>
              <ModelSelect
                models={models}
                allowSelect={sessionStatus === 'new'}
                value={model_id}
                onChange={(model_id) => setSessionConfig({ model_id })}
                className={cn('font-medium', { hidden: mode === 'h3c-test' })}
              />
              {sendBtn}
            </div>
          </>
        ) : (
          <>
            {/* 问答页面 */}
            {mode === 'model' || mode === 'graph_search' ? (
              <>
                {/* 大模型问答：左侧显示上传文件和联网搜索按钮（graph_search 模式不显示） */}
                {mode === 'model' && renderLeftActions()}
                {/* 大模型问答：模型选择与发送按钮在右侧 */}
                <div className='ml-auto flex items-center gap-3'>
                  <ModelSelect
                    models={models}
                    allowSelect={false}
                    showArrow={false}
                    value={model_id}
                    onChange={(model_id) => setSessionConfig({ model_id })}
                    className={cn(
                      'bg-[#fbe9ff] border-[1.5px] border-[#CC5DE833] rounded-[20px] h-[24px]',
                      'px-[10px] py-[5px] text-[12px] text-[#cc5de8] opacity-[0.5]',
                      'font-medium',
                    )}
                  />
                  {sendBtn}
                </div>
              </>
            ) : (
              <>
                {/* 其他模式：左侧 模型选择 + 已选资源/已选表格 */}
                <div className='flex items-center gap-[5px]'>
                  <ModelSelect
                    models={models}
                    allowSelect={false}
                    showArrow={false}
                    value={model_id}
                    onChange={(model_id) => setSessionConfig({ model_id })}
                    className={cn(
                      'bg-[#fbe9ff] border-[1.5px] border-[#CC5DE833] rounded-[20px] h-[24px]',
                      'px-[10px] py-[5px] text-[12px] text-[#cc5de8] opacity-[0.5]',
                      'font-medium',
                      { hidden: mode === 'h3c-test' },
                    )}
                  />
                  {/* 文档模式时显示问答模式选择器 */}
                  {mode === 'knowledge' ? (
                    <PromptModeSelect
                      value={prompt_key || 'normal'}
                      onChange={(val) => setSessionConfig({ prompt_key: val })}
                      allowSelect={false}
                      showArrow={false}
                      className={cn(
                        'bg-[#fbe9ff] border-[1.5px] border-[#CC5DE833] rounded-full h-[24px]',
                        'px-[10px] py-[5px] text-[12px] text-[#cc5de8] opacity-[0.5]',
                        'font-[500]',
                      )}
                    />
                  ) : null}
                  {/* 知识库问答数据源 */}
                  {/* {externalIds?.length ? (
                    <ExternalDataSourceSelect
                      list={externalDataSourceList}
                      checkedList={externalIds}
                      onChange={(externalIds) =>
                        setSessionConfig({ externalIds, knowledge: undefined })
                      }
                      allowSelect={false}
                      className='font-medium'
                    />
                  ) : null} */}
                  {/* 单文件问答页面 */}
                  {type === 'single-file' && mode === 'knowledge' ? (
                    isFirstQuestion ? (
                      // 第一次提问之前：可以修改，无透明度
                      <Popover
                        arrow={false}
                        placement='topLeft'
                        trigger='click'
                        content={
                          <KnowledgeSelect className='max-h-[40vh] min-w-100' />
                        }
                      >
                        <KnowledgeStatus
                          active={true}
                          className={cn('px-[10px] py-[5px] h-[24px]')}
                          title={selectedKnowledgeText || '已选择当前文件'}
                        />
                      </Popover>
                    ) : (
                      // 第一次提问之后：不能修改，有透明度
                      <KnowledgeStatus
                        className={cn(
                          'bg-[#fbe9ff] border-[1.5px] border-[#CC5DE833] rounded-[20px] h-[24px]',
                          'px-[10px] py-[5px] text-[12px] text-[#cc5de8] opacity-[0.5]',
                          'font-medium',
                        )}
                        active={true}
                        title={selectedKnowledgeText || '已选择当前文件'}
                      />
                    )
                  ) : (
                    <Popover
                      arrow={false}
                      placement='topLeft'
                      trigger='hover'
                      content={
                        <KnowledgeList
                          items={knowledge}
                          title={selectedKnowledgeText}
                        />
                      }
                    >
                      {knowledge?.length ? (
                        <KnowledgeStatus
                          className={cn(
                            'bg-[#fbe9ff] border-[1.5px] border-[#CC5DE833] rounded-[20px] h-[24px]',
                            'px-[10px] py-[5px] text-[12px] text-[#cc5de8] opacity-[0.5]',
                            'font-medium',
                          )}
                          active={Boolean(knowledge?.length)}
                          title={selectedKnowledgeText}
                        />
                      ) : null}
                    </Popover>
                  )}

                  {/* 图谱问答数据源 */}
                  <Popover
                    arrow={false}
                    placement='topLeft'
                    trigger='hover'
                    content={
                      <KnowledgeList
                        items={graphKnowledgeBase}
                        title={'已选资源'}
                        tooltipText='最多可展示200篇已选资源'
                      />
                    }
                  >
                    {graphKnowledgeBase?.length ? (
                      <KnowledgeStatus
                        className={cn(
                          'bg-[#fbe9ff] border-[1.5px] border-[#CC5DE833] rounded-[20px] h-[24px]',
                          'px-[10px] py-[5px] text-[12px] text-[#cc5de8] opacity-[0.5]',
                          'font-medium',
                        )}
                        active={Boolean(graphKnowledgeBase?.length)}
                        title={'已选资源'}
                      />
                    ) : null}
                  </Popover>

                  {/* 表格问答数据源 */}
                  {mode === 'table' && (
                    <Popover
                      arrow={false}
                      placement='topLeft'
                      trigger='hover'
                      content={
                        <KnowledgeList
                          items={tableKnowledgeBase}
                          title={'已选表格'}
                          tooltipText='最多可展示200个已选表格'
                        />
                      }
                    >
                      {tableKnowledgeBase?.length ? (
                        <KnowledgeStatus
                          className={cn(
                            'bg-[#fbe9ff] border-[1.5px] border-[#CC5DE833] rounded-[20px] h-[24px]',
                            'px-[10px] py-[5px] text-[12px] text-[#cc5de8] opacity-[0.5]',
                            'font-medium',
                          )}
                          active={Boolean(tableKnowledgeBase?.length)}
                          title={selectedTableText}
                          defaultTitle={'表格知识库'}
                        />
                      ) : null}
                    </Popover>
                  )}
                  {mode === 'database' && (
                    <Popover
                      arrow={false}
                      placement='topLeft'
                      trigger='hover'
                      content={
                        <KnowledgeList
                          items={databaseKnowledgeBase}
                          title={'已选数据表'}
                          tooltipText='最多可展示200个已选数据表'
                        />
                      }
                    >
                      {databaseKnowledgeBase?.length ? (
                        <KnowledgeStatus
                          className={cn(
                            'bg-[#fbe9ff] border-[1.5px] border-[#CC5DE833] rounded-[20px] h-[24px]',
                            'px-[10px] py-[5px] text-[12px] text-[#cc5de8] opacity-[0.5]',
                            'font-medium',
                          )}
                          active={Boolean(databaseKnowledgeBase?.length)}
                          title={'已选数据表'}
                        />
                      ) : null}
                    </Popover>
                  )}
                  {/* 外接数据模式：始终显示已选数据源状态（接口调整中，暂不展示详情，点击/hover 无反应） */}
                  {mode === 'external_data' ? (
                    <KnowledgeStatus
                      className={cn(
                        'bg-[#fbe9ff] border-[1.5px] border-[#CC5DE833] rounded-[20px] h-[24px]',
                        'px-[10px] py-[5px] text-[12px] text-[#cc5de8] opacity-[0.5]',
                        'font-medium',
                        'cursor-default pointer-events-none',
                      )}
                      active={true}
                      title={'已选数据源'}
                    />
                  ) : null}
                  {/* AI润色按钮 - 放在所有左侧功能按钮的最后 */}
                  {showPolishButton && (
                    <AIPolishButton
                      disabled={!allowPolish}
                      loading={isPolishing}
                      onClick={handlePolish}
                    />
                  )}
                </div>
                {/* 右侧：发送按钮 */}
                <div className='ml-auto flex items-center'>{sendBtn}</div>
              </>
            )}
          </>
        )}
      </DialogInput>
      <div
        className={cn('mt-12 flex gap-10 mx-auto', {
          hidden: sessionStatus !== 'new' || type === 'single-file',
        })}
      >
        {modeItems.map((item) => {
          return (
            <QAMode
              key={item.mode}
              text={item.text}
              icon={item.icon}
              active={mode === item.mode}
              onClick={() =>
                setSessionConfig({
                  mode: item.mode,
                })
              }
            />
          )
        })}
      </div>
      {/* 图谱模式下知识库为空的提示弹窗 */}
      <GraphEmptyModal
        open={graphEmptyModalOpen}
        onCancel={() => setGraphEmptyModalOpen(false)}
      />
    </div>
  )
})

const QAMode: FC<{
  onClick?: () => void
  icon?: ReactNode
  active?: boolean
  text?: string
}> = (props) => {
  const { onClick, icon, active, text } = props
  return (
    <div
      className={cn(
        'text-[#6E757F] cursor-pointer',
        'flex flex-col gap-2 items-center',
        {
          'text-[#CC5DE8]': active,
        },
      )}
      onClick={onClick}
    >
      <div
        className={cn('p-3 border-2 border-[#EFF1F4] bg-[#F7F7F7] rounded-xl', {
          'bg-[#FBE9FF] border-[#CC5DE833]': active,
        })}
      >
        {icon}
      </div>
      {text}
    </div>
  )
}
