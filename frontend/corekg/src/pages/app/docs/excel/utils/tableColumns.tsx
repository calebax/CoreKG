import React, { ReactNode } from 'react'
import { Input, Tooltip } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import FileActions from '../components/FileActions'
import { FileItem, SortItem } from '../types'
import { getFileIcon } from './fileUtils'

interface GetColumnsProps {
  disabled?: boolean
  editingId: number | null
  editInputRef: React.RefObject<any>
  sortItems: SortItem[]
  handleDelete: (id: number) => void
  startEditing: (id: number) => void
  saveEditing: (id: number, newName: string) => void
  handleFolderClick?: (folder: FileItem) => void
  currentLevel?: number
  handleMove?: (id: number) => void
  handleFileClick?: (file: FileItem) => void
  forest_id: number
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
  handleFileClick,
  forest_id,
}: GetColumnsProps): ColumnsType<FileItem> => [
  {
    title: <div className='font-[400] text-[#0A1A3A] text-base mr-1'>名称</div>,
    dataIndex: 'name',
    key: 'name',
    sorter: true,
    sortOrder: sortItems.find((item) => item.field === 'name')?.order,
    width: '25%',
    render: (text, record) => {
      const icon = getFileIcon(record.fileType, record.isFolder)
      if (editingId === record.id) {
        return (
          <div className='flex items-center'>
            {icon}
            <Input
              ref={editInputRef}
              defaultValue={text}
              size='small'
              className='w-full !border-none !outline-none !shadow-none'
              maxLength={25}
              onPressEnter={(e) =>
                saveEditing(record.id, (e.target as HTMLInputElement).value)
              }
              onBlur={(e) => saveEditing(record.id, e.target.value)}
            />
          </div>
        )
      }
      const title: ReactNode = text.length > 25 ? text : null

      return (
        <div className='flex items-center min-w-0'>
          {icon}
          <Tooltip title={title}>
            <span
              className={`truncate ${record.isFolder && handleFolderClick ? 'cursor-pointer hover:text-[#4080FF]' : !record.isFolder && handleFileClick ? 'cursor-pointer hover:text-[#4080FF]' : ''}`}
              onClick={() => {
                if (record.isFolder && handleFolderClick) {
                  handleFolderClick(record)
                } else if (!record.isFolder && handleFileClick) {
                  handleFileClick(record)
                }
              }}
            >
              {text.length > 25 ? `${text.substring(0, 25)}...` : text}
            </span>
          </Tooltip>
        </div>
      )
    },
  },

  {
    title: <div className='font-[400] text-[#0A1A3A] text-base mr-1'>大小</div>,
    dataIndex: 'size',
    key: 'size',
    sorter: true,
    sortOrder: sortItems.find((item) => item.field === 'size')?.order,
    width: '15%',
  },
  {
    title: (
      <div className='font-[400] text-[#0A1A3A] text-base mr-1'>文件类型</div>
    ),
    dataIndex: 'fileType',
    key: 'fileType',
    sorter: false,
    sortOrder: sortItems.find((item) => item.field === 'fileType')?.order,
    width: '15%',
    render: (_, record) => <div>{record.isFolder ? '-' : record.fileType}</div>,
  },
  {
    title: (
      <div className='font-[400] text-[#0A1A3A] text-base mr-1'>创建时间</div>
    ),
    dataIndex: 'updatedAt',
    key: 'updatedAt',
    sorter: true,
    sortOrder: sortItems.find((item) => item.field === 'updatedAt')?.order,
    width: '15%',
  },
  {
    title: (
      <div className='font-[400] text-[#0A1A3A] text-base mr-1'>资源状态</div>
    ),
    dataIndex: 'file_status',
    key: 'file_status',
    sorter: false,
    sortOrder: sortItems.find((item) => item.field === 'parseStatus')?.order,
    width: '15%',
    render: (_, record) => {
      const { file_status, file_progress, isFolder } = record
      if (isFolder) return '-'
      switch (file_status) {
        case 'unsupported':
          return (
            <span className='inline-flex items-center px-2.5 py-1 rounded-sm text-xs font-normal bg-[#F5F5F5] text-[#8C8C8C]'>
              暂不支持
            </span>
          )
        case 'pending':
          return (
            <span className='inline-flex items-center px-2.5 py-1 rounded-sm text-xs font-normal bg-[#F2F4F6] text-[#576275]'>
              待开始
            </span>
          )
        case 'success':
          return (
            <span className='inline-flex items-center px-2.5 py-1 rounded-sm text-xs font-normal bg-[#E1FADE] text-[#4BC144]'>
              已完成
            </span>
          )
        case 'running':
          return (
            <span className='inline-flex items-center px-2.5 py-1 rounded-sm text-xs font-normal bg-[#EAF2FF] text-[#3D7FFF]'>
              {file_progress ? `${file_progress}` : '解析中'}
            </span>
          )
        case 'fail':
          return (
            <span className='inline-flex items-center px-2.5 py-1 rounded-sm text-xs font-normal bg-[#FFEAE7] text-[#FF3B33]'>
              {file_progress ? `${file_progress}` : '解析失败'}
            </span>
          )
      }
    },
  },
  {
    title: <div className='font-[400] text-[#0A1A3A] text-base'>操作管理</div>,
    key: 'action',
    width: '15%',
    render: (_, record) => (
      <FileActions
        disabled={disabled}
        record={record}
        onDelete={handleDelete}
        onEdit={() => startEditing(record.id)}
        onMove={handleMove}
        forest_id={forest_id}
      />
    ),
  },
]
