import { FC } from 'react'
import { cn } from '@/utils'
import SendIcon from '../../images/send.svg'

interface SendButtonProps {
  active?: boolean
  onClick?: () => void
  className?: string
}

export const SendButton: FC<SendButtonProps> = ({
  active = false,
  onClick,
  className,
}) => {
  return (
    <div
      className={cn(
        'w-[24px] h-[24px] rounded flex items-center justify-center cursor-pointer transition-colors',
        active ? 'bg-[#1e1f28]' : 'bg-[#dfe0eb]',
        className,
      )}
      onClick={onClick}
    >
      <div className='relative w-4 h-4 flex items-center justify-center'>
        <img src={SendIcon} alt='send' />
      </div>
    </div>
  )
}
