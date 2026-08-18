import React, { ReactNode } from 'react'
import {
  Input,
  Tooltip,
  Tag,
  Dropdown,
  Button,
  App,
  Switch,
  MenuProps,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { MoreOutlined, TagOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import { useTranslation } from 'react-i18next'
import { retryParse } from '@/api/knowledge'
import DeleteIcon from '@/assets/icons/docs/delete-file.svg?react'
import EditIcon from '@/assets/icons/docs/edit-file.svg?react'
import PermissionIcon from '@/assets/icons/docs/permission-icon.svg?react'
import RefreshIcon from '@/pages/app/docs/components/ForestUploadBtn/upload/ForestAnalyzeFiles/refresh.svg?react'
import HelpIcon from '../images/help-tip.svg?react'
import MoveIcon from '../images/move.svg?react'
import { FileItem, SortItem } from '../types'
import { getFileIcon } from './fileUtils'

interface GetColumnsProps {
  disabled?: boolean
  editingId: number | null
  editInputRef: React.RefObject<any>
  sortItems: SortItem[]
  handleDelete: (id: number) => void
  startEditing: (id: number) => void
  saveEditing: (id: number, newName: string, ext?: string) => void
  handleFolderClick?: (folder: FileItem) => void
  currentLevel?: number
  handleMove?: (id: number) => void
  handlePermission?: (file: FileItem) => void
  handleFileClick?: (file: FileItem) => void
  handleFileEdit?: (file: FileItem) => void
  knowledgeBaseType?: 'file' | 'excel' | 'qa' | 'data'
  handleEnableChange?: (enable: boolean, file: FileItem) => void
  showFilePermission?: boolean // 是否显示文件权限功能
  showTagColumn?: boolean // 是否显示标签列
  handleTag?: (file: FileItem) => void // 处理标签选择
  enablingFileId?: number | null // 正在处理启用状态的文件ID
}

// 渲染解析状态标签
const renderParseStatus = (
  status: string,
  progress?: string,
  knowledgeBaseType?: 'file' | 'excel' | 'qa' | 'data',
  t: any = (k: string) => k,
) => {
  const statusMap: Record<
    string,
    { textKey: string; color: string; bg: string }
  > = {
    pending: {
      textKey: 'app.docs.detail.parseStatus.pending',
      color: '#576275',
      bg: '#F2F4F6',
    },
    parsing: {
      textKey: 'app.docs.detail.parseStatus.running',
      color: '#3D7FFF',
      bg: '#EBF2FF',
    },
    indexing: {
      textKey: 'app.docs.detail.parseStatus.indexing',
      color: '#D24BD2',
      bg: '#FFECFF',
    },
    // success 不展示任何标签
    success: {
      textKey: 'app.docs.detail.parseStatus.success',
      color: '#13A374',
      bg: '#D7F5E5',
    },
    fail: {
      textKey: 'app.docs.detail.parseStatus.fail',
      color: '#FF4C4C',
      bg: '#FFE7E7',
    },
    // unsupported 状态前端不展示标签（按需求简化为仅三种可见状态）
  }

  // if (status === 'success') return null

  const config = statusMap[status]
  if (!config) return null

  // For running status with progress, show only percentage (avoid duplicated wording like "进行中")
  let displayText: string

  if (progress) {
    displayText = progress
  } else {
    displayText = t(config.textKey)
  }

  if (status === 'fail' && knowledgeBaseType === 'excel') {
    displayText = t('app.docs.detail.parseStatus.tableFail')
  }

  return (
    <Tag
      className='border-0'
      style={{
        color: config.color,
        backgroundColor: config.bg,
        padding: '0px 8px',
        fontSize: '12px',
        borderRadius: '10px',
        fontWeight: '400',
      }}
    >
      {displayText}
    </Tag>
  )
}

// 获取表格列定义
export const getColumns = ({
  disabled,
  editingId,
  editInputRef,
  sortItems,
  handleDelete,
  startEditing,
  saveEditing,
  handleFolderClick,
  currentLevel = 0,
  handleMove,
  handlePermission,
  handleFileClick,
  handleFileEdit,
  knowledgeBaseType,
  handleEnableChange,
  showFilePermission = false,
  showTagColumn = false,
  handleTag,
  enablingFileId = null,
}: GetColumnsProps): ColumnsType<FileItem> => {
  const { t } = useTranslation('pages')
  const { message } = App.useApp()

  // 构建列数组
  const columns: ColumnsType<FileItem> = []

  // 名称列
  columns.push({
    title: (
      <div className='font-medium text-[#919497] text-sm leading-[22px]'>
        {t('app.docs.detail.name')}
      </div>
    ),
    dataIndex: 'name',
    key: 'name',
    width: '20%',
    // sorter: true,
    sortOrder: sortItems.find((item) => item.field === 'name')?.order,
    render: (text, record) => {
      const icon = getFileIcon(record.fileType, record.isFolder)

      // 编辑状态
      if (editingId === record.id) {
        // 编辑时也去除文件后缀
        let editName = text
        if (!record.isFolder && record.fileType && record.fileType !== '-') {
          // 去除文件扩展名
          const lastDotIndex = text.lastIndexOf('.')
          if (lastDotIndex > 0) {
            editName = text.substring(0, lastDotIndex)
          }
        }

        return (
          <div className='flex items-center'>
            {icon}
            <Input
              ref={editInputRef}
              defaultValue={editName}
              size='small'
              className='w-full !rounded-[4px] h-6 py-1 !bg-[#F7F9FC] !border !border-[#E3E6ED] !outline-none !shadow-none'
              maxLength={50}
              onPressEnter={(e) =>
                saveEditing(
                  record.id,
                  (e.target as HTMLInputElement).value,
                  record.ext,
                )
              }
              onBlur={(e) => saveEditing(record.id, e.target.value, record.ext)}
            />
          </div>
        )
      }

      // 正常显示状态 - 去除文件后缀
      let displayName = text
      if (!record.isFolder && record.fileType && record.fileType !== '-') {
        // 去除文件扩展名
        const lastDotIndex = text.lastIndexOf('.')
        if (lastDotIndex > 0) {
          displayName = text.substring(0, lastDotIndex)
        }
      }

      const title: ReactNode = displayName.length > 20 ? displayName : null
      const displayText =
        displayName.length > 20
          ? `${displayName.substring(0, 20)}...`
          : displayName

      return (
        <div className='flex items-center'>
          {icon}
          <Tooltip title={title}>
            <span
              className={`truncate ${
                record.isFolder && handleFolderClick
                  ? 'cursor-pointer hover:text-[#3D7FFF]'
                  : !record.isFolder && handleFileClick
                    ? 'cursor-pointer hover:text-[#3D7FFF]'
                    : ''
              }`}
              onClick={() => {
                if (record.isFolder && handleFolderClick) {
                  handleFolderClick(record)
                } else if (!record.isFolder && handleFileClick) {
                  handleFileClick(record)
                }
              }}
            >
              {displayText}
            </span>
          </Tooltip>
        </div>
      )
    },
  })

  // 状态列
  columns.push({
    title: (
      <div className='flex items-center gap-[4px] font-medium text-[#919497] text-sm leading-[22px]'>
        状态
        <Tooltip
          title={
            knowledgeBaseType === 'excel'
              ? '当资源已处于"资源已就绪"状态时，您使用该资源进行智能问答。'
              : '资源上传后将依次进入 "排队待处理"、"资源解析中"、"资源索引中"和"资源已就绪"四个阶段。当处理完成后，您使用该资源进行智能问答。'
          }
        >
          <HelpIcon />
        </Tooltip>
      </div>
    ),
    dataIndex: 'file_status',
    key: 'file_status',
    width: '10%',
    render: (status, record) => {
      if (record.isFolder) return <div className='text-[20px]'>-</div>

      return (
        <div className='flex items-center text-xs'>
          {/* 文件解析状态（仅文件显示，文件夹不显示） */}
          {!record.isFolder && record.file_status && (
            <>
              {renderParseStatus(
                status,
                record.file_progress,
                knowledgeBaseType,
                t,
              )}
              <Tooltip title='点击重试'>
                {record.file_status === 'fail' &&
                  knowledgeBaseType !== 'excel' && (
                    <RefreshIcon
                      className='ml-2 cursor-pointer'
                      onClick={async (e) => {
                        e.stopPropagation()
                        try {
                          await retryParse(record.id as any)
                          message.success('已触发重试')
                        } catch (err) {
                          message.error('重试失败，请稍后再试')
                        }
                      }}
                    />
                  )}
              </Tooltip>
            </>
          )}
        </div>
      )
    },
  })

  // 分段规则列
  columns.push({
    title: (
      <div className='flex items-center gap-[4px] font-medium text-[#919497] text-sm leading-[22px]'>
        分段规则
        <Tooltip
          title='1、系统将使用默认分段规则解析导入文件：
                   2、您可在文件详情页手动调整分段规则。'
        >
          <HelpIcon />
        </Tooltip>
      </div>
    ),
    dataIndex: 'file_config',
    key: 'file_config',
    width: '10%',
    render: (config, record) => {
      let typeText = '用户自定义'

      if (!config?.split_config) {
        typeText = '系统默认'
      }

      if (config?.split_config?.split_mode === 'auto') {
        typeText = '系统默认'
      }

      if (config?.split_config?.split_mode === 'rule') {
        typeText = '自定义分段'
      }

      if (record.isFolder) return <div className='text-[20px]'>-</div>

      return (
        <div className='flex items-center text-[14px] font-[500] text-[#3C4149]'>
          {typeText}
        </div>
      )
    },
  })

  // 标签列（仅在指定环境下显示）
  if (showTagColumn) {
    columns.push({
      title: (
        <div className='font-medium text-[#919497] text-sm leading-[22px]'>
          标签
        </div>
      ),
      dataIndex: 'tag_list',
      key: 'tag_list',
      width: '15%',
      render: (tagList, record) => {
        if (record.isFolder) return <div className='text-[20px]'>-</div>

        if (!tagList || !Array.isArray(tagList) || tagList.length === 0) {
          return <span className='text-[#919497]'>-</span>
        }

        // 获取所有标签名称
        const tagNames = tagList
          .map((tag: any) => tag.tag_name || tag.name || '')
          .filter(Boolean)
        const tagText = tagNames.join('、')

        return (
          <Tooltip title={tagText}>
            <div
              className='max-w-full truncate text-[14px] text-[#3C4149]'
              style={{ maxWidth: '150px' }}
            >
              {tagText}
            </div>
          </Tooltip>
        )
      },
    })
  }

  // 创建时间列
  columns.push({
    title: (
      <div className='font-medium text-[#919497] text-sm leading-[22px]'>
        {t('app.docs.detail.createTime')}
      </div>
    ),
    dataIndex: 'CreatedAt',
    key: 'CreatedAt',
    // sorter: true,
    sortOrder: sortItems.find((item) => item.field === 'createdAt')?.order,
    width: '20%',
    render: (time) => {
      if (!time) return '-'
      return dayjs(time).format('YYYY-MM-DD HH:mm')
    },
  })

  // 启用状态列
  columns.push({
    title: (
      <div className='flex items-center gap-[4px] font-medium text-[#919497] text-sm leading-[22px]'>
        启用状态
        <Tooltip title='启用后，系统将在问答时检索并引用该资源。'>
          <HelpIcon />
        </Tooltip>
      </div>
    ),
    dataIndex: 'enable',
    key: 'enable',
    width: '15%',
    render: (enable, record) => {
      // 如果该文件正在处理启用状态，禁用开关
      const isProcessing = enablingFileId === record.id
      return (
        <div onClick={(e) => e.stopPropagation()}>
          <Switch
            disabled={!!disabled || isProcessing}
            size='small'
            onChange={(val) => handleEnableChange?.(val, record)}
            value={enable}
          />
        </div>
      )
    },
  })

  // 操作列
  columns.push({
    title: (
      <div className='font-medium text-[#919497] text-sm leading-[22px]'>
        {t('app.docs.detail.actions')}
      </div>
    ),
    key: 'actions',
    width: '15%',
    render: (_, record) => {
      const isReadOnly = !!disabled
      const sharedItemClass =
        'text-[#2D2D2D] hover:!bg-[#F5F5F5] px-[10px] py-[7px] rounded'
      const disabledItemClass =
        'cursor-not-allowed !text-[#C0C4CC] hover:!bg-transparent'

      const menuItems: MenuProps['items'] = [
        {
          key: 'edit',
          label: (
            <span
              className='flex-1 min-w-0 truncate font-medium'
              title={t('app.docs.detail.edit')}
            >
              {t('app.docs.detail.edit')}
            </span>
          ),
          icon: (
            <EditIcon className={`w-4 h-4 ${isReadOnly ? 'opacity-40' : ''}`} />
          ),
          onClick: (e) => {
            e.domEvent.stopPropagation()
            if (!isReadOnly) {
              startEditing(record.id)
            }
          },
          className: `${sharedItemClass} ${isReadOnly ? disabledItemClass : ''}`,
          disabled: isReadOnly,
        },
        {
          key: 'move',
          label: (
            <span
              className='flex-1 min-w-0 truncate font-medium'
              title={t('app.docs.detail.edit')}
            >
              移动
            </span>
          ),
          icon: (
            <MoveIcon className={`w-4 h-4 ${isReadOnly ? 'opacity-40' : ''}`} />
          ),
          onClick: (e) => {
            e.domEvent.stopPropagation()
            if (!isReadOnly) {
              handleMove?.(record.id)
            }
          },
          className: `${sharedItemClass} ${isReadOnly ? disabledItemClass : ''}`,
          disabled: isReadOnly,
        },
        ...(!record.isFolder && showFilePermission
          ? [
              {
                key: 'permission',
                label: (
                  <span
                    className='flex-1 min-w-0 truncate font-medium'
                    title='权限'
                  >
                    权限
                  </span>
                ),
                icon: (
                  <PermissionIcon
                    className={`w-4 h-4 ${isReadOnly ? 'opacity-40' : ''}`}
                  />
                ),
                onClick: (e: any) => {
                  e.domEvent.stopPropagation()
                  if (!isReadOnly) {
                    handlePermission?.(record)
                  }
                },
                className: `${sharedItemClass} ${isReadOnly ? disabledItemClass : ''}`,
                disabled: isReadOnly,
              },
            ]
          : []),
        ...(!record.isFolder && showTagColumn
          ? [
              {
                key: 'tag',
                label: (
                  <span
                    className='flex-1 min-w-0 truncate font-medium'
                    title='标签'
                  >
                    标签
                  </span>
                ),
                icon: (
                  <TagOutlined
                    className={`w-4 h-4 ${isReadOnly ? 'opacity-40' : ''}`}
                  />
                ),
                onClick: (e: any) => {
                  e.domEvent.stopPropagation()
                  if (!isReadOnly) {
                    handleTag?.(record)
                  }
                },
                className: `${sharedItemClass} ${isReadOnly ? disabledItemClass : ''}`,
                disabled: isReadOnly,
              },
            ]
          : []),
        {
          key: 'delete',
          label: (
            <span
              className='flex-1 min-w-0 truncate font-medium'
              title={t('app.docs.detail.delete')}
            >
              {t('app.docs.detail.delete')}
            </span>
          ),
          icon: (
            <DeleteIcon
              className={`w-4 h-4 ${isReadOnly ? 'opacity-40' : ''}`}
            />
          ),
          onClick: (e) => {
            e.domEvent.stopPropagation()
            if (!isReadOnly) {
              handleDelete(record.id)
            }
          },
          className: `${sharedItemClass} ${isReadOnly ? disabledItemClass : ''}`,
          disabled: isReadOnly,
        },
      ]

      return (
        <div
          onClick={(e) => {
            e.stopPropagation()
          }}
        >
          <Dropdown
            menu={{
              items: menuItems,
              className:
                'px-[10px] py-[10px] min-w-[190px] max-w-[230px] bg-white',
            }}
            trigger={['click']}
            placement='bottomRight'
          >
            <Button
              type='text'
              icon={<MoreOutlined style={{ transform: 'rotate(90deg)' }} />}
              className='!px-2'
              onClick={(e) => {
                e.preventDefault()
                e.stopPropagation()
              }}
              aria-label={t('app.docs.detail.actions')}
            />
          </Dropdown>
        </div>
      )
    },
  })

  return columns
}
