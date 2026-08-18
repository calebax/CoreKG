import { useState, useMemo } from 'react'
import { useParams, useSearchParams, useNavigate } from 'react-router-dom'
import { Button, Input, message, Spin } from 'antd'
import { useMount, useUpdateEffect } from 'ahooks'
import { getAgentDetail, getAgentInfo, reviewQuestion } from '@/api'
import { cn, isJsonString } from '@/utils'
import MessageIcon from '@/assets/icons/message.svg?react'
import MarkdownPreview from '@/components/common/MarkdownPreview'
import useLocalStore from '@/stores/local'
import BaseInfo from './BaseInfo'
import CreateCard from './CreateCard'
import History from './History'

export default function EditAgent() {
  const userInfo = useLocalStore((state) => state.userInfo)
  const { id } = useParams()
  const agentId = useMemo(() => Number(id), [id])
  const [searchParams] = useSearchParams()
  const sessionId = searchParams.get('sessionId')
  const newSessionId = searchParams.get('newSessionId')

  const getDetail = async () => {
    const res = await getAgentInfo(agentId)
    const data = {
      agentId: agentId,
      showName: res.show_name,
      name: res.name,
      description: res.description,
      avatar: res.avatar_url,
      greetingMessage: res.greeting_message,
      promptTemplate: res.prompt_template,
      publicScope: res.public_scope,
      temperature: res.temperature,
      publishStatus: res.publish_status,
    }
    setAgentDetail(data)
    return data
  }

  useMount(() => {
    getDetail()
  })

  const [list, setList] = useState([])
  const [sendLoading, setSendLoading] = useState(false)
  const chatContainerRef = useRef<HTMLDivElement>(null)
  const [refreshHistoryFlag, setRefreshHistoryFlag] = useState(false)
  const [agentDetail, setAgentDetail] = useState({
    id: '',
    showName: '',
    name: '',
    description: '',
    avatar: '',
    greetingMessage: '',
    promptTemplate: '',
    publicScope: '',
    temperature: '',
    publishStatus: '',
    params: [],
  })

  const [currentSession, setCurrentSession] = useState(null)
  const handleCreate = () => {
    setCurrentSession(null)
  }

  const updateMessage = (index: number, data: any) => {
    if (data.code && data.code !== 0) {
      message.error(data.message)
    } else {
      if (data.content) {
        // update message
        setList((prev) => {
          const newList = [...prev]
          newList[index] = {
            ...newList[index],
            content: prev[index].content + data.content,
          }
          return newList
        })
      }
    }
  }

  // 移除强制滚动到底部的逻辑，让用户可以自由查看历史消息
  // useUpdateEffect(() => {
  //   scrollToEnd()
  // }, [list])

  const handleSendMessage = async (data) => {
    const currentIndex = 1
    setList([
      {
        role: 'question',
        content: data.previewUrl,
      },
      {
        role: 'answer',
        content: '',
      },
    ])
    setCurrentSession(data)
    setSendLoading(true)
    try {
      const response = await reviewQuestion({
        file: data.file,
      })

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      while (true) {
        const { done, value } = await reader.read()
        if (done) {
          setSendLoading(false)
          break
        }
        const text = decoder.decode(value, { stream: true })

        if (isJsonString(text)) {
          updateMessage(currentIndex, JSON.parse(text))
        } else {
          // const matches = text.match(/\{"content":\s*"[^"]*"\}/g)
          const matches = text.match(/({.*?})/g)
          if (matches && matches.length) {
            matches.forEach((match) => {
              if (isJsonString(match)) {
                updateMessage(currentIndex, JSON.parse(match))
              } else {
                console.error('error')
              }
            })
          }
        }
      }
    } finally {
      setSendLoading(false)
    }
  }

  const scrollToEnd = () => {
    if (chatContainerRef.current) {
      chatContainerRef.current.scrollTo({
        top: chatContainerRef.current.scrollHeight,
        behavior: 'smooth',
      })
    }
  }

  const handleSelect = (item) => {
    setCurrentSession(item)
    setList([
      {
        role: 'question',
        content: item.input,
      },
      {
        role: 'answer',
        content: item.result,
      },
    ])
  }

  return (
    <div
      className='w-full h-full pr-4 bg-white
    flex
  '
    >
      {/* left menu */}
      <div className='flex-none w-[210px] h-full border border-[#CCE9FF] flex flex-col gap-2 overflow-hidden'>
        <div className='w-full p-4 pb-0 flex flex-col gap-6'>
          <BaseInfo agentDetail={agentDetail} />

          <Button
            color='primary'
            variant='filled'
            className='w-full'
            onClick={() => {
              handleCreate()
            }}
          >
            <MessageIcon className='w-6 h-6' />
            <svg
              xmlns='http://www.w3.org/2000/svg'
              width='18'
              height='18'
              viewBox='0 0 18 18'
              fill='none'
            >
              <path
                d='M9.60252 1.76855H8.39716C8.29002 1.76855 8.23645 1.82213 8.23645 1.92927V8.2373H2.25031C2.14317 8.2373 2.0896 8.29088 2.0896 8.39802V9.60338C2.0896 9.71052 2.14317 9.76409 2.25031 9.76409H8.23645V16.0721C8.23645 16.1793 8.29002 16.2328 8.39716 16.2328H9.60252C9.70967 16.2328 9.76324 16.1793 9.76324 16.0721V9.76409H15.7503C15.8575 9.76409 15.911 9.71052 15.911 9.60338V8.39802C15.911 8.29088 15.8575 8.2373 15.7503 8.2373H9.76324V1.92927C9.76324 1.82213 9.70967 1.76855 9.60252 1.76855Z'
                fill='#165DFF'
              />
            </svg>
            <span>新建会话</span>
          </Button>
          <div className='w-full flex items-center gap-1 px-2'>
            <svg
              xmlns='http://www.w3.org/2000/svg'
              width='18'
              height='18'
              viewBox='0 0 18 18'
              fill='none'
            >
              <path
                d='M9 15.75C7.275 15.75 5.772 15.1782 4.491 14.0347C3.21 12.8912 2.4755 11.463 2.2875 9.75H3.825C4 11.05 4.57825 12.125 5.55975 12.975C6.54125 13.825 7.688 14.25 9 14.25C10.4625 14.25 11.7032 13.7407 12.7222 12.7222C13.7412 11.7037 14.2505 10.463 14.25 9C14.2495 7.537 13.7402 6.2965 12.7222 5.2785C11.7042 4.2605 10.4635 3.751 9 3.75C8.1375 3.75 7.33125 3.95 6.58125 4.35C5.83125 4.75 5.2 5.3 4.6875 6H6.75V7.5H2.25V3H3.75V4.7625C4.3875 3.9625 5.16575 3.34375 6.08475 2.90625C7.00375 2.46875 7.9755 2.25 9 2.25C9.9375 2.25 10.8158 2.42825 11.6347 2.78475C12.4537 3.14125 13.1663 3.62225 13.7723 4.22775C14.3783 4.83325 14.8595 5.54575 15.216 6.36525C15.5725 7.18475 15.7505 8.063 15.75 9C15.7495 9.937 15.5715 10.8152 15.216 11.6347C14.8605 12.4542 14.3793 13.1668 13.7723 13.7723C13.1653 14.3778 12.4527 14.859 11.6347 15.216C10.8167 15.573 9.9385 15.751 9 15.75ZM11.1 12.15L8.25 9.3V5.25H9.75V8.7L12.15 11.1L11.1 12.15Z'
                fill='#1D2129'
              />
            </svg>
            <span className='text-title font-medium'>历史会话</span>
          </div>
        </div>

        <History
          sessionId={sessionId}
          sendLoading={sendLoading}
          typePath='question'
          refreshHistoryFlag={refreshHistoryFlag}
          onSelect={handleSelect}
        />
      </div>

      {/* right content */}
      {currentSession ? (
        <div className='flex-grow max-w-[800px] mx-auto h-full bg-[#F8FCFF] flex flex-col relative'>
          <div
            className='flex-none w-full py-7 flex items-center bg-white px-6'
            style={{
              boxShadow: '0px 4px 9.9px 0px rgba(0, 0, 0, 0.01)',
            }}
          >
            <h1 className='flex-grow text-title font-bold text-[28px] text-center'>
              {agentDetail.showName}
            </h1>
          </div>
          <div
            ref={chatContainerRef}
            className='flex-grow w-full p-4 overflow-y-auto'
          >
            {list.map((item, index) => (
              <div
                key={index}
                className={cn(
                  'flex gap-5 mb-10',
                  item.role === 'question' &&
                    'flex-row-reverse justify-start text-black/90 pl-13',
                  item.role === 'answer' && 'flex-row text-black/60 pr-13',
                )}
              >
                <img
                  className='flex-none w-8 h-8 rounded-full object-cover'
                  src={
                    item.role === 'question'
                      ? userInfo.avatar
                      : agentDetail.avatar
                  }
                  alt=''
                />
                <div
                  className={cn(
                    'pt-1',
                    item.role === 'answer'
                      ? 'bg-white rounded-lg py-3 px-4'
                      : 'text-[#00000099]',
                  )}
                >
                  {item.role === 'question' ? (
                    <img
                      src={item.content}
                      alt=''
                      className='w-full h-full object-cover'
                    />
                  ) : (
                    <>
                      {sendLoading && item.content === '' ? (
                        <Spin />
                      ) : (
                        <MarkdownPreview content={item.content} disableReference={true} />
                      )}
                    </>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      ) : (
        <div
          className='flex-grow h-full bg-[#F6FCFF] flex flex-col items-center justify-center p-6'
          style={{
            background:
              'linear-gradient(90deg, #F7F7FF 0.03%, #F5FCFF 101.25%)',
          }}
        >
          <h1 className='text-title text-center font-bold text-[35px] mb-14'>
            {agentDetail.showName}
          </h1>
          <CreateCard onSend={handleSendMessage} />
        </div>
      )}
    </div>
  )
}
