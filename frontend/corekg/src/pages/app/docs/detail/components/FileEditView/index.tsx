import React, { useState, useMemo, useEffect } from 'react'
import { useParams, useNavigate, useSearchParams } from 'react-router-dom'
import { Button, Switch, message, Breadcrumb, Tooltip, Spin, Input } from 'antd'
import { LoadingOutlined } from '@ant-design/icons'
import {
  getFileSegments,
  updateFileSegment,
  deleteFileSegment,
  deleteFile,
  modifyFileSegmentRule,
  getFileInfo,
} from '@/api/knowledge'
import configurationIcon from '@/assets/icons/docs/configuration.svg'
import deleteIcon from '@/assets/icons/docs/delete.svg'
import editIcon from '@/assets/icons/docs/edit.svg'
import SegmentationIcon from '@/assets/icons/docs/segmentation.svg'
import scrollStyles from '@/styles/scroll/styles.module.scss'
import DeleteConfirmModal from '../DeleteConfirmModal'
import FilePreview from '../FilePreview'
import {
  ModifySegmentRuleModal,
  ModifySegmentRule,
} from './ModifySegmentRuleModal'
import SegmentEditModal from './SegmentEditModal'

// 分段数据接口
interface FileSegment {
  id: string
  content: string
  chunk_number: number
  charCount: number
}

// 文件信息接口
interface FileInfo {
  id: number
  name: string
  knowledgeBaseName: string
}

