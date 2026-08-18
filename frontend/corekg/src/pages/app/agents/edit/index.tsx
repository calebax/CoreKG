import { FC } from 'react'
import { Button, Drawer, Form, FormInstance, Input, InputRef } from 'antd'
import { Agent } from 'Agent'
import { useBoolean } from 'ahooks'
import { cn } from '@/utils'
import { withAgentStyle } from '../AgentStyleProvider'
import { getAvatar } from '../utils/getAvatar'
import AgentAcces from './access'
import { EditProvider } from './components/AgentContext'
import { EditForm } from './components/EditForm'
import { EditableAvatar } from './components/EditableAvatar'
import { Test } from './components/Test'
import EditIcon from './images/edit.svg?react'
import Icon from './images/icon.svg?react'
import styles from './styles.module.scss'

// 为param增加一个配置时可用的唯一key
type Param = NonNullable<Agent['params']>[number]
type FormParams = (Param & { key: string })[]

/** 在表单中编辑的数据 */
export type AgentEditValue = Omit<Agent, 'params'> & {
  params: FormParams
}
const AgentEdit: FC = withAgentStyle((props) => {
  const [form] = Form.useForm<AgentEditValue>()
  return (
    <EditProvider>
      <div
        className={cn(
          'h-full overflow-hidden flex flex-col',
          props.extraClassName,
        )}
      >
        <Header form={form} />
        <div className='flex-1 flex py-[20px] px-[15px] gap-[16px] bg-[#FAFAFA] overflow-hidden'>
          <EditForm
            form={form}
            className={cn(
              'w-138 min-w-138 p-[15px] overflow-y-auto',
              'bg-[#FFFFFF] border border-[#E3E6ED] rounded-[10px] ',
              styles.scroll,
            )}
          />
          <Test form={form} className='flex-1' />
        </div>
      </div>
    </EditProvider>
  )
})

export default AgentEdit

const Header: FC<{ form: FormInstance<AgentEditValue> }> = (props) => {
  const [open, { toggle }] = useBoolean()
  const inputRef = useRef<InputRef>(null)
  const [edit, changeEdit] = useState<boolean>(false)
  const avatar = Form.useWatch('avatar', {
    form: props.form,
    preserve: true,
  })
  const title = Form.useWatch('title', { form: props.form, preserve: true })
  const type = Form.useWatch('type', { form: props.form, preserve: true })
  const handleChangeMode = () => {
    changeEdit(!edit)
  }

  const handleBlur = (newVal: string) => {
    changeEdit(!edit)
    if (newVal?.trim?.()?.length) {
      // setTitle(newVal.trim())
      props.form.setFieldValue('title', newVal)
    }
  }

  return (
    <>
      <div className='flex-none h-12 flex px-[32px] items-center justify-between border-b border-[#EFF1F4]'>
        <div className='flex gap-[14px] items-center'>
          <EditableAvatar
            value={getAvatar(avatar, type)}
            onChange={(src) => {
              props.form.setFieldValue('avatar', src)
            }}
            className='w-10 h-10 cursor-pointer flex-shrink-0 border-0 rounded-[10px]'
          />
          {!edit ? (
            <span className='text-[#0C1F17]  text-[14px] font-[500]'>
              {title}
            </span>
          ) : (
            <Input
              ref={inputRef}
              autoFocus
              defaultValue={title}
              onBlur={(e) => handleBlur(e.target.value)}
              onPressEnter={() => inputRef.current?.blur()}
              maxLength={50}
            />
          )}

          <EditIcon
            className='cursor-pointer flex-shrink-0'
            onClick={handleChangeMode}
          />
        </div>
        <Button
          onClick={toggle}
          className='px-6 py-3  h-9 border-0'
          style={{
            background:
              'linear-gradient(90deg, rgba(246, 214, 255, 0.3) 0%, rgba(208, 191, 255, 0.3) 100%)',
          }}
          icon={<Icon />}
        >
          <span
            className='text-transparent bg-transparent'
            style={{
              backgroundImage:
                'linear-gradient(90deg, #845EF6 0%, #CC5DE8 100%)',
              backgroundClip: 'text',
            }}
          >
            访问设置
          </span>
        </Button>
      </div>
      <Drawer
        open={open}
        onClose={toggle}
        title='访问设置'
        width={926}
        className={styles.drawer}
      >
        <AgentAcces />
      </Drawer>
    </>
  )
}
