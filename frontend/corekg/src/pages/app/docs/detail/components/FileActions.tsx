import React, { useState } from 'react'
import { Button, message } from 'antd'
import { useBoolean } from 'ahooks'
import { cn } from '@/utils'
import { createSession } from '@/api/agent'
import KnowledgeMove from '@/assets/icons/knowledge-move.svg?react'
import { FileItem } from '../types'

interface FileActionsProps {
  disabled?: boolean
  record: FileItem
  onDelete: (id: number) => void
  onEdit: () => void
  onMove?: (id: number) => void
  onFileEdit?: (file: FileItem) => void
}

const FileActions: React.FC<FileActionsProps> = ({
  disabled,
  record,
  onDelete,
  onEdit,
  onMove,
  onFileEdit,
}) => {
  const [showActions, setShowActions] = useState(false)
  // 处理移动
  const handleMove = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (onMove) {
      onMove(record.id)
    } else {
      message.info('移动功能待实现')
    }
  }
  const { loading, startQA } = useStartQA(record)
  return (
    <div className={'relative'}>
      <div className={cn('flex items-center')}>
        {
          <>
            <Button
              className={cn(
                'cursor-pointer text-[#3D7FFF] font-normal text-sm p-0 bg-transparent!',
                record.file_status === 'success'
                  ? 'cursor-pointer text-[#3D7FFF] hover:text-[#3D7FFF]'
                  : 'cursor-not-allowed text-[#B9D3FF]',
              )}
              type='text'
              loading={loading}
              onClick={startQA}
              disabled={record.file_status !== 'success'}
            >
              问答
            </Button>
            {(!record.isFolder || onFileEdit) && (
              <div className='mx-2 h-3 w-px bg-[#EAEAEA]'></div>
            )}
          </>
        }

        {/* 编辑按钮 - 仅在文件解析完成时可点击 */}
        {!record.isFolder && (
          <>
            <Button
              className={cn(
                'font-normal text- p-0 bg-transparent!',
                record.file_status === 'success'
                  ? 'cursor-pointer text-[#3D7FFF] hover:text-[#3D7FFF]'
                  : 'cursor-not-allowed text-[#B9D3FF]',
              )}
              type='text'
              disabled={record.file_status !== 'success'}
              onClick={(e) => {
                e.stopPropagation()
                if (record.file_status === 'success' && onFileEdit) {
                  onFileEdit(record)
                }
              }}
            >
              编辑
            </Button>
            <div className='mx-2 h-3 w-px bg-[#E5E7EB]'></div>
          </>
        )}

        {/* 更多图标及下拉菜单容器 */}
        <div
          onClick={() => {
            if (!disabled) return
            message.error('无权限，请联系知识库管理员')
          }}
          onMouseEnter={() => {
            if (disabled) return
            setShowActions(true)
          }}
          onMouseLeave={() => setShowActions(false)}
        >
          <div
            className={cn('cursor-pointer text-[#3D7FFF] font-normal text-sm', {
              'cursor-not-allowed': disabled,
            })}
          >
            {/* <KnowledgeMore /> */}
            更多
          </div>

          {/* 更多操作下拉菜单 */}
          {showActions && (
            <div className='absolute left-0 top-0 z-50 pt-5'>
              <div className='bg-white rounded-[10px] shadow-[0px_4px_10px_0px_rgba(0,0,0,0.14)] flex flex-col gap-1 p-1 w-[123px]'>
                <div
                  className='bg-[#f8f9fd] px-2 py-[3px] h-8 cursor-pointer rounded-[3px] flex items-center'
                  onClick={handleMove}
                >
                  <span className='text-[#1e1f28] font-medium text-base leading-[22px]'>
                    移动到
                  </span>
                </div>
                <div
                  className='px-2 py-[3px] h-8 cursor-pointer rounded-[3px] flex items-center hover:bg-gray-50'
                  onClick={(e) => {
                    e.stopPropagation()
                    onEdit()
                  }}
                >
                  <span className='text-[#1e1f28] font-normal text-base leading-6'>
                    重命名
                  </span>
                </div>
                <div
                  className='px-2 py-[3px] h-8 cursor-pointer rounded-[3px] flex items-center hover:bg-gray-50'
                  onClick={(e) => {
                    e.stopPropagation()
                    onDelete(record.id)
                  }}
                >
                  <span className='text-[#ff3b33] font-normal text-base leading-6'>
                    删除
                  </span>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default FileActions

const useStartQA = (record: FileItem) => {
  const [loading, { toggle }] = useBoolean()
  const startQA = async () => {
    toggle()
    try {
      const { ID: session_id } = (await createSession({
        base_type: 'standard',
        ids: [record.id],
        resource_type: 'file_list',
      })) as any
      const searchParams = new URLSearchParams()
      searchParams.append('session_id', session_id)
      window.open(`${import.meta.env.BASE_URL}QA?${searchParams.toString()}`)
    } finally {
      toggle()
    }
  }
  return { startQA, loading }
}