const FileEditView: React.FC = () => {
  const { id, fileId } = useParams()
  const navigate = useNavigate()

  // 获取路由参数中的动态数据
  const [searchParams] = useSearchParams()
  const kbNameFromUrl = searchParams.get('kbName') || ''
  const fileNameFromUrl = searchParams.get('fileName') || ''

  // 预览状态管理
  const [isPreviewMode, setIsPreviewMode] = useState(true)

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
  })

  // 配置弹窗状态
  const [configModal, setConfigModal] = useState({
    visible: false,
  })

  // 分段数据状态管理
  const [segments, setSegments] = useState<FileSegment[]>([])
  const [segmentTotal, setSegmentTotal] = useState(0) // 分段总数
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // 文件配置状态管理
  const [fileConfig, setFileConfig] = useState<{
    split_config: any | null
  } | null>(null)

  // 文件信息
  const fileInfo: FileInfo = {
    id: Number(fileId) || 1,
    name: fileNameFromUrl || '未知文件',
    knowledgeBaseName: kbNameFromUrl || '知识库',
  }

  // 获取文件信息和分段数据
  useEffect(() => {
    const fetchData = async () => {
      if (!fileId) return

      setLoading(true)
      setError(null)

      try {
        // 同时获取文件信息和分段数据
        const [fileInfoRes, segmentsRes] = await Promise.all([
          getFileInfo({
            file_id: Number(fileId),
          }),
          getFileSegments({
            file_id: Number(fileId),
            forest_id: Number(id),
          }),
        ])

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
        if (segmentsRes) {
          const chunks = segmentsRes.chunks || []
          const formattedSegments = chunks.map((chunk: any) => ({
            id: chunk._id,
            chunk_number: chunk._source.sequence || 0,
            content: chunk._source.description || '',
            charCount: chunk._source.chunk_size || chunk._source.tokens,
          }))
          setSegments(formattedSegments)
          setSegmentTotal(formattedSegments.length)
        } else {
          setError('获取分段数据失败')
        }
      } catch (err) {
        setError('获取数据失败')
        console.error('获取数据失败:', err)
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [fileId, id])

  // 面包屑导航项 - 三级结构，使用Figma设计规范
  const breadcrumbItems = useMemo(
    () => [
      {
        title: (
          <span
            className="font-['PingFang_SC'] text-[#616373] text-[16px] leading-[24px] cursor-pointer hover:text-[#0C99FF]"
            onClick={() => navigate('/docs')}
          >
            知识库
          </span>
        ),
      },
      {
        title: (
          <span
            className="font-['PingFang_SC'] text-[#616373] text-[16px] leading-[24px] cursor-pointer hover:text-[#0C99FF]"
            onClick={() => navigate(`/docs/detail/${id}`)}
          >
            {fileInfo.knowledgeBaseName}
          </span>
        ),
      },
      {
        title: (
          <span className="font-['PingFang_SC'] text-[#1e1f28] text-[16px] leading-[24px]">
            {fileInfo.name}
          </span>
        ),
      },
    ],
    [navigate, id, fileInfo.knowledgeBaseName, fileInfo.name],
  )

  // 处理分段编辑
  const handleSegmentEdit = (segmentId: string) => {
    const segment = segments.find((s) => s.id === segmentId)
    if (segment) {
      // 动态生成分段标题
      const segmentTitle = `分段 ${segment.chunk_number} -- ${segment.charCount} 字符数`
      setEditModal({
        visible: true,
        segmentId: segmentId,
        segmentTitle: segmentTitle,
        content: segment.content,
      })
    }
  }

  // 处理分段编辑确认
  const handleSegmentEditConfirm = async (
    segmentId: string,
    newContent: string,
  ) => {
    try {
      const res = await updateFileSegment({
        chunk_id: segmentId,
        description: newContent,
        file_id: Number(fileId),
      })

      if (res) {
        message.success('分段编辑保存成功')
        // 重新获取分段数据
        const res = await getFileSegments({
          file_id: Number(fileId),
          forest_id: Number(id),
        })
        if (res) {
          // 处理API响应数据
          const chunks = res.chunks || []
          const formattedSegments = chunks.map((chunk: any) => ({
            id: chunk._id,
            chunk_number: chunk._source.sequence || 0,
            content: chunk._source.description || '',
            charCount: chunk._source.chunk_size || chunk._source.tokens,
          }))
          setSegments(formattedSegments)
          setSegmentTotal(formattedSegments.length)
        }
      } else {
        message.error('保存失败，请重试')
      }
    } catch (error) {
      message.error('保存失败，请重试')
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

  // 处理文件删除确认
  const handleFileDeleteConfirm = () => {
    setDeleteModal({
      visible: true,
      type: 'file',
      targetId: fileId || '',
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
          message.success('分段删除成功')
          // 重新获取分段数据
          const segmentsRes = await getFileSegments({
            file_id: Number(fileId),
            forest_id: Number(id),
          })
          if (segmentsRes) {
            // 处理API响应数据
            const chunks = segmentsRes.chunks || []
            const formattedSegments = chunks.map((chunk: any) => ({
              id: chunk._id,
              chunk_number: chunk._source.sequence || 0,
              content: chunk._source.description || '',
              charCount: chunk._source.chunk_size || chunk._source.tokens,
            }))
            setSegments(formattedSegments)
            setSegmentTotal(formattedSegments.length)
          }
        } else {
          // message.error('删除失败，请重试')
        }
      } else {
        const res = await deleteFile({
          file_id: [Number(deleteModal.targetId)],
        })
        if (res) {
          message.success('文件删除成功')
          // 删除成功后返回上级页面
          navigate(`/docs/detail/${id}`)
        } else {
          // message.error('删除失败，请重试')
        }
      }
    } catch (error) {
      // message.error('删除失败，请重试')
      console.error('删除失败:', error)
    } finally {
      setDeleteModal({ visible: false, type: 'segment', targetId: '' })
    }
  }

  // 处理调优
  const handleOptimize = (segmentId: string) => {
    message.info(`调优分段 ${segmentId} - 待实现`)
  }

  // 将API的split_config转换为弹窗需要的格式
  const convertApiConfigToModalConfig = (
    splitConfig: any,
  ): ModifySegmentRule => {
    if (!splitConfig) {
      return { type: 'default' }
    }

    // 根据split_mode判断是默认规则还是自定义规则
    if (splitConfig.split_mode === 'auto') {
      return { type: 'default' }
    }

    // 自定义规则，返回完整的配置信息
    return {
      type: 'custom',
      segmentLength: splitConfig.chunk_size || 256,
      segmentSeparator: splitConfig.split_mark?.[0] || '\n',
      segmentOverlap: Math.round((splitConfig.split_overlap ?? 0.25) * 100), // 转换回百分比
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
      // 构建分段规则配置，匹配真实API格式
      const split_config = {
        split_mode:
          config.type === 'default' ? ('auto' as const) : ('rule' as const),
        // 无论是默认规则还是自定义规则，都需要传递完整参数
        chunk_size:
          config.type === 'custom' ? config.segmentLength || 256 : 256,
        split_mark:
          config.type === 'custom' && config.segmentSeparator
            ? [config.segmentSeparator]
            : ['\n'],
        split_overlap:
          config.type === 'custom' ? (config.segmentOverlap ?? 30) / 100 : 0.25, // 转换为小数
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

      message.success('分段规则修改成功')
      setConfigModal({ visible: false })

      // 修改成功后返回上一级的多模态知识库详情表格页面
      navigate(`/docs/detail/${id}`)
    } catch (error) {
      message.error('修改分段规则失败，请重试')
      console.error('修改分段规则失败:', error)
    }
  }

  // 渲染分段列表 - 按Figma设计规范
  const renderSegmentList = (isCompact = false) => (
    <div className={`${isCompact ? 'pr-4' : ''}`}>
      {/* 分段总数 */}
      <div className="font-['PingFang_SC'] text-[#1e1f28] text-base font-medium leading-[24px] mb-3">
        分段总数：{segmentTotal}
      </div>

      {/* 分段列表 */}
      <div className='space-y-3'>
        {segments.map((segment, index) => (
          <div key={segment.id} className='space-y-2'>
            {/* 第一部分：分段标题行 */}
            <div className='flex items-center justify-between'>
              <div className='flex items-center gap-1'>
                <div className='w-3 h-3 flex items-center justify-center'>
                  <img
                    src={SegmentationIcon}
                    alt='segmentation'
                    className='w-3 h-3'
                  />
                </div>
                <span className="font-['PingFang_SC'] text-[#616373] text-sm leading-[24px] font-normal">
                  分段 {segment.chunk_number} -- {segment.charCount} 字符数
                </span>
              </div>
              <div className='flex items-center gap-2.5'>
                <Tooltip title='编辑'>
                  <Button
                    type='text'
                    size='small'
                    icon={<img src={editIcon} alt='edit' className='w-4 h-4' />}
                    onClick={() => handleSegmentEdit(segment.id)}
                    className='hover:bg-[#FCFCFE] hover:shadow-[0px_0px_3.3px_0px_rgba(0,0,0,0.15)] w-4 h-4 p-0 rounded transition-all duration-200'
                  />
                </Tooltip>
                <Tooltip title='删除'>
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

            {/* 第二部分：文本内容框 */}
            <div className='bg-white border border-[#d7d9e5] rounded-[4px] px-2.5 py-[5px] w-full'>
              <Input.TextArea
                value={segment.content}
                readOnly
                autoSize={{ minRows: 1, maxRows: 10 }}
                bordered={false}
                className={`font-['PingFang_SC'] text-[#1e1f28] text-base leading-[22px] ${scrollStyles.scroll}`}
                style={{
                  padding: 0,
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
          tip='加载中...'
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
            重新加载
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className='h-screen bg-[#fcfcfe] overflow-hidden'>
      {/* 面包屑导航 */}
      <div className='bg-white p-6'>
        <Breadcrumb
          separator={
            <span className="font-['Poppins'] text-[#616373] text-sm leading-[22px]">
              {' '}
              /{' '}
            </span>
          }
          items={breadcrumbItems}
        />
      </div>

      {/* 主内容区域 */}
      <div className='pb-6 px-4 h-[calc(100vh-104px)] overflow-auto bg-white'>
        <div className='bg-[#F8F9FD] rounded-[10px] border border-[#d7d9e5] h-full p-4 overflow-hidden flex flex-col'>
          {/* 文件标题和操作区域 */}
          <div className='flex items-center justify-between mb-4 flex-shrink-0'>
            {/* 左侧：文件名称 */}
            <div className='flex items-center gap-1'>
              <h1 className="font-['PingFang_SC'] text-[#1e1f28] text-base leading-[24px] font-normal">
                {fileInfo.name}
              </h1>
              {fileConfig && (
                <div
                  className={`px-2.5 py-0 rounded-[2px] min-w-5 h-6 flex items-center justify-center ${
                    fileConfig.split_config?.split_mode === 'auto'
                      ? 'bg-[#ffefd2]'
                      : 'bg-[#F7DDFF]'
                  }`}
                >
                  <span
                    className={`font-['PingFang_SC'] text-[12px] leading-[24px] font-normal ${
                      fileConfig.split_config?.split_mode === 'auto'
                        ? 'text-[#ff6600]'
                        : 'text-[#B638DD]'
                    }`}
                  >
                    {fileConfig.split_config?.split_mode === 'auto'
                      ? '默认分段'
                      : '自定义分段'}
                  </span>
                </div>
              )}
            </div>

            {/* 右侧：预览开关 + 文字 + 删除按钮 */}
            <div className='flex items-center gap-3'>
              <Switch
                checked={isPreviewMode}
                onChange={setIsPreviewMode}
                size='small'
              />
              <span className="font-['PingFang_SC'] text-[#616373] text-sm leading-[22px]">
                预览原始文档
              </span>
              {/* <Tooltip title='配置'> */}
              <div
                className='flex items-center gap-1 cursor-pointer'
                onClick={handleConfiguration}
              >
                <Button
                  type='text'
                  icon={
                    <img
                      src={configurationIcon}
                      alt='configuration'
                      className='w-4 h-4'
                    />
                  }
                  className='w-4 h-4 p-0 hover:bg-[#FCFCFE] hover:shadow-[0px_0px_3.3px_0px_rgba(0,0,0,0.15)] rounded transition-all duration-200'
                />
                <span className="font-['PingFang_SC'] text-[#616373] text-sm leading-[22px]">
                  修改分段规则
                </span>
              </div>
              {/* </Tooltip> */}
              {/* <Tooltip title='删除文件'> */}
              <div
                className='flex items-center gap-1 cursor-pointer'
                onClick={handleFileDeleteConfirm}
              >
                <Button
                  type='text'
                  icon={
                    <img src={deleteIcon} alt='delete' className='w-4 h-4 ' />
                  }
                  className='w-4 h-4 p-0 hover:bg-[#FCFCFE] hover:shadow-[0px_0px_3.3px_0px_rgba(0,0,0,0.15)] rounded transition-all duration-200'
                />
                <span className="font-['PingFang_SC'] text-[#616373] text-sm leading-[22px]">
                  删除当前文件
                </span>
              </div>
              {/* </Tooltip> */}
            </div>
          </div>

          {/* 分隔线 */}
          <div className='h-0 w-full border-b border-[#d7d9e5] border-dashed mb-4'></div>

          {/* 核心内容区域 */}
          <div className='flex-1 overflow-hidden'>
            {isPreviewMode ? (
              /* 预览模式：真正的50/50左右分栏 */
              <div className='flex h-full gap-4'>
                {/* 左侧预览区域 - 50% */}
                <div className='w-1/2 bg-[#f8f9fd] rounded-lg border border-[#d7d9e5] overflow-hidden'>
                  <FilePreview
                    fileName={fileInfo.name}
                    fileType={fileInfo.name.split('.').pop()}
                  />
                </div>

                {/* 右侧分段区域 - 50% */}
                <div
                  className={`w-1/2 bg-[#f8f9fd] rounded-lg overflow-auto ${scrollStyles.scroll}`}
                >
                  {renderSegmentList(true)}
                </div>
              </div>
            ) : (
              /* 非预览模式：完整分段列表 */
              <div
                className={`h-full overflow-auto pr-2 ${scrollStyles.scroll}`}
              >
                {renderSegmentList()}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* 删除确认弹窗 */}
      <DeleteConfirmModal
        visible={deleteModal.visible}
        isFolder={false}
        customText={
          deleteModal.type === 'segment'
            ? '删除后，该分段将无法恢复，请谨慎操作。'
            : '删除后，该文件将无法恢复，请谨慎操作。'
        }
        customTitle='确认删除'
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
    </div>
  )
}

export default FileEditView
