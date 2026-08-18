import { useState, useRef } from 'react'
import { uploadImage } from '@/api/common'

interface UploadCoverProps {
  value: string
  onChange: (value: string) => void
}

export default function UploadCover({ value, onChange }: UploadCoverProps) {
  const [loading, setLoading] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const [progress, setProgress] = useState(0)

  const handleUpload = async () => {
    inputRef.current?.click()
  }

  const handleChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setLoading(true)
    try {
      const data = {
        file,
        purpose: 'yg-chat',
      }
      const res = await uploadImage(data, {
        timeout: 0,
        headers: { 'Content-Type': 'multipart/form-data' },
        onUploadProgress: (e: any) => {
          if (e.total > 0) {
            const percent = Number(((e.loaded / e.total) * 100).toFixed(2))
            if (percent >= 99) {
              setProgress(99)
            } else {
              setProgress(percent)
            }
          }
        },
      })
      if (res?.url) {
        onChange(res.url)
      }
    } catch (error) {
      console.error('Upload failed:', error)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className='w-58 aspect-[234/131] bg-[#E5E9EF] rounded-md mx-auto relative overflow-hidden'>
      {value && (
        <img src={value} alt='cover' className='w-full h-full object-cover' />
      )}
      <div
        className='absolute bottom-1 left-1/2 -translate-x-1/2 rounded bg-[#165DFF82] backdrop-blur-[2px] px-2.5 py-1 text-white font-semibold cursor-pointer'
        onClick={handleUpload}
      >
        {loading ? '上传中...' : value ? '修改封面' : '上传封面'}
      </div>

      <input
        ref={inputRef}
        type='file'
        className='hidden!'
        accept='image/*'
        onChange={handleChange}
      />
    </div>
  )
}
