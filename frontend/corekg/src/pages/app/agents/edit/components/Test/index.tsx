import { FC } from 'react'
import { Button, Form, FormInstance } from 'antd'
import { RefreshIcon } from 'tdesign-icons-react'
import { match, P } from 'ts-pattern'
import { cn } from '@/utils'
import { AgentEditValue } from '../..'
import { PromptTest } from './components/PromptTest'
import { RoleTest } from './components/RoleTest'
import { SubmitAgentBtn } from './components/SubmitAgentBtn'

export type { AgentEditValue } from '../..'

export type TestInstance = {
  refresh: () => void
}

export const Test: FC<
  Style & {
    form: FormInstance<AgentEditValue>
  }
> = (props) => {
  const { form } = props
  const [tested, setTested] = useState(false)
  const type = Form.useWatch('type', { form, preserve: true })
  const testInstance = useRef<TestInstance>(null)
  return (
    <div
      className={cn(
        'bg-[#F8F9FB] border border-[#0000001A]',
        'flex flex-col h-full',
        props.className,
      )}
      style={props.style}
    >
      <div className='py-3 px-4 flex items-center bg-white'>
        <div className='text-lg font-medium'>调试与预览</div>
        <Button
          icon={<RefreshIcon />}
          className={cn(
            'ml-auto mr-4',
            ' text-description border-none shadow-none bg-[#EFF0F6]',
          )}
          onClick={() => testInstance.current?.refresh()}
        ></Button>
        <SubmitAgentBtn
          tested={tested}
          className='text-base font-medium'
          getAgentEditValue={async () => {
            await form.validateFields()
            return form.getFieldsValue(true)
          }}
        />
      </div>
      {match(type)
        .with('prompt', () => (
          <PromptTest
            setTested={setTested}
            className='px-4'
            form={form}
            ref={testInstance}
          />
        ))
        .with('workflow', () => (
          <PromptTest
            className='px-4'
            setTested={setTested}
            form={form}
            ref={testInstance}
            workflow
          />
        ))
        .with(P.union('knowledge', 'role_play'), () => (
          <RoleTest
            setTested={setTested}
            className='px-4'
            form={form}
            ref={testInstance}
          />
        ))
        .otherwise(() => null)}
    </div>
  )
}
