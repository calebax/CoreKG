import { App, Button, Form, FormInstance } from 'antd'
import { cn } from '@/utils'
import { AIDialog, UserDialog } from '@/components/dialog'
import { AgentEditValue, TestInstance } from '../..'
import { useEditContext } from '../../../AgentContext'
import styles from '../../styles.module.scss'
import { Param } from './Param'
import { useDialog } from './useDialog'

export const PromptTest = forwardRef<
  TestInstance,
  Style & {
    form: FormInstance<AgentEditValue>
    workflow?: boolean
    setTested: (val: boolean) => void
  }
>((props, ref) => {
  const { message } = App.useApp()
  const { agent } = useEditContext()
  const { form, workflow, setTested } = props
  const formParams = Form.useWatch('params', form)
  const params = workflow ? agent.params : formParams
  // 按照key关联param
  const [inputForm] = Form.useForm()
  const { dialog, setDialog, startQA, isAnswering } = useDialog()
  const container = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const dom = container.current
    if (!dom) return
    dom.scrollTo({
      top: dom.scrollHeight,
      behavior: 'smooth',
    })
  }, [dialog])

  const startTest = async () => {
    try {
      const data = workflow
        ? null
        : await form.validateFields([
            'chat_models',
            'id',
            'prompt_template',
            'params',
            'temperature',
          ])
      const _input = await inputForm.validateFields()
      const input: { name: string; value: string }[] = []
      Object.entries(_input).forEach((item) => {
        const [key, value] = item
        const par = params.find((item) => item.key === key)
        if (!par) return
        input.push({
          name: par.input,
          value: value as string,
        })
      })
      startQA(
        {
          ...data,
          chat_model_ids: data?.chat_models?.map((item) => item.id),
          input,
        },
        {
          onEnd: () => setTested(true),
        },
      )
    } catch {
      message.warning('请完善智能体配置')
    }
  }

  const refresh = () => {
    if (isAnswering) return
    setDialog([])
    inputForm.resetFields()
  }
  useImperativeHandle(ref, () => ({
    refresh,
  }))
  return (
    <div
      className={cn('flex-1 overflow-hidden', 'flex flex-col', props.className)}
      style={props.style}
    >
      <div
        className={cn(
          'flex-1 overflow-auto pt-2 pr-2',
          'flex flex-col gap-2',
          styles.scroll,
        )}
        ref={container}
      >
        {params && params.length > 0 ? (
          <Form
            preserve={false}
            form={inputForm}
            layout='vertical'
            className='flex flex-col gap-2 mb-2'
          >
            <span className='text-base text-title'>配置参数</span>
            {params.map((item) => {
              const { name, input, is_required } = item
              const label = (
                <span className='text-title flex gap-2 items-center'>
                  <span>{name}</span>
                  <span className={cn('text-description')}>({input})</span>
                </span>
              )
              return (
                <Form.Item
                  key={item.key ?? item.name}
                  name={item.key}
                  label={label}
                  rules={[
                    { required: is_required, message: '这个变量是必填的' },
                  ]}
                  className='m-0'
                >
                  <Param param={item}></Param>
                </Form.Item>
              )
            })}
          </Form>
        ) : null}
        {dialog.map((item, i) => {
          switch (item.role) {
            case 'question':
              return <UserDialog key={i} value={item} className='self-end' />
            case 'answer':
              return <AIDialog key={i} value={item} className='' />
          }
        })}
      </div>
      <div className='p-4 flex justify-center'>
        <Button
          className=' mx-auto w-139'
          type='primary'
          loading={isAnswering}
          onClick={startTest}
        >
          测试
        </Button>
      </div>
    </div>
  )
})
