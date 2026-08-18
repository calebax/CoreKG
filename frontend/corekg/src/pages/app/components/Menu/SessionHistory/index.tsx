import { FC, useMemo } from 'react'
import {
  Link,
  useLocation,
  useNavigate,
  useSearchParams,
} from 'react-router-dom'
import { App, Button, Form, Input, Popover, Skeleton, Typography } from 'antd'
import { DeleteOutlined, EditOutlined, UploadOutlined } from '@ant-design/icons'
import { useBoolean, useRequest } from 'ahooks'
import { cn } from '@/utils'
import { updateSessionName, deleteSession } from '@/api/agent'
import ArrowDown from '@/assets/icons/arrow-down.svg?react'
import { useGlobalSessionHistory } from '@/stores/GlobalSessionHistory'
import { scroll } from '@/styles/scroll'
import Operators from './images/operators.svg?react'
import Talking from './images/talking.svg?react'
import styles from './styles.module.scss'

type SessionHistory = {
  className?: string
  collapsed?: boolean
}
/** 首页问答历史记录 */
export const SessionHistory: FC<SessionHistory> = (props) => {
  const { className, collapsed } = props
  const { history, loadData } = useGlobalSessionHistory()
  const { loading } = useRequest(loadData)

  const [expaned, { toggle }] = useBoolean(true)
  const currentSessionId = useCurrentSession()
  return (
    <div
      className={cn(
        'flex-1 mt-2.5',
        'overflow-hidden flex flex-col gap-2',
        className,
      )}
    >
      <div className={cn('text-base', collapsed ? 'mx-3' : 'mx-4')}>
        <div
          className={cn(
            'h-10 w-10 rounded flex items-center justify-center cursor-pointer',
            'bg-transparent text-[#1E1F28]',
            'hover:bg-[#F0F2F7] transition-colors duration-200',
            { hidden: !collapsed },
          )}
          onClick={toggle}
        >
          <Talking className={cn('w-4 h-4')} />
        </div>
        <div
          className={cn(
            'w-50 h-10 rounded pl-3 flex items-center gap-3 cursor-pointer',
            'bg-transparent text-[#1E1F28]',
            'hover:bg-[#E6E8F0] transition-colors duration-200',
            ' font-normal',
            { hidden: collapsed },
          )}
          onClick={toggle}
        >
          <Talking className={cn('w-4 h-4')} />
          <span className='text-base'>最近会话</span>
        </div>
      </div>
      <div
        className={cn(
          'flex-1 overflow-y-auto overflow-x-hidden',
          'flex flex-col gap-1 mr-1',
          scroll,
          {
            hidden: collapsed,
          },
        )}
      >
        {loading ? <Skeleton active /> : null}
        {expaned && !loading
          ? history.map((item) => {
              const { id, name } = item
              return (
                <div
                  className={cn('text-base group', collapsed ? 'mx-3' : 'mx-4')}
                >
                  <div
                    key={id}
                    className={cn(
                      styles.historyItem,
                      'w-50 h-9 rounded pl-3 pr-1',
                      'flex gap-2 items-center',
                      'text-[#616373]',
                      'font-normal',
                      { [styles.active]: currentSessionId === id },
                    )}
                  >
                    <Link
                      to={`/QA?session_id=${id}`}
                      className='flex-1 min-w-0 text-inherit overflow-hidden mr-2'
                    >
                      {item.nameLoading ? (
                        <span className={styles.loadingName} />
                      ) : (
                        <Typography.Paragraph
                          ellipsis={{ rows: 1, tooltip: name }}
                          className='m-0 text-inherit'
                        >
                          {name}
                        </Typography.Paragraph>
                      )}
                    </Link>
                    <div className='flex-shrink-0'>
                      <Popover
                        arrow={false}
                        placement='rightTop'
                        content={
                          <HistoryOperators
                            id={id}
                            isCurrentId={currentSessionId === id}
                          />
                        }
                      >
                        <Operators
                          className={cn(
                            styles.operator,
                            'opacity-0 group-hover:opacity-100 transition-opacity',
                          )}
                        />
                      </Popover>
                    </div>
                  </div>
                </div>
              )
            })
          : null}
      </div>
    </div>
  )
}

const HistoryOperators: FC<{ id: number; isCurrentId: boolean }> = (props) => {
  const { id, isCurrentId } = props
  const navigate = useNavigate()
  const { modal, message } = App.useApp()
  const [form] = Form.useForm()
  const { history, set, del, rename } = useGlobalSessionHistory()

  return (
    <div className='p-2 flex flex-col gap-2'>
      <Button
        type='text'
        size='small'
        className='justify-start'
        icon={<EditOutlined />}
        onClick={() => {
          const currentItem = history.find((item) => item.id === id)
          const currentName = currentItem?.name || ''
          form.setFieldsValue({ name: currentName })
          modal.confirm({
            icon: null,
            title: '请输入新名称',
            content: (
              <Form form={form}>
                <Form.Item
                  rules={[
                    { required: true, message: '新名称不能为空' },
                    { max: 1, message: '名称不能少于1个字符' },
                    { max: 50, message: '名称不能超过50个字符' },
                  ]}
                  name='name'
                >
                  <Input />
                </Form.Item>
              </Form>
            ),
            onOk: async () => {
              const newName = form.getFieldValue('name')
              await updateSessionName(id, newName)
              rename(id, newName)
              message.success('重命名成功')
            },
          })
        }}
      >
        重命名
      </Button>
      {/* <Button
        type='text'
        size='small'
        icon={<UploadOutlined />}
        className='justify-start'
        onClick={() => {
          const item = history.find((item) => item.id === id)!
          set([item, ...history.filter((item) => item.id !== id)])
        }}
      >
        置顶
      </Button> */}
      <Button
        type='text'
        size='small'
        icon={<DeleteOutlined />}
        className='justify-start'
        onClick={() => {
          modal.confirm({
            content: '确定删除此会话?',
            onOk: async () => {
              try {
                await deleteSession(id)
                del(id)
                message.success('删除成功')
                if (isCurrentId) {
                  navigate('/')
                }
              } catch (error) {
                message.error('删除失败')
              }
            },
          })
        }}
      >
        删除
      </Button>
    </div>
  )
}

/** 当前的session_id */
const useCurrentSession = () => {
  const { pathname } = useLocation()
  const [searchParams] = useSearchParams()
  const sessionId = useMemo(() => {
    const _sessionId = Number(searchParams.get('session_id'))
    if (Number.isInteger(_sessionId)) return _sessionId
    return null
  }, [searchParams])
  if (sessionId && pathname === '/QA') return sessionId
  return null
}
