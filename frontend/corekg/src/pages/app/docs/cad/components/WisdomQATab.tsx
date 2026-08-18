import { useState, useEffect, useRef } from 'react'
import { useParams } from 'react-router-dom'
import { Input, message, Spin } from 'antd'
import { SendOutlined, DeleteOutlined } from '@ant-design/icons'
import { useBoolean, useRequest } from 'ahooks'
import { getFileInfo, /* releaseChat, */ deleteFileQA } from '@/api/knowledge'
import { AIDialog, UserDialog } from '@/components/dialog'
import DeleteConfirmModal from './DeleteConfirmModal'
import { RecommendQuestions } from './RecommendQuestions'
import { useFileDialog } from './useFileDialog'

const { TextArea } = Input

export default function WisdomQATab() {
  const params = useParams<{ id: string; fileId: string }>()
  const file_id = Number(params.fileId)
  const forest_id = Number(params.id)
  const knowledgeStatus = useRequest(
    async () => {
      const { knowledge_status } = await getFileInfo({
        file_id,
      })
      return knowledge_status
    },
    { refreshDeps: [file_id] },
  )

  const [showDeleteModal, { toggle: toggleModal }] = useBoolean(false)

  const [text, setText] = useState<string>('')
  const container = useRef<HTMLDivElement>(null)
  const { historyLoading, isAnswering, dialog, setDialog, startQA } =
    useFileDialog(file_id, forest_id)
  useEffect(() => {
    const dom = container.current
    if (!dom) return
    dom.scrollTo({
      top: dom.scrollHeight,
      behavior: 'smooth',
    })
  }, [dialog])

  const onQA = () => {
    startQA({ text })
    setText('')
    setHidden(true)
  }

  const loading = historyLoading || knowledgeStatus.loading || isAnswering

  // 添加滚动条样式
  useEffect(() => {
    const style = document.createElement('style')
    style.textContent = `
      .custom-scrollbar::-webkit-scrollbar {
        width: 6px;
      }
      .custom-scrollbar::-webkit-scrollbar-track {
        background: transparent;
      }
      .custom-scrollbar::-webkit-scrollbar-thumb {
        background-color: #cbd5e0;
        border-radius: 3px;
      }
    `
    document.head.appendChild(style)

    return () => {
      document.head.removeChild(style)
    }
  }, [])

  // 确认清空历史记录
  const handleConfirmClearHistory = async () => {
    try {
      await deleteFileQA({
        file_id,
        forest_id,
      })
      setDialog([])
      message.success('历史记录已清空')
    } catch (error) {
      message.error('清空失败，请重试')
      console.error('清空历史记录失败', error)
    } finally {
      toggleModal()
    }
  }
  const [hidden, setHidden] = useState(false)
  return (
    <div className='h-full flex flex-col gap-3 overflow-hidden relative'>
      <RecommendQuestions
        file_id={file_id}
        className='top-0 left-2'
        onSelectQusetion={setText}
        hidden={hidden}
        setHidden={setHidden}
      />
      {/* 聊天记录区域 */}
      <div
        className='flex-1 overflow-y-auto custom-scrollbar'
        ref={container}
        style={{
          scrollbarWidth: 'thin',
          scrollbarColor: '#cbd5e0 transparent',
        }}
      >
        {dialog.length === 0 && knowledgeStatus.data === 'success' ? (
          <div className='flex flex-col items-center justify-center h-full text-gray-500'>
            <div className='text-center max-w-md'>
              <div className='w-16 h-16 bg-blue-50 rounded-full flex items-center justify-center mx-auto mb-4'>
                <div className='w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center'>
                  <span className='text-blue-600 font-medium text-sm'>AI</span>
                </div>
              </div>
              <h3 className='text-lg font-medium text-gray-800 mb-2'>
                开始智慧问答
              </h3>
              <p className='text-sm text-gray-500 leading-relaxed'>
                在下方输入框中提出您的问题
              </p>
            </div>
          </div>
        ) : (
          <div className='py-6 space-y-6'>
            <div className='flex flex-col gap-2 pr-2'>
              {dialog.map((item) => {
                if (item.role === 'question') return <UserDialog value={item} />
                return <AIDialog value={item} showReference={true} />
              })}
            </div>
            {knowledgeStatus.data !== 'success' && (
              <div className='flex flex-col items-center justify-center text-gray-500 py-20'>
                <Spin size='large' className='mb-4' />
                <div className='text-base font-medium text-gray-800'>
                  知识库正在构建中
                </div>
                <div className='text-sm text-gray-400 mt-2'>
                  请稍后再来提问吧
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {/* 输入区域 */}
      <div className='flex-shrink-0'>
        <div className='max-w-4xl mx-auto'>
          <div className='relative flex items-center gap-3'>
            {/* 清空历史记录按钮 */}
            {dialog.length > 0 && (
              <button
                onClick={toggleModal}
                disabled={loading}
                className='flex-shrink-0 w-12 h-12 text-gray-600 hover:text-red-500 rounded-full flex items-center justify-center transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed border border-gray-200 hover:border-red-300 cursor-pointer'
                title='清空历史记录'
              >
                <DeleteOutlined className='text-lg' />
              </button>
            )}

            <div className='flex-1 relative'>
              <TextArea
                value={text}
                onChange={(e) => setText(e.target.value)}
                onPressEnter={onQA}
                autoSize={{ minRows: 1, maxRows: 4 }}
                className='text-sm resize-none pr-14'
                disabled={loading || knowledgeStatus.data !== 'success'}
                style={{
                  borderRadius: '999px',
                  paddingLeft: '20px',
                  paddingRight: '56px',
                  paddingTop: '20px',
                  paddingBottom: '20px',
                  fontSize: '16px',
                  lineHeight: '22px',
                  minHeight: '80px',
                  borderWidth: '1px',
                  borderColor: '#0052D9',
                  boxShadow: 'none',
                }}
              />
              <button
                onClick={onQA}
                disabled={
                  !text.trim() || loading || knowledgeStatus.data !== 'success'
                }
                className={`absolute right-2 top-1/2 transform -translate-y-1/2 w-12 h-12 rounded-full flex items-center justify-center transition-all duration-200 shadow-sm ${loading ? 'bg-[#0052D9] opacity-50 text-white cursor-not-allowed' : 'bg-[#0052D9] hover:bg-[#003db8] cursor-pointer text-white disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:bg-[#0052D9]'}`}
              >
                {loading ? (
                  <div className='w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin'></div>
                ) : (
                  <SendOutlined className='text-xl rotate-270' />
                )}
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* 删除确认弹窗 */}
      <DeleteConfirmModal
        visible={showDeleteModal}
        customTitle='确认清空历史记录'
        customText='清空后，所有历史聊天记录将无法恢复，请谨慎操作。'
        customOkText='确认清空'
        onCancel={toggleModal}
        onConfirm={handleConfirmClearHistory}
      />
    </div>
  )
}
