import { useState, useRef, useMemo } from 'react'
import { NavLink } from 'react-router-dom'
import { useNavigate } from 'react-router-dom'
import { Dropdown, message, Spin, Input, InputRef } from 'antd'
import { removeChatSession, renameChatSession, setTopChatSession } from '@/api'
import { cn } from '@/utils'
import DeleteIcon from '@/assets/icons/delete2.svg?react'
import EditIcon from '@/assets/icons/edit2.svg?react'
import HandleMenuIcon from '@/assets/icons/handle-menu.svg?react'
import TopTitle from '@/assets/icons/top-title.svg?react'
import useConfirm from '@/hooks/useConfirm'

interface HistoryItemProps {
  agentId: string
  currentSessionId: string
  typePath: boolean
  item: {
    id: string
    name: string
    resourceType: string
    input?: any
    is_top?: boolean
    isDeleted?: boolean
  }
  onClick?: (item: any) => void
  setSessionList: React.Dispatch<React.SetStateAction<any[]>>
  refresh: () => Promise<void>
}

export default function HistoryItem({
  agentId,
  currentSessionId,
  typePath,
  item,
  onClick,
  setSessionList,
  refresh,
}: HistoryItemProps) {
  const navigator = useNavigate()
  const { confirmDelete } = useConfirm()

  const [editingId, setEditingId] = useState<string | null>(null)
  const [editingName, setEditingName] = useState('')
  const [editingSubmitLoading, setEditingSubmitLoading] = useState(false)
  const inputRef = useRef<InputRef>(null)

  const isEdit = useMemo(() => {
    return editingId === item.id
  }, [editingId, item.id])

  const handleDelete = (id: string) => {
    confirmDelete({
      title: '确认删除该会话？',
      content: '您确定要删除此会话',
      onOk: async () => {
        await removeChatSession(id)
        message.success('删除成功')
        setSessionList((prev) => prev.filter((x) => x.id !== id))
        setEditingId(null)
        if (Number(currentSessionId) === Number(item.id)) {
          navigator(`/agents/detail/${typePath}/${agentId}`)
        }
      },
    })
  }

  const handleRename = () => {
    if (editingSubmitLoading) return
    if (editingName === item.name) {
      setEditingId(null)
      return
    }
    if (editingName.trim()) {
      setEditingSubmitLoading(true)
      renameChatSession({ id: item.id, name: editingName })
        .then(() => {
          setSessionList((prev) =>
            prev.map((x) =>
              x.id === item.id ? { ...x, name: editingName } : x,
            ),
          )
          setEditingId(null)
        })
        .finally(() => {
          setEditingSubmitLoading(false)
        })
    } else {
      setEditingId(null)
    }
  }

  const handleTop = async (id: string) => {
    await setTopChatSession(id)
    message.success('置顶成功')
    // 置顶之后影响数据库排序，需要重新加载数据
    refresh()
  }

  if (item.isDeleted) return null
  //浅色 E8F3FF  深色 bad8ff
  return (
    <NavLink
      key={item.id}
      className={cn(
        'flex-none px-2 h-8 flex items-center gap-1 rounded group',
        Number(currentSessionId) === Number(item.id)
          ? 'bg-[#E8F3FF] text-title'
          : 'text-black/60',
        isEdit && 'px-0',
      )}
      to={`/agents/detail/${typePath}/${agentId}?sessionId=${item.id}`}
      onClick={() => onClick && onClick(item)}
    >
      {isEdit ? (
        <div className='w-full h-full flex items-center'>
          <div className='flex-grow h-full'>
            <Input
              ref={inputRef}
              className='w-full h-full px-1 border border-[#bad8ff] rounded-md outline-none'
              value={editingName}
              onChange={(e) => setEditingName(e.target.value)}
              onBlur={handleRename}
              onPressEnter={handleRename}
              onClick={(e) => {
                e.stopPropagation()
                e.preventDefault()
              }}
              suffix={editingSubmitLoading && <Spin size='small' />}
              autoFocus
            />
          </div>
        </div>
      ) : (
        <span className='flex-grow truncate'>{item.name}</span>
      )}
      <Dropdown
        menu={{
          items: [
            {
              key: 'rename',
              label: '编辑标题',
              icon: <EditIcon />,
              className: 'text-[#165DFF]!',
              onClick: (e) => {
                e.domEvent.stopPropagation()
                setEditingId(item.id)
                setEditingName(item.name)
                setTimeout(() => inputRef.current?.focus(), 0)
              },
            },
            {
              key: 'top',
              label: item.is_top ? '取消置顶' : '置顶',
              icon: <TopTitle />,
              onClick: (e) => {
                e.domEvent.stopPropagation()
                handleTop(item.id)
              },
            },
            {
              key: 'delete',
              label: '删除',
              className: 'text-[#F56C6C]!',
              icon: <DeleteIcon />,
              onClick: (e) => {
                e.domEvent.stopPropagation()
                handleDelete(item.id)
              },
            },
          ],
        }}
      >
        <div
          className={cn(
            `flex-none text-transparent p-0.5 rounded
          hidden
        `,
            Number(currentSessionId) === Number(item.id) && 'block',
            isEdit ? 'hidden' : 'group-hover:block hover:bg-[#bad8ff]',
          )}
        >
          <HandleMenuIcon />
        </div>
      </Dropdown>
    </NavLink>
  )
}
