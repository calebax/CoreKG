import { Button } from 'antd'
import { useTranslation } from 'react-i18next'
import Icon from '@/assets/react.svg?react'
import Copy from '@/components/common/Copy'

export default function Test() {
  const { t } = useTranslation('pages')
  const [count, setCount] = useState(0)

  useEffect(() => {
    console.log('count', count)
  }, [count])

  return (
    <div className='w-screen h-screen flex flex-col items-center justify-center'>
      <Button type='primary' onClick={() => setCount(count + 1)}>
        {t('other.click')}
      </Button>
      <p>{count}</p>
      <Icon />

      <Copy text={count} className='px-2 py-1 border rounded' />
    </div>
  )
}
