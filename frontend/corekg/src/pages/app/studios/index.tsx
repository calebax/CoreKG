import { useTranslation } from 'react-i18next'
import Alert from '@/components/common/Alert'

export default function Home() {
  const { t } = useTranslation('pages')

  useEffect(() => {
    console.log('Home组件已加载')
  }, [])

  return (
    <div className='w-full h-full p-4'>
      <Alert message={t('app.studios.agentHandleAppDataEditExport')} />
    </div>
  )
}
