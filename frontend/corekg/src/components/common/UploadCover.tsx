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
      console.log('upload res', res)
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
    <div className='w-30 h-30 bg-[#E5E9EF] rounded-[20px] mx-auto relative overflow-hidden'>
      {value && (
        <img src={value} alt='cover' className='w-full h-full object-cover' />
      )}
      <div
        className='absolute bottom-3.5 left-4 right-4 rounded bg-[#FAFEFF] text-center py-1 text-primary font-semibold cursor-pointer'
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
