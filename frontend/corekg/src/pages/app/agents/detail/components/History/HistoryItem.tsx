import { FC } from 'react'
import { Dropdown, Input, App, Form } from 'antd'
import { cn } from '@/utils'
import DeleteIcon from '@/assets/icons/delete2.svg?react'
import EditIcon from '@/assets/icons/edit2.svg?react'
import HandleMenuIcon from '@/assets/icons/handle-menu.svg?react'
import TopTitle from '@/assets/icons/top-title.svg?react'
import styles from './styles.module.scss'

type HistoryItem = {
  name: string
  session_id: number
  active: boolean
  isTop: boolean
  onDel: () => Promise<void>
  onRename: (newName: string) => Promise<void>
  onToTop: () => Promise<void>
}
const HistoryItem: FC<HistoryItem> = ({
  name,
  session_id,
  active,
  isTop,
  onDel,
  onRename,
  onToTop,
}) => {
  const { modal, message } = App.useApp()
  const [form] = Form.useForm()
  return (
    <Link
      to={`?session_id=${session_id}`}
      className={cn(
        'flex-none px-2 h-8 cursor-pointer',
        'flex items-center gap-1 rounded',
        active ? 'bg-[#E8F3FF] text-title' : 'text-black/60',
        styles.historyItem,
        {
          [styles.active]: active,
        },
      )}
    >
      <span className='flex-grow truncate'>{name}</span>
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
                modal.confirm({
                  icon: null,
                  title: '请输入新名称',
                  content: (
                    <Form form={form} preserve={false} initialValues={{ name }}>
                      <Form.Item name='name'>
                        <Input />
                      </Form.Item>
                    </Form>
                  ),
                  onOk: async () => {
                    const newName = form.getFieldValue('name')
                    await onRename(newName)
                    message.success('操作成功')
                  },
                })
              },
            },
            {
              key: 'top',
              label: isTop ? '取消置顶' : '置顶',
              icon: <TopTitle />,
              onClick: async (e) => {
                e.domEvent.stopPropagation()
                await onToTop()
                message.success('操作成功')
              },
            },
            {
              key: 'delete',
              label: '删除',
              className: 'text-[#F56C6C]!',
              icon: <DeleteIcon />,
              onClick: async () => {
                modal.confirm({
                  title: '确认删除?',
                  onOk: async () => {
                    await onDel()
                    message.success('操作成功')
                  },
                })
              },
            },
          ],
        }}
      >
        <div
          className={cn(
            'flex-none text-transparent p-0.5 rounded hidden',
            styles.operator,
          )}
          onClick={(e) => {
            e.stopPropagation()
            e.preventDefault()
          }}
        >
          <HandleMenuIcon />
        </div>
      </Dropdown>
    </Link>
  )
}
export default HistoryItem
