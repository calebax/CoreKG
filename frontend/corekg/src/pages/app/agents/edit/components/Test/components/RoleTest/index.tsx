import { App, FormInstance } from 'antd'
import { cn } from '@/utils'
import { AIDialog, UserDialog } from '@/components/dialog'
import SendIcon from '@/pages/app/home/images/send.svg'
import { AgentEditValue, TestInstance } from '../..'
import styles from '../../styles.module.scss'
import { HomeInput } from './HomeInput'
import { useDialog } from './useDialog'

export const RoleTest = forwardRef<
  TestInstance,
  Style & {
    form: FormInstance<AgentEditValue>
    setTested: (val: boolean) => void
  }
>((props, ref) => {
  const { message } = App.useApp()
  const { form, setTested } = props

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

  const [search, setSearch] = useState('')
  const startTest = async () => {
    form
      .validateFields([
        'chat_models',
        'id',
        'prompt_template',
        'greeting_message',
        'forests',
        'temperature',
      ])
      .then(
        (data) => {
          startQA(
            {
              ...data,
              chat_model_ids: data.chat_models!.map((item) => item.id),
              forest_ids: data.forests?.map((item) => item.id),
              question: search,
            },
            {
              onEnd: () => setTested(true),
            },
          )
          setSearch('')
        },
        () => {
          message.warning('请完善智能体配置')
        },
      )
  }

  const refresh = useCallback(() => {
    if (isAnswering) return
    setDialog([])
    setSearch('')
  }, [isAnswering, setDialog])
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
        {dialog.map((item, i) => {
          switch (item.role) {
            case 'question':
              return <UserDialog key={i} value={item} className='self-end' />
            case 'answer':
              return (
                <AIDialog
                  key={i}
                  value={item}
                  showReference={item.withKnowledge}
                />
              )
          }
        })}
      </div>
      <div className='p-4 flex-shrink-0'>
        <HomeInput
          className='w-auto'
          value={search}
          onChange={setSearch}
          onSubmit={startTest}
        >
          {isAnswering || !search.trim() ? (
            <div className='w-[24px] h-[24px] rounded flex items-center justify-center bg-[#dfe0eb] cursor-not-allowed ml-auto'>
              <div className='relative w-4 h-4 flex items-center justify-center'>
                <img src={SendIcon} alt='send' />
              </div>
            </div>
          ) : (
            <div
              className='w-[24px] h-[24px] rounded flex items-center justify-center bg-[#1e1f28] cursor-pointer transition-colors hover:bg-[#2a2b36] ml-auto'
              onClick={startTest}
            >
              <div className='relative w-4 h-4 flex items-center justify-center'>
                <img src={SendIcon} alt='send' />
              </div>
            </div>
          )}
        </HomeInput>
      </div>
    </div>
  )
})
