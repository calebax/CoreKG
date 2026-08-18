import { useState, useRef } from 'react'
import { Button, message } from 'antd'
import { cn } from '@/utils'

export default function CreateCard({ onSend }) {
  const inputRef = useRef(null)
  const [file, setFile] = useState(null)
  const [previewUrl, setPreviewUrl] = useState(null)

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) {
      setFile(null)
      setPreviewUrl(null)
      return
    }

    // 检查文件类型是否为图片
    if (!file.type.startsWith('image/')) {
      message.error('请选择图片文件')
      return
    }

    setFile(file)
    // 创建预览URL
    const url = URL.createObjectURL(file)
    setPreviewUrl(url)
  }

  // 清除选择的图片
  const handleClearImage = () => {
    setFile(null)
    setPreviewUrl(null)
    if (inputRef.current) {
      inputRef.current.value = ''
    }
  }

  const [submitLoading, setSubmitLoading] = useState(false)
  const handleSubmit = async () => {
    if (!file) {
      message.error('请先选择图片')
      return
    }

    try {
      setSubmitLoading(true)
      const data = {
        file: file,
        previewUrl: previewUrl,
      }
      onSend(data)
    } catch (error) {
      console.error('Validate Failed:', error)
    } finally {
      setSubmitLoading(false)
    }
  }
  return (
    <div className='max-w-[620px] w-full overflow-auto'>
      <div
        className={cn(
          'w-full min-h-60 border rounded-2xl border-dotted overflow-hidden flex items-center justify-center gap-10 cursor-pointer',
          submitLoading && 'pointer-events-none opacity-80',
        )}
        onClick={() => inputRef.current?.click()}
      >
        {previewUrl ? (
          <div className='relative w-full h-full'>
            <img
              src={previewUrl}
              alt='预览图片'
              className='w-full h-full object-contain rounded-lg'
            />
            <button
              onClick={(e) => {
                e.stopPropagation()
                handleClearImage()
              }}
              className='absolute top-2 right-2 bg-red-500 text-white rounded-full w-6 h-6 flex items-center justify-center text-sm hover:bg-red-600'
            >
              ×
            </button>
          </div>
        ) : (
          <span>选择图片</span>
        )}
      </div>

      <input
        type='file'
        className='hidden'
        ref={inputRef}
        onChange={handleFileChange}
        accept='image/*'
      />

      <Button
        type='primary'
        className='w-full mt-14 h-11! flex items-center justify-center gap-2'
        loading={submitLoading}
        onClick={handleSubmit}
      >
        <span className='text-base font-medium'>立即生成</span>
      </Button>
    </div>
  )
}
