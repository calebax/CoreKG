import { FC } from 'react'
import { useTranslation } from 'react-i18next'
import emptyStateIcon from '@/assets/icons/EmptyState.svg'

interface EmptyStateProps {
  message?: string
}

const EmptyState: FC<EmptyStateProps> = ({ message }) => {
  const { t } = useTranslation('common')
  const defaultMessage = t('empty.noFind')
  const displayMessage = message || defaultMessage
  return (
    <div className='flex flex-col items-center justify-center h-full text-center text-gray-500'>
      <img
        src={emptyStateIcon}
        alt={t('empty.emptyState')}
        className='w-40 h-40 mb-3'
      />
      <p className='text-xl text-[#616373] font-normal'>{displayMessage}</p>
    </div>
  )
}

export default EmptyState
