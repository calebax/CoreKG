import { FC } from 'react'
import { App, Button, Popover } from 'antd'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import { createSession, createStream } from '@/api/agent'
import { useGlobalSessionHistory } from '@/stores/GlobalSessionHistory'
import ArrowDown from '@/assets/icons/arrow-down.svg?react'
import { DialogInitData } from '../../QA/type'
import SendIcon from '../../images/send.svg'
import { CustomDialogInput } from '../CustomDialogInput'
import { useKnowledge } from '../KnowledgeContext'
import { ModelSelect } from '../ModelSelect'
import { Knowledge, KnowledgeSelect } from './KnowledgeSelect'

export const QAMode: FC<{ hidden?: boolean }> = (props) => {
  const { hidden } = props
  const { message } = App.useApp()
  const navigate = useNavigate()
  const [text, setText] = useState('')
  const [model, setModel] = useState<number>()
  // type===undefined时 以所有知识库进行提问 但是可以继续选择知识库或文件
  const [type, setType] = useState<DialogInitData['type'] | undefined>()
  const [knowledge, setKnowledge] = useState<Knowledge>([])

  const { t } = useTranslation('pages')
  const { add } = useGlobalSessionHistory()

  const { forestList } = useKnowledge()

  const onQA = () => {
    if (!model || !forestList) return
    const toQA = async (search: DialogInitData) => {
      const sessionResult = await createSession({
        base_type: 'standard',
        ids: search.knowledge.map((item) => item.id),
        resource_type: search.type,
        model_id: search.model,
        resource_id: search.model,
      })
      const { ID: session_id } = sessionResult as any
      add({ id: session_id, name: '', nameLoading: true })
      const result = await createStream({
        session_id,
        question: search.text,
      })
      const { question_id } = result
      const searchParams = new URLSearchParams()
      searchParams.append('session_id', session_id)
      searchParams.append('question_id', question_id)
      navigate(`/QA?${searchParams.toString()}`)
    }
    if (!type) {
      if (forestList.length === 0) {
        message.warning(t('app.home.createKnowledgeBase'))
        return
      }
      toQA({
        text,
        model,
        type: 'forest',
        knowledge: forestList,
      })
      return
    }
    if (knowledge.length === 0) {
      message.warning(t('app.home.selectFileOrKb'))
      return
    }
    toQA({
      text,
      model,
      type,
      knowledge,
    })
  }

  /** 选择知识类型 */
  const selectKnowledgeType = (type: DialogInitData['type']) => {
    setKnowledge([])
    setType(type)
    setOpen(true)
    setTypeOpen(false)
  }

  const [typeSelectOpen, setTypeOpen] = useState<boolean>()
  const [knowledgeSelectOpen, setOpen] = useState(false)

  return (
    <CustomDialogInput
      className={cn({ hidden })}
      value={text}
      onChange={setText}
      onSubmit={onQA}
      mode='QA'
    >
      {/* 左侧原有按钮组 */}
      <div className='flex items-center gap-2'>
        <Popover
          arrow={false}
          placement='bottomLeft'
          trigger={[]}
          open={knowledgeSelectOpen}
          styles={{ root: { padding: 0, marginTop: 0 }, body: { padding: 0 } }}
          getPopupContainer={(trigger) =>
            trigger?.parentElement || document.body
          }
          content={
            type ? (
              <KnowledgeSelect
                type={type}
                selectedKnoledge={knowledge}
                onSelectKnowledge={(type, val) => {
                  setType(type)
                  setKnowledge(val)
                }}
                onClose={() => {
                  setOpen(false)
                }}
                onDestory={() => {
                  setKnowledge([])
                  setOpen(false)
                  setTypeOpen(undefined)
                  setTimeout(() => {
                    setType(undefined)
                  }, 0)
                }}
              />
            ) : null
          }
        >
          <Popover
            arrow={false}
            placement='bottom'
            trigger={['hover']}
            open={typeSelectOpen}
            content={
              type ? null : (
                <div
                  className='py-3.5 px-2.5 w-32.5 flex flex-col gap-2 mt-4 rounded-[10px] ml-[-8px]'
                  onClick={(e) => e.stopPropagation()}
                >
                  <Button
                    type='text'
                    size='small'
                    className={cn(
                      'ml-[8px] justify-start px-2 py-1 h-8 text-[#000000E5]/90 font-normal text-base leading-[22px] hover:bg-[#F8F9FD] hover:text-[#1E1F28] hover:font-medium',
                    )}
                    onClick={() => selectKnowledgeType('forest')}
                  >
                    {t('app.home.knowledgeBase')}
                  </Button>
                  <Button
                    type='text'
                    size='small'
                    className='ml-[8px] justify-start px-2 py-1 h-8 text-[#000000E5]/90 font-normal text-base leading-[22px] hover:bg-[##F8F9FD] hover:text-[#1E1F28] hover:font-medium'
                    onClick={() => selectKnowledgeType('file_list')}
                  >
                    {t('app.home.file')}
                  </Button>
                </div>
              )
            }
          >
            <Button
              // icon={<ArrowDown />}
              iconPosition='end'
              className={cn(
                'w-full rounded-[28px] px-2.5',
                ' text-[#653ec4] border-none bg-[#F2F0FF]',
                'hover:bg-[#c0b3ff] hover:text-[#341f68]',
                'focus:bg-[#c0b3ff] focus:text-[#341f68]',
              )}
              onClick={(e) => {
                e.stopPropagation()
                if (type) setOpen((val) => !val)
              }}
            >
              {(() => {
                if (!knowledge.length) return t('app.home.allKnowledgeBase')
                return type === 'file_list'
                  ? t('app.home.linked', { target: t('app.home.file') })
                  : t('app.home.linked', {
                      target: t('app.home.knowledgeBase'),
                    })
              })()}
            </Button>
          </Popover>
        </Popover>
        <ModelSelect value={model} onChange={setModel} />
      </div>

      {/* 右侧发送按钮 */}
      <div className='flex items-center'>
        <div
          className={cn(
            'w-[24px] h-[24px] rounded flex items-center justify-center cursor-pointer transition-colors',
            text.trim() && model && forestList
              ? 'bg-[#1e1f28]'
              : 'bg-[#dfe0eb]',
          )}
          onClick={() => {
            if (text.trim() && model && forestList) {
              onQA()
            } else {
              message.warning(t('app.home.enterTextChooseKbModel'))
            }
          }}
        >
          <div className='relative w-4 h-4 flex items-center justify-center'>
            <img src={SendIcon} alt='send' />
          </div>
        </div>
      </div>
    </CustomDialogInput>
  )
}
