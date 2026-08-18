import { useTranslation } from 'react-i18next'

export default function Home() {
  const { t } = useTranslation('pages')
  return (
    <div className='w-full h-full flex flex-col items-center justify-center'>
      <h1>{t('profile.purchaseRecord')}</h1>
    </div>
  )
}
