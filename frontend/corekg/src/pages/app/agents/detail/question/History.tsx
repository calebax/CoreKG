import { useState, forwardRef, useImperativeHandle } from 'react'
import { Spin } from 'antd'
import { useMount, useUpdateEffect } from 'ahooks'
import { getReviewQuestion } from '@/api'
import { cn } from '@/utils'

export interface HistoryRef {
  refresh: () => Promise<void>
}

const History = forwardRef<
  HistoryRef,
  {
    sendLoading: boolean
    sessionId: string
    newSessionId: string
    typePath: boolean
    refreshHistoryFlag: boolean
    onSelect: (item: any) => void
  }
>(
  (
    { sendLoading, sessionId, newSessionId, refreshHistoryFlag, onSelect },
    ref,
  ) => {
    const [sessionList, setSessionList] = useState<
      Array<{
        id: number
        requestId: string
        status: string
        input: string
        result: string
      }>
    >([])
    const [loading, setLoading] = useState(false)

    const getData = async () => {
      setLoading(true)
      const res = await getReviewQuestion()
      let list = res.Data || []
      list = list.map((item: any) => {
        return {
          id: item.ID,
          requestId: item.request_id,
          status: item.status,
          input: item.input, // 图片地址
          result: item.cad_results,
        }
      })
      setSessionList(list)
      setLoading(false)
    }

    useMount(() => {
      getData()
    })

    useUpdateEffect(() => {
      getData()
    }, [refreshHistoryFlag])

    useImperativeHandle(ref, () => ({
      refresh: getData,
    }))

    return (
      <div
        className={cn(
          'flex-grow w-full overflow-hidden relative',
          sendLoading && 'pointer-events-none opacity-80',
        )}
      >
        <div className='w-full h-full overflow-y-auto flex flex-col gap-6 p-4'>
          {sessionList.map((item) => {
            return (
              <div
                key={item.id}
                className={cn(
                  'flex-none px-2 h-8 flex items-center gap-1 rounded',
                  Number(sessionId || newSessionId) === Number(item.id)
                    ? 'bg-[#E8F3FF] text-title'
                    : 'text-black/60',
                )}
                onClick={() => {
                  onSelect(item)
                }}
              >
                <span className='flex-grow truncate'>{item.requestId}</span>
                <svg
                  className='flex-none'
                  xmlns='http://www.w3.org/2000/svg'
                  width='16'
                  height='16'
                  viewBox='0 0 16 16'
                  fill='none'
                >
                  <path
                    d='M13.3333 8.33073C13.4217 8.33073 13.5064 8.29561 13.569 8.2331C13.6315 8.17059 13.6666 8.0858 13.6666 7.9974C13.6666 7.90899 13.6315 7.82421 13.569 7.76169C13.5064 7.69918 13.4217 7.66406 13.3333 7.66406C13.2448 7.66406 13.1601 7.69918 13.0975 7.76169C13.035 7.82421 12.9999 7.90899 12.9999 7.9974C12.9999 8.0858 13.035 8.17059 13.0975 8.2331C13.1601 8.29561 13.2448 8.33073 13.3333 8.33073ZM7.99992 8.33073C8.08832 8.33073 8.17311 8.29561 8.23562 8.2331C8.29813 8.17059 8.33325 8.0858 8.33325 7.9974C8.33325 7.90899 8.29813 7.82421 8.23562 7.76169C8.17311 7.69918 8.08832 7.66406 7.99992 7.66406C7.91151 7.66406 7.82673 7.69918 7.76422 7.76169C7.7017 7.82421 7.66658 7.90899 7.66658 7.9974C7.66658 8.0858 7.7017 8.17059 7.76422 8.2331C7.82673 8.29561 7.91151 8.33073 7.99992 8.33073ZM2.66659 8.33073C2.75499 8.33073 2.83978 8.29561 2.90229 8.2331C2.9648 8.17059 2.99992 8.0858 2.99992 7.9974C2.99992 7.90899 2.9648 7.82421 2.90229 7.76169C2.83978 7.69918 2.75499 7.66406 2.66659 7.66406C2.57818 7.66406 2.4934 7.69918 2.43088 7.76169C2.36837 7.82421 2.33325 7.90899 2.33325 7.9974C2.33325 8.0858 2.36837 8.17059 2.43088 8.2331C2.4934 8.29561 2.57818 8.33073 2.66659 8.33073Z'
                    fill='black'
                    stroke='black'
                    strokeWidth='1.25'
                    strokeLinecap='round'
                    strokeLinejoin='round'
                  />
                </svg>
              </div>
            )
          })}
        </div>

        {sessionId && loading && (
          <div className='absolute inset-0 flex items-center justify-center bg-white/50'>
            <Spin />
          </div>
        )}
      </div>
    )
  },
)

export default History
