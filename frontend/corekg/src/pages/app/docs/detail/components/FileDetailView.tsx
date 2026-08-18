import React, {
  useState,
  useMemo,
  useEffect,
  createContext,
  useContext,
  useRef,
} from 'react'
import {
  useParams,
  useNavigate,
  useSearchParams,
  useLocation,
} from 'react-router-dom'
import { Button, Switch, message, Breadcrumb, Tooltip, Spin, Input } from 'antd'
import { LoadingOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import {
  getFileSegments,
  updateFileSegment,
  deleteFileSegment,
  deleteFile,
  modifyFileSegmentRule,
  getFileInfo,
  getKnowledgeBaseDetail,
  disableFileChunk,
} from '@/api/knowledge'
import ChangeRuleIcon from '@/assets/icons/docs/change-rules.svg'
import deleteIcon from '@/assets/icons/docs/delete.svg'
import editIcon from '@/assets/icons/docs/edit.svg'
import KnowledgeBaseIcon from '@/assets/icons/docs/knowledge-base-icon.svg?react'
import NavigationIcon from '@/assets/icons/docs/navigation.svg?react'
import '@/styles/markdown.css'
import { useFileLocation } from '../../utils'
import styles from '../styles.module.scss'
import DeleteConfirmModal from './DeleteConfirmModal'
import {
  ModifySegmentRuleModal,
  ModifySegmentRule,
} from './FileDetailView/ModifySegmentRuleModal'
import SegmentEditModal from './FileEditView/SegmentEditModal'
import FilePreview from './FilePreview'
import RightPanel from './RightPanel'

const NAVIGATION_GESTURE_THRESHOLD = 30

interface TabItem {
  key: string
  label: string
}

const findScrollableParent = (
  element: HTMLElement | null,
): HTMLElement | null => {
  let current = element
  while (current) {
    const { overflowX, overflow } = window.getComputedStyle(current)
    const canScrollX =
      overflowX === 'auto' ||
      overflowX === 'scroll' ||
      overflow === 'auto' ||
      overflow === 'scroll'

    if (canScrollX && current.scrollWidth > current.clientWidth) {
      return current
    }
    current = current.parentElement
  }
  return null
}

// 分段数据接口
interface FileSegment {
  id: string
  content: string
  table?: string
  type?: 'chunk' | 'table' | 'image'
  chunk_number: number
  charCount: number
  location?: number[] // 用于定位文件位置，如PDF页码
  imageUrl?: string
}

// 文件信息接口
interface FileInfo {
  id: number
  name: string
  knowledgeBaseName: string
}
const FileDetailViewContext = createContext<any | null>(null)

// eslint-disable-next-line react-refresh/only-export-components
export function useFileDetailViewProject<T = any>() {
  const project = useContext(FileDetailViewContext)
  if (project) return project as T
  return null
}

const FileDetailView: React.FC = () => {
  const { id, fileId } = useParams()
  const navigate = useNavigate()
  const { t } = useTranslation('pages')
  const fileLocation = useFileLocation()

  // 监听 fileId 变化，切换文档时清除旧文档的 sessionStorage
  const prevFileIdRef = useRef<string>()
  // 用于跟踪location参数是否已处理，避免重复处理导致tab自动切换
  const processedLocationRef = useRef<string>()
  
  useEffect(() => {
    const prevFileId = prevFileIdRef.current
    if (prevFileId && prevFileId !== fileId) {
      // fileId 变化了，说明切换到了其他文档，清除旧文档的 session 记录
      const storageKey = `file_session_${prevFileId}`
      sessionStorage.removeItem(storageKey)
      // 切换文档时重置location处理标记
      processedLocationRef.current = undefined
    }
    prevFileIdRef.current = fileId

    // 组件卸载时清除当前文档的 session 记录
    return () => {
      if (fileId) {
        const storageKey = `file_session_${fileId}`
        sessionStorage.removeItem(storageKey)
      }
    }
  }, [fileId])

  const location = useLocation()
  // 从URL查询参数中获取当前选中的tab，如果没有则默认为'document-analysis'
  const getActiveTabFromUrl = (): string => {
    try {
      const searchParams = new URLSearchParams(location.search)
      const tabFromUrl = searchParams.get('tab')
      return tabFromUrl &&
        [
          'document-analysis',
          'intelligent-analysis',
          'segmentRule',
          'wisdom-qa',
        ].includes(tabFromUrl)
        ? tabFromUrl
        : 'wisdom-qa'
    } catch (error) {
      console.error('解析tab参数失败', error)
      return 'wisdom-qa'
    }
  }

  const [activeTab, setActiveTab] = useState<string>(getActiveTabFromUrl())
  // 监听URL变化，同步activeTab状态
  useEffect(() => {
    const newActiveTab = getActiveTabFromUrl()
    setActiveTab(newActiveTab)
  }, [location.search])
  const handleTabChange = (key: string) => {
    // 更新URL查询参数
    const searchParams = new URLSearchParams(location.search)
    searchParams.set('tab', key)

    // 使用replace而不是push，避免产生过多的历史记录
    navigate(
      {
        pathname: location.pathname,
        search: searchParams.toString(),
      },
      { replace: true },
    )

    // 本地状态会通过useEffect自动更新
  }

  // 获取路由参数中的动态数据
  const [searchParams] = useSearchParams()
  const kbNameFromUrl = searchParams.get('kbName') || ''
  const fileNameFromUrl = searchParams.get('fileName') || ''
  const folderPathFromUrl = searchParams.get('folderPath') || '[]'

  // 状态管理
  const [isPreviewMode, setIsPreviewMode] = useState(
    () => true,
    // Boolean(fileLocation),
  ) // 默认不预览
  const [isQAEnabled, setIsQAEnabled] = useState(false) // 启用问答开关
  const [isQADisabled, setIsQADisabled] = useState(false) // 问答开关是否禁用

  // 删除确认弹窗状态
  const [deleteModal, setDeleteModal] = useState({
    visible: false,
    type: 'segment' as 'segment' | 'file',
    targetId: '',
  })

  // 编辑弹窗状态
  const [editModal, setEditModal] = useState({
    visible: false,
    segmentId: '',
    segmentTitle: '',
    content: '',
    imageUrl: '',
    segmentType: 'chunk' as 'chunk' | 'table' | 'image',
  })

  // 配置弹窗状态
  const [configModal, setConfigModal] = useState({
    visible: false,
  })

  // 分段数据状态管理
  const [segments, setSegments] = useState<FileSegment[]>([])
  const [activeChunkId, setActiveChunkId] = useState<string | null>(null)
  const [segmentTotal, setSegmentTotal] = useState(0)

  const formatSegments = (chunks: any[]): FileSegment[] =>
    chunks.map((chunk: any) => {
      const segmentType = chunk._source.type || 'chunk'
      const description = chunk._source.description || ''
      const table = chunk._source.table || ''
      const imageUrl =
        segmentType === 'image' ? chunk._source.image_url || '' : ''

      return {
        id: chunk._id,
        type: segmentType,
        chunk_number: chunk._source.sequence || 0,
        content: description,
        table,
        charCount: chunk._source.chunk_size || chunk._source.tokens,
        location: chunk._source.location || undefined,
        imageUrl: imageUrl || undefined,
      }
    })

  // 监听 URL 中的 location 参数，实现自动选中和标签切换
  // 只在location参数真正变化时处理，避免切换tab时重复触发
  useEffect(() => {
    const locationParam = searchParams.get('location')
    if (!locationParam || !segments.length) return

    // 如果这个location参数已经处理过了，跳过
    if (processedLocationRef.current === locationParam) {
      return
    }

    try {
      const targetLocation = JSON.parse(decodeURIComponent(locationParam))
      // 匹配分块：通常 location 的前几个数值（如页码、坐标）能唯一标识一个分块
      const matchedSegment = segments.find((s) => {
        if (!s.location || !Array.isArray(s.location)) return false
        // 比较关键坐标：页码, x1, y1
        return (
          Number(s.location[0]) === Number(targetLocation[0]) &&
          Number(s.location[1]) === Number(targetLocation[1]) &&
          Number(s.location[2]) === Number(targetLocation[2])
        )
      })

      if (matchedSegment) {
        // 1. 设置激活的分块 ID
        setActiveChunkId(matchedSegment.id)
        // 2. 自动切换到分段规则标签页
        if (activeTab !== 'segmentRule') {
          handleTabChange('segmentRule')
        }
        // 3. 标记这个location参数已处理，避免重复处理
        processedLocationRef.current = locationParam
      }
    } catch (e) {
      console.error('解析 location 参数或匹配分块失败:', e)
    }
  }, [searchParams, segments]) // 仅在参数或数据变化时执行

  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dataSourceType, setDataSourceType] = useState<string>('')

  // 文件配置状态管理
  const [fileConfig, setFileConfig] = useState<{
    split_config: any | null
  } | null>(null)

  // 知识库名称状态
  const [knowledgeBaseName, setKnowledgeBaseName] = useState<string>(
    kbNameFromUrl || '知识库',
  )
  const [fileInfo, setFileInfo] = useState<FileInfo>({
    id: Number(fileId) || 1,
    name: fileNameFromUrl || '未知文件',
    knowledgeBaseName: knowledgeBaseName,
  })

  const isTableKnowledgeBase = dataSourceType === 'excel'

  useEffect(() => {
    if (!isTableKnowledgeBase) return

    const styleId = 'file-detail-disable-navigation-styles'
    const style = document.createElement('style')
    style.id = styleId
    style.textContent = `
      html, body, .excel-preview-container {
        overscroll-behavior-x: contain !important;
      }
    `
    document.head.appendChild(style)

    const originalStyles = {
      body: document.body.style.overscrollBehaviorX,
      html: document.documentElement.style.overscrollBehaviorX,
    }

    document.body.style.overscrollBehaviorX = 'contain'
    document.documentElement.style.overscrollBehaviorX = 'contain'

    const handleWheel = (event: WheelEvent) => {
      const isHorizontalSwipe = Math.abs(event.deltaX) > Math.abs(event.deltaY)
      const isNavigationGesture =
        Math.abs(event.deltaX) > NAVIGATION_GESTURE_THRESHOLD

      if (isHorizontalSwipe && isNavigationGesture) {
        event.preventDefault()
        event.stopPropagation()

        const scrollable = findScrollableParent(event.target as HTMLElement)
        if (scrollable) {
          scrollable.scrollLeft += event.deltaX
        }
        return false
      }
      return true
    }

    const handleKeydown = (event: KeyboardEvent) => {
      const isNavigationShortcut =
        (event.altKey &&
          (event.key === 'ArrowLeft' || event.key === 'ArrowRight')) ||
        ((event.metaKey || event.ctrlKey) &&
          (event.key === '[' || event.key === ']'))

      if (isNavigationShortcut) {
        event.preventDefault()
        return false
      }
      return true
    }

    const wheelOptions: AddEventListenerOptions = {
      passive: false,
      capture: true,
    }
    document.addEventListener('wheel', handleWheel, wheelOptions)
    window.addEventListener('wheel', handleWheel, wheelOptions)
    document.addEventListener('keydown', handleKeydown, { capture: true })

    return () => {
      document.getElementById(styleId)?.remove()
      document.body.style.overscrollBehaviorX = originalStyles.body
      document.documentElement.style.overscrollBehaviorX = originalStyles.html
      document.removeEventListener('wheel', handleWheel, true)
      window.removeEventListener('wheel', handleWheel, true)
      document.removeEventListener('keydown', handleKeydown, true)
    }
  }, [isTableKnowledgeBase])

  // 获取文件信息和分段数据
  useEffect(() => {
    const fetchData = async () => {
      if (!fileId) return

      setLoading(true)
      setError(null)

      try {
        // 同时获取知识库信息、文件信息和分段数据
        const [kbInfoRes, fileInfoRes, segmentsRes] = await Promise.all([
          getKnowledgeBaseDetail({ id: Number(id) }),
          getFileInfo({
            file_id: Number(fileId),
          }),
          getFileSegments({
            file_id: Number(fileId),
            forest_id: Number(id),
          }),
        ])
        const kbData = kbInfoRes?.data
        const kbType = kbData?.data_source_type || ''
        const isExcelType = kbType === 'excel'

        setFileInfo({
          id: Number(id),
          name: fileInfoRes.name,
          knowledgeBaseName: kbData.name,
        })
        // 处理知识库信息
        if (kbData) {
          setKnowledgeBaseName(kbData.name)
        }
        setDataSourceType(kbType)

        // 处理文件配置信息
        if (fileInfoRes) {
          // 兼容 split_config 为 null 的情况，默认为 auto 模式
          const splitConfig = fileInfoRes.file_config?.split_config
          const defaultConfig = {
            split_mode: 'auto',
            chunk_size: 256,
            split_mark: ['\n'],
            split_overlap: 0.25,
            preprocessing_rules: {
              remove_empty_line: false,
              remove_url: false,
              remove_email: false,
            },
          }

          setFileConfig({
            split_config: splitConfig || defaultConfig,
          })
        }

        // 处理分段数据
        if (isExcelType) {
          setSegments([])
          setSegmentTotal(0)
          setIsQADisabled(true)
          setIsQAEnabled(false)
        } else if (segmentsRes) {
          const chunks = segmentsRes.chunks || []
          const formattedSegments = formatSegments(chunks)

          setSegments(formattedSegments)
          setSegmentTotal(formattedSegments.length)

          // 根据chunks数据设置启用问答状态
          if (chunks.length === 0) {
            // 如果chunks为空，禁用滑块
            setIsQADisabled(true)
            setIsQAEnabled(false)
          } else {
            // 根据第一个chunk的is_disable属性设置状态
            setIsQADisabled(false)
            const firstChunk = chunks[0]
            // is_disable为false表示启用，true表示禁用
            setIsQAEnabled(firstChunk._source?.is_disable)
          }
        } else {
          setError(t('app.docs.fileDetail.loadSegmentsFailed'))
          setIsQADisabled(true)
          setIsQAEnabled(false)
        }
      } catch (err) {
        setError(t('app.docs.fileDetail.loadDataFailed'))
        console.error('获取数据失败:', err)
      } finally {
        setLoading(false)
      }
    }

    fetchData()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [fileId, id])

  // 面包屑导航项
  const breadcrumbItems = useMemo(() => {
    const items = [
      {
        title: (
          <span
            className='flex items-center gap-2 text-sm cursor-pointer hover:text-[#0C99FF] font-medium text-[#3C4149]'
            onClick={() => navigate('/docs')}
          >
            <KnowledgeBaseIcon className='w-4 h-4' />
            <span>{t('app.docs.fileDetail.knowledgeBase')}</span>
          </span>
        ),
      },
      {
        title: (
          <span
            className='text-sm cursor-pointer hover:text-[#0C99FF] font-medium text-[#3C4149]'
            onClick={() => navigate(`/docs/detail/${id}`)}
          >
            {fileInfo.knowledgeBaseName}
          </span>
        ),
      },
    ]

    // 解析文件夹路径并添加到面包屑中
    try {
      const folderPath = JSON.parse(folderPathFromUrl)
      if (Array.isArray(folderPath) && folderPath.length > 0) {
        folderPath.forEach((pathItem, index) => {
          // 构建到该层级的路径
          const prevPath = folderPath.slice(0, index)
          const pathParam = encodeURIComponent(JSON.stringify(prevPath))

          items.push({
            title: (
              <span
                className='text-sm cursor-pointer hover:text-[#0C99FF] font-medium text-[#3C4149]'
                onClick={() =>
                  navigate(
                    `/docs/detail/${id}/folder/${pathItem.id}?path=${pathParam}`,
                  )
                }
              >
                {pathItem.name}
              </span>
            ),
          })
        })
      }
    } catch (error) {
      console.error('解析文件夹路径失败:', error)
    }

    // 最后添加文件名
    items.push({
      title: (
        <span className='text-sm font-medium text-[#3C4149]'>
          {fileInfo.name}
        </span>
      ),
    })

    return items
  }, [
    navigate,
    id,
    fileInfo.knowledgeBaseName,
    fileInfo.name,
    folderPathFromUrl,
    t,
  ])

  // 处理分段编辑
  const handleSegmentEdit = (segmentId: string) => {
    const segment = segments.find((s) => s.id === segmentId)
    if (segment) {
      const segmentTitle = t('app.docs.fileDetail.segmentTitle', {
        number: segment.chunk_number,
        count: segment.charCount,
      })
      const isTableSegment = segment.type === 'table'
      setEditModal({
        visible: true,
        segmentId: segmentId,
        segmentTitle: segmentTitle,
        content: isTableSegment ? segment.table || '' : segment.content,
        imageUrl: segment.imageUrl || '',
        segmentType: segment.type || 'chunk',
      })
    }
  }

  // 处理分段编辑确认
  const handleSegmentEditConfirm = async (
    segmentId: string,
    newContent: string,
  ) => {
    try {
      // 根据编辑弹窗中的分段类型决定传递哪个字段
      const segmentType = editModal.segmentType
      const updateData: {
        chunk_id: string
        description?: string
        table?: string
        file_id: number
      } = {
        chunk_id: segmentId,
        file_id: Number(fileId),
      }

      // 如果是表格类型，传递 table 字段；否则传递 description 字段
      if (segmentType === 'table') {
        updateData.table = newContent
      } else {
        updateData.description = newContent
      }

      const res = await updateFileSegment(updateData)

      if (res) {
        message.success(t('app.docs.fileDetail.editSuccess'))
        // 重新获取分段数据
        const res = await getFileSegments({
          file_id: Number(fileId),
          forest_id: Number(id),
        })
        if (res) {
          const chunks = res.chunks || []
          const formattedSegments = formatSegments(chunks)
          setSegments(formattedSegments)
          setSegmentTotal(formattedSegments.length)
        }
      } else {
        message.error(t('app.docs.fileDetail.editFailed'))
      }
    } catch (error) {
      message.error(t('app.docs.fileDetail.editFailed'))
      console.error('保存失败:', error)
    }
  }

  // 处理编辑弹窗取消
  const handleEditCancel = () => {
    setEditModal({
      visible: false,
      segmentId: '',
      segmentTitle: '',
      content: '',
      imageUrl: '',
      segmentType: 'chunk',
    })
  }

  // 处理分段删除确认
  const handleSegmentDeleteConfirm = (segmentId: string) => {
    setDeleteModal({
      visible: true,
      type: 'segment',
      targetId: segmentId,
    })
  }

  // 处理删除操作
  const handleDeleteConfirm = async () => {
    try {
      if (deleteModal.type === 'segment') {
        const res = await deleteFileSegment({
          chunk_id: deleteModal.targetId,
          file_id: Number(fileId),
        })
        if (res) {
          message.success(t('app.docs.fileDetail.deleteSuccess'))
          // 重新获取分段数据
          const segmentsRes = await getFileSegments({
            file_id: Number(fileId),
            forest_id: Number(id),
          })
          if (segmentsRes) {
            const chunks = segmentsRes.chunks || []
            const formattedSegments = formatSegments(chunks)
            setSegments(formattedSegments)
            setSegmentTotal(formattedSegments.length)
          }
        } else {
          message.error(t('app.docs.fileDetail.deleteFailed'))
        }
      }
    } catch (error) {
      message.error(t('app.docs.fileDetail.deleteFailed'))
      console.error('删除失败:', error)
    } finally {
      setDeleteModal({ visible: false, type: 'segment', targetId: '' })
    }
  }

  // 将API的split_config转换为弹窗需要的格式
  const convertApiConfigToModalConfig = (
    splitConfig: any,
  ): ModifySegmentRule => {
    if (!splitConfig) {
      return { type: 'default' }
    }

    if (splitConfig.split_mode === 'auto') {
      return { type: 'default' }
    }

    return {
      type: 'custom',
      segmentLength: splitConfig.chunk_size || 256,
      segmentSeparator: splitConfig.split_mark?.[0] || '\n',
      segmentOverlap: Math.round((splitConfig.split_overlap ?? 0.25) * 100),
      textPreprocessing: {
        removeExtraSpaces:
          splitConfig.preprocessing_rules?.remove_empty_line || false,
        removeLineBreaks: splitConfig.preprocessing_rules?.remove_url || false,
        removeSpecialChars:
          splitConfig.preprocessing_rules?.remove_email || false,
      },
    }
  }

  // 处理配置按钮点击
  const handleConfiguration = () => {
    setConfigModal({ visible: true })
  }

  // 处理配置弹窗取消
  const handleConfigCancel = () => {
    setConfigModal({ visible: false })
  }

  // 处理配置确认
  const handleConfigConfirm = async (config: ModifySegmentRule) => {
    try {
      const split_config = {
        split_mode:
          config.type === 'default' ? ('auto' as const) : ('rule' as const),
        chunk_size:
          config.type === 'custom' ? config.segmentLength || 256 : 256,
        split_mark:
          config.type === 'custom' && config.segmentSeparator
            ? [config.segmentSeparator]
            : ['\n'],
        split_overlap:
          config.type === 'custom' ? (config.segmentOverlap ?? 30) / 100 : 0.25,
        preprocessing_rules: {
          remove_empty_line:
            config.type === 'custom'
              ? config.textPreprocessing?.removeExtraSpaces || false
              : false,
          remove_url:
            config.type === 'custom'
              ? config.textPreprocessing?.removeLineBreaks || false
              : false,
          remove_email:
            config.type === 'custom'
              ? (config.textPreprocessing?.removeSpecialChars ?? false)
              : false,
        },
      }

      await modifyFileSegmentRule({
        file_id: Number(fileId),
        forest_id: Number(id),
        split_config,
      })

      message.success(t('app.docs.fileDetail.rulesUpdateSuccess'))
      setConfigModal({ visible: false })

      // 修改成功后返回上一级
      navigate(`/docs/detail/${id}`)
    } catch (error) {
      message.error(t('app.docs.fileDetail.rulesUpdateFailed'))
      console.error('修改分段规则失败:', error)
    }
  }

  // 处理问答开关变化
  const handleQAToggle = async (checked: boolean) => {
    if (isQADisabled) return // 如果禁用状态，不执行任何操作

    const res = await disableFileChunk({
      file_id: Number(fileId),
      is_disable: checked,
    })

    if (res) {
      setIsQAEnabled(checked)
      message.success(
        checked
          ? t('app.docs.fileDetail.disableQASuccess')
          : t('app.docs.fileDetail.enableQASuccess'),
      )
    }
  }

  // 渲染分段列表
  const renderSegmentList = () => (
    <div className='h-full'>
      <div className='space-y-3 h-full overflow-auto custom-preview-scroll pr-1'>
        {segments.map((segment) => (
          <div key={segment.id} className='mb-[16px]'>
            {/* 分段标题行 */}
            <div className='flex items-center  mb-[6px] justify-between'>
              <div className='flex items-center gap-1'>
                <span className=' text-[#919497] text-sm  font-normal'>
                  {t('app.docs.fileDetail.segmentTitle', {
                    number: segment.chunk_number,
                    count: segment.charCount,
                  })}
                </span>
              </div>
              <div className='flex items-center gap-2.5'>
                <Tooltip title={t('app.docs.fileDetail.edit')}>
                  <Button
                    type='text'
                    size='small'
                    icon={<img src={editIcon} alt='edit' className='w-4 h-4' />}
                    onClick={() => handleSegmentEdit(segment.id)}
                    className='hover:bg-[#FCFCFE] hover:shadow-[0px_0px_3.3px_0px_rgba(0,0,0,0.15)] w-4 h-4 p-0 rounded transition-all duration-200'
                  />
                </Tooltip>
                <Tooltip title={t('app.docs.fileDetail.delete')}>
                  <Button
                    type='text'
                    size='small'
                    icon={
                      <img src={deleteIcon} alt='delete' className='w-4 h-4' />
                    }
                    onClick={() => handleSegmentDeleteConfirm(segment.id)}
                    className='hover:bg-[#FCFCFE] hover:shadow-[0px_0px_3.3px_0px_rgba(0,0,0,0.15)] w-4 h-4 p-0 rounded transition-all duration-200'
                  />
                </Tooltip>
              </div>
            </div>

            {/* 文本内容框 */}
            <div className='bg-white border border-[#d7d9e5] rounded-md px-2.5 py-[5px] w-full'>
              <Input.TextArea
                value={segment.content}
                readOnly
                autoSize={{ minRows: 1, maxRows: 10 }}
                bordered={false}
                className='text-[#1e1f28] text-base leading-[22px] custom-preview-scroll'
                style={{
                  padding: '0 4px 0 0',
                  resize: 'none',
                  cursor: 'default',
                  backgroundColor: 'transparent',
                }}
              />
            </div>
          </div>
        ))}
      </div>
    </div>
  )

  // 加载中状态
  if (loading) {
    return (
      <div className='flex justify-center items-center h-screen bg-[#fcfcfe]'>
        <Spin
          indicator={<LoadingOutlined style={{ fontSize: 48 }} spin />}
          tip={t('app.docs.fileDetail.loading')}
        />
      </div>
    )
  }

  // 错误状态
  if (error) {
    return (
      <div className='flex justify-center items-center h-screen bg-[#fcfcfe]'>
        <div className='text-center'>
          <div className='text-red-500 text-lg mb-4'>{error}</div>
          <Button type='primary' onClick={() => window.location.reload()}>
            {t('app.docs.fileDetail.reload')}
          </Button>
        </div>
      </div>
    )
  }

  const tabItems: TabItem[] = [
    {
      key: 'wisdom-qa',
      label: '智能问答',
    },
    {
      key: 'intelligent-analysis',
      label: '智能摘要',
    },

    {
      key: 'document-analysis',
      label: '文档解析',
    },
    {
      key: 'segmentRule',
      label: '分段规则',
    },
  ]

  return (
    <div className='h-screen bg-[#fcfcfe] overflow-hidden'>
      {/* 面包屑导航 */}
      <div className='w-full h-12 bg-white flex items-center px-5 pt-[2px] font-medium'>
        <Breadcrumb
          separator={<NavigationIcon className='inline-block' />}
          items={breadcrumbItems}
        />
      </div>

      {/* 主内容区域 */}
      <div className=' h-[calc(100vh-50px)] overflow-auto bg-white p-[10px]'>
        <div className=' h-full p-0 overflow-hidden flex flex-col gap-4'>
          {isTableKnowledgeBase ? (
            <div className='flex-1 overflow-hidden excel-preview-container'>
              <div className='h-full rounded-lg overflow-hidden bg-white'>
                <FilePreview
                  fileName={fileInfo.name}
                  fileType={fileInfo.name.split('.').pop()}
                  activeChunkId={activeChunkId}
                  segments={segments}
                  onChunkClick={setActiveChunkId}
                  activeTab={activeTab}
                />
              </div>
            </div>
          ) : (
            <>
              <div className='flex items-center justify-between flex-shrink-0  '>
                <div className='flex items-center gap-1 rounded-[6px] px-2.5 py-2 '>
                  <Switch
                    checked={isPreviewMode}
                    onChange={setIsPreviewMode}
                    rootClassName={styles.previewSwitch}
                    className=''
                    size='small'
                  />
                  <span className='text-[#0C1F17] text-sm font-medium'>
                    {t('app.docs.fileDetail.previewArticle')}
                  </span>
                </div>

                {/* <div className=' text-[#0C1F17] text-base font-medium '>
                  {t('app.docs.fileDetail.segmentationCount', {
                    count: segmentTotal,
                  })}
                </div> */}
                {/* Tab按钮组 */}
                <div className='flex h=[40px] p-[4px] gap-[10px] bg-[#F5F5F5] rounded-[6px]  flex-shrink-0 overflow-x-auto custom-tab-scroll'>
                  {tabItems.map((item, index) => {
                    const isActive = activeTab === item.key
                    const isFirst = index === 0
                    const isLast = index === tabItems.length - 1
                    return (
                      <Button
                        key={item.key}
                        onClick={() => handleTabChange(item.key)}
                        className={` px-4  text-sm font-medium !rounded-[4px] bg-[transparent] transition-all duration-200 !border-none flex-shrink-0
                                ${isActive ? '!bg-[#fff] !text-[#0C1F17]' : '!text-[#0C1F17] hover:!bg-[#FFFFFF]'}
                                ${isFirst ? '!rounded-l-md' : ''}
                                ${isLast ? '!rounded-r-md' : ''}
                              `}
                        style={{
                          boxShadow: 'none',
                        }}
                      >
                        {item.label}
                      </Button>
                    )
                  })}
                </div>
                {/* <div className='flex items-center gap-3'>
                  <Button
                    type='text'
                    onClick={handleConfiguration}
                    className='flex items-center gap-1 px-2.5 py-2 bg-[#9194971A] rounded-[6px] hover:bg-[#9194971A] text-[#0C1F17] text-sm font-medium'
                  >
                    {t('app.docs.fileDetail.editRules')}
                    <span>
                      <img
                        src={ChangeRuleIcon}
                        alt='change rule'
                        className='w-4 h-4'
                      />
                    </span>
                  </Button>

                  <div
                    className={`flex items-center gap-1 bg-[#9194971A] rounded-[6px] px-2.5 py-2 ${
                      isQADisabled ? 'cursor-not-allowed opacity-60' : ''
                    }`}
                  >
                    <span
                      className={`text-[#0C1F17] text-sm font-medium ${
                        isQADisabled ? 'text-gray-400' : ''
                      }`}
                    >
                      {isQAEnabled
                        ? t('app.docs.fileDetail.disableQA')
                        : t('app.docs.fileDetail.enableQA')}
                    </span>
                    <Switch
                      checked={!isQAEnabled}
                      onChange={(checked) => handleQAToggle(!checked)}
                      size='small'
                      disabled={isQADisabled}
                    />
                  </div>
                </div> */}
              </div>

              <div className='flex-1 overflow-auto'>
                <div className='flex h-full gap-3'>
                  {isPreviewMode && (
                    <div className='w-[40%] rounded-md border border-[#EFF1F4] overflow-hidden shadow-[0_10px_15px_0_#0000000D]'>
                      <FilePreview
                        fileName={fileInfo.name}
                        fileType={fileInfo.name.split('.').pop()}
                        activeChunkId={activeChunkId}
                        segments={segments}
                        onChunkClick={setActiveChunkId}
                        activeTab={activeTab}
                      />
                    </div>
                  )}

                  {/* <div className='w-1/2 rounded-lg overflow-hidden'>
                      {renderSegmentList()}
                      </div> */}
                  <FileDetailViewContext.Provider
                    value={{
                      fileId: Number(fileId),
                      fileInfo,
                      segments,
                      activeChunkId,
                      setActiveChunkId,
                      handleSegmentEdit,
                      handleSegmentDeleteConfirm,
                      segmentTotal,
                      handleConfiguration,
                    }}
                  >
                    <RightPanel activeKey={activeTab} />
                  </FileDetailViewContext.Provider>
                </div>
              </div>
            </>
          )}
        </div>
      </div>

      {!isTableKnowledgeBase && (
        <>
          {/* 删除确认弹窗 */}
          <DeleteConfirmModal
            visible={deleteModal.visible}
            isFolder={false}
            customText={t('app.docs.fileDetail.deleteConfirmText')}
            customTitle={t('app.docs.fileDetail.deleteConfirmTitle')}
            onCancel={() =>
              setDeleteModal({ visible: false, type: 'segment', targetId: '' })
            }
            onConfirm={handleDeleteConfirm}
          />

          {/* 分段编辑弹窗 */}
          <SegmentEditModal
            visible={editModal.visible}
            segmentId={editModal.segmentId}
            segmentTitle={editModal.segmentTitle}
            initialContent={editModal.content}
            imageUrl={editModal.imageUrl}
            segmentType={editModal.segmentType}
            onCancel={handleEditCancel}
            onConfirm={handleSegmentEditConfirm}
          />

          {/* 修改分段规则弹窗 */}
          <ModifySegmentRuleModal
            open={configModal.visible}
            onCancel={handleConfigCancel}
            onOk={handleConfigConfirm}
            initialRule={
              fileConfig
                ? convertApiConfigToModalConfig(fileConfig.split_config)
                : { type: 'default' }
            }
          />
        </>
      )}
    </div>
  )
}

export default FileDetailView
