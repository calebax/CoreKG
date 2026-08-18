import { useEffect, useState } from 'react'
import { Button, Empty, Spin } from 'antd'
import { useTranslation } from 'react-i18next'
import apiKeyQuantity from '@/assets/icons/usage-quantity.svg'
import PurchaseModal from './components/PurchaseModal'

interface UsageItem {
  id: string
  title: string
  currentQuota: number
  historicalUsage: number
}

export default function Usage() {
  const { t } = useTranslation('pages')
  // 状态管理
  const [loading, setLoading] = useState<boolean>(false)
  const [usageList, setUsageList] = useState<UsageItem[]>([])
  const [modalOpen, setModalOpen] = useState(false)
  const [currentCard, setCurrentCard] = useState<UsageItem | null>(null)

  // 模拟接口获取数据
  const fetchUsageData = async () => {
    try {
      setLoading(true)
      // 这里是接口调用的位置，暂时用模拟数据
      // const response = await fetch('/api/usage')
      // const data = await response.json()
      // setUsageList(data)

      // 模拟数据
      setTimeout(() => {
        const mockData: UsageItem[] = [
          {
            id: '1',
            title: t('profile.generalQAMatching'),
            currentQuota: 25,
            historicalUsage: 850,
          },
          {
            id: '2',
            title: t('profile.intelligentCustomerServiceSupport'),
            currentQuota: 15,
            historicalUsage: 620,
          },
          {
            id: '3',
            title: t('profile.faqAutoReply'),
            currentQuota: 40,
            historicalUsage: 310,
          },
          {
            id: '4',
            title: t('profile.documentContentRetrieval'),
            currentQuota: 10,
            historicalUsage: 1200,
          },
          {
            id: '5',
            title: t('profile.knowledgeGraphQuery'),
            currentQuota: 5,
            historicalUsage: 95,
          },
          {
            id: '6',
            title: t('profile.voiceQARecognition'),
            currentQuota: 20,
            historicalUsage: 470,
          },
        ]
        setUsageList(mockData)
        setLoading(false)
      }, 500)
    } catch (error) {
      console.error('获取用量数据失败', error)
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchUsageData()
  }, [])

  const handleBuyClick = (item: UsageItem) => {
    setCurrentCard(item)
    setModalOpen(true)
  }

  return (
    <div className='w-full h-full flex flex-col p-4'>
      {/* 账号余量按钮区域 - 固定在顶部 */}
      <div className='w-full flex-none flex items-center mb-5'>
        <Button
          type='primary'
          className='!rounded flex items-center justify-center !gap-2.5 !py-4.5 hover:!bg-[#165DFF] hover:!text-white'
        >
          <img
            src={apiKeyQuantity}
            alt={t('profile.accountBalance')}
            className='w-5 h-5'
          />
          <span className='text-base'>{t('profile.accountBalance')}</span>
        </Button>
      </div>

      {/* 内容区域 - 可滚动 */}
      <div className='flex-grow w-full overflow-auto'>
        {loading ? (
          <div className='w-full h-full flex items-center justify-center'>
            <Spin size='large' />
          </div>
        ) : usageList.length === 0 ? (
          <div className='w-full h-full flex items-center justify-center'>
            <Empty description={t('profile.noUsageDataForNow')} />
          </div>
        ) : (
          <div className='w-full grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6'>
            {usageList.map((item) => (
              <div
                key={item.id}
                className='bg-white rounded-xl p-4 flex flex-col'
                style={{ boxShadow: '0px 0px 10px rgba(0, 0, 0, 0.05)' }}
              >
                <div className='text-2xl font-medium text-center mb-8 text-[#1A1A1A]'>
                  {item.title}
                </div>

                <div className='flex flex-row justify-center items-center gap-4 mb-8 text-base font-medium text-[#3780F1]'>
                  <span>
                    {t('profile.currentBalance', { target: ':' })}{' '}
                    {item.currentQuota}
                  </span>
                  <span>
                    {t('profile.historicalUsage', { target: ':' })}{' '}
                    {item.historicalUsage}
                  </span>
                </div>

                <button
                  className='w-full p-3 rounded-full text-white font-medium cursor-pointer'
                  style={{
                    background:
                      'linear-gradient(90deg, #008BFC 0%, #008BFE 100%)',
                    boxShadow: 'inset 0 0 15.1px 0 rgba(255, 255, 255, 0.71)',
                  }}
                  onClick={() => handleBuyClick(item)}
                  tabIndex={0}
                  aria-label={t('profile.buyImmediately')}
                  onKeyDown={(e) =>
                    (e.key === 'Enter' || e.key === ' ') && handleBuyClick(item)
                  }
                >
                  <span className='text-xl font-medium'>
                    {t('profile.buyImmediately')}
                  </span>
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
      <PurchaseModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        card={currentCard}
      />
    </div>
  )
}
