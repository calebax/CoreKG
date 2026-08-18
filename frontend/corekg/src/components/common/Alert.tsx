import { useState } from 'react'
import { cn } from '@/utils'
import AlertMessageClose from '@/assets/icons/alert-message-close.svg?react'

interface AlertProps {
  message: string
  className?: string
  onClose?: () => void
}

export default function Alert({ message, className, onClose }: AlertProps) {
  const [isClose, setIsClose] = useState(false)

  const handleClose = () => {
    setIsClose(true)
    onClose?.()
  }

  return (
    <div
      className={cn(
        'w-full bg-[#165DFF1F] border border-[#165DFF] rounded-md flex items-center px-4 py-5 gap-2',
        className,
        isClose && 'hidden',
      )}
    >
      <svg
        className='flex-none'
        xmlns='http://www.w3.org/2000/svg'
        width='24'
        height='24'
        viewBox='0 0 24 24'
        fill='none'
      >
        <path
          d='M12.0002 23C18.0754 23 23.0002 18.0752 23.0002 12C23.0002 5.9249 18.0754 1.00003 12.0002 1.00003C5.92511 1.00003 1.00024 5.9249 1.00024 12C1.00024 18.0752 5.92511 23 12.0002 23ZM7.5001 10.5858L10.5001 13.5858L16.5001 7.58581L17.9143 9.00002L10.5001 16.4142L6.08588 12L7.5001 10.5858Z'
          fill='#5769FF'
        />
      </svg>
      <span className='flex-grow text-[#0A1A3A] text-base font-normal'>
        {message}
      </span>
      <div
        className='flex-none w-6 h-6 p-0.5 rounded-full bg-[#F8F7FF] cursor-pointer hover:bg-[#F8F7FF]/80'
        onClick={handleClose}
      >
        <AlertMessageClose className='w-5 h-5' />
      </div>
    </div>
  )
}
